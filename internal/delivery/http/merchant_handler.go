package http

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"be_customer/internal/domain"
	"be_customer/internal/usecase"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type MerchantHandler struct {
	usecase *usecase.MerchantUsecase
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewMerchantHandler(r *mux.Router, u *usecase.MerchantUsecase) {
	handler := &MerchantHandler{usecase: u}

	// Create a subrouter with prefix "/api"
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/merchants", handler.Create).Methods("POST")
	api.HandleFunc("/merchants", handler.GetAll).Methods("GET")
	api.HandleFunc("/merchants/{id}", handler.GetByID).Methods("GET")
	api.HandleFunc("/merchants/{id}", handler.Update).Methods("PUT")
	api.HandleFunc("/merchants/{id}", handler.Delete).Methods("DELETE")
	api.HandleFunc("/merchants/login", handler.Login).Methods("POST")
}

func (h *MerchantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Password string `json:"password"` // hex SHA-512 from frontend
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	pwBytes, err := hex.DecodeString(req.Password)
	if err != nil || len(pwBytes) != 64 {
		http.Error(w, "invalid password format", http.StatusBadRequest)
		return
	}

	log.Println(pwBytes)
	log.Println(bcrypt.DefaultCost)
	hashed, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	merchant := domain.Merchant{
		C_NM:       req.Name,
		C_PHONE:    &req.Phone,
		C_EMAIL:    &req.Email,
		C_USERNAME: req.Username,
		C_PASSWORD: string(hashed), // store bcrypt hash
	}

	if err := h.usecase.Create(&merchant); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build the response DTO (do NOT include password)
	resp := struct {
		ResponseCode    string      `json:"responseCode"`
		ResponseMessage string      `json:"responseMessage"`
		Data            interface{} `json:"data"`
	}{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data: struct {
			Name     string `json:"name"`
			Phone    string `json:"phone"`
			Email    string `json:"email"`
			Username string `json:"username"`
		}{
			Name:     merchant.C_NM,
			Phone:    *merchant.C_PHONE,
			Email:    *merchant.C_EMAIL,
			Username: merchant.C_USERNAME,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
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

func (h *MerchantHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	merchant, err := h.usecase.FindByUsername(req.Username)
	log.Println(merchant)
	if err != nil || merchant == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if merchant.C_PASSWORD == "" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Quick debug helpers (remove in prod)
	log.Printf("stored pwd prefix: %.4s, len stored: %d", merchant.C_PASSWORD, len(merchant.C_PASSWORD))
	// A valid bcrypt hash usually starts with "$2"

	storedHash := []byte(merchant.C_PASSWORD)

	pwBytes, err := hex.DecodeString(req.Password)
	if err != nil || len(pwBytes) != 64 {
		http.Error(w, "invalid password format", http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword(storedHash, pwBytes); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := struct {
		ResponseCode    string      `json:"responseCode"`
		ResponseMessage string      `json:"responseMessage"`
		Data            interface{} `json:"data"`
	}{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data: struct {
			Name     string  `json:"name"`
			Phone    *string `json:"phone"`
			Email    *string `json:"email"`
			Username string  `json:"username"`
		}{
			Name:     merchant.C_NM,
			Phone:    merchant.C_PHONE,
			Email:    merchant.C_EMAIL,
			Username: merchant.C_USERNAME,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
