#!/bin/bash
set -e

echo "Generating Protocol Buffers..."
# buf.yaml / buf.gen.yaml live at the repo root — do not cd into backend/,
# buf would fail to find its config there.
buf generate
echo "✅ Proto generation complete."
