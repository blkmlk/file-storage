.PHONY: test
test: unit-test

.PHONY: unit-test
unit-test:
	@echo 'Running tests...'
	go test -cover -count=1 -p 1 ./internal/...

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
	docker-compose up storage-1 storage-2 storage-3 storage-4 storage-5 --build -d

.PHONY: client-test
client-test:
	@echo 'Running client test'
	go run cmd/client/main.go

.PHONY: stop
stop:
	@echo 'Running local...'
	docker-compose down
