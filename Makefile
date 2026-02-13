include .env

.PHONY: dev
dev:
	air

.PHONY: build
build:
	go build -o ./bin/main ./cmd/api

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_DIR) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path $(MIGRATIONS_DIR) -database $(APP_DATABASE_URL) up

.PHONY: migrate-down
migrate-down:
	@migrate -path $(MIGRATIONS_DIR) -database $(APP_DATABASE_URL) down
