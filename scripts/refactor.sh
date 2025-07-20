#!/bin/bash

# Backup original main.go
cp main.go main.go.backup

echo "Starting refactoring process..."
echo "1. Created backup: main.go.backup"
echo "2. New structure created in internal/ and cmd/"
echo "3. Next steps:"
echo "   - Extract handlers to respective packages"
echo "   - Move services to internal/services"
echo "   - Update imports and dependencies"
echo "   - Test each module independently"

echo "Refactoring structure:"
echo "├── cmd/server/main.go (new clean main)"
echo "├── internal/"
echo "│   ├── handlers/"
echo "│   │   ├── telegram/"
echo "│   │   ├── market/"
echo "│   │   ├── solana/"
echo "│   │   ├── portfolio/"
echo "│   │   ├── arbitration/"
echo "│   │   ├── sniping/"
echo "│   │   ├── token/"
echo "│   │   └── system/"
echo "│   ├── services/"
echo "│   ├── middleware/"
echo "│   ├── websocket/"
echo "│   └── server/"
echo "└── pkg/ (existing packages)"