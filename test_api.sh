#!/bin/bash

echo "Testing Books API endpoints..."
echo ""

echo "1. Creating a book..."
BOOK_ID=$(curl -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language","author":"Alan Donovan","isbn":"978-0134190440"}' \
  | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

echo "   Created book with ID: $BOOK_ID"
echo ""

echo "2. Getting all books..."
curl -s http://localhost:8080/books | jq .
echo ""

echo "3. Getting book by ID..."
curl -s http://localhost:8080/books/$BOOK_ID | jq .
echo ""

echo "4. Updating the book..."
curl -s -X PUT http://localhost:8080/books/$BOOK_ID \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language - 2nd Edition","author":"Alan Donovan & Brian Kernighan","isbn":"978-0134190440"}' \
  | jq .
echo ""

echo "5. Getting updated book..."
curl -s http://localhost:8080/books/$BOOK_ID | jq .
echo ""

echo "6. Deleting the book..."
curl -s -X DELETE http://localhost:8080/books/$BOOK_ID
echo "   Book deleted"
echo ""

echo "7. Verifying deletion..."
curl -s http://localhost:8080/books/$BOOK_ID
echo ""

echo "API test completed!"
