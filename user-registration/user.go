package user_registration

import "time"

type User struct {
	Email            string
	Password         string
	ConfirmationCode string
	CreatedAt        time.Time
	ConfirmedAt      *time.Time
}
