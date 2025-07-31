#!/bin/sh

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
  sleep 1
done

echo "Database is ready!"

# Run migrations
echo "Running database migrations..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f /app/migrations/001_create_streams_table.sql

echo "Migrations completed!"

# Start the encoder service
echo "Starting encoder service..."
exec ./encoder-service 