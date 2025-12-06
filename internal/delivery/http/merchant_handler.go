package http

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
    "be_customer/internal/domain"
    "be_customer/internal/usecase"
)

type MerchantHandler struct {
    usecase *usecase.MerchantUsecase
}

func NewMerchantHandler(r *mux.Router, u *usecase.MerchantUsecase) {
    handler := &MerchantHandler{usecase: u}

    r.HandleFunc("/merchants", handler.Create).Methods("POST")
    r.HandleFunc("/merchants", handler.GetAll).Methods("GET")
    r.HandleFunc("/merchants/{id}", handler.GetByID).Methods("GET")
    r.HandleFunc("/merchants/{id}", handler.Update).Methods("PUT")
    r.HandleFunc("/merchants/{id}", handler.Delete).Methods("DELETE")
}

func (h *MerchantHandler) Create(w http.ResponseWriter, r *http.Request) {
    var merchant domain.Merchant
    json.NewDecoder(r.Body).Decode(&merchant)
    if err := h.usecase.Create(&merchant); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    json.NewEncoder(w).Encode(merchant)
}

func (h *MerchantHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    merchants, _ := h.usecase.GetAll()
    json.NewEncoder(w).Encode(merchants)
}

func (h *MerchantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
	merchant, err := h.usecase.GetByID(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(merchant)
}

func (h *MerchantHandler) Update(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    var merchant domain.Merchant
    json.NewDecoder(r.Body).Decode(&merchant)
    merchant.C_ID = id
    if err := h.usecase.Update(&merchant); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(merchant)
}

func (h *MerchantHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    if err := h.usecase.Delete(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}