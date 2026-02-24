.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build application
	mkdir -p ./bin
	go build -v -o ./bin/app ./cmd/app

.PHONY: migrate-create
migrate-create: ## create new migration, pass name by NAME env var
	migrate create -ext sql -dir migrations $(NAME)

.PHONY: migrate-up
migrate-up: ### migration up
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up

.PHONY: migrate-down
migrate-down: ### migration down
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' down

.PHONY: compose-up-integration-test
compose-up-integration-test: ### Run docker-compose with integration test
	#docker volume rm -f data ### remove old db
	docker compose up --build --abort-on-container-exit --exit-code-from app-integration

.PHONY: lint
lint: ### Run check style with go linter
	golangci-lint run --timeout 5m0s

.DEFAULT_GOAL := help