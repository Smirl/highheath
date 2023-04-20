package highheath

import (
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/gorilla/schema"
	"google.golang.org/api/gmail/v1"
)

type AppContext struct {
	Decoder      *schema.Decoder
	GmailClient  *gmail.Service
	GithubClient *github.Client
	Recaptcha    Recaptcha
}

type Handler struct {
	AppContext AppContext
	H          func(c AppContext, w http.ResponseWriter, r *http.Request)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.H(h.AppContext, w, r)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HandleContactForm(c AppContext, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var form Contact
	if err := c.Decoder.Decode(&form, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := ValidateForm(c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/contact-us/failure/")
		return
	}

	if err := SendMessages(c.GmailClient, &form); err != nil {
		HandleFormError(w, r, &form, err, "/contact-us/failure/")
		return
	}

	http.Redirect(w, r, "/contact-us/success/", http.StatusFound)
}

func HandleBookingForm(c AppContext, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var form Booking
	if err := c.Decoder.Decode(&form, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := ValidateForm(c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/book-now/failure/")
		return
	}

	if err := SendMessages(c.GmailClient, &form); err != nil {
		HandleFormError(w, r, &form, err, "/book-now/failure/")
		return
	}

	http.Redirect(w, r, "/book-now/success/", http.StatusFound)
}

func HandleCommentForm(c AppContext, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Unable to parse form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var form Comment
	if err := c.Decoder.Decode(&form, r.Form); err != nil {
		log.Printf("Unable to decode form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	form.Date = time.Now()

	if err := ValidateForm(c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	if err := CreateComment(r.Context(), c.GithubClient, &form); err != nil {
		log.Printf("Error creating comment pull request: %v", err)
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	if err := SendMessages(c.GmailClient, &form); err != nil {
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	http.Redirect(w, r, "/comments/success/", http.StatusFound)
}
