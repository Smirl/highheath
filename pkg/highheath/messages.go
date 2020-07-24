package highheath

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-gomail/gomail"
	"github.com/matcornic/hermes/v2"
	"google.golang.org/api/gmail/v1"
)

var hermesConfig = hermes.Hermes{
	Theme: new(HighheathTheme),
	Product: hermes.Product{
		Name: "High Heath Farm Cattery",
		Link: "https://highheathcattery.co.uk/",
		Logo: "http://localhost:8080/img/header_email.png",
	},
}

func ToDict(message interface{}) []hermes.Entry {
	var dict []hermes.Entry
	value := reflect.ValueOf(message)
	for i := 0; i < value.NumField(); i++ {
		dict = append(dict, hermes.Entry{
			Key:   value.Type().Field(i).Name,
			Value: fmt.Sprintf("%s", value.Field(i).Interface()),
		})
	}
	return dict
}

func ToTable(message interface{}) hermes.Table {
	var table hermes.Table
	value := reflect.ValueOf(message)
	for i := 0; i < value.NumField(); i++ {
		table.Data = append(table.Data, []hermes.Entry{
			{Key: "Field", Value: value.Type().Field(i).Name},
			{Key: "Value", Value: fmt.Sprintf("%s", value.Field(i).Interface())},
		})
	}
	return table
}

type Contact struct {
	Name    string
	Email   string
	Message string
}

func (contact *Contact) GetEmail() *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: contact.Name,
			Intros: []string{
				"Thank you for your message. We will get back to you soon.",
			},
			Dictionary: ToDict(*contact),
		},
	}
}

type Booking struct {
	CatsNames       string
	CatsAges        int
	CatsSexs        string
	CatsMC          string
	CatsColours     string
	CatsFood        string
	ArrivalDate     string
	TimeOfDayA      string
	DepartureDate   string
	TimeOfDayD      string
	Name            string
	Address         string
	Postcode        string
	Email           string
	Number          string
	EmergencyName   string
	EmergencyNumber string
	VetName         string
	VetNumber       string
	KnowAllergies   string
	Meds            string
	Relevant        string
	VaccinationDate string
	Sharing         bool
	TC              bool
}

func (booking *Booking) GetEmail() *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: booking.Name,
			Intros: []string{
				"Thank you for your message. We will get back to you soon.",
			},
			Table: ToTable(*booking),
		},
	}
}

func SendUserMessage(client *gmail.Service, name, emailAddress, subject string, email *hermes.Email) error {
	message, err := createMessageFromEmail("High Heath Farm Cattery", "smirlie@googlemail.com", name, emailAddress, subject, email)
	if err != nil {
		return err
	}
	if _, err := gmailClient.Users.Messages.Send("me", message).Do(); err != nil {
		return err
	}
	return nil
}

func SendAdminMessage(client *gmail.Service, name, emailAddress, subject string, email *hermes.Email) error {
	message, err := createMessageFromEmail(name, emailAddress, "Alex Williams", "smirlie@googlemail.com", subject, email)
	if err != nil {
		return err
	}
	if _, err := gmailClient.Users.Messages.Insert("me", message).Do(); err != nil {
		return err
	}
	return nil
}

func createMessageFromEmail(from, fromEmail, to, toEmail, subject string, email *hermes.Email) (*gmail.Message, error) {
	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := hermesConfig.GenerateHTML(*email)
	if err != nil {
		return nil, err
	}
	// Generate the plaintext version of the e-mail (for clients that do not support xHTML)
	emailText, err := hermesConfig.GeneratePlainText(*email)
	if err != nil {
		return nil, err
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", fromEmail, from)
	m.SetAddressHeader("Reply-To", fromEmail, from)
	m.SetAddressHeader("To", toEmail, to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", emailText)
	m.AddAlternative("text/html", emailBody)

	var raw strings.Builder
	w := base64.NewEncoder(base64.StdEncoding, &raw)
	if _, err := m.WriteTo(w); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	var message gmail.Message
	message.Raw = raw.String()
	message.LabelIds = []string{"INBOX", "UNREAD"}
	return &message, nil
}
