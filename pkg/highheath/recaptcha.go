package highheath

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

var RECAPTCHA_URL = "https://www.google.com/recaptcha/api/siteverify"

type Recaptcha struct {
	secret string
}

func (r *Recaptcha) VerifyToken(token string) (success bool, err error) {
	values := url.Values{}
	values.Set("secret", r.secret)
	values.Set("response", token)
	resp, err := http.PostForm(RECAPTCHA_URL, values)
	if err != nil {
		return false, err
	}
	var response map[string]interface{}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}
	log.Println(response)

	return response["success"].(bool), nil
}

func NewRecaptcha() *Recaptcha {
	if secret, ok := os.LookupEnv("RECAPTCHA_SECRET"); ok {
		log.Printf("Recaptcha: %s", secret)
		return &Recaptcha{secret: secret}
	}
	log.Fatal("Unable to read RECAPTCHA_SECRET")
	return nil
}
