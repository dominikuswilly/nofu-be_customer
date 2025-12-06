package usecase

import (
    "errors"
    "be_customer/internal/domain"
)

type AdminUsecase struct {
    repo domain.AdminRepository
}

func NewAdminUsecase(r domain.AdminRepository) *AdminUsecase {
    return &AdminUsecase{repo: r}
}

func (u *AdminUsecase) Create(admin *domain.Admin) error {
    if admin.C_NM == "" {
        return errors.New("name is required")
    }
    return u.repo.Create(admin)
}

func (u *AdminUsecase) GetAll() ([]domain.Admin, error) {
    return u.repo.GetAll()
}

func (u *AdminUsecase) GetByID(id string) (*domain.Admin, error) {
    return u.repo.GetByID(id)
}

func (u *AdminUsecase) Update(admin *domain.Admin) error {
    return u.repo.Update(admin)
}

func (u *AdminUsecase) Delete(id string) error {
    return u.repo.Delete(id)
}