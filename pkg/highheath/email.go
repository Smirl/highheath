package highheath

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/matcornic/hermes/v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Overrides some CSS from the hermes.Default theme
type HighheathTheme struct {
	hermes.Default
}

func (e *HighheathTheme) HTMLTemplate() string {
	return strings.ReplaceAll(e.Default.HTMLTemplate(), "#F2F4F6;", "#FCFBE6;")
}

// Retrieve a token, saves the token, then returns the generated client.
func getToken() (*oauth2.Config, *oauth2.Token) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailInsertScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	if tf, ok := os.LookupEnv("TOKEN_FILE"); ok {
		tokFile = tf
	}
	token, err := tokenFromFile(tokFile)
	if err != nil {
		log.Printf("Cannot find token in %s", tokFile)
		token = getTokenFromWeb(config)
		saveToken(tokFile, token)
	} else {
		log.Printf("Using token found in %s", tokFile)
	}
	return config, token
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Make a channel to wait for the code to be sent
	ch := make(chan string)
	// A random state is needed for security
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	// Start a local server to listen for the code after redirect
	// Used to be able to use special redirect UI that just printed the token
	// this is now removed by google
	// See https://github.com/googleapis/google-api-go-client/blob/main/examples/main.go
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	// Wait for code
	var authCode string
	select {
	case authCode = <-ch:
	case <-time.After(5 * time.Minute):
		log.Fatalf("Timeout waiting for code")
	}

	// Swap code for a token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Fatalf("Unable write token to %s: %v", path, err)
	}
}

func GmailClient(ctx context.Context) *gmail.Service {
	config, token := getToken()
	srv, err := gmail.NewService(
		ctx,
		option.WithTokenSource(config.TokenSource(ctx, token)),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	return srv
}
