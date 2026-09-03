.PHONY: backend-run backend-test frontend-dev frontend-build db-up db-down docker-build docker-up docker-down docker-logs

backend-run:
	cd backend && go run main.go

backend-test:
	cd backend && go test ./...

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

db-up:
	docker compose --env-file clubmax.env up -d db

db-down:
	docker compose --env-file clubmax.env down

docker-build:
	docker compose --env-file clubmax.production.env -f compose.production.yaml build --pull

docker-up:
	docker compose --env-file clubmax.production.env -f compose.production.yaml up -d --wait

docker-down:
	docker compose --env-file clubmax.production.env -f compose.production.yaml down

docker-logs:
	docker compose --env-file clubmax.production.env -f compose.production.yaml logs -f --tail=100
