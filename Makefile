include .envrc

MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))
migrate-create:
	@echo "Creating a new migration..."
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@echo "Applying all up migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_URL) up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_URL) down $(filter-out $@,$(MAKECMDGOALS))
	
.PHONY: migrate-force
migrate-force:
	@echo "Forcing migration to a specific version..."
	@if [ -z "$(ARG)" ]; then echo "Error: Missing ARG. Usage: make migrate-force ARG=<version>"; exit 1; fi
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_URL) force $(ARG)

.PHONY: migrate-version
migrate-version:
	@echo "Checking current migration version..."
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DATABASE_URL) version

.PHONY: seed
seed:
	@echo "Seeding the database..."
	@DB_ADDR=$(DATABASE_URL) go run ./cmd/migrate/seed/seed.go
	@echo "Seed completed successfully."

.PHONY: clear
clear:
	@echo "Clearing the database..."
	@DB_ADDR=$(DATABASE_URL) go run ./cmd/migrate/clear/clear.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: test
test:
	@go test -v ./...