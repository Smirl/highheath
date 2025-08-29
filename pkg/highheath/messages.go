package highheath

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/matcornic/hermes/v2"
	"google.golang.org/api/gmail/v1"
)

var hermesConfig = hermes.Hermes{
	Theme: new(HighheathTheme),
	Product: hermes.Product{
		Name:      "Lyn at High Heath Farm Cattery",
		Link:      "https://highheathcattery.co.uk/",
		Copyright: fmt.Sprintf("Copyright © %s High Heath Farm Cattery. All rights reserved.", time.Now().Format("2006")),
		Logo:      "https://www.highheathcattery.co.uk/img/header_email.png",
	},
}

var commentTemplateBody = `---
author: {{ .Name }}
date: {{ .Date.Format "2006-01-02 15:04:05" }}
title: "Comment from {{ .Name }} | {{ .Date.Format "2006-01-02" }}"
---
{{ .Message }}

`
var commentTemplate = template.Must(template.New("comment").Parse(commentTemplateBody))

func ToDict(message interface{}) []hermes.Entry {
	var dict []hermes.Entry
	for _, row := range ToTable(message).Data {
		field, value := row[0], row[1]
		dict = append(dict, hermes.Entry{Key: field.Value, Value: value.Value})
	}
	return dict
}

func ToTable(message interface{}) hermes.Table {
	var table hermes.Table
	value := reflect.ValueOf(message)
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)

		if field.Kind() == reflect.Struct && fieldType.Anonymous {
			table.Data = append(table.Data, ToTable(field.Interface()).Data...)
		} else {
			fieldName, ok := fieldType.Tag.Lookup("name")
			if !ok {
				fieldName = fieldType.Name
			}
			if fieldType.Tag.Get("omit") != "" {
				continue
			}
			table.Data = append(table.Data, []hermes.Entry{
				{Key: "Field", Value: fieldName},
				{Key: "Value", Value: fmt.Sprintf("%v", field.Interface())},
			})
		}
	}
	return table
}

type EmailableMessage interface {
	GetName() string
	GetEmailAddress() string
	GetEmailCheck() string
	GetSubject() string
	GetEmail() *hermes.Email
	GetToken() string
}

type Contact struct {
	Name       string
	Email      string
	EmailCheck string `omit:"true"`
	Message    string
	Token      string `omit:"true"`
}

func (contact *Contact) GetName() string {
	return contact.Name
}

func (contact *Contact) GetEmailAddress() string {
	return contact.Email
}

func (contact *Contact) GetEmailCheck() string {
	return contact.EmailCheck
}

func (contact *Contact) GetSubject() string {
	return fmt.Sprintf("Thank you for your message, %s.", contact.Name)
}

func (contact *Contact) GetToken() string {
	return contact.Token
}

func (contact *Contact) GetEmail() *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: contact.Name,
			Intros: []string{
				"Thank you for your message. We will get back to you soon.",
			},
			Dictionary: ToDict(*contact),
			Signature:  "Yours",
		},
	}
}

// Ensure Contact implements EmailableMessage
var _ EmailableMessage = &Contact{}

type Comment struct {
	Contact
	Date time.Time
}

func (comment *Comment) GetEmail() *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: comment.Name,
			Intros: []string{
				"Thank you for your review!",
				"It will be visible on the website shortly.",
			},
			Dictionary: ToDict(*comment),
			Signature:  "Yours",
		},
	}
}

func (comment *Comment) GetFileContent() []byte {
	var b bytes.Buffer
	if err := commentTemplate.Execute(&b, comment); err != nil {
		log.Fatalf("Error getting comment file content: %v", err)
	}
	return b.Bytes()
}

// Ensure Comment implements EmailableMessage
var _ EmailableMessage = &Comment{}

type Booking struct {
	Pens            int    `name:"Number of pens"`
	CatsNames       string `name:"Cat Name(s)"`
	CatsAges        string `name:"Cat Ages(s)"`
	CatsSexs        string `name:"Cat Sex(s)"`
	CatsMC          string `name:"Cat Microchip Number(s)"`
	CatsColours     string `name:"Cat Colour(s)"`
	CatsFood        string `name:"Cat Preferred Food"`
	Insured         bool
	ArrivalDate     string `name:"Arrival Date"`
	TimeOfDayA      string `name:"Arrival Time"`
	DepartureDate   string `name:"Departure Date"`
	TimeOfDayD      string `name:"Arrival Time"`
	Name            string
	Address         string
	Postcode        string
	Email           string
	EmailCheck      string `omit:"true"`
	Number          string
	EmergencyName   string `name:"Emergency Contact Name"`
	EmergencyNumber string `name:"Emergency Contact Number"`
	VetName         string `name:"Vet Name"`
	VetNumber       string `name:"Vet Number"`
	KnowAllergies   string `name:"Known Allergies"`
	Meds            string `name:"Medication Details"`
	Relevant        string `name:"Other Medical Details"`
	VaccinationDate string `name:"Vaccination Date"`
	FleaDate        string `name:"Flea Treatment Date"`
	WormingDate     string `name:"Worming Date"`
	TC              bool   `name:"Terms & Conditions"`
	Token           string `omit:"true"`
}

func (booking *Booking) GetName() string {
	return booking.Name
}

func (booking *Booking) GetEmailAddress() string {
	return booking.Email
}

func (booking *Booking) GetEmailCheck() string {
	return booking.EmailCheck
}

func (booking *Booking) GetSubject() string {
	return "[High Heath Farm Cattery] Booking received"
}

func (booking *Booking) GetToken() string {
	return booking.Token
}

func (booking *Booking) GetEmail() *hermes.Email {
	return &hermes.Email{
		Body: hermes.Body{
			Name: booking.Name,
			Intros: []string{
				"Thank you for booking with us. Your message has been received and we will contact you to confirm your booking as soon as we can.",
				fmt.Sprintf("Please note this is not a confirmation of %v's stay.", booking.CatsNames),
			},
			Table:     ToTable(*booking),
			Signature: "Yours",
		},
	}
}

// Ensure Booking implements EmailableMessage
var _ EmailableMessage = &Booking{}

func ValidateForm(ctx context.Context, r Recaptcha, email EmailableMessage) error {
	if email.GetEmailCheck() != "" {
		return fmt.Errorf("Bot Suspected from email check")
	}
	if ok, err := r.VerifyToken(ctx, email.GetToken()); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("Bot Suspected from token")
	}
	return nil
}

func SendMessages(ctx context.Context, client *gmail.Service, email EmailableMessage) (err error) {
	company := "High Heath Farm Cattery"
	companyEmailAddress := "highheath@googlemail.com"
	if os.Getenv("TOKEN_FILE") != "token.json" {
		companyEmailAddress = "smirlie@googlemail.com"
	}

	name := email.GetName()
	emailAddress := email.GetEmailAddress()
	subject := email.GetSubject()
	hermesEmail := email.GetEmail()

	var message *gmail.Message
	message, err = createMessageFromEmail(company, companyEmailAddress, name, emailAddress, subject, hermesEmail)
	if err != nil {
		return err
	}
	_, err = client.Users.Messages.Send("me", message).Context(ctx).Do()
	if err != nil {
		return err
	}
	message, err = createMessageFromEmail(name, emailAddress, company, companyEmailAddress, subject, hermesEmail)
	if err != nil {
		return err
	}
	_, err = client.Users.Messages.Insert("me", message).Context(ctx).Do()
	if err != nil {
		return err
	}
	return nil
}

func HandleFormError(w http.ResponseWriter, r *http.Request, email EmailableMessage, e error, redirectURL string) {
	output, err := json.Marshal(email)
	if err != nil {
		log.Printf("Error in HandleSendMessagesError: %v", err)
	} else {
		log.Println(string(output))
	}
	log.Printf("Error sending %T message: %v", email, e)
	http.Redirect(w, r, redirectURL, http.StatusFound)
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
