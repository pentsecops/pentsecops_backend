package domain

import "time"

// Stakeholder defines stakeholder profile information
type Stakeholder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Address   string    `json:"address"`
	About     string    `json:"about"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StakeholderResponse is the response model for stakeholder profile
type StakeholderResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Address   string    `json:"address"`
	About     string    `json:"about"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Stakeholder to StakeholderResponse
func (s *Stakeholder) ToResponse() *StakeholderResponse {
	return &StakeholderResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		Email:     s.Email,
		Company:   s.Company,
		Address:   s.Address,
		About:     s.About,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
