.PHONY: help build run run-local stop down restart logs health test test-e2e test-load-smoke test-load-stress lint clean clean-all

APP_NAME=pr-reviewer-service
BUILD_DIR=./bin
MAIN_PATH=./cmd/api/main.go

# Определяем команду docker compose (v2 или v1)
DC := $(shell command -v docker-compose 2> /dev/null)
ifeq ($(DC),)
    DC := docker compose
endif

help: ## Показать список доступных команд
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать Go-бинарник локально
	@echo "Сборка $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Готово: $(BUILD_DIR)/$(APP_NAME)"

run: ## Запустить ВСЕ сервисы в Docker (App + DB + Migrator) - как в README
	@echo "Запуск полного стека в Docker..."
	@$(DC) up -d --build
	@echo "Сервис доступен по адресу: http://localhost:8080"

run-local: ## Запустить приложение локально (Go), а БД в Docker
	@echo "Подготовка БД в Docker..."
	@$(DC) up -d postgres migrator
	@echo "Ожидание готовности БД..."
	@sleep 5
	@echo "Запуск локального приложения..."
	@go run $(MAIN_PATH)

down: ## Остановить и удалить контейнеры
	@echo "Остановка сервисов..."
	@$(DC) down

stop: down ## Алиас для down

restart: down run ## Перезапустить все сервисы (Docker)

logs: ## Показать логи сервиса API
	@$(DC) logs -f api

health: ## Проверить здоровье API (curl)
	@echo "Проверка healthcheck..."
	@curl -s http://localhost:8080/health | jq . || echo "API недоступен"

test: ## Запустить Unit-тесты
	@echo "Запуск unit тестов..."
	@go test -v -race ./internal/...

test-e2e: ## Запустить E2E тесты
	@echo "Запуск E2E тестов..."
	@go test -v ./tests/e2e/...

test-load-smoke: ## Запустить Smoke тест (k6, 10 секунд, 1 VU)
	@echo "Запуск k6 smoke test..."
	@docker run --rm -i -v $(PWD)/tests:/tests --network host -e BASE_URL=http://localhost:8080 -e SMOKE=true grafana/k6:latest run /tests/k6/load_test.js

test-load-stress: ## Запустить Stress тест (k6, 100 VUs, 5 минут)
	@echo "Запуск k6 stress test..."
	@docker run --rm -i -v $(PWD)/tests:/tests --network host -e BASE_URL=http://localhost:8080 grafana/k6:latest run --vus 100 --duration 5m /tests/k6/load_test.js

lint: ## Запустить линтер (golangci-lint)
	@echo "Запуск линтера..."
	@golangci-lint run ./...

clean: ## Очистить артефакты сборки
	@rm -rf $(BUILD_DIR)

clean-all: clean ## Полная очистка (включая Docker volumes и данные БД)
	@echo "Полная очистка данных..."
	@$(DC) down -v

cover: ## Показать покрытие кода тестами
	@echo "Запуск тестов с покрытием..."
	@go test -v -coverprofile=coverage.out ./internal/...
	@go tool cover -func=coverage.out
	@echo "Генерация HTML отчета..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет сохранен в coverage.html"

.DEFAULT_GOAL := help
