.PHONY: backend-run backend-test frontend-dev frontend-build db-up db-down

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
