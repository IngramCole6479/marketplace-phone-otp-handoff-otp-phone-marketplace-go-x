package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubPhoneAuth struct{}

func (s stubPhoneAuth) Verify(context.Context, string, string, string) error { return nil }

func TestVerifiedBuyerOrderDecision(t *testing.T) {
	tests := []struct {
		name       string
		orderID    string
		phone      string
		wantStatus int
		wantState  string
	}{
		{name: "paid order is handed off", orderID: "ord_1042", phone: "+14155550123", wantStatus: http.StatusOK, wantState: "ready_for_handoff"},
		{name: "unpaid order stays blocked", orderID: "ord_1043", phone: "+14155550124", wantStatus: http.StatusOK, wantState: "payment_required"},
		{name: "different buyer is rejected", orderID: "ord_1042", phone: "+14155550199", wantStatus: http.StatusForbidden},
		{name: "unknown order is hidden", orderID: "ord_missing", phone: "+14155550123", wantStatus: http.StatusNotFound},
	}

	service := NewService(stubPhoneAuth{}, "123456")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge := httptest.NewRequest(http.MethodPost, "/login/code", strings.NewReader(`{"phone":"`+tt.phone+`","captcha_token":"valid-token","widget_record_id":"widget-test"}`))
			service.Handler().ServeHTTP(httptest.NewRecorder(), challenge)
			body := `{"phone":"` + tt.phone + `","code":"123456","order_id":"` + tt.orderID + `"}`
			req := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(body))
			res := httptest.NewRecorder()
			service.Handler().ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
			if tt.wantState != "" && !strings.Contains(res.Body.String(), `"state":"`+tt.wantState+`"`) {
				t.Fatalf("body %s does not contain state %q", res.Body.String(), tt.wantState)
			}
		})
	}
}

func TestInfraiRequestBoundary(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["widget_record_id"] != "widget-test" || body["token"] != "bad-token" {
			t.Errorf("captcha request body = %v, decode error = %v", body, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"challenge_declined","message":"captcha declined"},"metadata":{}}`))
	}))
	defer server.Close()

	client := NewInfraiCaptcha("test-key")
	client.baseURL = server.URL
	err := client.Verify(context.Background(), "widget-test", "bad-token", "203.0.113.9")
	apiErr, ok := err.(*InfraiError)
	if !ok || apiErr.Code != "challenge_declined" || apiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("error = %#v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/captcha/verify" || gotAuth != "Bearer test-key" {
		t.Fatalf("boundary = %s %s %s", gotMethod, gotPath, gotAuth)
	}
}
