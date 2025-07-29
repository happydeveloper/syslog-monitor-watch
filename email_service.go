package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService 이메일 전송 서비스
type EmailService struct {
	config *EmailConfig
	logger Logger
}

// Logger 인터페이스 정의
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewEmailService 새로운 이메일 서비스 생성
func NewEmailService(config *EmailConfig, logger Logger) *EmailService {
	return &EmailService{
		config: config,
		logger: logger,
	}
}

// SendEmail 이메일 전송 (Gmail 자동 감지)
func (es *EmailService) SendEmail(subject, body string) error {
	if !es.config.Enabled {
		return nil
	}

	// Gmail SMTP 서버 자동 감지 및 최적화된 전송
	if es.config.SMTPServer == DefaultSMTPServer {
		return es.sendGmailEmail(subject, body)
	}

	// 일반 SMTP 서버 전송
	return es.sendGenericEmail(subject, body)
}

// sendGmailEmail Gmail SMTP 최적화 전송
func (es *EmailService) sendGmailEmail(subject, body string) error {
	// Gmail SMTP 서버로 전송 (포트 587, STARTTLS)
	serverName := DefaultSMTPServer + ":" + DefaultSMTPPort

	// 인증 설정
	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, DefaultSMTPServer)

	// 이메일 메시지 구성
	message := es.buildEmailMessage(subject, body)

	// Gmail SMTP 전송
	err := smtp.SendMail(serverName, auth, es.config.From, es.config.To, []byte(message))
	if err != nil {
		return fmt.Errorf("%s: %v", ErrEmailSendFailed, err)
	}

	es.logger.Infof("✅ Gmail email sent successfully to: %s", strings.Join(es.config.To, ", "))
	return nil
}

// sendGenericEmail 범용 SMTP 서버 전송
func (es *EmailService) sendGenericEmail(subject, body string) error {
	message := es.buildEmailMessage(subject, body)
	serverName := es.config.SMTPServer + ":" + es.config.SMTPPort

	// 인증 설정
	var auth smtp.Auth
	if es.config.Username != "" && es.config.Password != "" {
		auth = smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.SMTPServer)
	}

	// TLS 설정
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         es.config.SMTPServer,
	}

	// 포트에 따라 다른 연결 방식 사용
	if es.config.SMTPPort == SMTPPortSSL {
		return es.sendWithSSL(serverName, auth, message, tlsConfig)
	}

	// STARTTLS 연결 (포트 587)
	return es.sendWithSTARTTLS(serverName, auth, message, tlsConfig)
}

// sendWithSSL SSL/TLS 직접 연결 (포트 465)
func (es *EmailService) sendWithSSL(serverName string, auth smtp.Auth, message string, tlsConfig *tls.Config) error {
	conn, err := tls.Dial("tcp", serverName, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server (SSL): %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, es.config.SMTPServer)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%s: %v", ErrSMTPAuth, err)
		}
	}

	return es.sendEmailMessage(client, message)
}

// sendWithSTARTTLS STARTTLS 연결 (포트 587)
func (es *EmailService) sendWithSTARTTLS(serverName string, auth smtp.Auth, message string, tlsConfig *tls.Config) error {
	client, err := smtp.Dial(serverName)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Quit()

	// STARTTLS 시작
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %v", err)
		}
	}

	// 인증
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("%s: %v", ErrSMTPAuth, err)
		}
	}

	return es.sendEmailMessage(client, message)
}

// sendEmailMessage SMTP 클라이언트를 통한 메시지 전송
func (es *EmailService) sendEmailMessage(client *smtp.Client, message string) error {
	// 발신자 설정
	if err := client.Mail(es.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// 수신자 설정
	for _, to := range es.config.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", to, err)
		}
	}

	// 메시지 전송
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	es.logger.Infof("✅ Email sent successfully to: %s", strings.Join(es.config.To, ", "))
	return nil
}

// buildEmailMessage 이메일 메시지 구성
func (es *EmailService) buildEmailMessage(subject, body string) string {
	message := fmt.Sprintf("From: %s\r\n", es.config.From)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(es.config.To, ","))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "\r\n"
	message += body
	return message
}

// SendTestEmail 테스트 이메일 전송
func (es *EmailService) SendTestEmail() error {
	subject := fmt.Sprintf("[TEST] %s - Test Email", AppName)
	body := fmt.Sprintf(`📧 테스트 이메일
==============

%s v%s 테스트 이메일입니다.

이 이메일을 받으셨다면 이메일 설정이 올바르게 구성되었습니다.

🕐 전송 시간: %s
📧 수신자: %s
🔧 SMTP 서버: %s:%s

정상적으로 이메일이 전송되고 있습니다! ✅
`,
		AppName,
		AppVersion,
		fmt.Sprintf("%s", strings.Join(es.config.To, ", ")),
		strings.Join(es.config.To, ", "),
		es.config.SMTPServer,
		es.config.SMTPPort,
	)

	return es.SendEmail(subject, body)
}

// GetRecipientsCount 수신자 수 반환
func (es *EmailService) GetRecipientsCount() int {
	return len(es.config.To)
}

// GetRecipientsList 수신자 목록 반환
func (es *EmailService) GetRecipientsList() string {
	return strings.Join(es.config.To, ", ")
}

// IsEnabled 이메일 서비스 활성화 여부 확인
func (es *EmailService) IsEnabled() bool {
	return es.config.Enabled
} 