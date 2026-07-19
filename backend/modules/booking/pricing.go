package booking

import "math"

// rentalUnits converts an inclusive day count into billable periods for a
// rental rate. Week/month bill per *started* period (a 9-day weekly rental
// bills 2 weeks). Unknown/empty period is treated as per-day.
func rentalUnits(period string, days int) int {
	if days < 1 {
		days = 1
	}
	switch period {
	case "week":
		return int(math.Ceil(float64(days) / 7))
	case "month":
		return int(math.Ceil(float64(days) / 30))
	default: // day
		return days
	}
}

// rentalTotal is the total cost for `days` inclusive days at the given
// per-period rate.
func rentalTotal(price float64, period string, days int) float64 {
	return price * float64(rentalUnits(period, days))
}
