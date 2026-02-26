#!/bin/bash
# Kafka Topics Creation Script
# Run this script once on first start to create all required topics
# Usage: ./topics.sh

KAFKA_BROKER="kafka:29092"

# Define topics with partition counts and retention settings
declare -a TOPICS=(
    # Equipment events
    "equipment.created:3:86400000"
    "equipment.updated:3:86400000"
    "equipment.deleted:3:86400000"
    
    # Listing events
    "listing.created:3:86400000"
    "listing.deactivated:3:86400000"
    
    # Company verification events
    "company.verified:3:604800000"
    "company.rejected:3:604800000"
    
    # Deal events
    "deal.status.changed:3:2592000000"
    "deal.completed:3:2592000000"
    
    # Payment events
    "payment.completed:3:2592000000"
    "payment.failed:3:604800000"
    
    # Review events
    "review.created:3:2592000000"
    
    # Message events
    "message.sent:9:604800000"
    
    # Media events
    "media.uploaded:3:86400000"
    "media.processed:3:86400000"
    
    # Dispute events
    "dispute.filed:3:2592000000"
    "dispute.resolved:3:2592000000"
    
    # Subscription events
    "subscription.activated:3:604800000"
    "subscription.expired:3:604800000"
    
    # Delivery events
    "delivery.status.changed:3:604800000"
    
    # Moderation events
    "moderation.action.taken:3:2592000000"
    
    # Favorite events
    "favorite.price_dropped:3:86400000"
)

echo "Creating Kafka topics..."

for topic_config in "${TOPICS[@]}"; do
    IFS=':' read -r topic_name partitions retention <<< "$topic_config"
    
    kafka-topics.sh \
        --bootstrap-server "$KAFKA_BROKER" \
        --create \
        --topic "$topic_name" \
        --partitions "$partitions" \
        --replication-factor 1 \
        --config retention.ms="$retention" \
        --config cleanup.policy=delete \
        2>/dev/null || echo "Topic $topic_name already exists or creation failed"
    
    echo "Created/Verified topic: $topic_name (partitions: $partitions, retention: ${retention}ms)"
done

echo "Kafka topics creation complete!"
echo ""
echo "Listing all topics:"
kafka-topics.sh --bootstrap-server "$KAFKA_BROKER" --list
