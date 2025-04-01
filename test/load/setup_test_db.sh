#!/bin/bash
# Script to setup dedicated load test database

DB_NAME="flights_load_test"
DB_USER="loadtester"
DB_PASS="testpass123"

echo "Creating load test database..."
psql -U postgres -c "CREATE DATABASE $DB_NAME;"
psql -U postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

echo "Running migrations on load test database..."
go run db/seed.go -dbname=$DB_NAME -dbuser=$DB_USER -dbpass=$DB_PASS

echo "Load test database setup complete:"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
