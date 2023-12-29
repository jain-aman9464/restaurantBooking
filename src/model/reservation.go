package model

import (
	"time"
)

// Reservation represents a reservation entity
type Reservation struct {
	ID           int
	UserID       int
	RestaurantID int
	TableID      int
	DateTime     time.Time
	Status       string
}
