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

var hermesConfig = hermes.Hermes{
	Theme: new(HighheathTheme),
	Product: hermes.Product{
		Name: "High Heath Farm Cattery",
		Link: "https://highheathcattery.co.uk/",
		Logo: "http://localhost:8080/img/header_email.png",
	},
}

func ToDict(message interface{}) []hermes.Entry {
	log.Println(message)
	var dict []hermes.Entry
	value := reflect.ValueOf(message)
	log.Println(value.NumField())
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

func CreateUserMessage(name, emailAddress, subject string, email *hermes.Email) (*gmail.Message, error) {
	return createMessageFromEmail("High Heath Farm Cattery", "smirlie@googlemail.com", name, emailAddress, subject, email)
}

func CreateAdminMessage(name, emailAddress, subject string, email *hermes.Email) (*gmail.Message, error) {
	return createMessageFromEmail(name, emailAddress, "High Heath Farm Cattery", "highheath@googlemail.com", subject, email)
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
	m.SetAddressHeader("Sender", "smirlie@googlemail.com", "Alex Williams")
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
	return &message, nil
}
