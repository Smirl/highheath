package highheath

import (
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/gorilla/schema"
	"google.golang.org/api/gmail/v1"
)

var decoder *schema.Decoder
var gmailClient *gmail.Service
var githubClient *github.Client

func init() {
	decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	gmailClient = GmailClient()
	githubClient = GithubClient()
}

func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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

	if err := ValidateForm(&contact); err != nil {
		HandleFormError(w, r, &contact, err, "/contact-us/failure/")
		return
	}

	if err := SendMessages(gmailClient, &contact); err != nil {
		HandleFormError(w, r, &contact, err, "/contact-us/failure/")
		return
	}

	http.Redirect(w, r, "/contact-us/success/", http.StatusFound)
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

	if err := ValidateForm(&booking); err != nil {
		HandleFormError(w, r, &booking, err, "/book-now/failure/")
		return
	}

	if err := SendMessages(gmailClient, &booking); err != nil {
		HandleFormError(w, r, &booking, err, "/book-now/failure/")
		return
	}

	http.Redirect(w, r, "/book-now/success/", http.StatusFound)
}

func HandleCommentForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var comment Comment
	if err := decoder.Decode(&comment, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	comment.Date = time.Now()

	if err := ValidateForm(&comment); err != nil {
		HandleFormError(w, r, &comment, err, "/comments/failure/")
		return
	}

	if err := CreateComment(r.Context(), githubClient, &comment); err != nil {
		log.Printf("Error creating comment pull request: %v", err)
		HandleFormError(w, r, &comment, err, "/comments/failure/")
		return
	}

	if err := SendMessages(gmailClient, &comment); err != nil {
		HandleFormError(w, r, &comment, err, "/comments/failure/")
		return
	}

	http.Redirect(w, r, "/comments/success/", http.StatusFound)
}
