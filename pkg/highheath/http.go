package highheath

import (
	"log"
	"net/http"

	"github.com/gorilla/schema"
)

var decoder *schema.Decoder

func init() {
	decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
}

func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func HandleContactForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var contact Contact
	if err := decoder.Decode(&contact, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message, err := CreateUserMessage(contact.Name, contact.Email, "Thank you for your message", contact.GetEmail())
	if err != nil {
		log.Printf("Error creating message from inputs: %v", err)
	}
	SendMessage(message)
	http.Redirect(w, r, "/contact-us/", http.StatusFound)
}
