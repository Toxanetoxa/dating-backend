package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"gopkg.in/mail.v2"
)

// Email сервис отправки емейлов с кодом авторизации
type Email struct {
	l      *slog.Logger
	enable bool
	dialer *mail.Dialer
	from   string
}

func NewEmailService(l *slog.Logger, enable bool, smtpHost string, smtpPort int, smtpUser, smtpPassword, from string) EmailService {
	d := mail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return &Email{
		l:      l,
		enable: enable,
		dialer: d,
		from:   from,
	}
}

func (e *Email) SendCode(ctx context.Context, to string, code string) error {
	if !e.enable {
		e.l.Info("fake sending", "code", code, "to", to)

		return nil
	}

	m := mail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Код авторизации")
	m.SetBody("text/plain", fmt.Sprintf("Ваш код для авторизации: %s", code)) // utf-8 must be by default
	err := e.dialer.DialAndSend(m)
	if err != nil {
		e.l.Error("could not send email", "error", err)

		return err
	}

	return nil
}
