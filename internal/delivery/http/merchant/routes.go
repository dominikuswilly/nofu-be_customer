package merchant

import (
	"be_customer/internal/middleware"
	"be_customer/internal/usecase"
	"net/http"

	"github.com/gorilla/mux"
)

// NewMerchantHandler registers merchant routes under /api
func NewMerchantHandler(r *mux.Router, u *usecase.MerchantUsecase) {
	handler := &MerchantHandler{usecase: u}
	api := r.PathPrefix("/api/customer").Subrouter()

	// Public routes
	api.HandleFunc("/merchants/login", handler.Login).Methods(http.MethodPost)

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/merchants", handler.Create).Methods(http.MethodPost)
	protected.HandleFunc("/merchants", handler.GetAll).Methods(http.MethodGet)
	protected.HandleFunc("/merchants/{id}", handler.GetByID).Methods(http.MethodGet)
	protected.HandleFunc("/merchants/{id}", handler.Update).Methods(http.MethodPut)
	protected.HandleFunc("/merchants/{id}", handler.Delete).Methods(http.MethodDelete)
}
