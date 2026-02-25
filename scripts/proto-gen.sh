#!/bin/bash

set -e

# Proto generation script for Industrix
# Generates Go and TypeScript code from .proto files
# Requires: buf, protoc, protoc-gen-go, protoc-gen-go-grpc, protoc-gen-ts

echo "🔧 Generating protobuf code..."

# Navigate to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Create gen directories if they don't exist
mkdir -p backend/proto/gen/go
mkdir -p backend/proto/gen/ts

# Check if buf is installed
if ! command -v buf &> /dev/null; then
    echo "⚠️  buf not found. Attempting to use protoc directly..."
    
    # Fall back to protoc
    PROTOS_DIR="./backend/proto"
    GEN_GO_DIR="./backend/proto/gen/go"
    GEN_TS_DIR="./backend/proto/gen/ts"
    
    # Generate Go code
    echo "📦 Generating Go code..."
    protoc --proto_path="$PROTOS_DIR" \
           --go_out="$GEN_GO_DIR" \
           --go_opt=paths=source_relative \
           --go-grpc_out="$GEN_GO_DIR" \
           --go-grpc_opt=paths=source_relative \
           "$PROTOS_DIR"/**/*.proto
    
    # Generate TypeScript code (if protoc-gen-ts is available)
    if command -v protoc-gen-ts &> /dev/null; then
        echo "📦 Generating TypeScript code..."
        protoc --proto_path="$PROTOS_DIR" \
               --ts_out="$GEN_TS_DIR" \
               --ts_opt=paths=source_relative \
               "$PROTOS_DIR"/**/*.proto
    else
        echo "⚠️  protoc-gen-ts not found. Skipping TypeScript generation."
    fi
else
    # Use buf for generation
    echo "📦 Using buf to generate code..."
    
    # Generate using buf
    buf generate
    
    # If buf.yaml doesn't exist, create a basic one
    if [ ! -f "buf.yaml" ]; then
        echo "📝 Creating buf.yaml configuration..."
        cat > buf.yaml << 'EOF'
version: v1
name: buf.build/industrix/industrix
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc/grpc
plugins:
  - plugin: buf.build/protocolbuffers/go:v1.31.0
    out: backend/proto/gen/go
    opt: paths=source_relative
  - plugin: buf.build/grpc/go:v1.3.0
    out: backend/proto/gen/go
    opt: paths=source_relative
EOF
    fi
    
    # Generate again with config
    buf generate
fi

echo "✅ Proto generation complete!"
echo "📁 Generated files:"
echo "   - Go: backend/proto/gen/go/"
echo "   - TypeScript: backend/proto/gen/ts/"

# List generated files
echo ""
echo "📄 Go files generated:"
find backend/proto/gen/go -name "*.pb.go" 2>/dev/null | head -20 || echo "   (none found)"
