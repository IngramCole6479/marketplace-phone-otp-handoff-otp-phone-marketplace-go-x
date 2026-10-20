package main

import (
	"log"
	"net/http"
	"os"

	"example.com/marketplace-phone-otp/internal/marketplace"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	loginCode := os.Getenv("OTP_LOGIN_CODE")
	if loginCode == "" {
		log.Fatal("OTP_LOGIN_CODE is required")
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	service := marketplace.NewService(marketplace.NewInfraiCaptcha(apiKey), loginCode)
	server := &http.Server{Addr: addr, Handler: service.Handler()}
	log.Printf("marketplace OTP service listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
