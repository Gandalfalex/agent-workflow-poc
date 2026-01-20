#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$PROJECT_DIR")"

echo "Building Go implement-ticket binary..."
cd "$ROOT_DIR"
go build -o "$PROJECT_DIR/bin/implement-ticket" ./codex-agent/cmd/implement-ticket/

echo "Binary created at: $PROJECT_DIR/bin/implement-ticket"
