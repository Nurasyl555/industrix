#!/bin/bash

# Wait for Kafka to be ready
echo "Waiting for Kafka to be ready..."
cub kafka-ready -b kafka:29092 1 20

# Create topics
topics=(
  "user.profile.updated"
  "company.verified"
  "company.rejected"
  "equipment.created"
  "equipment.updated"
  "equipment.deleted"
  "listing.submitted"
  "listing.published"
  "listing.deactivated"
  "listing.price_changed"
  "deal.status.changed"
  "payment.completed"
  "payment.failed"
  "payment.refunded"
  "booking.confirmed"
  "booking.cancelled"
  "message.sent"
  "notification.dispatch"
  "media.uploaded"
  "media.processed"
  "review.created"
  "subscription.expired"
  "subscription.activated"
  "dispute.filed"
  "dispute.resolved"
)

for topic in "${topics[@]}"; do
  kafka-topics --create --if-not-exists --bootstrap-server kafka:29092 --partitions 1 --replication-factor 1 --topic "$topic"
  echo "Created topic: $topic"
done

echo "All topics created."
