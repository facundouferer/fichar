#!/bin/bash

echo "Validating docker-compose.yml configuration..."
docker-compose config

if [ $? -eq 0 ]; then
  echo "✓ docker-compose.yml is valid"
  exit 0
else
  echo "✗ docker-compose.yml has errors"
  exit 1
fi
