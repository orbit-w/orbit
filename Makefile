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
	go run tools/gotools/gen/main.go routergen --proto-dir=$(CS_PROTO_DIR) --output-dir=app/proto/pb ${CS_PROTO_DIR:+--proto-file=$(CS_PROTO_DIR)}

# 生成 Go structs
GenGoStructs:
	go run tools/gotools/gen/main.go go_structs --proto-dir=$(PROTO_DIR) --output-dir=app/structs --package=pb

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

# 帮助信息
.PHONY: help
help:
	@echo "可用的make命令："
	@echo "  make Build              - 构建项目"
	@echo "  make BuildLinux         - 构建Linux版本"
	@echo "  make GenProto [CS_PROTO_DIR=path/to/proto] - 使用routergen生成proto相关代码（可选指定文件）"
	@echo "  make GenGoStructs       - 生成Go structs从proto"
	@echo "  make BuildPackageLinux  - 为Linux打包，用法: make BuildPackageLinux ENV=prod"
	@echo "  make GoBenchmark        - 运行基准测试"
	@echo "  make help               - 显示此帮助信息"