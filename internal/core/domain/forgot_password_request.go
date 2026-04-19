package domain

import "time"

type ForgotPasswordRequest struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	UserEmail         string    `json:"user_email"`
	Status            string    `json:"status"` // pending, accepted, rejected
	TemporaryPassword string    `json:"temporary_password,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
