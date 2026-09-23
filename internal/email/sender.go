package email

import "log"

type Sender interface {
	SendPasswordResetEmail(to string, resetToken string) error
	SendVerificationEmail(to string, verificationToken string) error // ← tambah
}

type LogSender struct{}

func NewLogSender() Sender {
	return &LogSender{}
}

func (s *LogSender) SendPasswordResetEmail(to string, resetToken string) error {
	log.Printf("=== EMAIL (dummy) ===\nTo: %s\nReset token: %s\n=====================", to, resetToken)
	return nil
}

func (s *LogSender) SendVerificationEmail(to string, verificationToken string) error {
	log.Printf("=== EMAIL (dummy) ===\nTo: %s\nVerification token: %s\n=====================", to, verificationToken)
	return nil
}