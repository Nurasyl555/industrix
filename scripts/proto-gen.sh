#!/bin/bash
set -e

echo "Generating Protocol Buffers..."
cd backend
buf generate
echo "✅ Proto generation complete."
