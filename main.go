package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/smtp"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
)

type EmailConfig struct {
	SMTPServer   string
	SMTPPort     string
	Username     string
	Password     string
	To           []string
	From         string
	Enabled      bool
}

type SyslogMonitor struct {
	logFile     string
	filters     []string
	keywords    []string
	outputFile  string
	logger      *logrus.Logger
	emailConfig *EmailConfig
}

func NewSyslogMonitor(logFile, outputFile string, filters, keywords []string, emailConfig *EmailConfig) *SyslogMonitor {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		}
	}

	return &SyslogMonitor{
		logFile:     logFile,
		filters:     filters,
		keywords:    keywords,
		outputFile:  outputFile,
		logger:      logger,
		emailConfig: emailConfig,
	}
}

func (sm *SyslogMonitor) shouldFilter(line string) bool {
	if len(sm.filters) == 0 {
		return false
	}

	for _, filter := range sm.filters {
		matched, _ := regexp.MatchString(filter, line)
		if matched {
			return true
		}
	}
	return false
}

func (sm *SyslogMonitor) containsKeyword(line string) bool {
	if len(sm.keywords) == 0 {
		return true // 키워드가 없으면 모든 라인을 포함
	}

	lowLine := strings.ToLower(line)
	for _, keyword := range sm.keywords {
		if strings.Contains(lowLine, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (sm *SyslogMonitor) parseSyslogLine(line string) map[string]string {
	result := make(map[string]string)
	result["raw"] = line
	result["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	// 기본적인 syslog 파싱 (간단한 버전)
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		result["month"] = parts[0]
		result["day"] = parts[1]
		result["time"] = parts[2]
		if len(parts) >= 4 {
			result["host"] = parts[3]
			if len(parts) >= 5 {
				result["service"] = parts[4]
				result["message"] = strings.Join(parts[5:], " ")
			}
		}
	}

	return result
}

func (sm *SyslogMonitor) sendEmail(subject, body string) error {
	if !sm.emailConfig.Enabled {
		return nil
	}

	// Gmail SMTP 사용 시 간단한 방법 사용
	if sm.emailConfig.SMTPServer == "smtp.gmail.com" {
		return sm.sendGmailEmail(subject, body)
	}

	// 기타 SMTP 서버용 일반적인 방법
	return sm.sendGenericEmail(subject, body)
}

func (sm *SyslogMonitor) sendGmailEmail(subject, body string) error {
	// Gmail SMTP 설정
	auth := smtp.PlainAuth("", sm.emailConfig.Username, sm.emailConfig.Password, sm.emailConfig.SMTPServer)

	// 이메일 메시지 구성
	message := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		sm.emailConfig.From,
		strings.Join(sm.emailConfig.To, ","),
		subject,
		body))

	// Gmail SMTP 서버로 전송 (포트 587, STARTTLS)
	err := smtp.SendMail(
		sm.emailConfig.SMTPServer+":"+sm.emailConfig.SMTPPort,
		auth,
		sm.emailConfig.From,
		sm.emailConfig.To,
		message,
	)

	if err != nil {
		return fmt.Errorf("failed to send Gmail email: %v", err)
	}

	sm.logger.Infof("✅ Gmail email sent successfully to: %s", strings.Join(sm.emailConfig.To, ", "))
	return nil
}

func (sm *SyslogMonitor) sendGenericEmail(subject, body string) error {
	// 이메일 메시지 구성
	message := fmt.Sprintf("From: %s\r\n", sm.emailConfig.From)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(sm.emailConfig.To, ","))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// SMTP 서버 연결
	serverName := sm.emailConfig.SMTPServer + ":" + sm.emailConfig.SMTPPort

	// 인증 설정
	var auth smtp.Auth
	if sm.emailConfig.Username != "" && sm.emailConfig.Password != "" {
		auth = smtp.PlainAuth("", sm.emailConfig.Username, sm.emailConfig.Password, sm.emailConfig.SMTPServer)
	}

	// TLS 설정
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         sm.emailConfig.SMTPServer,
	}

	// 포트에 따라 다른 연결 방식 사용
	if sm.emailConfig.SMTPPort == "465" {
		// SSL/TLS 직접 연결 (포트 465)
		conn, err := tls.Dial("tcp", serverName, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server (SSL): %v", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, sm.emailConfig.SMTPServer)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %v", err)
		}
		defer client.Quit()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %v", err)
			}
		}

		return sm.sendEmailMessage(client, message)

	} else {
		// STARTTLS 연결 (포트 587)
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

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %v", err)
			}
		}

		return sm.sendEmailMessage(client, message)
	}
}

func (sm *SyslogMonitor) sendEmailMessage(client *smtp.Client, message string) error {
	// 발신자 설정
	if err := client.Mail(sm.emailConfig.From); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// 수신자 설정
	for _, to := range sm.emailConfig.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", to, err)
		}
	}

	// 메시지 전송
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %v", err)
	}
	defer writer.Close()

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	sm.logger.Infof("Email alert sent successfully to %s", strings.Join(sm.emailConfig.To, ","))
	return nil
}

func (sm *SyslogMonitor) processLine(line string) {
	// 필터링 체크
	if sm.shouldFilter(line) {
		return
	}

	// 키워드 체크
	if !sm.containsKeyword(line) {
		return
	}

	// 로그 파싱
	parsed := sm.parseSyslogLine(line)

	// 경고나 에러 레벨 감지
	lowLine := strings.ToLower(line)
	if strings.Contains(lowLine, "error") || strings.Contains(lowLine, "err") {
		sm.logger.WithFields(logrus.Fields{
			"level": "ERROR",
			"host":  parsed["host"],
			"service": parsed["service"],
		}).Error(parsed["message"])
		
		// 에러 발생 시 이메일 알림 전송
		if sm.emailConfig.Enabled {
			subject := fmt.Sprintf("[SYSLOG ERROR] %s - %s", parsed["host"], parsed["service"])
			body := fmt.Sprintf("시간: %s\n호스트: %s\n서비스: %s\n메시지: %s\n원본 로그: %s", 
				parsed["timestamp"], parsed["host"], parsed["service"], parsed["message"], line)
			
			sm.logger.Infof("📧 Sending ERROR alert to: %s", strings.Join(sm.emailConfig.To, ", "))
			go func() {
				if err := sm.sendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send email alert to %s: %v", strings.Join(sm.emailConfig.To, ", "), err)
				}
			}()
		}
		
	} else if strings.Contains(lowLine, "warn") || strings.Contains(lowLine, "warning") {
		sm.logger.WithFields(logrus.Fields{
			"level": "WARNING",
			"host":  parsed["host"],
			"service": parsed["service"],
		}).Warn(parsed["message"])
		
	} else if strings.Contains(lowLine, "fail") || strings.Contains(lowLine, "critical") {
		sm.logger.WithFields(logrus.Fields{
			"level": "CRITICAL",
			"host":  parsed["host"],
			"service": parsed["service"],
		}).Fatal(parsed["message"])
		
		// 크리티컬 에러 발생 시 이메일 알림 전송
		if sm.emailConfig.Enabled {
			subject := fmt.Sprintf("[SYSLOG CRITICAL] %s - %s", parsed["host"], parsed["service"])
			body := fmt.Sprintf("🚨 CRITICAL ALERT 🚨\n\n시간: %s\n호스트: %s\n서비스: %s\n메시지: %s\n원본 로그: %s", 
				parsed["timestamp"], parsed["host"], parsed["service"], parsed["message"], line)
			
			sm.logger.Warnf("🚨 Sending CRITICAL alert to: %s", strings.Join(sm.emailConfig.To, ", "))
			go func() {
				if err := sm.sendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send critical email alert to %s: %v", strings.Join(sm.emailConfig.To, ", "), err)
				}
			}()
		}
		
	} else {
		sm.logger.WithFields(logrus.Fields{
			"level": "INFO",
			"host":  parsed["host"],
			"service": parsed["service"],
		}).Info(parsed["message"])
	}
}

func (sm *SyslogMonitor) Start() error {
	// syslog 파일이 존재하는지 확인
	if _, err := os.Stat(sm.logFile); os.IsNotExist(err) {
		return fmt.Errorf("syslog file not found: %s", sm.logFile)
	}

	sm.logger.Infof("Starting syslog monitor for file: %s", sm.logFile)

	// tail을 사용해 파일을 실시간으로 감시
	t, err := tail.TailFile(sm.logFile, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
		Location: &tail.SeekInfo{Offset: 0, Whence: 2}, // 파일 끝에서 시작
	})
	if err != nil {
		return fmt.Errorf("failed to tail file: %v", err)
	}

	// 종료 신호 처리
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sm.logger.Info("Syslog monitor started. Press Ctrl+C to stop.")

	for {
		select {
		case line := <-t.Lines:
			if line.Err != nil {
				sm.logger.Errorf("Error reading line: %v", line.Err)
				continue
			}
			sm.processLine(line.Text)

		case <-sigChan:
			sm.logger.Info("Shutting down syslog monitor...")
			t.Stop()
			return nil
		}
	}
}

func main() {
	var (
		logFile       = flag.String("file", "/var/log/syslog", "Path to syslog file")
		outputFile    = flag.String("output", "", "Output file for filtered logs (default: stdout)")
		filterList    = flag.String("filters", "", "Comma-separated list of regex filters to exclude")
		keywordList   = flag.String("keywords", "", "Comma-separated list of keywords to include")
		showHelp      = flag.Bool("help", false, "Show help message")
		emailTo       = flag.String("email-to", "", "Email address to send alerts (comma-separated)")
		emailFrom     = flag.String("email-from", "", "Email sender address")
		smtpServer    = flag.String("smtp-server", "", "SMTP server address")
		smtpPort      = flag.String("smtp-port", "", "SMTP server port")
		smtpUser      = flag.String("smtp-user", "", "SMTP username")
		smtpPassword  = flag.String("smtp-password", "", "SMTP password")
		testEmail     = flag.Bool("test-email", false, "Send test email and exit")
	)
	flag.Parse()

	// 환경변수에서 이메일 설정 읽기
	if *emailTo == "" {
		*emailTo = os.Getenv("SYSLOG_EMAIL_TO")
		if *emailTo == "" {
			// 기본 설정: 여러 명에게 자동 전송
			*emailTo = "robot@lambda-x.ai,enfn2001@gmail.com"
		}
	}
	if *emailFrom == "" {
		*emailFrom = os.Getenv("SYSLOG_EMAIL_FROM")
		if *emailFrom == "" {
			*emailFrom = "enfn2001@gmail.com"
		}
	}
	if *smtpServer == "" {
		*smtpServer = os.Getenv("SYSLOG_SMTP_SERVER")
		if *smtpServer == "" {
			*smtpServer = "smtp.gmail.com"
		}
	}
	if *smtpPort == "" {
		*smtpPort = os.Getenv("SYSLOG_SMTP_PORT")
		if *smtpPort == "" {
			*smtpPort = "587"
		}
	}
	if *smtpUser == "" {
		*smtpUser = os.Getenv("SYSLOG_SMTP_USER")
		if *smtpUser == "" {
			// 기본 SMTP 사용자
			*smtpUser = "enfn2001@gmail.com"
		}
	}
	if *smtpPassword == "" {
		*smtpPassword = os.Getenv("SYSLOG_SMTP_PASSWORD")
		if *smtpPassword == "" {
			// 기본 App Password (테스트에서 성공한 값)
			*smtpPassword = "lcsn auno hcqx zozp"
		}
	}

	if *showHelp {
		fmt.Println("Syslog Monitor - Real-time syslog monitoring service")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  syslog-monitor [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Monitor default syslog with all messages")
		fmt.Println("  ./syslog-monitor")
		fmt.Println()
		fmt.Println("  # Monitor specific file with keyword filtering")
		fmt.Println("  ./syslog-monitor -file=/var/log/auth.log -keywords=failed,error")
		fmt.Println()
		fmt.Println("  # Monitor with output to file and filtering")
		fmt.Println("  ./syslog-monitor -output=monitor.log -filters=systemd,kernel")
		fmt.Println()
		fmt.Println("  # Monitor with default email alerts (multiple recipients)")
		fmt.Println("  ./syslog-monitor")
		fmt.Println()
		fmt.Println("  # Monitor with custom multiple recipients")
		fmt.Println("  ./syslog-monitor -email-to=\"admin@company.com,security@company.com,ops@company.com\"")
		fmt.Println()
		fmt.Println("  # Monitor with email alerts (using command line)")
		fmt.Println("  ./syslog-monitor -email-to=admin@example.com -smtp-user=your@gmail.com -smtp-password=yourapppassword")
		fmt.Println()
		fmt.Println("  # Monitor with email alerts (using environment variables)")
		fmt.Println("  export SYSLOG_EMAIL_TO=\"admin@company.com,security@company.com\"")
		fmt.Println("  export SYSLOG_SMTP_USER=your@gmail.com")
		fmt.Println("  export SYSLOG_SMTP_PASSWORD=yourapppassword")
		fmt.Println("  ./syslog-monitor")
		fmt.Println()
		fmt.Println("  # Test email configuration (multiple recipients)")
		fmt.Println("  ./syslog-monitor -test-email -email-to=\"user1@test.com,user2@test.com\"")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  SYSLOG_EMAIL_TO      - Email addresses to send alerts (comma-separated)")
		fmt.Println("  SYSLOG_EMAIL_FROM    - Email sender address")
		fmt.Println("  SYSLOG_SMTP_SERVER   - SMTP server (default: smtp.gmail.com)")
		fmt.Println("  SYSLOG_SMTP_PORT     - SMTP port (default: 587)")
		fmt.Println("  SYSLOG_SMTP_USER     - SMTP username")
		fmt.Println("  SYSLOG_SMTP_PASSWORD - SMTP password")
		fmt.Println()
		fmt.Println("Gmail Setup:")
		fmt.Println("  1. Enable 2-Step Verification in your Google Account")
		fmt.Println("  2. Generate App Password at: https://myaccount.google.com/apppasswords")
		fmt.Println("  3. Use the App Password instead of your regular password")
		return
	}

	// 필터와 키워드 파싱
	var filters []string
	var keywords []string

	if *filterList != "" {
		filters = strings.Split(*filterList, ",")
		for i := range filters {
			filters[i] = strings.TrimSpace(filters[i])
		}
	}

	if *keywordList != "" {
		keywords = strings.Split(*keywordList, ",")
		for i := range keywords {
			keywords[i] = strings.TrimSpace(keywords[i])
		}
	}

	// 이메일 설정 (기본값으로 항상 활성화)
	emailConfig := &EmailConfig{
		SMTPServer: *smtpServer,
		SMTPPort:   *smtpPort,
		Username:   *smtpUser,
		Password:   *smtpPassword,
		From:       *emailFrom,
		Enabled:    true, // 기본값으로 항상 활성화
	}

	// 이메일 주소 파싱
	emails := strings.Split(*emailTo, ",")
	for i := range emails {
		emails[i] = strings.TrimSpace(emails[i])
	}
	emailConfig.To = emails

	// 사용자 알림
	if (*emailTo == "robot@lambda-x.ai,enfn2001@gmail.com" || *emailTo == "robot@lambda-x.ai" || *emailTo == "enfn2001@gmail.com") && *smtpUser == "enfn2001@gmail.com" {
		fmt.Printf("📧 Email alerts enabled with DEFAULT settings\n")
		fmt.Printf("    📨 Recipients (%d): %s\n", len(emailConfig.To), strings.Join(emailConfig.To, ", "))
		fmt.Printf("    🔑 Using built-in Gmail credentials (enfn2001@gmail.com)\n")
		fmt.Printf("    💡 To add more recipients: -email-to=\"user1@example.com,user2@example.com\"\n")
	} else {
		fmt.Printf("📧 Email alerts enabled with CUSTOM settings\n")
		fmt.Printf("    📨 Recipients (%d): %s\n", len(emailConfig.To), strings.Join(emailConfig.To, ", "))
		
		if *smtpUser == "" || *smtpPassword == "" {
			fmt.Println("⚠️  Warning: SMTP username or password not provided. Email alerts may not work.")
			fmt.Println("    For Gmail, generate an App Password at: https://myaccount.google.com/apppasswords")
			fmt.Println("    Use: ./email-setup.sh for easy configuration")
		}
	}

	// 테스트 이메일 전송
	if *testEmail {
		if !emailConfig.Enabled {
			fmt.Println("Error: Email configuration required for test email")
			fmt.Println("Please provide -email-to and SMTP credentials")
			os.Exit(1)
		}

		fmt.Println("Sending test email...")
		
		monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig)
		subject := "[TEST] Syslog Monitor Email Test"
		body := fmt.Sprintf(`이것은 syslog 모니터의 테스트 이메일입니다.

테스트 시간: %s
SMTP 서버: %s:%s
발신자: %s
수신자: %s

이 이메일을 받으셨다면 이메일 설정이 올바르게 구성되었습니다.

Syslog Monitor
`, time.Now().Format("2006-01-02 15:04:05"), *smtpServer, *smtpPort, *emailFrom, strings.Join(emailConfig.To, ", "))

		if err := monitor.sendEmail(subject, body); err != nil {
			fmt.Printf("Test email failed: %v\n", err)
			fmt.Println("\nTroubleshooting:")
			fmt.Println("1. Check your Gmail App Password")
			fmt.Println("2. Ensure 2-Step Verification is enabled")
			fmt.Println("3. Verify SMTP server and port settings")
			os.Exit(1)
		}

		fmt.Printf("✅ Test email sent successfully to %d recipients: %s\n", len(emailConfig.To), strings.Join(emailConfig.To, ", "))
		return
	}

	// 감시 서비스 생성 및 시작
	monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig)
	
	if err := monitor.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
} 