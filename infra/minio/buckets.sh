#!/bin/bash
# MinIO Buckets Creation Script
# Run this script once on first start to create all required buckets
# Usage: ./buckets.sh

# MinIO configuration
MINIO_ENDPOINT="minio:9000"
MINIO_ROOT_USER="minio"
MINIO_ROOT_PASSWORD="minio123"

# Wait for MinIO to be ready
echo "Waiting for MinIO to be ready..."
until mc alias set myminio http://${MINIO_ENDPOINT} ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} 2>/dev/null; do
    echo "Waiting..."
    sleep 2
done

echo "MinIO is ready!"

# Create buckets
declare -a BUCKETS=(
    "equipment-media"
    "documents"
    "chat-files"
    "dispute-evidence"
)

echo "Creating buckets..."

for bucket in "${BUCKETS[@]}"; do
    # Create bucket if not exists
    mc mb myminio/${bucket} 2>/dev/null && echo "Created bucket: ${bucket}" || echo "Bucket ${bucket} already exists"
    
    # Set appropriate policies
    case $bucket in
        "equipment-media")
            # Public read for equipment images
            mc anonymous set download myminio/${bucket} 2>/dev/null
            echo "Set public download policy for: ${bucket}"
            ;;
        "documents"|"dispute-evidence")
            # Private - only authenticated access
            mc anonymous set none myminio/${bucket} 2>/dev/null
            echo "Set private policy for: ${bucket}"
            ;;
        "chat-files")
            # Private with specific access pattern
            mc anonymous set none myminio/${bucket} 2>/dev/null
            echo "Set private policy for: ${bucket}"
            ;;
    esac
done

echo ""
echo "Setting up CORS configuration for equipment-media bucket..."
mc cors set myminio/equipment-media <<EOF
[
    {
        "AllowedOrigins": ["*"],
        "AllowedMethods": ["GET", "PUT", "HEAD"],
        "AllowedHeaders": ["*"],
        "ExposeHeaders": [],
        "MaxAgeSeconds": 3600
    }
]
EOF

echo ""
echo "MinIO buckets setup complete!"
echo ""
echo "Listing all buckets:"
mc ls myminio
