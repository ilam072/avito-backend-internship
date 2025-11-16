up:
	docker-compose up -d
down:
	docker-compose down

test:
	go test ./integration_tests/... -v