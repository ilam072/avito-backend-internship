up:
	docker-compose up -d
	go run ./cmd/server/main.go
down:
	docker-compose down

test:
	go test ./integration_tests/... -v