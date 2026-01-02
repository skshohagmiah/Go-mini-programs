#!/bin/bash

# Test script for Task Manager API
# Make sure the server is running first with: go run main.go

echo "=== Task Manager API Test ==="
echo ""

BASE_URL="http://localhost:8080"

echo "1. Creating first task..."
curl -X POST $BASE_URL/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go packages","description":"Study fmt, os, http, json, etc."}'
echo -e "\n"

echo "2. Creating second task..."
curl -X POST $BASE_URL/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Build REST API","description":"Create a mini project using essential packages"}'
echo -e "\n"

echo "3. Getting all tasks..."
curl $BASE_URL/tasks
echo -e "\n"

echo "4. Getting task with ID 1..."
curl $BASE_URL/tasks/1
echo -e "\n"

echo "5. Updating task 1 (mark as completed)..."
curl -X PUT $BASE_URL/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'
echo -e "\n"

echo "6. Getting all tasks again (task 1 should be completed)..."
curl $BASE_URL/tasks
echo -e "\n"

echo "7. Deleting task 2..."
curl -X DELETE $BASE_URL/tasks/2
echo -e "\n"

echo "8. Final task list..."
curl $BASE_URL/tasks
echo -e "\n"

echo "=== Test Complete ==="