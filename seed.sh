#!/bin/bash

echo "Building seed tool..."
go build -o bin/seed ./cmd/seed

echo ""
echo "Running database seed..."
./bin/seed

echo ""
echo "Seed completed!"
