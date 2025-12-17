pwd:=$(shell pwd)
APP_NAME:=game

# Default proto file (leave empty to process all)
CS_PROTO_DIR ?=../protocol/cspb/
# Default proto directory
PROTO_DIR ?= ../protocol/pt

blueprint_DIR ?= ../protocol/blueprint
protocol_DIR ?= ../protocol/protocol

# 拼接完整路径（使用当前工作目录）
BLUEPRINT_FULL_DIR := $(shell realpath $(blueprint_DIR) 2>/dev/null || echo $(pwd)/$(blueprint_DIR))
PROTOCOL_FULL_DIR := $(shell realpath $(protocol_DIR) 2>/dev/null || echo $(pwd)/$(protocol_DIR))

GoBenchmark:
	go test ./benchmark/... -v -run=^$ -benchmem -bench=.

Build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) main.go

BuildLinux:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME) main.go

# 生成所有proto相关的代码（protobuf、协议ID和胶水代码）
GenProto:
	@echo "Generating proto files from $(CS_PROTO_DIR) to app/proto/pb..."
	go run tools/gotools/gen/main.go routergen --proto-dir=$(CS_PROTO_DIR) --output-dir=app/proto/pb --gen-pb-go=true --gen-proto-code=true --gen-proto-ids=true ${CS_PROTO_DIR:+--proto-file=$(CS_PROTO_DIR)}

# 生成 Go structs
GenGoStructs:
	go run tools/gotools/gen/main.go go_structs --proto-dir=$(PROTO_DIR) --output-dir=app/structs --package=pb

# 生成 Blueprint 代码（从 YAML 生成 proto 和 Go 代码）
GenBlueprint:
	@echo "Generating proto and Go code from blueprint YAML files..."
	@echo "Blueprint directory: $(BLUEPRINT_FULL_DIR)"
	@echo "Protocol output directory: $(PROTOCOL_FULL_DIR)"
	go run tools/gotools/gen/main.go blueprintgen \
		--blueprint-dir=$(BLUEPRINT_FULL_DIR) \
		--proto-output=$(PROTOCOL_FULL_DIR) \
		--go-output=internal/game/mme \
		--debug

# Build for Linux with specified config file
# Usage: make BuildPackageLinux ENV=prod (or other environment name without the 'config_' prefix and '.toml' suffix)
# If ENV is not specified, it will use the default config.toml
BuildPackageLinux:
	mkdir -p package
	# Check if ENV parameter is provided and the corresponding config file exists
	if [ -n "$(ENV)" ] && [ -f "configs/config_$(ENV).toml" ]; then \
		echo "Using config_$(ENV).toml for build"; \
		cp configs/config_$(ENV).toml package/config.toml; \
	else \
		echo "Using default config.toml for build"; \
		cp configs/config.toml package/config.toml; \
	fi
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o package/$(APP_NAME) main.go

# 安装/更新 protoc-gen-go (需要 >= 1.22.0 以支持 proto3 optional 字段)
InstallProtocGenGo:
	@echo "Installing/updating protoc-gen-go (requires >= 1.22.0 for proto3 optional support)..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "protoc-gen-go installed/updated successfully"
	@protoc-gen-go --version

# 检查 protoc-gen-go 版本是否支持 proto3 optional 字段
check-protoc-gen-go:
	@echo "Checking protoc-gen-go version..."
	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "Error: protoc-gen-go not found."; \
		echo "Please install it with: make InstallProtocGenGo"; \
		exit 1; \
	fi
	@VERSION=$$(protoc-gen-go --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ -z "$$VERSION" ]; then \
		echo "Warning: Could not determine protoc-gen-go version. Attempting to proceed..."; \
	else \
		MAJOR=$$(echo $$VERSION | cut -d. -f1); \
		MINOR=$$(echo $$VERSION | cut -d. -f2); \
		if [ $$MAJOR -lt 1 ] || ([ $$MAJOR -eq 1 ] && [ $$MINOR -lt 22 ]); then \
			echo "Error: protoc-gen-go version $$VERSION is too old."; \
			echo "proto3 optional fields require protoc-gen-go >= 1.22.0"; \
			echo "Please update with: make InstallProtocGenGo"; \
			exit 1; \
		fi; \
		echo "protoc-gen-go version $$VERSION is compatible with proto3 optional fields"; \
	fi

# 生成protocol目录下所有proto文件的pb.go文件到app/proto目录
g: check-protoc-gen-go
	@echo "Generating proto Go files from generated proto files..."
	@echo "Using protocol directory: $(PROTOCOL_FULL_DIR)"
	@if [ ! -d "$(PROTOCOL_FULL_DIR)" ]; then \
		echo "Error: protocol directory does not exist: $(PROTOCOL_FULL_DIR)"; \
		exit 1; \
	fi
	@GOPATH_BIN=$$(go env GOPATH)/bin; \
	GOBIN_VAL=$$(go env GOBIN 2>/dev/null || echo ""); \
	if [ -n "$$GOBIN_VAL" ]; then \
		export PATH=$$GOBIN_VAL:$$PATH; \
	else \
		export PATH=$$GOPATH_BIN:$$PATH; \
	fi; \
	echo "Using protoc-gen-go from: $$(which protoc-gen-go)"; \
	for proto_file in $(PROTOCOL_FULL_DIR)/*.proto; do \
		if [ -f "$$proto_file" ]; then \
			echo "Generating $$proto_file..."; \
			protoc --proto_path=$(PROTOCOL_FULL_DIR) \
				--go_out=paths=import:. \
				--go_opt=paths=import \
				--go_opt=module=gitee.com/orbit-w/orbit \
				$$proto_file; \
		fi; \
	done

# 生成 Blueprint 代码并编译 proto（完整流程）
GenBlueprintFull: GenBlueprint
	@echo "Generating proto Go files from generated proto files..."
	$(MAKE) g

# 帮助信息
.PHONY: help InstallProtocGenGo check-protoc-gen-go
help:
	@echo "可用的make命令："
	@echo "  make Build              - 构建项目"
	@echo "  make BuildLinux         - 构建Linux版本"
	@echo "  make GenProto [CS_PROTO_DIR=path/to/proto] - 生成pb.go文件、协议ID和胶水代码（可选指定文件）"
	@echo "  make GenGoStructs       - 生成Go structs从proto"
	@echo "  make GenBlueprint       - 从blueprint YAML文件生成proto和Go代码"
	@echo "  make BuildPackageLinux  - 为Linux打包，用法: make BuildPackageLinux ENV=prod"
	@echo "  make InstallProtocGenGo - 安装/更新 protoc-gen-go (需要 >= 1.22.0 以支持 proto3 optional)"
	@echo "  make GoBenchmark        - 运行基准测试"
	@echo "  make help               - 显示此帮助信息"