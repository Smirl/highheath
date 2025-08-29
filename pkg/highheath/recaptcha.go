package highheath

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var RECAPTCHA_URL = "https://www.google.com/recaptcha/api/siteverify"

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

type Recaptcha interface {
	VerifyToken(ctx context.Context, token string) (succes bool, err error)
}

type recaptcha struct {
	secret string
}

func (r *recaptcha) VerifyToken(ctx context.Context, token string) (success bool, err error) {
	// Create request
	values := url.Values{}
	values.Set("secret", r.secret)
	values.Set("response", token)
	req, err := http.NewRequest("POST", RECAPTCHA_URL, strings.NewReader(values.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request with context
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	// Parse json response
	var response RecaptchaResponse
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}
	log.Println(response)

	return response.Success, nil
}

func NewRecaptcha() Recaptcha {
	if secret, ok := os.LookupEnv("RECAPTCHA_SECRET"); ok {
		log.Printf("Recaptcha: %s", secret)
		return &recaptcha{secret: secret}
	}
	log.Fatal("Unable to read RECAPTCHA_SECRET")
	return nil
}
