package booking

import (
	"context"
	"time"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

const dateLayout = "2006-01-02"

// Service defines the booking service interface.
type Service interface {
	CreateBooking(ctx context.Context, renterID string, req CreateBookingRequest) (*Booking, error)
	BookedRanges(ctx context.Context, listingID string) ([]*DateRange, error)
	ListMine(ctx context.Context, renterID string) ([]*Booking, error)
	Cancel(ctx context.Context, id, userID string) error
	// Quote estimates the rental cost for a listing over a date range.
	Quote(ctx context.Context, listingID, startDate, endDate string) (*Quote, error)
}

type service struct {
	repo     *Repository
	listings contracts.ListingProvider
	notifier contracts.Notifier
}

func NewService(repo *Repository, listings contracts.ListingProvider, notifier contracts.Notifier) Service {
	return &service{repo: repo, listings: listings, notifier: notifier}
}

// parseRange validates a date range and returns the inclusive day count.
func parseRange(startStr, endStr string) (start, end time.Time, days int, err error) {
	start, err = time.Parse(dateLayout, startStr)
	if err != nil {
		return start, end, 0, errors.New(errors.CodeValidation, "Invalid start date")
	}
	end, err = time.Parse(dateLayout, endStr)
	if err != nil {
		return start, end, 0, errors.New(errors.CodeValidation, "Invalid end date")
	}
	if end.Before(start) {
		return start, end, 0, errors.New(errors.CodeValidation, "End date must be on or after start date")
	}
	// Inclusive of both ends — a same-day rental is one day (matches the DB
	// exclusion constraint's '[]' range).
	days = int(end.Sub(start).Hours()/24) + 1
	return start, end, days, nil
}

func (s *service) CreateBooking(ctx context.Context, renterID string, req CreateBookingRequest) (*Booking, error) {
	if req.ListingID == "" {
		return nil, errors.New(errors.CodeValidation, "Listing is required")
	}
	start, _, days, err := parseRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	// Compare against today's date (midnight) so booking for today is allowed.
	today := time.Now().Truncate(24 * time.Hour)
	if start.Before(today) {
		return nil, errors.New(errors.CodeValidation, "Cannot book dates in the past")
	}

	l, err := s.listings.GetListingBasic(ctx, req.ListingID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Listing does not exist")
	}
	if l.Status != "active" {
		return nil, errors.New(errors.CodeValidation, "Listing is not active")
	}
	if l.ListingType != "rental" {
		return nil, errors.New(errors.CodeValidation, "This listing is not for rent")
	}
	if l.SellerID == renterID {
		return nil, errors.New(errors.CodeValidation, "You cannot book your own listing")
	}

	b := &Booking{
		ListingID:  req.ListingID,
		RenterID:   renterID,
		OwnerID:    l.SellerID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TotalPrice: rentalTotal(l.Price, l.PricePeriod, days),
	}
	// The DB exclusion constraint is the race-free guard against double booking.
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	if s.notifier != nil {
		s.notifier.Notify(ctx, b.OwnerID, "booking", "Your rental was booked for "+b.StartDate+" → "+b.EndDate, "/shop/bookings")
	}
	return b, nil
}

func (s *service) Quote(ctx context.Context, listingID, startDate, endDate string) (*Quote, error) {
	if listingID == "" {
		return nil, errors.New(errors.CodeValidation, "Listing is required")
	}
	_, _, days, err := parseRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	l, err := s.listings.GetListingBasic(ctx, listingID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Listing does not exist")
	}
	if l.ListingType != "rental" {
		return nil, errors.New(errors.CodeValidation, "This listing is not for rent")
	}
	return &Quote{
		ListingID: listingID,
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
		Period:    l.PricePeriod,
		UnitPrice: l.Price,
		Units:     rentalUnits(l.PricePeriod, days),
		Total:     rentalTotal(l.Price, l.PricePeriod, days),
	}, nil
}

func (s *service) BookedRanges(ctx context.Context, listingID string) ([]*DateRange, error) {
	return s.repo.ConfirmedRanges(ctx, listingID)
}

func (s *service) ListMine(ctx context.Context, renterID string) ([]*Booking, error) {
	return s.repo.ListByRenter(ctx, renterID)
}

func (s *service) Cancel(ctx context.Context, id, userID string) error {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// Either party (renter or owner) may cancel.
	if b.RenterID != userID && b.OwnerID != userID {
		return errors.New(errors.CodeUnauthorized, "You are not part of this booking")
	}
	return s.repo.Cancel(ctx, id)
}
