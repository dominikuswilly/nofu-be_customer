package domain

type Merchant struct {
	C_ID   string `json:"id"`
	C_NM string `json:"name"`
}

type MerchantRepository interface {
    Create(u *Merchant) error
    GetAll() ([]Merchant, error)
    GetByID(id string) (*Merchant, error)
    Update(u *Merchant) error
    Delete(id string) error
}