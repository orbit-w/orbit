pwd:=$(shell pwd)
APP_NAME:=game

# Default proto file (leave empty to process all)
CS_PROTO_DIR ?=../protocol/cspb/
# Default proto directory
PROTO_DIR ?= ../protocol/pt

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
	go run tools/gotools/gen/main.go blueprintgen \
		--blueprint-dir=blueprint \
		--proto-output=protocol \
		--go-output=app/mme \
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

# 生成protocol目录下所有proto文件的pb.go文件到app/proto/mme目录
g:
	@echo "Generating proto files to app/proto/mme..."
	@if [ ! -d "protocol" ]; then \
		echo "Error: protocol directory does not exist"; \
		exit 1; \
	fi
	@for proto_file in protocol/*.proto; do \
		if [ -f "$$proto_file" ]; then \
			echo "Generating $$proto_file..."; \
			protoc --proto_path=. \
				--proto_path=protocol \
				--go_out=app/proto \
				$$proto_file; \
		fi \
	done

# 生成 Blueprint 代码并编译 proto（完整流程）
GenBlueprintFull: GenBlueprint
	@echo "Generating proto Go files from generated proto files..."
	$(MAKE) g

# 帮助信息
.PHONY: help
help:
	@echo "可用的make命令："
	@echo "  make Build              - 构建项目"
	@echo "  make BuildLinux         - 构建Linux版本"
	@echo "  make GenProto [CS_PROTO_DIR=path/to/proto] - 生成pb.go文件、协议ID和胶水代码（可选指定文件）"
	@echo "  make GenGoStructs       - 生成Go structs从proto"
	@echo "  make GenBlueprint       - 从blueprint YAML文件生成proto和Go代码"
	@echo "  make BuildPackageLinux  - 为Linux打包，用法: make BuildPackageLinux ENV=prod"
	@echo "  make GoBenchmark        - 运行基准测试"
	@echo "  make help               - 显示此帮助信息"