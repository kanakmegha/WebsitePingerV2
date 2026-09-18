.PHONY: dev down logs build help

dev:
	docker-compose up --build

down:
	docker-compose down

logs:
	docker-compose logs -f

build:
	docker-compose build
