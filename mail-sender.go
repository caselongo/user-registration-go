package main

import (
	"fmt"
	"github.com/caselongo/user-registration-go/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"os"
	"strings"
	"time"
)

const (
	noReplyEmail   string = "YOUR NO-REPLY EMAIL"
	smtpHost       string = "YOUR SMTP HOST"
	smtpPort       int    = 0 // YOUR SMTP PORT
	smtpUsername   string = "YOUR SMTP USERNAME"
	smtpPassword   string = "YOUR SMTP PASSWORD"
	smtpEncryption        = mail.EncryptionNone // YOUR SMTP ENCRYPTION
)

type MailSender struct {
	MailChan chan models.MailData
}

func NewMailSender() *MailSender {
	return &MailSender{MailChan: make(chan models.MailData)}
}

func (ms *MailSender) ListenForMail() {
	go func() {
		for {
			m := <-ms.MailChan
			sendMail(m)
		}
	}()
}

func (ms *MailSender) Close() {
	close(ms.MailChan)
}

func sendMail(data models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = smtpHost
	server.Port = smtpPort
	server.Encryption = smtpEncryption
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.Username = smtpUsername
	server.Password = smtpPassword

	client, err := server.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	email := mail.NewMSG()
	email.SetFrom(data.From).AddTo(data.To).SetSubject(data.Subject)
	if data.Template == "" {
		email.SetBody(mail.TextHTML, data.Content)
	} else {
		d, err := os.ReadFile(fmt.Sprintf("./email-templates/%s", data.Template))
		if err != nil {
			fmt.Println(err)
			return
		}

		mailTemplate := string(d)
		msgToSend := strings.ReplaceAll(mailTemplate, "[%url%]", data.Content)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("mail sent!")
	}
}

func (ms *MailSender) Confirm(email, code string) error {
	d, err := os.ReadFile("./email-templates/confirm.html")
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/confirm/%s", app.Host(), code)

	mailTemplate := string(d)
	content := strings.ReplaceAll(mailTemplate, "[%url%]", url)

	ms.MailChan <- models.MailData{
		To:      email,
		From:    noReplyEmail,
		Subject: "Confirm your e-mail address",
		Content: content,
	}
	return nil
}

func (ms *MailSender) Reset(email, code string) error {
	d, err := os.ReadFile("./email-templates/reset.html")
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/reset/%s", app.Host(), code)

	mailTemplate := string(d)
	content := strings.ReplaceAll(mailTemplate, "[%url%]", url)

	ms.MailChan <- models.MailData{
		To:      email,
		From:    noReplyEmail,
		Subject: "Reset your password",
		Content: content,
	}
	return nil
}
