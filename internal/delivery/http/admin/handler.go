package admin

import (
	"be_customer/internal/domain"
	"be_customer/internal/usecase"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	usecase *usecase.AdminUsecase
}

func NewAdminHandler(r *mux.Router, u *usecase.AdminUsecase) {
	handler := &AdminHandler{usecase: u}

	// Create a subrouter with prefix "/api"
	api := r.PathPrefix("/api/customer").Subrouter()

	// Frontend routes
	frontend := api.PathPrefix("/frontend").Subrouter()

	// Legacy routes
	api.HandleFunc("/admins/login", handler.Login).Methods(http.MethodPost, http.MethodOptions)
	api.HandleFunc("/admins", handler.Create).Methods("POST")
	api.HandleFunc("/admins", handler.GetAll).Methods("GET")
	api.HandleFunc("/admins/{id}", handler.GetByID).Methods("GET")
	api.HandleFunc("/admins/{id}", handler.Update).Methods("PUT")
	api.HandleFunc("/admins/{id}", handler.Delete).Methods("DELETE")

	// Frontend routes
	frontend.HandleFunc("/admins/login", handler.Login).Methods(http.MethodPost, http.MethodOptions)
	frontend.HandleFunc("/admins", handler.Create).Methods("POST")
	frontend.HandleFunc("/admins", handler.GetAll).Methods("GET")
	frontend.HandleFunc("/admins/{id}", handler.GetByID).Methods("GET")
	frontend.HandleFunc("/admins/{id}", handler.Update).Methods("PUT")
	frontend.HandleFunc("/admins/{id}", handler.Delete).Methods("DELETE")
}

func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Password string `json:"password"` // hex SHA-512 from frontend
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// Basic validation
	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.TrimSpace(req.Username)
	if req.Name == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "name, username and password are required")
		return
	}

	// decode hex SHA-512 (expected length 64 bytes)
	pwBytes, err := hex.DecodeString(req.Password)
	if err != nil || len(pwBytes) != 64 {
		writeError(w, http.StatusBadRequest, "invalid password format")
		return
	}

	// Hash with bcrypt (store the bcrypt hash)
	hashed, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt generate error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	admin := domain.Admin{
		C_NM:       req.Name,
		C_PHONE:    &req.Phone,
		C_EMAIL:    &req.Email,
		C_USERNAME: req.Username,
		C_PASSWORD: string(hashed),
	}

	if err := h.usecase.Create(&admin); err != nil {
		// propagate usecase error message (consider mapping to safe messages)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	respData := itemDetailAdmin{
		ID:       fmt.Sprintf("%v", admin.C_ID),
		Name:     admin.C_NM,
		Phone:    derefString(admin.C_PHONE),
		Email:    derefString(admin.C_EMAIL),
		Username: admin.C_USERNAME,
	}
	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            respData,
	})
}

func (h *AdminHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	admins, _ := h.usecase.GetAll()
	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            admins,
	})
}

func (h *AdminHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	admin, err := h.usecase.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            admin,
	})
}

func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var admin domain.Admin
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	admin.C_ID = id
	if err := h.usecase.Update(&admin); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            admin,
	})
}

func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.usecase.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	defer r.Body.Close()
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	admin, err := h.usecase.FindByUsername(req.Username)
	if err != nil || admin == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if admin.C_PASSWORD == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	pwBytes, err := hex.DecodeString(req.Password)
	if err != nil || len(pwBytes) != 64 {
		writeError(w, http.StatusBadRequest, "invalid password format")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.C_PASSWORD), pwBytes); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("missing JWT_SECRET")
		writeError(w, http.StatusInternalServerError, "server misconfiguration")
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "nofu-customer-api"
	}

	claims := jwt.MapClaims{
		"iss":      issuer,
		"sub":      admin.C_ID,
		"username": admin.C_USERNAME,
		"role":     "admin",
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("jwt sign error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	var userData loginResponseData
	userData.Token = signedToken
	userData.ExpiresAt = expiresAt
	userData.User.Name = admin.C_NM
	userData.User.Email = derefString(admin.C_EMAIL)
	userData.User.Username = admin.C_USERNAME

	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            userData,
	})
}
