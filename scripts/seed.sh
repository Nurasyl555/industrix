#!/bin/bash
set -e
echo "🌱 Seeding development data..."
node "$(dirname "$0")/seed.mjs"
echo "✅ Seeding complete!"
