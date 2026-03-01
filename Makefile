-include .env

.PHONY: dev
dev:
	air

.PHONY: test
test:
	go test ./...

.PHONY: test-integration
test-integration:
	go test -tags=integration ./tests/integration/...

.PHONY: coverage
coverage:
	go test -count=1 -tags=integration -covermode=atomic -coverpkg=./internal/service/... -coverprofile=coverage-service.out ./internal/service ./tests/integration/...
	bash -lc "set -o pipefail; go tool cover -func=coverage-service.out | grep '^total:'"

.PHONY: coverage-full
coverage-full:
	go test -count=1 -tags=integration -covermode=atomic -coverpkg=./internal/api/...,./internal/platform/...,./internal/service/... -coverprofile=coverage-full.out ./internal/api/... ./internal/platform/... ./internal/service ./tests/integration/...
	bash -lc "set -o pipefail; go tool cover -func=coverage-full.out | grep '^total:'"

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

.PHONY: seed
seed:
	go run ./cmd/seed

.PHONY: reset-seed
reset-seed:
	@migrate -path $(MIGRATIONS_DIR) -database $(APP_DATABASE_URL) down -all
	@migrate -path $(MIGRATIONS_DIR) -database $(APP_DATABASE_URL) up
	go run ./cmd/seed
