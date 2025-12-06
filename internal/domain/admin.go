package domain

type Admin struct {
	C_ID   string `json:"id"`
	C_NM string `json:"name"`
	C_PHONE string `json:"phone"`
	C_EMAIL string `json:"email"`
}

type AdminRepository interface {
    Create(u *Admin) error
    GetAll() ([]Admin, error)
    GetByID(id string) (*Admin, error)
    Update(u *Admin) error
    Delete(id string) error
}