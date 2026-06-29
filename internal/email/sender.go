package email

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To      string
	Subject string
	Body    string
}

type Config struct {
	From          string
	ResendAPIKey  string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	DevLog        bool
}

func NewSender(cfg Config) Sender {
	if cfg.DevLog || (cfg.ResendAPIKey == "" && cfg.SMTPHost == "") {
		return devSender{}
	}
	if cfg.ResendAPIKey != "" {
		return resendSender{apiKey: cfg.ResendAPIKey, from: cfg.From}
	}
	return smtpSender{cfg: cfg}
}

type devSender struct{}

func (devSender) Send(_ context.Context, msg Message) error {
	log.Printf("email (dev log) to=%s subject=%q\n%s", msg.To, msg.Subject, msg.Body)
	return nil
}

type resendSender struct {
	apiKey string
	from   string
}

func (s resendSender) Send(ctx context.Context, msg Message) error {
	payload := fmt.Sprintf(`{"from":%q,"to":[%q],"subject":%q,"text":%q}`,
		s.from, msg.To, msg.Subject, msg.Body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("resend: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

type smtpSender struct {
	cfg Config
}

func (s smtpSender) Send(_ context.Context, msg Message) error {
	addr := s.cfg.SMTPHost + ":" + s.cfg.SMTPPort
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.cfg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", msg.To))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	buf.WriteString(msg.Body)

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	return smtp.SendMail(addr, auth, s.cfg.From, []string{msg.To}, buf.Bytes())
}

func PasswordResetMessage(publicURL, token string) Message {
	link := strings.TrimRight(publicURL, "/") + "/reset-password?token=" + token
	return Message{
		Subject: "Reset your Rubrical password",
		Body:    "Use this link to reset your password (valid for 1 hour):\n\n" + link + "\n",
	}
}
