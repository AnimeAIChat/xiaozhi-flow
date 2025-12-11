# Makefile for building the Go project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
MAIN_PKG=./cmd/xiaozhi-server
BINARY_NAME=xiaozhi-server
SWAG_MAIN=main.go
SWAG_DIRS=cmd/xiaozhi-server,internal/transport/http/webapi,internal/transport/http/vision,internal/transport/http/ota,internal/transport/http,internal/platform/storage,internal/transport/http/v1
SWAG_OUT=internal/platform/docs
SWAG_FLAGS=--parseDependency=false --parseGoList=false

# 插件管理相关参数
PROTO_DIR=api/proto
GEN_DIR=gen/go
PLUGIN_PROTO=$(PROTO_DIR)/plugin.proto

BUILD_DEPS := swag proto-gen

all: build

build: $(BUILD_DEPS)
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PKG)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run: $(BUILD_DEPS)
	$(GOCMD) run $(MAIN_PKG)

test:
	$(GOCMD) test ./...

# 生成 Swagger 文档；若未安装 swag 或失败，忽略错误继续
swag:
	swag init -g $(SWAG_MAIN) -d $(SWAG_DIRS) -o $(SWAG_OUT) $(SWAG_FLAGS) || (echo "swag init failed, continuing..." && exit 0)

# 生成 Protocol Buffers 代码
proto-gen:
	@echo "Generating protobuf code..."
	@buf generate || echo "Protobuf generation skipped or failed"

# 验证 Protocol Buffers
proto-lint:
	@echo "Linting protobuf files..."
	@where buf >nul 2>nul && ( \
		buf lint; \
		echo "Protobuf linting completed"; \
	) || ( \
		echo "Warning: buf not found, please install buf for protobuf linting"; \
	)

# 格式化 Protocol Buffers
proto-format:
	@echo "Formatting protobuf files..."
	@where buf >nul 2>nul && ( \
		buf format -w; \
		echo "Protobuf formatting completed"; \
	) || ( \
		echo "Warning: buf not found, please install buf for protobuf formatting"; \
	)

# 构建插件protoc相关
proto: proto-gen proto-lint proto-format

# 启动服务器并测试插件API
run-with-plugins: $(BUILD_DEPS)
	@echo "Starting server with plugin management..."
	$(GOCMD) run $(MAIN_PKG)

# 测试插件API
test-plugins:
	@echo "Testing plugin APIs..."
	@echo "1. Testing plugin list API..."
	@curl -s "http://localhost:8080/api/v1/plugins/" > temp_plugins.json 2>nul || echo "Plugin list API request failed"
	@if exist temp_plugins.json ( \
		echo "Plugin list API response received"; \
		del temp_plugins.json; \
	) else ( \
		echo "Plugin list API test failed"; \
	)
	@echo "2. Testing plugin control API..."
	@curl -s -X POST "http://localhost:8080/api/v1/plugins/ollama/control" -H "Content-Type: application/json" -d "{\"action\": \"start\"}" > temp_control.json 2>nul || echo "Plugin control API request failed"
	@if exist temp_control.json ( \
		echo "Plugin control API response received"; \
		del temp_control.json; \
	) else ( \
		echo "Plugin control API test failed"; \
	)

# 显示插件管理帮助
plugins-help:
	@echo "Plugin Management Commands:"
	@echo "  make proto          - Generate, lint and format protobuf files"
	@echo "  make proto-gen      - Generate protobuf code"
	@echo "  make proto-lint     - Lint protobuf files"
	@echo "  make proto-format   - Format protobuf files"
	@echo "  make run-with-plugins - Start server with plugin management"
	@echo "  make test-plugins   - Test plugin APIs"
	@echo "  make plugins-help   - Show this help message"
	@echo ""
	@echo "Plugin API Endpoints:"
	@echo "  GET    /api/v1/plugins/              - List all plugins"
	@echo "  GET    /api/v1/plugins/?capability_type=llm - Filter plugins by capability"
	@echo "  POST   /api/v1/plugins/{id}/control - Control plugin (start/stop/restart/reallocate_port)"

# 开发环境完整启动
dev: $(BUILD_DEPS)
	@echo "Starting development environment..."
	@echo "Server will be available at: http://localhost:8080"
	@echo "Plugin API will be available at: http://localhost:8080/api/v1/plugins/"
	@echo "API Documentation: http://localhost:8080/docs"
	$(GOCMD) run $(MAIN_PKG)

.PHONY: all build clean run test swag proto-gen proto-lint proto-format proto run-with-plugins test-plugins plugins-help dev
