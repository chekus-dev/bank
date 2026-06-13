#!/bin/bash
set -e

echo "📦 Installing PostgreSQL..."
sudo apt-get update -qq
sudo apt-get install -y -qq postgresql postgresql-contrib

echo "🚀 Starting PostgreSQL..."
sudo service postgresql start
sleep 2

echo "🔧 Creating database and user..."
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'password';" 2>/dev/null || true
sudo -u postgres psql -c "CREATE DATABASE transfer_app;" 2>/dev/null || echo "DB may already exist, continuing..."

echo "📋 Running migrations..."
export PGPASSWORD=password
psql -U postgres -h localhost -d transfer_app -f migrations/001_create_users.sql
psql -U postgres -h localhost -d transfer_app -f migrations/002_create_wallets.sql
psql -U postgres -h localhost -d transfer_app -f migrations/003_create_transactions.sql

echo ""
echo "✅ Database ready!"
echo "✅ Run: go run cmd/server/main.go"
