package highheath

import (
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-gomail/gomail"
	"github.com/matcornic/hermes/v2"
	"google.golang.org/api/gmail/v1"
)

type Contact struct {
	Name    string
	Email   string
	Message string
}

func ToTable(obj interface{}) hermes.Table {
	var table hermes.Table
	value := reflect.ValueOf(obj)
	for i := 0; i < value.NumField(); i++ {
		table.Data = append(table.Data, []hermes.Entry{
			{Key: "Field", Value: value.Type().Field(i).Name},
			{Key: "Value", Value: fmt.Sprintf("%s", value.Field(i).Interface())},
		})
	}
	return table
}

func contactMessage(contact *Contact) *gmail.Message {
	h := hermes.Hermes{
		Theme: new(HighheathTheme),
		Product: hermes.Product{
			Name: "High Heath Farm Cattery",
			Link: "https://highheathcattery.co.uk/",
		},
	}
	email := hermes.Email{
		Body: hermes.Body{
			Name: contact.Name,
			Intros: []string{
				"Thank you for your message. We will get back to you soon.",
			},
			Table: ToTable(*contact),
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		log.Fatalf("Failed to generate html for message: %v", err)
	}

	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	emailText, err := h.GeneratePlainText(email)
	if err != nil {
		log.Fatalf("Failed to generate plaintext for message: %v", err)
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", "smirlie@googlemail.com", "Alex Williams")
	m.SetAddressHeader("To", "smirlie@googlemail.com", "Alex Williams")
	m.SetAddressHeader("Reply-To", "alex.williams@skyscanner.net", "Alex Williams")
	m.SetHeader("Subject", "This is a test hermes email")

	m.SetBody("text/plain", emailText)
	m.AddAlternative("text/html", emailBody)

	var raw strings.Builder
	w := base64.NewEncoder(base64.StdEncoding, &raw)
	if _, err := m.WriteTo(w); err != nil {
		log.Fatalf("Failed to get raw message: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Failed to write all message to base64 encoder: %v", err)
	}

	var message gmail.Message
	message.Raw = raw.String()
	return &message
}
