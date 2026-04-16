COMPOSE := docker compose

.PHONY: help build up down restart logs ps cli cli-up cli-shell api db clean

help:
	@echo "Available targets:"
	@echo "  make build   - Build Docker images"
	@echo "  make up      - Start db and api in background"
	@echo "  make down    - Stop and remove containers"
	@echo "  make restart - Restart stack (down + up)"
	@echo "  make logs    - Follow logs for all services"
	@echo "  make ps      - Show running services"
	@echo "  make cli     - Build, start DB, and run CLI app in terminal"
	@echo "  make cli-up  - Start CLI container in background"
	@echo "  make cli-shell - Enter running CLI container shell"
	@echo "  make api     - Build and start only API + DB"
	@echo "  make db      - Start only DB"
	@echo "  make clean   - Stop stack and remove volumes"

build:
	$(COMPOSE) build

up: build
	$(COMPOSE) up -d --build db api

down:
	$(COMPOSE) down

restart: down up

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

cli: build
	$(COMPOSE) up -d --build db
	$(COMPOSE) run --rm app

cli-up: build
	$(COMPOSE) up -d --build app

cli-shell:
	$(COMPOSE) exec app sh

api: build
	$(COMPOSE) up -d db api

db:
	$(COMPOSE) up -d db

clean:
	$(COMPOSE) down -v
