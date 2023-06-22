.PHONY: start
start:
	go run main.go start

.PHONY: migrate
migrate:
	go run main.go migrate

.PHONY: test
test: unit-test

.PHONY: unit-test
unit-test:
	@echo 'Running tests...'
	docker-compose -p test up -d && sleep 5
	(\
		export GCS_BUCKET_NAME="test"; \
		go test -race -cover -timeout 1m -count=1 -p 1 ./internal/... \
	) || (docker-compose -p test down -v ; exit 1)
	docker-compose -p test down -v

.PHONY: generate
generate: go-generate

.PHONY: go-generate
go-generate:
	go generate ./...

.PHONY: local-run
local-run:
	@echo 'Running local...'
	docker-compose -p test up -d postgres && sleep 5
	(\
		export DATABASE_URL="postgres://root:root@127.0.0.1:25432/file-storage-test?sslmode=disable"; \
		export UPLOAD_HOST="127.0.0.1:8080"; \
		go run cmd/migration/main.go migrate; \
		go run cmd/uploader/main.go start; \
	) || (docker-compose -p test down -v ; exit 1)
	docker-compose -p test down -v
