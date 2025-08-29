package highheath

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/gorilla/handlers"
	"github.com/gorilla/schema"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/api/gmail/v1"
)

type AppContext struct {
	ServeMux     *http.ServeMux
	Decoder      *schema.Decoder
	GmailClient  *gmail.Service
	GithubClient *github.Client
	Recaptcha    Recaptcha
}

func (c *AppContext) HandleFunc(pattern string, handler func(c AppContext, w http.ResponseWriter, r *http.Request)) {
	// Create a regular http.HandlerFunc by passing in the AppContext.
	withoutAppContext := func(w http.ResponseWriter, r *http.Request) {
		handler(*c, w, r)
	}
	// Wrap the handler with OpenTelemetry instrumentation.
	handlerFunc := otelhttp.WithRouteTag(pattern, http.HandlerFunc(withoutAppContext))
	// Register the handler with the ServeMux, wrapping it with OpenTelemetry instrumentation.
	c.ServeMux.Handle(pattern, handlerFunc)
}

func (c *AppContext) WrapHandler() (handler http.Handler) {
	handler = handlers.CombinedLoggingHandler(os.Stdout, c.ServeMux)
	handler = handlers.ProxyHeaders(handler)
	handler = otelhttp.NewHandler(handler, "highheath")
	handler = handlers.RecoveryHandler()(handler)
	return handler
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

	if err := ValidateForm(r.Context(), c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/contact-us/failure/")
		return
	}

	if err := SendMessages(r.Context(), c.GmailClient, &form); err != nil {
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

	if err := ValidateForm(r.Context(), c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/book-now/failure/")
		return
	}

	if err := SendMessages(r.Context(), c.GmailClient, &form); err != nil {
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

	if err := ValidateForm(r.Context(), c.Recaptcha, &form); err != nil {
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	if err := CreateComment(r.Context(), c.GithubClient, &form); err != nil {
		log.Printf("Error creating comment pull request: %v", err)
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	if err := SendMessages(r.Context(), c.GmailClient, &form); err != nil {
		HandleFormError(w, r, &form, err, "/comments/failure/")
		return
	}

	http.Redirect(w, r, "/comments/success/", http.StatusFound)
}
