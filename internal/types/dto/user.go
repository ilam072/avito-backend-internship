package dto

type User struct {
	ID       string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

type Users struct {
	Users []User `json:"users"`
}

type TeamWithMembers struct {
	TeamName string `json:"team_name" validate:"required"`
	Members  []User `json:"members" validate:"required"`
}

type SetUserIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type UpdateUserResponse struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
