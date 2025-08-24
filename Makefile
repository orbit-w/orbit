pwd:=$(shell pwd)
APP_NAME:=game

GoBenchmark:
	go test ./benchmark/... -v -run=^$ -benchmem -bench=.

Build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) main.go

BuildLinux:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME) main.go

# Default proto file (leave empty to process all)
PROTO_FILE ?=../protocol/cspb/

# 生成所有proto相关的代码（protobuf、协议ID和胶水代码）
GenProto:
	./scripts/genproto.sh --all ${PROTO_FILE:+--proto-file=$(PROTO_FILE)}

# 只生成协议ID
GenProtoID:
	./scripts/genproto.sh --ids-only ${PROTO_FILE:+--proto-file=$(PROTO_FILE)}

# 只生成胶水代码
GenGlueCode:
	./scripts/genproto.sh --glue-only ${PROTO_FILE:+--proto-file=$(PROTO_FILE)}

# 调试模式生成所有proto相关代码
GenProtoDebug:
	./scripts/genproto.sh --all --debug ${PROTO_FILE:+--proto-file=$(PROTO_FILE)}

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
	@echo "  make GenProto [PROTO_FILE=path/to/proto] - 生成所有proto相关代码（可选指定文件）"
	@echo "  make GenProtoID [PROTO_FILE=path/to/proto] - 只生成协议ID（可选指定文件）"
	@echo "  make GenGlueCode [PROTO_FILE=path/to/proto] - 只生成胶水代码（可选指定文件）"
	@echo "  make GenProtoDebug [PROTO_FILE=path/to/proto] - 调试模式生成proto代码（可选指定文件）"
	@echo "  make GenProtoFile PROTO_FILE=proto/path/file.proto - 生成指定proto文件的代码"
	@echo "  make BuildPackageLinux  - 为Linux打包，用法: make BuildPackageLinux ENV=prod"
	@echo "  make GoBenchmark        - 运行基准测试"
	@echo "  make help               - 显示此帮助信息"

# 生成指定proto文件的代码
GenProtoFile:
	./scripts/genproto.sh --proto-file=$(PROTO_FILE)