COMPOSE := docker compose -f docker-compose.yml --env-file classifier/.env

.PHONY: up api cli down clean

up:
	$(COMPOSE) up -d --build db api frontend

api:
	$(COMPOSE) up -d --build db api

cli:
	$(COMPOSE) up -d --build db
	$(COMPOSE) run --rm cli

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v --remove-orphans
