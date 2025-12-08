package http

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"be_customer/internal/domain"
	"be_customer/internal/usecase"

	jwt "github.com/golang-jwt/jwt/v5"
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

type loginResponseData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
}

type resGetMerchant struct {
	ResponseCode    string      `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	Data            interface{} `json:"data"`
}

type itemMerchant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type itemDetailMerchant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Username string `json:"username"`
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
	resp := make([]itemMerchant, 0, len(merchants))
	for _, m := range merchants {
		var phone, email string
		if m.C_PHONE != nil {
			phone = *m.C_PHONE
		}
		if m.C_EMAIL != nil {
			email = *m.C_EMAIL
		}
		resp = append(resp, itemMerchant{
			ID:       m.C_ID,
			Name:     m.C_NM,
			Phone:    phone,
			Email:    email,
			Username: m.C_USERNAME,
		})
	}

	out := resGetMerchant{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            resp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *MerchantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	merchant, err := h.usecase.GetByID(id)
	if err != nil {
		out := resGetMerchant{
			ResponseCode:    "404",
			ResponseMessage: "merchant not found",
			Data:            nil,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	// Safely extract optional fields
	var phone, email string
	if merchant.C_PHONE != nil {
		phone = *merchant.C_PHONE
	}
	if merchant.C_EMAIL != nil {
		email = *merchant.C_EMAIL
	}

	// Convert ID to string (works even if C_ID is numeric)
	item := itemDetailMerchant{
		ID:       fmt.Sprintf("%v", merchant.C_ID),
		Name:     merchant.C_NM,
		Phone:    phone,
		Email:    email,
		Username: merchant.C_USERNAME,
	}

	out := resGetMerchant{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            item,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
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
	w.Header().Set("Content-Type", "application/json")
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

	// Generate JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		http.Error(w, "server misconfiguration", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	claims := jwt.MapClaims{
		"sub":      merchant.C_ID,
		"username": merchant.C_USERNAME,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	// Build response (do NOT include password)
	var data loginResponseData
	data.Token = signedToken
	data.ExpiresAt = expiresAt
	data.User.Name = merchant.C_NM
	data.User.Email = *merchant.C_EMAIL
	data.User.Username = merchant.C_USERNAME

	resp := struct {
		ResponseCode    string            `json:"responseCode"`
		ResponseMessage string            `json:"responseMessage"`
		Data            loginResponseData `json:"data"`
	}{
		ResponseCode:    "200",
		ResponseMessage: "login success",
		Data:            data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
