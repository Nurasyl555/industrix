package contracts

import "context"

// UserProvider is implemented by the identity module, consumed by other modules
type UserProvider interface {
	GetUserBasic(ctx context.Context, userID string) (*UserBasic, error)
}

// CompanyProvider is implemented by the integrity module, consumed by other modules
type CompanyProvider interface {
	GetCompanyBasic(ctx context.Context, companyID string) (*CompanyBasic, error)
}

// EquipmentProvider is implemented by the catalog module, consumed by other modules
type EquipmentProvider interface {
	GetEquipmentBasic(ctx context.Context, equipmentID string) (*EquipmentBasic, error)
}

// ListingProvider is implemented by the listing module, consumed by other modules
type ListingProvider interface {
	GetListingBasic(ctx context.Context, listingID string) (*ListingBasic, error)
}

// DealProvider is implemented by the deal module, consumed by the payment
// module to validate participants and coordinate escrow with deal state.
type DealProvider interface {
	GetDealBasic(ctx context.Context, dealID string) (*DealBasic, error)
}

// SubscriptionProvider is implemented by the integrity module, consumed by the
// listing module to enforce per-plan limits.
type SubscriptionProvider interface {
	// ListingLimit returns the maximum number of live/pending listings allowed
	// for the user's current plan. -1 means unlimited.
	ListingLimit(ctx context.Context, userID string) int
}

// Notifier is implemented by the notification module and consumed by any module
// that emits user-facing events. Fire-and-forget: emitting a notification must
// never fail the underlying operation, so there's no error return.
type Notifier interface {
	Notify(ctx context.Context, userID, ntype, message, link string)
}

// EventPublisher is implemented by the platform's Kafka layer and consumed by
// any module that emits domain events onto the bus. Fire-and-forget, mirroring
// Notifier: publishing must never fail the underlying operation, so there's no
// error return — transport failures are logged by the implementation.
//
// key is the partition key (usually the entity ID) so all events for one entity
// land on the same partition and stay ordered. payload is JSON-marshalled by the
// implementation.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any)
}

// Kafka topic names — the canonical domain-event vocabulary shared across
// modules. Kept in sync with infra/kafka/topics.sh (topics are pre-created;
// auto-create is disabled on the broker).
const (
	TopicEquipmentCreated     = "equipment.created"
	TopicEquipmentUpdated     = "equipment.updated"
	TopicEquipmentDeleted     = "equipment.deleted"
	TopicListingSubmitted     = "listing.submitted"
	TopicListingPublished     = "listing.published"
	TopicListingDeactivated   = "listing.deactivated"
	TopicDealStatusChanged    = "deal.status.changed"
	TopicPaymentCompleted     = "payment.completed"
	TopicPaymentFailed        = "payment.failed"
	TopicPaymentRefunded      = "payment.refunded"
	TopicNotificationDispatch = "notification.dispatch"
)

// UserBasic is a minimal user DTO for cross-module communication
type UserBasic struct {
	ID        string
	FirstName string
	LastName  string
	AvatarURL string
}

// CompanyBasic is a minimal company DTO for cross-module communication
type CompanyBasic struct {
	ID       string
	Name     string
	Verified bool
}

// EquipmentBasic is a minimal equipment DTO for cross-module communication
type EquipmentBasic struct {
	ID      string
	Title   string
	OwnerID string
}

// ListingBasic is a minimal listing DTO for cross-module communication
type ListingBasic struct {
	ID          string
	EquipmentID string
	SellerID    string
	Status      string
	ListingType string // sale | rental
}

// DealBasic is a minimal deal DTO for cross-module communication
type DealBasic struct {
	ID        string
	ListingID string
	BuyerID   string
	SellerID  string
	Status    string
}
