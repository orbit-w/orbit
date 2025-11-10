.PHONY: gen clean

# 默认目标
all: gen

# 生成protobuf文件
gen:
	# 执行protoc命令生成Go代码
	protoc --proto_path=. \
		--proto_path=$(GOPATH)/src \
		--proto_path=$(GOPATH)/pkg/mod \
		--proto_path=../../../../vendor/github.com/asynkron/protoactor-go/actor \
		--go_out=. \
		actor_ref.proto

	@echo "Proto files generated successfully!"

# 清理生成的文件
clean:
	rm -f *.pb.go
	@echo "Generated files cleaned!"

# 执行和显示帮助
help:
	@echo "Available commands:"
	@echo "  make gen    - Generate Go files from proto files"
	@echo "  make clean  - Remove generated files"
	@echo "  make help   - Show this help message"
