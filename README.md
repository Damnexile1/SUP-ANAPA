# SUP Anapa Booking Service

Fullstack сервис бронирования SUP-прогулок у Анапы.

## Стек
- Frontend: Next.js + TypeScript + Tailwind
- Backend: Go 1.22 + Gin + GORM
- DB: PostgreSQL
- Миграции: golang-migrate
- API: REST + OpenAPI (`backend/docs/openapi.yaml`)

## Быстрый старт
```bash
cp backend/.env.example backend/.env
make up
make migrate
make seed
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

## Полезные команды
```bash
make test
```

## Структура
- `backend/` — API, weather scoring, модели, миграции
- `frontend/` — интерфейс на русском языке
- `deploy/` — docker-compose
