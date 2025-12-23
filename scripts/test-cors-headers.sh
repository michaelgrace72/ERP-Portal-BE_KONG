#!/bin/bash

echo "Testing CORS Headers from Kong Gateway"
echo "======================================="
echo ""

echo "1. Testing OPTIONS preflight request..."
curl -i -X OPTIONS http://localhost:3600/api/v1/auth/phantom-login \
  -H "Origin: http://localhost:5052" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type"

echo ""
echo ""
echo "2. Testing actual POST request with CORS..."
curl -i -X POST http://localhost:3600/api/v1/auth/phantom-login \
  -H "Origin: http://localhost:5052" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test"}'
