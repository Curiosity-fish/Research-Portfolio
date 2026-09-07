.PHONY: build test test-integration test-integration-down clean run generate migrate-up migrate-down migrate-down-one migrate-version

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/admin ./cmd/admin

test:
	go test ./...

test-integration:
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test

test-integration-down:
	docker compose -f docker-compose.test.yml down -v

clean:
	rm -rf bin/

run:
	go run ./cmd/server

generate:
	# Local Go 1.27 is newer than the ent generator currently supports;
	# pin the toolchain to the Go version declared in go.mod.
	GOTOOLCHAIN=go1.24.0 go generate ./ent/...

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-down-one:
	go run ./cmd/migrate down 1

migrate-version:
	go run ./cmd/migrate version
