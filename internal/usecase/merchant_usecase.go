package usecase

import (
	"be_customer/internal/domain"
	"errors"
)

type MerchantUsecase struct {
	repo domain.MerchantRepository
}

func NewMerchantUsecase(r domain.MerchantRepository) *MerchantUsecase {
	return &MerchantUsecase{repo: r}
}

func (u *MerchantUsecase) Create(merchant *domain.Merchant) error {
	if merchant.C_NM == "" {
		return errors.New("name is required")
	}
	return u.repo.Create(merchant)
}

func (u *MerchantUsecase) GetAll() ([]domain.Merchant, error) {
	return u.repo.GetAll()
}

func (u *MerchantUsecase) GetByID(id string) (*domain.Merchant, error) {
	return u.repo.GetByID(id)
}

func (u *MerchantUsecase) Update(merchant *domain.Merchant) error {
	return u.repo.Update(merchant)
}

func (u *MerchantUsecase) Delete(id string) error {
	return u.repo.Delete(id)
}

func (u *MerchantUsecase) FindByUsername(username string) (*domain.Merchant, error) {
	return u.repo.FindByUsername(username)
}
