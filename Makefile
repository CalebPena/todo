.PHONY: dev up down logs ps

dev:
	bash scripts/dev.sh

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps
