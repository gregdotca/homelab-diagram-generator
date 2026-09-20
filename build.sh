#!/bin/bash
set -e
cd "$(dirname "$0")"
go mod tidy
go build -o homelab-diagram .
echo "Built: ./homelab-diagram"
echo "Try: ./homelab-diagram -input example/homelab.yaml -output diagram.svg"
