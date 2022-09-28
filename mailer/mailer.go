package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer struct {
	client *mail.SMTPClient
	sender string
}

func New(host string, port int, username, password, sender string) (Mailer, error) {
	// Create email server
	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Connect to mail server
	client, err := server.Connect()
	if err != nil {
		return Mailer{}, err
	}

	return Mailer{client: client, sender: sender}, err
}

// Define a Send() method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates, and any
// dynamic data for the templates as an any parameter.
func (m Mailer) Send(recipient string, fileSystem embed.FS, templateFile string, data any) error {
	// Use the `ParseFS()` method to parse the required template file from the embedded
	// file system
	tmpl, err := template.New("email").ParseFS(fileSystem, "emails/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// And likewise with the "htmlBody" template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Setup email message
	email := mail.NewMSG()
	email.SetFrom(m.sender)
	email.AddTo(recipient)
	email.SetSubject(subject.String())
	email.SetBody(mail.TextPlain, plainBody.String())
	email.AddAlternative(mail.TextHTML, htmlBody.String())

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 milliseconds between each attempt.
	for i := 1; i <= 3; i++ {
		err = email.Send(m.client)
		// If everything worked, return nil
		if nil == err {
			return nil
		}

		// If it didn't work, sleep for a short time and retry.
		time.Sleep(500 * time.Millisecond)
	}

	return err
}
