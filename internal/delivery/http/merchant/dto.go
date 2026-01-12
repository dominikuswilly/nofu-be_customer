package merchant

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

type itemMerchant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Active   bool   `json:"active"`
}

type itemDetailMerchant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Active   bool   `json:"active"`
}

type stockData struct {
	ID         string `json:"id"`
	MerchantId string `json:"merchantId"`
	GivenBy    string `json:"givenBy"`
	CreatedBy  string `json:"createdBy"`
	CreatedAt  string `json:"createdAt"`
	Status     string `json:"status"`
}

type transactionStockResp struct {
	ResponseCode    string      `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            []stockData `json:"data"`
}

type itemMerchantStock struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Stocks []stockInfo `json:"stocks"`
}

type stockInfo struct {
	ID        string `json:"id"`
	GivenBy   string `json:"givenBy"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
	Status    string `json:"status"`
}
