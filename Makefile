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

# 生成所有proto相关的代码（protobuf、协议ID和胶水代码）
GenProto:
	# 删除 app/proto/pb 下的所有文件
	find app/proto/pb -type f -not -path "*/\.*" -delete
	find app/proto -name "*.proto" -type f | xargs -I{} protoc \
	       --proto_path=. \
	       --proto_path=$(GOPATH)/src \
	       --proto_path=$(GOPATH)/pkg/mod \
	       --proto_path=./vendor/github.com/asynkron/protoactor-go/actor \
	       --go_out=app/proto --go-grpc_out=app/proto {}
	# 生成协议ID和胶水代码
	go run lib/genproto/main.go --proto_dir=app/proto --quiet

# 只生成协议ID
GenProtoID:
	# 删除旧的协议ID文件
	find app/proto/pb -name "*_protocol_ids.go" -delete
	find app/proto/pb -name "protocol_ids.go" -delete
	go run lib/genproto/main.go --proto_dir=app/proto --gen_proto_code=false --quiet

# 只生成胶水代码
GenGlueCode:
	# 删除旧的胶水代码文件
	find app/proto/pb -name "*_request_glue.go" -delete
	find app/proto/pb -name "*_notify_glue.go" -delete
	go run lib/genproto/main.go --proto_dir=app/proto --gen_proto_ids=false --quiet

# 调试模式生成所有proto相关代码
GenProtoDebug:
	# 删除 app/proto/pb 下的所有文件
	find app/proto/pb -type f -not -path "*/\.*" -delete
	find app/proto -name "*.proto" -type f | xargs -I{} protoc \
	       --proto_path=. \
	       --proto_path=$(GOPATH)/src \
	       --proto_path=$(GOPATH)/pkg/mod \
	       --proto_path=./vendor/github.com/asynkron/protoactor-go/actor \
	       --go_out=app/proto --go-grpc_out=app/proto {}
	# 调试模式生成协议ID和胶水代码
	go run lib/genproto/main.go --proto_dir=app/proto --debug --quiet=false

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
	@echo "  make GenProto           - 生成所有proto相关代码"
	@echo "  make GenProtoID         - 只生成协议ID"
	@echo "  make GenGlueCode        - 只生成胶水代码"
	@echo "  make GenProtoDebug      - 调试模式生成proto代码"
	@echo "  make BuildPackageLinux  - 为Linux打包，用法: make BuildPackageLinux ENV=prod"
	@echo "  make GoBenchmark        - 运行基准测试"
	@echo "  make help               - 显示此帮助信息"