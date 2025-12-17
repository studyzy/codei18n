# Makefile for CodeI18n Core MVP
# 项目: CodeI18n - 代码注释国际化基础设施
# 版本: 0.1.0

# 变量定义
BINARY_NAME=codei18n
BINARY_PATH=./bin/$(BINARY_NAME)
GO=go
GOFLAGS=-v
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# 颜色输出（仅在支持的终端中使用）
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

# 默认目标
.DEFAULT_GOAL := help

# 检测操作系统
ifeq ($(OS),Windows_NT)
    BINARY_NAME := $(BINARY_NAME).exe
    RM := del /Q
    RMDIR := rmdir /S /Q
else
    RM := rm -f
    RMDIR := rm -rf
endif

.PHONY: help
help: ## 显示帮助信息
	@echo "$(COLOR_BOLD)CodeI18n Makefile 帮助$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)可用命令:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_YELLOW)示例:$(COLOR_RESET)"
	@echo "  make build          # 构建可执行文件"
	@echo "  make test           # 运行所有测试"
	@echo "  make coverage       # 生成覆盖率报告"
	@echo "  make lint           # 运行代码检查"
	@echo "  make all            # 运行所有检查并构建"
	@echo ""

.PHONY: deps
deps: ## 安装和更新依赖
	@echo "$(COLOR_BLUE)正在安装/更新依赖...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ 依赖安装完成$(COLOR_RESET)"

.PHONY: build
build: ## 构建可执行文件
	@echo "$(COLOR_BLUE)正在构建 $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p ./bin
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) ./cmd/codei18n
	@echo "$(COLOR_GREEN)✓ 构建完成: $(BINARY_PATH)$(COLOR_RESET)"

.PHONY: install
install: ## 安装到 $GOPATH/bin
	@echo "$(COLOR_BLUE)正在安装 $(BINARY_NAME)...$(COLOR_RESET)"
	$(GO) install $(GOFLAGS) ./cmd/codei18n
	@echo "$(COLOR_GREEN)✓ 已安装到 $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: test
test: ## 运行所有测试（跳过集成测试）
	@echo "$(COLOR_BLUE)正在运行测试...$(COLOR_RESET)"
	$(GO) test -short $(GOFLAGS) ./...
	@echo "$(COLOR_GREEN)✓ 测试完成$(COLOR_RESET)"

.PHONY: test-integration
test-integration: ## 运行所有测试（包括集成测试）
	@echo "$(COLOR_BLUE)正在运行测试（包括集成测试）...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)注意: 需要有效的 OPENAI_API_KEY 和 OPENAI_BASE_URL$(COLOR_RESET)"
	$(GO) test $(GOFLAGS) ./...
	@echo "$(COLOR_GREEN)✓ 测试完成$(COLOR_RESET)"

.PHONY: test-verbose
test-verbose: ## 运行测试（详细输出）
	@echo "$(COLOR_BLUE)正在运行测试（详细模式）...$(COLOR_RESET)"
	$(GO) test -v ./...
	@echo "$(COLOR_GREEN)✓ 测试完成$(COLOR_RESET)"

.PHONY: coverage
coverage: ## 生成测试覆盖率报告（跳过集成测试）
	@echo "$(COLOR_BLUE)正在生成覆盖率报告...$(COLOR_RESET)"
	$(GO) test -short -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "$(COLOR_YELLOW)提示: 使用 'make coverage-html' 查看 HTML 格式的覆盖率报告$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ 覆盖率报告已生成: $(COVERAGE_FILE)$(COLOR_RESET)"

.PHONY: coverage-html
coverage-html: coverage ## 生成并打开 HTML 覆盖率报告
	@echo "$(COLOR_BLUE)正在生成 HTML 覆盖率报告...$(COLOR_RESET)"
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✓ HTML 覆盖率报告已生成: $(COVERAGE_HTML)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)提示: 在浏览器中打开 $(COVERAGE_HTML) 查看详细报告$(COLOR_RESET)"

.PHONY: coverage-check
coverage-check: coverage ## 检查覆盖率是否达标（总体 ≥60%，核心模块 ≥80%）
	@echo "$(COLOR_BLUE)正在检查覆盖率达标情况...$(COLOR_RESET)"
	@total=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	core_comment=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep 'core/comment' | awk '{sum+=$$3; count++} END {if(count>0) print sum/count; else print 0}'); \
	core_mapping=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep 'core/mapping' | awk '{sum+=$$3; count++} END {if(count>0) print sum/count; else print 0}'); \
	core_translate=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep 'core/translate' | awk '{sum+=$$3; count++} END {if(count>0) print sum/count; else print 0}'); \
	echo ""; \
	echo "覆盖率检查结果:"; \
	echo "  总体覆盖率: $$total% (要求 ≥60%)"; \
	if [ -n "$$core_comment" ] && [ "$$(echo "$$core_comment > 0" | bc -l)" -eq 1 ]; then echo "  core/comment: $$core_comment% (要求 ≥80%)"; fi; \
	if [ -n "$$core_mapping" ] && [ "$$(echo "$$core_mapping > 0" | bc -l)" -eq 1 ]; then echo "  core/mapping: $$core_mapping% (要求 ≥80%)"; fi; \
	if [ -n "$$core_translate" ] && [ "$$(echo "$$core_translate > 0" | bc -l)" -eq 1 ]; then echo "  core/translate: $$core_translate% (要求 ≥80%)"; fi; \
	echo ""

.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(COLOR_BLUE)正在格式化代码...$(COLOR_RESET)"
	gofmt -w .
	@echo "$(COLOR_GREEN)✓ 代码格式化完成$(COLOR_RESET)"

.PHONY: lint
lint: ## 运行代码检查（golint 或 staticcheck）
	@echo "$(COLOR_BLUE)正在运行代码检查...$(COLOR_RESET)"
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "使用 staticcheck 进行检查..."; \
		staticcheck ./...; \
	elif command -v golint >/dev/null 2>&1; then \
		echo "使用 golint 进行检查..."; \
		golint ./...; \
	else \
		echo "$(COLOR_YELLOW)警告: 未找到 staticcheck 或 golint$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)请安装其中之一:$(COLOR_RESET)"; \
		echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		echo "  或"; \
		echo "  go install golang.org/x/lint/golint@latest"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)✓ 代码检查完成$(COLOR_RESET)"

.PHONY: vet
vet: ## 运行 go vet
	@echo "$(COLOR_BLUE)正在运行 go vet...$(COLOR_RESET)"
	$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ go vet 检查完成$(COLOR_RESET)"

.PHONY: check
check: fmt vet lint ## 运行所有代码质量检查（fmt + vet + lint）
	@echo "$(COLOR_GREEN)✓ 所有代码质量检查完成$(COLOR_RESET)"

.PHONY: clean
clean: ## 清理构建产物
	@echo "$(COLOR_BLUE)正在清理构建产物...$(COLOR_RESET)"
	$(RM) $(BINARY_PATH)
	$(RM) $(COVERAGE_FILE)
	$(RM) $(COVERAGE_HTML)
	@echo "$(COLOR_GREEN)✓ 清理完成$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## 清理所有生成文件（包括依赖缓存）
	@echo "$(COLOR_BLUE)正在清理所有生成文件...$(COLOR_RESET)"
	$(GO) clean -cache -testcache -modcache
	@echo "$(COLOR_GREEN)✓ 深度清理完成$(COLOR_RESET)"

.PHONY: run
run: build ## 构建并运行（显示帮助）
	@echo "$(COLOR_BLUE)正在运行 $(BINARY_NAME)...$(COLOR_RESET)"
	$(BINARY_PATH) --help

.PHONY: dev
dev: ## 开发模式：运行测试和代码检查
	@echo "$(COLOR_BLUE)开发模式检查...$(COLOR_RESET)"
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test
	@echo "$(COLOR_GREEN)✓ 开发模式检查完成$(COLOR_RESET)"

.PHONY: all
all: clean deps check test build ## 运行完整的 CI 流程（清理 + 依赖 + 检查 + 测试 + 构建）
	@echo ""
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ 完整 CI 流程完成$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)可执行文件: $(BINARY_PATH)$(COLOR_RESET)"
	@echo ""

.PHONY: ci
ci: all coverage-check ## CI/CD 模式：运行所有检查并验证覆盖率
	@echo ""
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ CI/CD 检查全部通过$(COLOR_RESET)"
	@echo ""

.PHONY: pre-commit
pre-commit: fmt vet test ## Pre-commit 检查（格式化 + vet + 测试）
	@echo "$(COLOR_GREEN)✓ Pre-commit 检查完成$(COLOR_RESET)"

.PHONY: version
version: ## 显示 Go 版本和项目信息
	@echo "$(COLOR_BOLD)项目信息$(COLOR_RESET)"
	@echo "  项目名称: CodeI18n"
	@echo "  二进制文件: $(BINARY_NAME)"
	@echo ""
	@echo "$(COLOR_BOLD)Go 环境$(COLOR_RESET)"
	@$(GO) version
	@echo ""
	@$(GO) env | grep -E 'GOPATH|GOROOT|GOOS|GOARCH'

# Docker 相关命令
DOCKER_IMAGE ?= codei18n
DOCKER_TAG ?= latest

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "$(COLOR_BLUE)正在构建 Docker 镜像...$(COLOR_RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(COLOR_GREEN)✓ Docker 镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "$(COLOR_BLUE)正在运行 Docker 容器...$(COLOR_RESET)"
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-test
docker-test: ## 在 Docker 容器中运行测试
	@echo "$(COLOR_BLUE)在 Docker 容器中运行测试...$(COLOR_RESET)"
	docker build --target builder -t $(DOCKER_IMAGE):test .
	docker run --rm $(DOCKER_IMAGE):test go test -v ./...
	@echo "$(COLOR_GREEN)✓ Docker 测试完成$(COLOR_RESET)"

.PHONY: docker-clean
docker-clean: ## 清理 Docker 镜像
	@echo "$(COLOR_BLUE)正在清理 Docker 镜像...$(COLOR_RESET)"
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	docker rmi $(DOCKER_IMAGE):test || true
	@echo "$(COLOR_GREEN)✓ Docker 镜像清理完成$(COLOR_RESET)"
