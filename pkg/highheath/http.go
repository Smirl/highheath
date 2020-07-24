package highheath

import (
	"log"
	"net/http"

	"github.com/gorilla/schema"
	"google.golang.org/api/gmail/v1"
)

var decoder *schema.Decoder
var gmailClient *gmail.Service

func init() {
	decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	gmailClient = GmailClient()
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

	if err := SendMessages(gmailClient, &contact); err != nil {
		log.Printf("Error creating message from inputs: %v", err)
	}

	http.Redirect(w, r, "/contact-us/", http.StatusFound)
}

func HandleBookingForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var booking Booking
	if err := decoder.Decode(&booking, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := SendMessages(gmailClient, &booking); err != nil {
		log.Printf("Error creating message from inputs: %v", err)
	}

	http.Redirect(w, r, "/book-now/", http.StatusFound)
}
