up:
	docker compose -f deploy/docker-compose.yml up --build -d

migrate:
	docker compose -f deploy/docker-compose.yml run --rm migrate

seed:
	docker compose -f deploy/docker-compose.yml exec -T postgres psql -U postgres -d sup_anapa < backend/seed.sql

test:
	cd backend && go test ./...
	cd frontend && npm run build
