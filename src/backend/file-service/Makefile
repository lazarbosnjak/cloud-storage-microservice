
.PHONY: test test-integration test-all

FILE_SERVICE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
BACKEND_DIR := $(abspath $(FILE_SERVICE_DIR)/..)

COMPOSE = docker compose -f $(BACKEND_DIR)/compose.yaml -p backend

test:
	cd $(FILE_SERVICE_DIR) && go test -v -count=1 ./...

test-integration:
	$(COMPOSE) up -d --wait mongo-file-service
	docker run --rm \
		--network backend_default \
		--env-file $(FILE_SERVICE_DIR)/.env \
		-e MONGO_DB=files_test_db \
		-e RUN_MONGO_INTEGRATION=1 \
		-v "$(BACKEND_DIR):/app" \
		-w /app/file-service \
		golang:1.26.10-alpine \
		go test -v -count=1 ./internal/infrastructure/repository

test-all: test test-integration
