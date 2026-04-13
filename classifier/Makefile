COMPOSE := docker compose

.PHONY: help build up down restart logs ps cli api db clean

help:
	@echo "Available targets:"
	@echo "  make build   - Build Docker images"
	@echo "  make up      - Start db and api in background"
	@echo "  make down    - Stop and remove containers"
	@echo "  make restart - Restart stack (down + up)"
	@echo "  make logs    - Follow logs for all services"
	@echo "  make ps      - Show running services"
	@echo "  make cli     - Run CLI app in terminal (interactive)"
	@echo "  make api     - Build and start only API + DB"
	@echo "  make db      - Start only DB"
	@echo "  make clean   - Stop stack and remove volumes"

build:
	$(COMPOSE) build

up: build
	$(COMPOSE) up -d db api

down:
	$(COMPOSE) down

restart: down up

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

cli: build
	$(COMPOSE) run --rm app

api: build
	$(COMPOSE) up -d db api

db:
	$(COMPOSE) up -d db

clean:
	$(COMPOSE) down -v
