#!/bin/bash
echo "Starting CandyPro OEM Backend..."
cd "$(dirname "$0")/backend"
go run cmd/api/main.go
