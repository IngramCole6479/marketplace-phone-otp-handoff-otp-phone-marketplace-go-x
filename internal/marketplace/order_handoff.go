package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
)

type CaptchaVerifier interface {
	Verify(ctx context.Context, widgetRecordID, token, ip string) error
}

type SellerAsset struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	OwnerID string `json:"owner_id"`
}

type BuyerUpdate struct {
	OrderID string `json:"order_id"`
	BuyerID string `json:"buyer_id"`
	State   string `json:"state"`
}

type Order struct {
	ID      string
	BuyerID string
	Phone   string
	Asset   SellerAsset
	Paid    bool
}

type Handoff struct {
	OrderID    string      `json:"order_id"`
	Asset      SellerAsset `json:"seller_asset"`
	Buyer      BuyerUpdate `json:"buyer_update"`
	State      string      `json:"state"`
	AuthResult string      `json:"auth_result"`
}

type Service struct {
	captcha    CaptchaVerifier
	loginCode  string
	orders     map[string]Order
	mu         sync.Mutex
	challenges map[string]bool
}

func NewService(captcha CaptchaVerifier, loginCode string) *Service {
	return &Service{captcha: captcha, loginCode: loginCode, challenges: make(map[string]bool), orders: map[string]Order{
		"ord_1042": {
			ID: "ord_1042", BuyerID: "buyer_7", Phone: "+14155550123", Paid: true,
			Asset: SellerAsset{ID: "asset_camera_9", Title: "Rangefinder camera", OwnerID: "seller_3"},
		},
		"ord_1043": {
			ID: "ord_1043", BuyerID: "buyer_8", Phone: "+14155550124", Paid: false,
			Asset: SellerAsset{ID: "asset_lens_4", Title: "Prime lens", OwnerID: "seller_3"},
		},
	}}
}

type sendCodeInput struct {
	Phone          string `json:"phone"`
	CaptchaToken   string `json:"captcha_token"`
	WidgetRecordID string `json:"widget_record_id"`
}

type verifyInput struct {
	Phone   string `json:"phone"`
	Code    string `json:"code"`
	OrderID string `json:"order_id"`
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/code", s.sendCode)
	mux.HandleFunc("POST /login/verify", s.verifyAndHandoff)
	return mux
}

func (s *Service) sendCode(w http.ResponseWriter, r *http.Request) {
	var in sendCodeInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Phone == "" || in.CaptchaToken == "" || in.WidgetRecordID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone, captcha_token, and widget_record_id are required"})
		return
	}
	if err := s.captcha.Verify(r.Context(), in.WidgetRecordID, in.CaptchaToken, clientIP(r)); err != nil {
		writeAuthError(w, err)
		return
	}
	s.mu.Lock()
	s.challenges[in.Phone] = true
	s.mu.Unlock()
	writeJSON(w, http.StatusAccepted, map[string]string{"state": "code_sent"})
}

func (s *Service) verifyAndHandoff(w http.ResponseWriter, r *http.Request) {
	var in verifyInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Phone == "" || in.Code == "" || in.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone, code, and order_id are required"})
		return
	}
	order, exists := s.orders[in.OrderID]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}
	if order.Phone != in.Phone {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "phone does not match order buyer"})
		return
	}
	s.mu.Lock()
	challenged := s.challenges[in.Phone]
	if challenged && in.Code == s.loginCode {
		delete(s.challenges, in.Phone)
	}
	s.mu.Unlock()
	if !challenged || in.Code != s.loginCode {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired login code"})
		return
	}
	state := "payment_required"
	if order.Paid {
		state = "ready_for_handoff"
	}
	writeJSON(w, http.StatusOK, Handoff{
		OrderID:    in.OrderID,
		Asset:      order.Asset,
		Buyer:      BuyerUpdate{OrderID: in.OrderID, BuyerID: order.BuyerID, State: state},
		State:      state,
		AuthResult: "phone_verified",
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	if apiErr, ok := err.(*InfraiError); ok && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		status = apiErr.HTTPStatus
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
