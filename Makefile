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