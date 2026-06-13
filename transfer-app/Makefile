.PHONY: run build migrate-up migrate-down test tidy setup

APP=transfer-app
BINARY=bin/$(APP)
DB_URL=postgres://postgres:password@localhost:5432/transfer_app?sslmode=disable

run:
	go run cmd/server/main.go

build:
	go build -o $(BINARY) ./cmd/server

tidy:
	go mod tidy

test:
	go test ./... -v -race

setup:
	@echo "📦 Installing PostgreSQL..."
	sudo apt-get update -qq && sudo apt-get install -y -qq postgresql postgresql-contrib
	@echo "🚀 Starting PostgreSQL..."
	sudo service postgresql start
	sleep 2
	@echo "🔧 Setting up database..."
	sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'password';" 2>/dev/null || true
	sudo -u postgres psql -c "CREATE DATABASE transfer_app;" 2>/dev/null || true
	@echo "✅ Database ready — run: make migrate-up"

migrate-up:
	@export PGPASSWORD=password && \
	psql -U postgres -h localhost -d transfer_app -f migrations/001_create_users.sql && \
	psql -U postgres -h localhost -d transfer_app -f migrations/002_create_wallets.sql && \
	psql -U postgres -h localhost -d transfer_app -f migrations/003_create_transactions.sql && \
	echo "✅ Migrations applied"

migrate-down:
	@export PGPASSWORD=password && \
	psql -U postgres -h localhost -d transfer_app -c \
	"DROP TABLE IF EXISTS transactions; DROP TABLE IF EXISTS beneficiaries; DROP TABLE IF EXISTS wallets; DROP TABLE IF EXISTS otps; DROP TABLE IF EXISTS users CASCADE; DROP TYPE IF EXISTS transaction_type; DROP TYPE IF EXISTS transaction_status;" && \
	echo "✅ Migrations rolled back"

pg-start:
	sudo service postgresql start

pg-stop:
	sudo service postgresql stop

.DEFAULT_GOAL := run
