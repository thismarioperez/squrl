package notify

import "time"

type Notification = struct {
	Title string
	Message string
	OnActivate func()
	Duration time.Duration
}