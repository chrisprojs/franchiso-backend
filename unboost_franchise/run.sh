#!/bin/bash

# Script untuk menjalankan unboost franchise service
# Pastikan environment variables sudah diset

# Set working directory
cd "$(dirname "$0")"

# Load environment variables jika file .env ada
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check required environment variables
if [ -z "$PG_ADDR" ] || [ -z "$PG_USER" ] || [ -z "$PG_PASSWORD" ] || [ -z "$PG_DATABASE" ]; then
    echo "Error: Required environment variables not set"
    echo "Please set PG_ADDR, PG_USER, PG_PASSWORD, PG_DATABASE"
    exit 1
fi

# Run the application
echo "Starting unboost franchise service..."
go run main.go

if [ $? -eq 0 ]; then
    echo "Unboost franchise service completed successfully"
else
    echo "Unboost franchise service failed"
    exit 1
fi 