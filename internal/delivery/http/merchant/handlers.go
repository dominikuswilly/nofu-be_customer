package merchant

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

type MerchantHandler struct {
	usecase *usecase.MerchantUsecase
}

func (h *MerchantHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	merchant := domain.Merchant{
		C_NM:       req.Name,
		C_PHONE:    &req.Phone,
		C_EMAIL:    &req.Email,
		C_USERNAME: req.Username,
		C_PASSWORD: string(hashed),
		I_ACTIVE:   1, // default to active on creation
	}

	if err := h.usecase.Create(&merchant); err != nil {
		// propagate usecase error message (consider mapping to safe messages)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	respData := itemDetailMerchant{
		ID:       fmt.Sprintf("%v", merchant.C_ID),
		Name:     merchant.C_NM,
		Phone:    derefString(merchant.C_PHONE),
		Email:    derefString(merchant.C_EMAIL),
		Username: merchant.C_USERNAME,
		Active:   merchant.I_ACTIVE == 1,
	}

	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            respData,
	})
}

func (h *MerchantHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	merchants, err := h.usecase.GetAll()
	if err != nil {
		log.Printf("GetAll usecase error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch merchants")
		return
	}

	resp := make([]itemMerchant, 0, len(merchants))
	for _, m := range merchants {
		resp = append(resp, itemMerchant{
			ID:       fmt.Sprintf("%v", m.C_ID),
			Name:     m.C_NM,
			Phone:    derefString(m.C_PHONE),
			Email:    derefString(m.C_EMAIL),
			Username: m.C_USERNAME,
			Active:   m.I_ACTIVE == 1,
		})
	}

	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            resp,
	})
}

func (h *MerchantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	merchant, err := h.usecase.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "merchant not found")
		return
	}

	item := itemDetailMerchant{
		ID:       fmt.Sprintf("%v", merchant.C_ID),
		Name:     merchant.C_NM,
		Phone:    derefString(merchant.C_PHONE),
		Email:    derefString(merchant.C_EMAIL),
		Username: merchant.C_USERNAME,
		Active:   merchant.I_ACTIVE == 1,
	}

	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            item,
	})
}

func (h *MerchantHandler) Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]

	var reqBody struct {
		domain.Merchant
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	merchant := reqBody.Merchant
	merchant.C_ID = id
	if reqBody.Active {
		merchant.I_ACTIVE = 1
	} else {
		merchant.I_ACTIVE = 0
	}

	if err := h.usecase.Update(&merchant); err != nil {
		// map the domain/usecase error appropriately
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Update the domain object for the response to reflect the boolean mapping if needed,
	// though we are returning the merchant struct directly here.
	// Actually, the merchant struct has I_ACTIVE now.
	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            merchant,
	})
}

func (h *MerchantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.usecase.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MerchantHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	merchant, err := h.usecase.FindByUsername(req.Username)
	if err != nil || merchant == nil {
		// avoid revealing whether the username exists
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if merchant.C_PASSWORD == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// decode hex SHA-512 (expected 64 bytes)
	pwBytes, err := hex.DecodeString(req.Password)
	if err != nil || len(pwBytes) != 64 {
		writeError(w, http.StatusBadRequest, "invalid password format")
		return
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(merchant.C_PASSWORD), pwBytes); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Create JWT using github.com/golang-jwt/jwt/v5
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("missing JWT_SECRET")
		writeError(w, http.StatusInternalServerError, "server misconfiguration")
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	// Get issuer from environment or use default
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "nofu-customer-api"
	}

	claims := jwt.MapClaims{
		"iss":      issuer,
		"sub":      merchant.C_ID,
		"username": merchant.C_USERNAME,
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
	userData.User.Name = merchant.C_NM
	userData.User.Email = derefString(merchant.C_EMAIL)
	userData.User.Username = merchant.C_USERNAME

	writeJSON(w, http.StatusOK, apiResp{
		ResponseCode:    "200",
		ResponseMessage: "success",
		Data:            userData,
	})
}
