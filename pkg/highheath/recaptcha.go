package highheath

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

var RECAPTURE_URL = "https://www.google.com/recaptcha/api/siteverify"
var RECAPTURE_SECRET string

func init() {
	if secret, ok := os.LookupEnv("RECAPTURE_SECRET"); ok {
		RECAPTURE_SECRET = secret
	} else {
		log.Fatal("Unable to read RECAPTURE_SECRET")
	}
}

func VerifyToken(token string) (success bool, err error) {
	values := url.Values{}
	values.Set("secret", RECAPTURE_SECRET)
	values.Set("response", token)
	resp, err := http.PostForm(RECAPTURE_URL, values)
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
