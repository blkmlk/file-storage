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

.PHONY: start
start:
	@echo 'Running local...'
	docker-compose up postgres --build -d
	docker-compose up migration --build -d
	docker-compose up uploader --build -d
	docker-compose up storage-1 --build -d
	docker-compose up storage-2 --build -d
