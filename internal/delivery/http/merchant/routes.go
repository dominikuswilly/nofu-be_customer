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

	// Frontend routes
	frontend := api.PathPrefix("/frontend").Subrouter()

	// Internal routes
	_ = api.PathPrefix("/internal").Subrouter()

	// Public routes (Legacy)
	api.HandleFunc("/merchants/login", handler.Login).Methods(http.MethodPost)
	// Public routes (Frontend)
	frontend.HandleFunc("/merchants/login", handler.Login).Methods(http.MethodPost)

	// Protected routes (Legacy)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/merchants", handler.Create).Methods(http.MethodPost)
	protected.HandleFunc("/merchants", handler.GetAll).Methods(http.MethodGet)
	protected.HandleFunc("/merchants/{id}", handler.GetByID).Methods(http.MethodGet)
	protected.HandleFunc("/merchants/{id}", handler.Update).Methods(http.MethodPut)
	protected.HandleFunc("/merchants/{id}", handler.Delete).Methods(http.MethodDelete)

	// Protected routes (Frontend)
	protectedFrontend := frontend.PathPrefix("").Subrouter()
	protectedFrontend.Use(middleware.AuthMiddleware)
	protectedFrontend.HandleFunc("/merchants", handler.Create).Methods(http.MethodPost)
	protectedFrontend.HandleFunc("/merchants", handler.GetAll).Methods(http.MethodGet)
	protectedFrontend.HandleFunc("/merchants/{id}", handler.GetByID).Methods(http.MethodGet)
	protectedFrontend.HandleFunc("/merchants/{id}", handler.Update).Methods(http.MethodPut)
	protectedFrontend.HandleFunc("/merchants/{id}", handler.Delete).Methods(http.MethodDelete)
	protectedFrontend.HandleFunc("/merchants-stock", handler.GetMerchantsStock).Methods(http.MethodGet)
}
