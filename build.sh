#!/bin/bash
set -e
cd "$(dirname "$0")"
go mod tidy
go build -o hdg .
echo "Built: ./hdg"
echo "Try: ./hdg -input sample.yaml -output sample.svg"
