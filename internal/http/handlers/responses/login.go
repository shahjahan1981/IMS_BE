package responses

import "myapp/internal/http/dto"

type LoginResponse struct {
	Status bool              `json:"status"`
	Token  string            `json:"token,omitempty"`
	User   *dto.UserDTO      `json:"user,omitempty"`
	Errors map[string]string `json:"errors"`
}