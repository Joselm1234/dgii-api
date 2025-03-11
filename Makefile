start:
	make start-services db_migrate

start-services:
    @if [ -f docker-compose.yml ]; then \
        docker compose up -d; \
    else \
        echo "docker-compose.yml not found, skipping docker compose"; \
    fi

stop-services:
	docker compose down

build-image:
	docker build -t dgii-api .

db_reset:
	sudo -u postgres psql -c "DROP DATABASE IF EXISTS dgii-api"
	sudo -u postgres psql -c "CREATE DATABASE dgii-api"

	make db_migrate

db_migrate:
	go run cmd/bun/main.go -env=dev db init
	go run cmd/bun/main.go -env=dev db migrate

# test:
# 	TZ= go test ./org
# 	TZ= go test ./blog

# api_test:
# 	TZ= go run cmd/bun/main.go -env=test api &
# 	APIURL=http://localhost:8000/api ./scripts/run-api-tests.sh

default: start
