.PHONY: help build run dev test clean docker-up docker-up-build docker-down migrate-up migrate-down migrate-create deploy-local deploy-prod logs-file admin-create

# Переменные
APP_NAME=sup-anapa
DOCKER_COMPOSE=docker compose
DOCKER_EXEC=$(DOCKER_COMPOSE) exec app
DOCKER_EXEC_DB=$(DOCKER_COMPOSE) exec postgres

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

deps: ## Установить зависимости Go (внутри контейнера)
	@echo "$(GREEN)Установка зависимостей внутри контейнера...$(NC)"
	@$(DOCKER_EXEC) go mod download
	@$(DOCKER_EXEC) go mod tidy

build: ## Собрать приложение (внутри контейнера)
	@echo "$(GREEN)Сборка приложения внутри контейнера...$(NC)"
	@$(DOCKER_EXEC) go build -o /app/bin/server ./cmd/server
	@echo "$(GREEN)Готово!$(NC)"

test: ## Запустить тесты (внутри контейнера)
	@echo "$(GREEN)Запуск тестов...$(NC)"
	@$(DOCKER_EXEC) go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	@$(DOCKER_EXEC) go test -v -coverprofile=coverage.out ./...
	@$(DOCKER_EXEC) go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет сохранен в coverage.html$(NC)"

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	@$(DOCKER_EXEC) go fmt ./...

vet: ## Проверить код
	@echo "$(GREEN)Проверка кода...$(NC)"
	@$(DOCKER_EXEC) go vet ./...

clean: ## Очистить сборочные файлы
	@echo "$(GREEN)Очистка...$(NC)"
	@$(DOCKER_EXEC) rm -rf /app/bin/
	@$(DOCKER_EXEC) rm -f coverage.out coverage.html
	@echo "$(GREEN)Готово!$(NC)"

# Docker команды
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t $(APP_NAME):latest .

docker-up: ## Запустить все сервисы через Docker Compose
	@echo "$(GREEN)Запуск Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Контейнеры запущены!$(NC)"
	@echo "$(YELLOW)Приложение доступно на http://localhost:8080$(NC)"

docker-up-build: ## Пересобрать образы и запустить сервисы через Docker Compose
	@echo "$(GREEN)Сборка и запуск Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d --build
	@echo "$(GREEN)Контейнеры запущены с пересборкой!$(NC)"
	@echo "$(YELLOW)Приложение доступно на http://localhost:8080$(NC)"

docker-down: ## Остановить все сервисы
	@echo "$(GREEN)Остановка Docker контейнеров...$(NC)"
	$(DOCKER_COMPOSE) down

docker-logs: ## Показать логи контейнеров
	$(DOCKER_COMPOSE) logs -f

docker-logs-app: ## Показать логи приложения
	$(DOCKER_COMPOSE) logs -f app

logs-file: ## Показать file-логи приложения (logs/app.log)
	@mkdir -p logs
	@tail -f logs/app.log

docker-logs-db: ## Показать логи БД
	$(DOCKER_COMPOSE) logs -f postgres

docker-restart: docker-down docker-up ## Перезапустить все сервисы

docker-clean: docker-down ## Удалить контейнеры и volumes
	@echo "$(GREEN)Удаление контейнеров и volumes...$(NC)"
	$(DOCKER_COMPOSE) down -v
	docker system prune -f

docker-ps: ## Показать статус контейнеров
	@$(DOCKER_COMPOSE) ps

# База данных
db-shell: ## Подключиться к PostgreSQL через psql
	@echo "$(GREEN)Подключение к БД...$(NC)"
	$(DOCKER_EXEC_DB) psql -U postgres -d sup_anapa

# Миграции (выполняются внутри контейнера приложения)
migrate-up: ## Применить все миграции
	@echo "$(GREEN)Применение миграций...$(NC)"
	@$(DOCKER_EXEC) migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/sup_anapa?sslmode=disable" up
	@echo "$(GREEN)Миграции применены!$(NC)"

migrate-down: ## Откатить последнюю миграцию
	@echo "$(YELLOW)Откат миграции...$(NC)"
	@$(DOCKER_EXEC) migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/sup_anapa?sslmode=disable" down 1
	@echo "$(GREEN)Миграция откачена!$(NC)"

migrate-reset: ## Откатить все миграции и применить заново
	@echo "$(YELLOW)Сброс всех миграций...$(NC)"
	@$(DOCKER_EXEC) migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/sup_anapa?sslmode=disable" down -all || true
	@$(DOCKER_EXEC) migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/sup_anapa?sslmode=disable" up
	@echo "$(GREEN)Миграции пересозданы!$(NC)"

migrate-create: ## Создать новую миграцию (использование: make migrate-create NAME=имя_миграции)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Ошибка: укажите NAME=имя_миграции$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Создание миграции $(NAME)...$(NC)"
	@$(DOCKER_EXEC) migrate create -ext sql -dir /app/migrations -seq $(NAME)
	@echo "$(GREEN)Миграция создана!$(NC)"

migrate-status: ## Показать статус миграций
	@echo "$(GREEN)Статус миграций:$(NC)"
	@$(DOCKER_EXEC) migrate -path /app/migrations -database "postgres://postgres:postgres@postgres:5432/sup_anapa?sslmode=disable" version

# Локальный деплой
deploy-local: ## Полный деплой на локальной машине
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  Локальный деплой SUP-Anapa$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Шаг 1: Проверка .env файла...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW).env не найден, создаем из .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(RED)ВАЖНО: Отредактируйте .env файл с вашими настройками!$(NC)"; \
	fi
	@echo "$(GREEN)✓ .env найден$(NC)"
	@echo ""
	@echo "$(YELLOW)Шаг 2: Пересборка и запуск Docker контейнеров...$(NC)"
	@$(MAKE) docker-up-build
	@echo ""
	@echo "$(YELLOW)Шаг 3: Ожидание готовности БД...$(NC)"
	@sleep 5
	@echo ""
	@echo "$(YELLOW)Шаг 4: Применение миграций...$(NC)"
	@$(MAKE) migrate-up
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  Деплой завершен!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Приложение доступно на:$(NC)"
	@echo "  http://localhost:8080"
	@echo ""
	@echo "$(YELLOW)Просмотр логов:$(NC)"
	@echo "  make docker-logs"

# Продакшн деплой
deploy-prod: ## Деплой на продакшн через Docker
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  Продакшн деплой SUP-Anapa$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Шаг 1: Проверка .env файла...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Ошибка: .env файл не найден!$(NC)"; \
		echo "$(YELLOW)Создайте .env файл с продакшн настройками$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ .env найден$(NC)"
	@echo ""
	@echo "$(YELLOW)Шаг 2: Остановка старых контейнеров...$(NC)"
	@$(MAKE) docker-down
	@echo ""
	@echo "$(YELLOW)Шаг 3: Сборка Docker образа...$(NC)"
	@$(MAKE) docker-build
	@echo ""
	@echo "$(YELLOW)Шаг 4: Запуск контейнеров...$(NC)"
	@$(MAKE) docker-up
	@echo ""
	@echo "$(YELLOW)Шаг 5: Ожидание готовности БД...$(NC)"
	@sleep 5
	@echo ""
	@echo "$(YELLOW)Шаг 6: Применение миграций...$(NC)"
	@$(MAKE) migrate-up
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  Продакшн деплой завершен!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Приложение доступно на:$(NC)"
	@echo "  http://localhost:8080"
	@echo ""
	@echo "$(YELLOW)Просмотр логов:$(NC)"
	@echo "  make docker-logs"

# Быстрые команды
setup: ## Первоначальная настройка проекта
	@echo "$(GREEN)Настройка проекта...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(YELLOW)Создан .env файл. Отредактируйте его перед запуском!$(NC)"; \
	fi
	@echo "$(GREEN)Проект настроен! Теперь запустите:$(NC)"
	@echo "  $(YELLOW)make deploy-local$(NC)  - для локальной разработки"
	@echo "  $(YELLOW)make deploy-prod$(NC)   - для продакшн деплоя"

start: deploy-local ## Быстрый старт (setup + deploy)

restart: docker-restart ## Перезапустить приложение
	@echo "$(GREEN)Приложение перезапущено$(NC)"

status: ## Показать статус сервисов
	@echo "$(GREEN)Статус Docker контейнеров:$(NC)"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "$(GREEN)Статус миграций:$(NC)"
	@$(MAKE) migrate-status || true

backup-db: ## Создать бэкап базы данных
	@echo "$(GREEN)Создание бэкапа БД...$(NC)"
	@mkdir -p backups
	@$(DOCKER_EXEC_DB) pg_dump -U postgres sup_anapa > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Бэкап создан в backups/$(NC)"

restore-db: ## Восстановить БД из бэкапа (использование: make restore-db FILE=backups/backup.sql)
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)Ошибка: укажите FILE=путь_к_бэкапу$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Восстановление БД из $(FILE)...$(NC)"
	@cat $(FILE) | $(DOCKER_EXEC_DB) psql -U postgres sup_anapa
	@echo "$(GREEN)БД восстановлена!$(NC)"

shell: ## Открыть shell в контейнере приложения
	@echo "$(GREEN)Открытие shell в контейнере приложения...$(NC)"
	@$(DOCKER_EXEC) sh

admin-create: ## Создать или обновить админа: make admin-create USERNAME=admin PASSWORD='password'
	@if [ -z "$(USERNAME)" ]; then \
		echo "$(RED)Ошибка: укажите USERNAME=admin$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$(PASSWORD)" ]; then \
		echo "$(RED)Ошибка: укажите PASSWORD='password'$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Создание/обновление администратора $(USERNAME)...$(NC)"
	@docker run --rm \
		--network sup-anapa_sup-anapa-network \
		-v "$$(pwd)":/app \
		-w /app \
		-e DATABASE_URL="postgres://postgres:postgres@sup-anapa-db:5432/sup_anapa?sslmode=disable" \
		golang:alpine \
		sh -c 'go run ./cmd/admin -username "$(USERNAME)" -password "$(PASSWORD)"'
	@echo "$(GREEN)Готово!$(NC)"

.DEFAULT_GOAL := help
