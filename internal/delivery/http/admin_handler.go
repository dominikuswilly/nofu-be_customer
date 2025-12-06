package http

import (
	"encoding/json"
	"net/http"

	"be_customer/internal/domain"
	"be_customer/internal/usecase"

	"github.com/gorilla/mux"
)

type AdminHandler struct {
	usecase *usecase.AdminUsecase
}

func NewAdminHandler(r *mux.Router, u *usecase.AdminUsecase) {
	handler := &AdminHandler{usecase: u}

	// Create a subrouter with prefix "/api"
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/admins", handler.Create).Methods("POST")
	api.HandleFunc("/admins", handler.GetAll).Methods("GET")
	api.HandleFunc("/admins/{id}", handler.GetByID).Methods("GET")
	api.HandleFunc("/admins/{id}", handler.Update).Methods("PUT")
	api.HandleFunc("/admins/{id}", handler.Delete).Methods("DELETE")
}

func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var admin domain.Admin
	json.NewDecoder(r.Body).Decode(&admin)
	if err := h.usecase.Create(&admin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(admin)
}

func (h *AdminHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	admins, _ := h.usecase.GetAll()
	json.NewEncoder(w).Encode(admins)
}

func (h *AdminHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	admin, err := h.usecase.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(admin)
}

func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var admin domain.Admin
	json.NewDecoder(r.Body).Decode(&admin)
	admin.C_ID = id
	if err := h.usecase.Update(&admin); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(admin)
}

func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.usecase.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
