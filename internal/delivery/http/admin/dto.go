package admin

import "time"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponseData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
}

type apiResp struct {
	ResponseCode    string      `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"`
}

type itemAdmin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

type itemDetailAdmin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}
