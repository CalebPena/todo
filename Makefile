.PHONY: dev up down logs ps

dev:
	docker compose up --build --watch

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps
