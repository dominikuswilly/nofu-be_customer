package merchant

import (
	"be_customer/internal/usecase"
	"net/http"

	"github.com/gorilla/mux"
)

// NewMerchantHandler registers merchant routes under /api
func NewMerchantHandler(r *mux.Router, u *usecase.MerchantUsecase) {
	handler := &MerchantHandler{usecase: u}
	api := r.PathPrefix("/api/customer").Subrouter()

	api.HandleFunc("/merchants", handler.Create).Methods(http.MethodPost)
	api.HandleFunc("/merchants", handler.GetAll).Methods(http.MethodGet)
	api.HandleFunc("/merchants/{id}", handler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/merchants/{id}", handler.Update).Methods(http.MethodPut)
	api.HandleFunc("/merchants/{id}", handler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/merchants/login", handler.Login).Methods(http.MethodPost)
}
