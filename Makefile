COMPOSE := docker compose -f docker-compose.yml --env-file classifier/.env

.PHONY: up api cli seed down clean

up:
	$(COMPOSE) up -d --build db api frontend

api:
	$(COMPOSE) up -d --build db api

cli:
	$(COMPOSE) up -d --build db cli
	$(COMPOSE) run --rm cli

seed:
	$(COMPOSE) up -d db
	$(COMPOSE) exec -T db sh -c 'until pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" >/dev/null 2>&1; do sleep 1; done; psql -h 127.0.0.1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -f /seed.sql'

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v --remove-orphans
