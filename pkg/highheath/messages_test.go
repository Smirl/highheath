package highheath

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/matcornic/hermes/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var contact = Contact{
	Name:    "Alex Williams",
	Email:   "foo@example.com",
	Message: "this is a great contact form",
}

func TestContactGetName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(contact.GetName()).To(Equal("Alex Williams"))
}

func TestContactGetEmailAddress(t *testing.T) {
	g := NewWithT(t)
	g.Expect(contact.GetEmailAddress()).To(Equal("foo@example.com"))
}

func TestContactGetEmailCheck(t *testing.T) {
	g := NewWithT(t)
	g.Expect(contact.GetEmailCheck()).To(Equal(""))
}

func TestContactGetSubject(t *testing.T) {
	g := NewWithT(t)
	g.Expect(contact.GetSubject()).To(Equal("Thank you for your message, Alex Williams."))
}

func TestContactGetToken(t *testing.T) {
	g := NewWithT(t)
	g.Expect(contact.GetToken()).To(Equal(""))
}

func TestContactGetEmail(t *testing.T) {
	g := NewWithT(t)
	email := contact.GetEmail()
	g.Expect(email.Body.Name).To(Equal("Alex Williams"))
	g.Expect(email.Body.Intros).To(ConsistOf(
		"Thank you for your message. We will get back to you soon.",
	))
	g.Expect(email.Body.Dictionary).To(BeEquivalentTo(ToDict(contact)))
	g.Expect(email.Body.Table).To(BeEquivalentTo(hermes.Table{}))
	g.Expect(email.Body.Signature).To(Equal("Yours"))
}

var expectedFileContent = `---
author: Alex Williams
date: 2021-01-01 00:00:00
title: "Comment from Alex Williams | 2021-01-01"
---
this is a great comment

`

var comment = Comment{
	Contact: Contact{
		Name:    "Alex Williams",
		Email:   "foo@example.com",
		Message: "this is a great comment",
	},
	Date: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
}

func TestCommentGetName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(comment.GetName()).To(Equal("Alex Williams"))
}

func TestCommentGetEmailAddress(t *testing.T) {
	g := NewWithT(t)
	g.Expect(comment.GetEmailAddress()).To(Equal("foo@example.com"))
}

func TestCommentGetEmailCheck(t *testing.T) {
	g := NewWithT(t)
	g.Expect(comment.GetEmailCheck()).To(Equal(""))
}

func TestCommentGetSubject(t *testing.T) {
	g := NewWithT(t)
	g.Expect(comment.GetSubject()).To(Equal("Thank you for your message, Alex Williams."))
}

func TestCommentGetToken(t *testing.T) {
	g := NewWithT(t)
	g.Expect(comment.GetToken()).To(Equal(""))
}

func TestCommentGetEmail(t *testing.T) {
	g := NewWithT(t)
	email := comment.GetEmail()
	g.Expect(email.Body.Name).To(Equal("Alex Williams"))
	g.Expect(email.Body.Intros).To(ConsistOf(
		"Thank you for your review!",
		"It will be visible on the website shortly.",
	))
	g.Expect(email.Body.Dictionary).To(BeEquivalentTo(ToDict(comment)))
	g.Expect(email.Body.Table).To(BeEquivalentTo(hermes.Table{}))
	g.Expect(email.Body.Signature).To(Equal("Yours"))
}

func TestCommentGetFileContent(t *testing.T) {
	g := NewWithT(t)
	actual := comment.GetFileContent()
	g.Expect(string(actual)).To(Equal(expectedFileContent))
}

var booking = Booking{
	Name:      "Alex Williams",
	Email:     "foo@example.com",
	CatsNames: "Bob",
}

func TestBookingGetName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(booking.GetName()).To(Equal("Alex Williams"))
}

func TestBookingGetEmailAddress(t *testing.T) {
	g := NewWithT(t)
	g.Expect(booking.GetEmailAddress()).To(Equal("foo@example.com"))
}

func TestBookingGetEmailCheck(t *testing.T) {
	g := NewWithT(t)
	g.Expect(booking.GetEmailCheck()).To(Equal(""))
}

func TestBookingGetSubject(t *testing.T) {
	g := NewWithT(t)
	g.Expect(booking.GetSubject()).To(Equal("[High Heath Farm Cattery] Booking received"))
}

func TestBookingGetToken(t *testing.T) {
	g := NewWithT(t)
	g.Expect(booking.GetToken()).To(Equal(""))
}

func TestBookingGetEmail(t *testing.T) {
	g := NewWithT(t)
	email := booking.GetEmail()
	g.Expect(email.Body.Name).To(Equal("Alex Williams"))
	g.Expect(email.Body.Intros).To(ConsistOf(
		"Thank you for booking with us. Your message has been received and we will contact you to confirm your booking as soon as we can.",
		"Please note this is not a confirmation of Bob's stay.",
	))
	g.Expect(email.Body.Dictionary).To(BeNil())
	g.Expect(email.Body.Table).To(BeEquivalentTo(ToTable(booking)))
	g.Expect(email.Body.Signature).To(Equal("Yours"))
}

// A fake Recapture client used for testing that always returns what you pass in
type fakeRecaptcha struct {
	success bool
	err     error
}

func (r *fakeRecaptcha) VerifyToken(ctx context.Context, token string) (success bool, err error) {
	return r.success, r.err
}

func TestValidateForm(t *testing.T) {
	testCases := []struct {
		name        string
		r           Recaptcha
		expectedErr string
	}{
		{
			"validation passes",
			&fakeRecaptcha{true, nil},
			"",
		},
		{
			"validation fails due to recapture",
			&fakeRecaptcha{false, nil},
			"Bot Suspected from token",
		},
		{
			"validation fails due to recapture error",
			&fakeRecaptcha{false, fmt.Errorf("some error")},
			"some error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			err := ValidateForm(context.TODO(), testCase.r, &contact)
			if testCase.expectedErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(testCase.expectedErr))
			}
		})
	}
}

func TestSendMessages(t *testing.T) {
	g := NewWithT(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("{}"))
		g.Expect(err).To(BeNil())
	}))
	client, err := gmail.NewService(context.Background(), option.WithEndpoint(ts.URL), option.WithAPIKey("abcd"))
	g.Expect(err).To(BeNil())

	err = SendMessages(context.TODO(), client, &contact)
	g.Expect(err).To(BeNil())
}

func TestHandleFormError(t *testing.T) {
	g := NewWithT(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "https://foo.com/contact", nil)

	HandleFormError(recorder, request, &contact, fmt.Errorf("some error"), "https://foo.com")

	response := recorder.Result()
	g.Expect(response.StatusCode).To(Equal(http.StatusFound))
	g.Expect(response.Header.Get("Location")).To(Equal("https://foo.com"))
}

func TestCreateMessageFromEmail(t *testing.T) {
	g := NewWithT(t)

	message, err := createMessageFromEmail("A", "a@foo.com", "B", "b@foo.com", "subject", contact.GetEmail())
	g.Expect(err).To(BeNil())

	// Decode the base64
	content, err := base64.StdEncoding.DecodeString(message.Raw)
	g.Expect(err).To(BeNil())

	// Parse the content into a mail.Message object
	reader := bytes.NewReader(content)
	mail, err := mail.ReadMessage(reader)
	g.Expect(err).To(BeNil())

	// Test the correct headers are set at the root of the mail message
	g.Expect(mail.Header.Get("From")).To(Equal("\"A\" <a@foo.com>"))
	g.Expect(mail.Header.Get("Reply-To")).To(Equal("\"A\" <a@foo.com>"))
	g.Expect(mail.Header.Get("Subject")).To(Equal("subject"))
	g.Expect(mail.Header.Get("To")).To(Equal("\"B\" <b@foo.com>"))

	// Parse the header to get the boundary which is dynamic
	_, params, err := mime.ParseMediaType(mail.Header.Get("Content-Type"))
	g.Expect(err).To(BeNil())

	// Read the multipart content in parts
	multiReader := multipart.NewReader(mail.Body, params["boundary"])

	// Test the first is plain text
	plain, err := multiReader.NextPart()
	g.Expect(err).To(BeNil())
	g.Expect(plain.Header.Get("Content-Type")).To(ContainSubstring("text/plain"))

	// Test the second part is html
	html, err := multiReader.NextPart()
	g.Expect(err).To(BeNil())
	g.Expect(html.Header.Get("Content-Type")).To(ContainSubstring("text/html"))

	// Test there are no more parts
	part, err := multiReader.NextPart()
	g.Expect(part).To(BeNil())
	g.Expect(err).To(MatchError("EOF"))
}
