package http

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
    "be_customer/internal/domain"
    "be_customer/internal/usecase"
)

type AdminHandler struct {
    usecase *usecase.AdminUsecase
}

func NewAdminHandler(r *mux.Router, u *usecase.AdminUsecase) {
    handler := &AdminHandler{usecase: u}

    r.HandleFunc("/admins", handler.Create).Methods("POST")
    r.HandleFunc("/admins", handler.GetAll).Methods("GET")
    r.HandleFunc("/admins/{id}", handler.GetByID).Methods("GET")
    r.HandleFunc("/admins/{id}", handler.Update).Methods("PUT")
    r.HandleFunc("/admins/{id}", handler.Delete).Methods("DELETE")
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