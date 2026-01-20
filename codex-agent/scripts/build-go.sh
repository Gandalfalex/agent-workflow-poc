#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building Go implement-ticket binary..."
cd "$PROJECT_DIR/cmd/implement-ticket"
go build -o "../../bin/implement-ticket"

echo "Binary created at: $PROJECT_DIR/bin/implement-ticket"
