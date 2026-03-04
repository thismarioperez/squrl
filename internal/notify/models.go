package notify

import "time"

// DefaultDuration is the duration used when Notification.Duration is zero.
const DefaultDuration = 5 * time.Second

// Notification holds the data for a desktop notification.
// Duration controls how long the notification is displayed:
//   - 0 (zero value): uses DefaultDuration (3 seconds)
//   - negative (e.g. -1): notification is displayed indefinitely until dismissed
//   - positive: notification is displayed for that duration
type Notification = struct {
	Title      string
	Message    string
	OnActivate func()
	Duration   time.Duration
}