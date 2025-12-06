package domain

type Merchant struct {
	C_ID   string `json:"id"`
	C_NM string `json:"name"`
    C_PHONE *string `json:"phone"`
	C_EMAIL *string `json:"email"`
}

type MerchantRepository interface {
    Create(u *Merchant) error
    GetAll() ([]Merchant, error)
    GetByID(id string) (*Merchant, error)
    Update(u *Merchant) error
    Delete(id string) error
}