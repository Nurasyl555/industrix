#!/bin/bash

# Alias for MinIO client
mc alias set myminio http://minio:9000 minio minio123

# Create buckets
buckets=(
  "equipment-media"
  "documents"
  "chat-files"
  "dispute-evidence"
)

for bucket in "${buckets[@]}"; do
  mc mb myminio/"$bucket" --ignore-existing
  echo "Created bucket: $bucket"
done

# Set policies
mc anonymous set public myminio/equipment-media
mc anonymous set none myminio/documents
mc anonymous set none myminio/chat-files
mc anonymous set none myminio/dispute-evidence

echo "All buckets created and policies set."
