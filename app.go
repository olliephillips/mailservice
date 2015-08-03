package main

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/sendgrid/sendgrid-go"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

const (
	DEFAULT_PORT = "1323"
	DEFAULT_HOST = "localhost"
)

const (
	// Auth
	appUser      string = "appuser"
	appPass      string = "apppassword"
	
	// Assumes Sendgrid
	sendgridUser string = "username"
	sendgridPass string = "password"

	// Defaults
	msgType      string = "System"
	msgFromEmail string = "from@email.com"
	msgTemplate  string = "default.tmpl"
	msgSubject2  string = ""
)

type Message struct {
	ToName  string
	ToEmail string
	Subject string
	Body    string

	FromEmail string // Optional
	Subject2  string // Optional
	Template  string // Optional
	CCEmail1  string // Optional
	CCEmail2  string // Optional
	Type      string // Optional
}

// This handles the message sending
func send(c *echo.Context) error {

	var msg Message
	var merge bytes.Buffer

	// Parse out data
	// Type
	if c.Form("type") != "" {
		msg.Type = c.Form("type")
	} else {
		msg.Type = msgType
	}

	// Template
	if c.Form("template") != "" {
		msg.Template = c.Form("template")
	} else {
		msg.Template = msgTemplate
	}

	// FromEmail
	if c.Form("fromEmail") != "" {
		msg.FromEmail = c.Form("fromEmail")
	} else {
		msg.FromEmail = msgFromEmail
	}

	// Subject2
	if c.Form("subject2") != "" {
		msg.Subject2 = c.Form("subject2")
	} else {
		msg.Subject2 = msgSubject2
	}

	// CC
	if c.Form("ccEmail1") != "" {
		msg.CCEmail1 = c.Form("ccEmail1")
	} else {
		msg.CCEmail1 = ""
	}

	if c.Form("ccEmail2") != "" {
		msg.CCEmail2 = c.Form("ccEmail2")
	} else {
		msg.CCEmail2 = ""
	}

	msg.ToName = c.Form("toName")
	msg.ToEmail = c.Form("toEmail")
	msg.Subject = c.Form("subject")
	msg.Body = c.Form("body")

	// Combine body with template
	t := template.Must(template.ParseFiles(filepath.Join("templates", msg.Template)))
	t.Execute(&merge, msg)

	// Send the message
	sg := sendgrid.NewSendGridClient(sendgridUser, sendgridPass)
	m := sendgrid.NewMail()
	m.AddTo(msg.ToEmail)
	m.AddToName(msg.ToName)
	m.SetSubject(msg.Subject)
	m.SetHTML(merge.String())
	m.SetFrom(msg.FromEmail)

	// Copy to sender using BCC
	m.AddBcc(msg.FromEmail)

	// Handle up to 2 CC addresses
	if msg.CCEmail1 != "" {
		m.AddCc(msg.CCEmail1)
	}

	if msg.CCEmail2 != "" {
		m.AddCc(msg.CCEmail2)
	}

	// Based on success log & respond
	if r := sg.Send(m); r == nil {
		fmt.Println("Email sent")
		return c.JSON(201, "Email sent")
	} else {
		fmt.Println(r)
		return c.JSON(400, "There was a problem, the email could not be sent. Please check the logs")
	}

}

func main() {

	// Bluemix or local config options -- just local in this repo
	var port string
	port = DEFAULT_PORT

	var host string
	host = DEFAULT_HOST

	e := echo.New()

	// Basic Authentication
	e.Use(mw.BasicAuth(func(usr, pwd string) bool {
		if usr == appUser && pwd == appPass {
			return true
		}
		return false
	}))

	// Routes
	e.Post("/send", send)

	log.Printf("Starting mailservice on %+v:%+v\n", host, port)

	// Start server
	e.Run(host + ":" + port)

}
