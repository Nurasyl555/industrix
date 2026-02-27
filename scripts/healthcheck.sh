#!/bin/bash

SERVICES=(
  "gateway:8080"
  "trust:8081"
  "inventory:8081"
  "transaction:8081"
  "content:8081"
  "communication:8081"
)

echo "Checking service health..."

for service in "${SERVICES[@]}"; do
  name=$(echo $service | cut -d: -f1)
  port=$(echo $service | cut -d: -f2)

  if curl -s "http://localhost:$port/health" > /dev/null; then
    echo "✅ $name is UP"
  else
    echo "❌ $name is DOWN"
  fi
done
