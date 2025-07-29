package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"regexp"
	"runtime"
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

type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
	Enabled    bool
}

type SlackMessage struct {
	Channel   string             `json:"channel,omitempty"`
	Username  string             `json:"username,omitempty"`
	Text      string             `json:"text,omitempty"`
	IconEmoji string             `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SyslogMonitor struct {
	logFile       string
	filters       []string
	keywords      []string
	outputFile    string
	logger        *logrus.Logger
	emailConfig   *EmailConfig
	slackConfig   *SlackConfig
	aiAnalyzer    *AIAnalyzer
	systemMonitor *SystemMonitor
	logParser     *LogParserManager
	aiEnabled     bool
	systemEnabled bool
}

func NewSyslogMonitor(logFile, outputFile string, filters, keywords []string, emailConfig *EmailConfig, slackConfig *SlackConfig, aiEnabled, systemEnabled bool) *SyslogMonitor {
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

	var aiAnalyzer *AIAnalyzer
	var systemMonitor *SystemMonitor
	
	if aiEnabled {
		aiAnalyzer = NewAIAnalyzer()
	}
	
	if systemEnabled {
		systemMonitor = NewSystemMonitor(time.Minute * 5) // 5분 간격으로 시스템 모니터링
	}

	return &SyslogMonitor{
		logFile:       logFile,
		filters:       filters,
		keywords:      keywords,
		outputFile:    outputFile,
		logger:        logger,
		emailConfig:   emailConfig,
		slackConfig:   slackConfig,
		aiAnalyzer:    aiAnalyzer,
		systemMonitor: systemMonitor,
		logParser:     NewLogParserManager(),
		aiEnabled:     aiEnabled,
		systemEnabled: systemEnabled,
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

func (sm *SyslogMonitor) sendSlackMessage(message SlackMessage) error {
	if !sm.slackConfig.Enabled {
		return nil
	}

	// 기본값 설정
	if message.Channel == "" && sm.slackConfig.Channel != "" {
		message.Channel = sm.slackConfig.Channel
	}
	if message.Username == "" && sm.slackConfig.Username != "" {
		message.Username = sm.slackConfig.Username
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %v", err)
	}

	resp, err := http.Post(sm.slackConfig.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send slack message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status: %d", resp.StatusCode)
	}

	sm.logger.Infof("✅ Slack message sent successfully to channel: %s", message.Channel)
	return nil
}

func (sm *SyslogMonitor) detectLoginPattern(line string) (bool, map[string]string) {
	patterns := map[string]*regexp.Regexp{
		"ssh_success": regexp.MustCompile(`(?i)sshd.*Accepted\s+(password|publickey|keyboard-interactive)\s+for\s+(\w+)\s+from\s+([\d\.]+)`),
		"ssh_failed":  regexp.MustCompile(`(?i)sshd.*Failed\s+(password|publickey)\s+for\s+(\w+)\s+from\s+([\d\.]+)`),
		"sudo_usage":  regexp.MustCompile(`(?i)sudo.*USER=(\w+).*COMMAND=(.+)`),
		"user_login":  regexp.MustCompile(`(?i)(login|session).*user\s+(\w+)`),
		"web_login":   regexp.MustCompile(`(?i)(login|authentication).*user[:\s]+(\w+).*from[:\s]+([\d\.]+)`),
	}

	for patternName, pattern := range patterns {
		matches := pattern.FindStringSubmatch(line)
		if matches != nil {
			result := map[string]string{
				"pattern": patternName,
				"line":    line,
			}

			switch patternName {
			case "ssh_success":
				result["method"] = matches[1]
				result["user"] = matches[2]
				result["ip"] = matches[3]
				result["status"] = "success"
			case "ssh_failed":
				result["method"] = matches[1]
				result["user"] = matches[2]
				result["ip"] = matches[3]
				result["status"] = "failed"
			case "sudo_usage":
				result["user"] = matches[1]
				result["command"] = matches[2]
				result["status"] = "sudo"
			case "user_login":
				result["user"] = matches[2]
				result["status"] = "login"
			case "web_login":
				result["user"] = matches[2]
				result["ip"] = matches[3]
				result["status"] = "web_login"
			}
			return true, result
		}
	}
	return false, nil
}

func (sm *SyslogMonitor) createLoginSlackMessage(loginInfo map[string]string, parsed map[string]string) SlackMessage {
	var color, title, emoji string
	var fields []SlackField

	switch loginInfo["status"] {
	case "success":
		color = "good"
		title = "✅ SSH Login Success"
		emoji = ":white_check_mark:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "IP Address", Value: loginInfo["ip"], Short: true},
			{Title: "Method", Value: loginInfo["method"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
		}
	case "failed":
		color = "danger"
		title = "❌ SSH Login Failed"
		emoji = ":x:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "IP Address", Value: loginInfo["ip"], Short: true},
			{Title: "Method", Value: loginInfo["method"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
		}
	case "sudo":
		color = "warning"
		title = "⚡ Sudo Command Executed"
		emoji = ":zap:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
			{Title: "Command", Value: loginInfo["command"], Short: false},
		}
	case "web_login":
		color = "good"
		title = "🌐 Web Login Detected"
		emoji = ":globe_with_meridians:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "IP Address", Value: loginInfo["ip"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
		}
	default:
		color = "#36a64f"
		title = "👤 User Activity"
		emoji = ":bust_in_silhouette:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
			{Title: "Activity", Value: loginInfo["status"], Short: true},
		}
	}

	attachment := SlackAttachment{
		Color:     color,
		Title:     title,
		Fields:    fields,
		Timestamp: time.Now().Unix(),
	}

	return SlackMessage{
		Text:      fmt.Sprintf("%s *%s*", emoji, title),
		IconEmoji: ":robot_face:",
		Username:  "Syslog Monitor",
		Attachments: []SlackAttachment{attachment},
	}
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

	// 기본 로그 파싱
	parsed := sm.parseSyslogLine(line)
	
	// 고급 로그 파싱 (AI 분석 활성화된 경우)
	var parsedLog *ParsedLog
	if sm.aiEnabled {
		parsedLog = sm.logParser.ParseLog(line)
	}

	// AI 분석 수행
	var aiResult *AIAnalysisResult
	if sm.aiEnabled && sm.aiAnalyzer != nil {
		aiResult = sm.aiAnalyzer.AnalyzeLog(line, parsed)
		
		// AI 분석 결과에 따른 알림
		if aiResult.AnomalyScore >= sm.aiAnalyzer.alertThreshold {
			sm.sendAIAlert(aiResult, parsedLog)
		}
	}

	// 로그인 패턴 감지 (우선 처리)
	if isLogin, loginInfo := sm.detectLoginPattern(line); isLogin {
		sm.logger.WithFields(logrus.Fields{
			"level": "LOGIN",
			"user":  loginInfo["user"],
			"host":  parsed["host"],
			"status": loginInfo["status"],
		}).Infof("User activity detected: %s", loginInfo["status"])

		// 슬랙 로그인 알림 전송
		if sm.slackConfig.Enabled {
			slackMsg := sm.createLoginSlackMessage(loginInfo, parsed)
			sm.logger.Infof("💬 Sending login notification to Slack: %s", loginInfo["user"])
			go func() {
				if err := sm.sendSlackMessage(slackMsg); err != nil {
					sm.logger.Errorf("❌ Failed to send Slack login notification: %v", err)
				}
			}()
		}
	}

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

		// 에러 시 슬랙 알림도 전송
		if sm.slackConfig.Enabled {
			slackMsg := SlackMessage{
				Text:      fmt.Sprintf("🔴 *ERROR Alert*"),
				IconEmoji: ":rotating_light:",
				Username:  "Syslog Monitor",
				Attachments: []SlackAttachment{
					{
						Color: "danger",
						Title: fmt.Sprintf("Error on %s", parsed["host"]),
						Fields: []SlackField{
							{Title: "Service", Value: parsed["service"], Short: true},
							{Title: "Host", Value: parsed["host"], Short: true},
							{Title: "Message", Value: parsed["message"], Short: false},
						},
						Timestamp: time.Now().Unix(),
					},
				},
			}
			go func() {
				if err := sm.sendSlackMessage(slackMsg); err != nil {
					sm.logger.Errorf("❌ Failed to send Slack error alert: %v", err)
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

		// 크리티컬 에러 시 슬랙 긴급 알림
		if sm.slackConfig.Enabled {
			slackMsg := SlackMessage{
				Text:      fmt.Sprintf("🚨 *CRITICAL ALERT* 🚨"),
				IconEmoji: ":warning:",
				Username:  "Syslog Monitor",
				Attachments: []SlackAttachment{
					{
						Color: "#ff0000",
						Title: fmt.Sprintf("CRITICAL ERROR on %s", parsed["host"]),
						Fields: []SlackField{
							{Title: "Service", Value: parsed["service"], Short: true},
							{Title: "Host", Value: parsed["host"], Short: true},
							{Title: "Message", Value: parsed["message"], Short: false},
						},
						Timestamp: time.Now().Unix(),
					},
				},
			}
			go func() {
				if err := sm.sendSlackMessage(slackMsg); err != nil {
					sm.logger.Errorf("❌ Failed to send Slack critical alert: %v", err)
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
		if runtime.GOOS == "darwin" {
			// macOS 사용자를 위한 상세한 안내
			sm.logger.Errorf("❌ 로그 파일을 찾을 수 없습니다: %s", sm.logFile)
			sm.logger.Info("🍎 macOS에서 사용 가능한 로그 파일들:")
			
			recommendations := getMacOSLogRecommendations()
			for _, rec := range recommendations {
				if rec == "" {
					sm.logger.Info("")
				} else {
					sm.logger.Infof("   %s", rec)
				}
			}
			
			sm.logger.Info("")
			sm.logger.Info("💡 사용법 예시:")
			sm.logger.Info("   # 설치 로그 모니터링")
			sm.logger.Info("   ./syslog-monitor -file=/var/log/install.log")
			sm.logger.Info("")
			sm.logger.Info("   # WiFi 로그 모니터링")  
			sm.logger.Info("   ./syslog-monitor -file=/var/log/wifi.log")
			sm.logger.Info("")
			sm.logger.Info("   # 실시간 시스템 로그 (sudo 필요)")
			sm.logger.Info("   sudo log stream | ./syslog-monitor -file=/dev/stdin")
			
			return fmt.Errorf("macOS에서는 다른 로그 파일 경로를 사용해주세요")
		} else {
			return fmt.Errorf("syslog file not found: %s", sm.logFile)
		}
	}

	sm.logger.Infof("Starting syslog monitor for file: %s", sm.logFile)
	
	// AI 분석 활성화 메시지
	if sm.aiEnabled {
		sm.logger.Infof("🤖 AI 로그 분석이 활성화되었습니다")
		sm.logger.Infof(sm.aiAnalyzer.GetAnalysisReport())
	}
	
	// 시스템 모니터링 시작
	if sm.systemEnabled && sm.systemMonitor != nil {
		sm.logger.Infof("🖥️  시스템 모니터링을 시작합니다")
		sm.systemMonitor.Start()
		
		// 시스템 알림 처리 고루틴
		go sm.handleSystemAlerts()
		
		sm.logger.Infof(sm.systemMonitor.GetSystemReport())
	}

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

// sendAIAlert AI 분석 결과 알림 전송
func (sm *SyslogMonitor) sendAIAlert(aiResult *AIAnalysisResult, parsedLog *ParsedLog) {
	// 이메일 알림
	if sm.emailConfig.Enabled {
		subject := fmt.Sprintf("[AI ALERT %s] %s", aiResult.ThreatLevel, "이상 징후 감지")
		
		body := fmt.Sprintf(`🤖 AI 로그 분석 결과

위협 레벨: %s
이상 점수: %.2f/10
신뢰도: %.1f%%
감지 시간: %s

📊 분석 결과:
- 영향받는 시스템: %s

🔮 예측 결과:`,
			aiResult.ThreatLevel,
			aiResult.AnomalyScore,
			aiResult.Confidence*100,
			aiResult.Timestamp.Format("2006-01-02 15:04:05"),
			strings.Join(aiResult.AffectedSystems, ", "),
		)
		
		for _, prediction := range aiResult.Predictions {
			body += fmt.Sprintf(`
- %s (확률: %.0f%%, 시간: %s)
  영향: %s`, prediction.Event, prediction.Probability*100, prediction.TimeFrame, prediction.Impact)
		}
		
		body += "\n\n💡 추천 조치사항:"
		for _, recommendation := range aiResult.Recommendations {
			body += fmt.Sprintf("\n- %s", recommendation)
		}
		
		if parsedLog != nil {
			body += fmt.Sprintf(`

📋 로그 상세 정보:
- 로그 타입: %s
- 레벨: %s
- 메시지: %s
- 원본: %s`, parsedLog.LogType, parsedLog.Level, parsedLog.Message, parsedLog.RawLog)
		}
		
		sm.logger.Infof("🚨 Sending AI alert to: %s", strings.Join(sm.emailConfig.To, ", "))
		go func() {
			if err := sm.sendEmail(subject, body); err != nil {
				sm.logger.Errorf("❌ Failed to send AI alert email: %v", err)
			}
		}()
	}
	
	// 슬랙 알림
	if sm.slackConfig.Enabled {
		color := "warning"
		if aiResult.AnomalyScore >= 8.0 {
			color = "danger"
		}
		
		fields := []SlackField{
			{Title: "위협 레벨", Value: aiResult.ThreatLevel, Short: true},
			{Title: "이상 점수", Value: fmt.Sprintf("%.2f/10", aiResult.AnomalyScore), Short: true},
			{Title: "신뢰도", Value: fmt.Sprintf("%.1f%%", aiResult.Confidence*100), Short: true},
			{Title: "영향 시스템", Value: strings.Join(aiResult.AffectedSystems, ", "), Short: false},
		}
		
		// 예측 결과 추가
		if len(aiResult.Predictions) > 0 {
			predictionText := ""
			for _, prediction := range aiResult.Predictions {
				predictionText += fmt.Sprintf("• %s (%.0f%%)\n", prediction.Event, prediction.Probability*100)
			}
			fields = append(fields, SlackField{Title: "예측", Value: predictionText, Short: false})
		}
		
		slackMsg := SlackMessage{
			Text:      fmt.Sprintf("🤖 *AI 이상 징후 감지* %s", aiResult.ThreatLevel),
			IconEmoji: ":robot_face:",
			Username:  "AI Log Analyzer",
			Attachments: []SlackAttachment{
				{
					Color:     color,
					Title:     "AI 분석 결과",
					Fields:    fields,
					Timestamp: time.Now().Unix(),
				},
			},
		}
		
		go func() {
			if err := sm.sendSlackMessage(slackMsg); err != nil {
				sm.logger.Errorf("❌ Failed to send AI alert to Slack: %v", err)
			}
		}()
	}
}

// handleSystemAlerts 시스템 알림 처리
func (sm *SyslogMonitor) handleSystemAlerts() {
	for alert := range sm.systemMonitor.GetAlertChannel() {
		sm.logger.WithFields(logrus.Fields{
			"level": "SYSTEM_ALERT",
			"type":  alert.Type,
			"value": alert.Value,
		}).Warnf("System alert: %s", alert.Message)
		
		// 이메일 알림
		if sm.emailConfig.Enabled {
			subject := fmt.Sprintf("[SYSTEM ALERT %s] %s", alert.Level, alert.Type)
			
			body := fmt.Sprintf(`🖥️  시스템 알림

알림 레벨: %s
타입: %s
메시지: %s
현재 값: %.2f
임계값: %.2f
시간: %s

💡 추천 조치사항:`,
				alert.Level,
				alert.Type,
				alert.Message,
				alert.Value,
				alert.Threshold,
				alert.Timestamp.Format("2006-01-02 15:04:05"),
			)
			
			for _, suggestion := range alert.Suggestions {
				body += fmt.Sprintf("\n- %s", suggestion)
			}
			
			body += fmt.Sprintf(`

📊 시스템 상태:
- CPU 사용률: %.1f%%
- 메모리 사용률: %.1f%%
- 로드 평균: %.2f`,
				alert.Metrics.CPU.UsagePercent,
				alert.Metrics.Memory.UsagePercent,
				alert.Metrics.LoadAverage.Load1Min,
			)
			
			sm.logger.Infof("🖥️  Sending system alert to: %s", strings.Join(sm.emailConfig.To, ", "))
			go func() {
				if err := sm.sendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send system alert email: %v", err)
				}
			}()
		}
		
		// 슬랙 알림
		if sm.slackConfig.Enabled {
			color := "warning"
			if alert.Level == "CRITICAL" {
				color = "danger"
			} else if alert.Level == "MEDIUM" {
				color = "warning"
			} else {
				color = "good"
			}
			
			fields := []SlackField{
				{Title: "알림 레벨", Value: alert.Level, Short: true},
				{Title: "타입", Value: alert.Type, Short: true},
				{Title: "현재 값", Value: fmt.Sprintf("%.2f", alert.Value), Short: true},
				{Title: "임계값", Value: fmt.Sprintf("%.2f", alert.Threshold), Short: true},
				{Title: "CPU", Value: fmt.Sprintf("%.1f%%", alert.Metrics.CPU.UsagePercent), Short: true},
				{Title: "메모리", Value: fmt.Sprintf("%.1f%%", alert.Metrics.Memory.UsagePercent), Short: true},
			}
			
			slackMsg := SlackMessage{
				Text:      fmt.Sprintf("🖥️  *시스템 알림* - %s", alert.Message),
				IconEmoji: ":warning:",
				Username:  "System Monitor",
				Attachments: []SlackAttachment{
					{
						Color:     color,
						Title:     fmt.Sprintf("%s Alert", alert.Type),
						Fields:    fields,
						Timestamp: time.Now().Unix(),
					},
				},
			}
			
			go func() {
				if err := sm.sendSlackMessage(slackMsg); err != nil {
					sm.logger.Errorf("❌ Failed to send system alert to Slack: %v", err)
				}
			}()
		}
	}
}

// getDefaultLogFile 운영체제에 따른 기본 로그 파일 경로 반환
func getDefaultLogFile() string {
	switch runtime.GOOS {
	case "darwin": // macOS
		// macOS에서 일반적으로 접근 가능한 로그 파일들을 순서대로 확인
		macOSLogFiles := []string{
			"/var/log/system.log",    // macOS 주요 시스템 로그
			"/var/log/install.log",   // 설치 로그
			"/var/log/wifi.log",      // WiFi 로그
			"/usr/local/var/log/messages", // Homebrew 환경
		}
		
		for _, logFile := range macOSLogFiles {
			if _, err := os.Stat(logFile); err == nil {
				return logFile
			}
		}
		
		// 기본값으로 system.log 반환 (존재하지 않아도)
		return "/var/log/system.log"
		
	case "linux":
		return "/var/log/syslog"
		
	default:
		return "/var/log/syslog"
	}
}

// getMacOSLogRecommendations macOS 사용자를 위한 로그 파일 추천
func getMacOSLogRecommendations() []string {
	return []string{
		"/var/log/system.log     # 주요 시스템 로그 (macOS Monterey 이전)",
		"/var/log/install.log    # 패키지 설치 로그",
		"/var/log/wifi.log       # WiFi 연결 로그",
		"/var/log/kernel.log     # 커널 로그",
		"/var/log/fsck_hfs.log   # 파일시스템 체크 로그",
		"",
		"💡 macOS Big Sur/Monterey 이후:",
		"   sudo log show --predicate 'process == \"kernel\"' --last 1h",
		"   sudo log show --predicate 'eventMessage contains \"error\"' --last 1h",
		"   sudo log stream --predicate 'process == \"syslogd\"'",
	}
}

func main() {
	defaultLogFile := getDefaultLogFile()
	
	var (
		logFile       = flag.String("file", defaultLogFile, "Path to syslog file")
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
		slackWebhook  = flag.String("slack-webhook", "", "Slack webhook URL for notifications")
		slackChannel  = flag.String("slack-channel", "", "Slack channel (default: webhook default)")
		slackUsername = flag.String("slack-username", "Syslog Monitor", "Slack bot username")
		testSlack     = flag.Bool("test-slack", false, "Send test Slack message and exit")
		loginWatch    = flag.Bool("login-watch", false, "Enable login monitoring (SSH, sudo, web)")
		aiEnabled     = flag.Bool("ai-analysis", false, "Enable AI-based log analysis and anomaly detection")
		systemEnabled = flag.Bool("system-monitor", false, "Enable system metrics monitoring (CPU, memory, disk, temperature)")
		_ = flag.String("log-type", "auto", "Log type for parsing (auto, apache, nginx, mysql, postgresql, application)") // Reserved for future use
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
	if *slackWebhook == "" {
		*slackWebhook = os.Getenv("SYSLOG_SLACK_WEBHOOK")
	}
	if *slackChannel == "" {
		*slackChannel = os.Getenv("SYSLOG_SLACK_CHANNEL")
	}
	if *slackUsername == "Syslog Monitor" {
		if env := os.Getenv("SYSLOG_SLACK_USERNAME"); env != "" {
			*slackUsername = env
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
		if runtime.GOOS == "darwin" {
			fmt.Println("  # macOS specific examples")
			fmt.Println("  ./syslog-monitor -file=/var/log/system.log -ai-analysis")
			fmt.Println("  ./syslog-monitor -file=/var/log/install.log -keywords=error")
			fmt.Println("  ./syslog-monitor -file=/var/log/wifi.log -system-monitor")
			fmt.Println("  sudo log stream | ./syslog-monitor -file=/dev/stdin -ai-analysis")
		}
		fmt.Println()
		fmt.Println("  # Test email configuration (multiple recipients)")
		fmt.Println("  ./syslog-monitor -test-email -email-to=\"user1@test.com,user2@test.com\"")
		fmt.Println()
		fmt.Println("  # Slack integration with login monitoring")
		fmt.Println("  ./syslog-monitor -slack-webhook=https://hooks.slack.com/... -login-watch")
		fmt.Println()
		fmt.Println("  # Combined email + Slack alerts")
		fmt.Println("  ./syslog-monitor -slack-webhook=https://hooks.slack.com/... -slack-channel=#alerts")
		fmt.Println()
		fmt.Println("  # Test Slack integration")
		fmt.Println("  ./syslog-monitor -test-slack -slack-webhook=https://hooks.slack.com/...")
		fmt.Println()
		fmt.Println("  # AI-powered log analysis with system monitoring")
		fmt.Println("  ./syslog-monitor -ai-analysis -system-monitor")
		fmt.Println()
		fmt.Println("  # Monitor web server logs with AI analysis")
		fmt.Println("  ./syslog-monitor -file=/var/log/nginx/access.log -log-type=nginx -ai-analysis")
		fmt.Println()
		fmt.Println("  # Database log monitoring with anomaly detection")
		fmt.Println("  ./syslog-monitor -file=/var/log/mysql/error.log -log-type=mysql -ai-analysis")
		fmt.Println()
		fmt.Println("  # Complete monitoring setup")
		fmt.Println("  ./syslog-monitor -ai-analysis -system-monitor -login-watch -slack-webhook=URL")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  SYSLOG_EMAIL_TO        - Email addresses to send alerts (comma-separated)")
		fmt.Println("  SYSLOG_EMAIL_FROM      - Email sender address")
		fmt.Println("  SYSLOG_SMTP_SERVER     - SMTP server (default: smtp.gmail.com)")
		fmt.Println("  SYSLOG_SMTP_PORT       - SMTP port (default: 587)")
		fmt.Println("  SYSLOG_SMTP_USER       - SMTP username")
		fmt.Println("  SYSLOG_SMTP_PASSWORD   - SMTP password")
		fmt.Println("  SYSLOG_SLACK_WEBHOOK   - Slack webhook URL")
		fmt.Println("  SYSLOG_SLACK_CHANNEL   - Slack channel")
		fmt.Println("  SYSLOG_SLACK_USERNAME  - Slack bot username")
		fmt.Println()
		fmt.Println("Gmail Setup:")
		fmt.Println("  1. Enable 2-Step Verification in your Google Account")
		fmt.Println("  2. Generate App Password at: https://myaccount.google.com/apppasswords")
		fmt.Println("  3. Use the App Password instead of your regular password")
		fmt.Println()
		fmt.Println("Slack Setup:")
		fmt.Println("  1. Create Slack App: https://api.slack.com/apps")
		fmt.Println("  2. Enable Incoming Webhooks")
		fmt.Println("  3. Copy webhook URL and use with -slack-webhook")
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

	// 로그인 모니터링이 활성화된 경우 관련 키워드 자동 추가
	if *loginWatch {
		loginKeywords := []string{"sshd", "sudo", "login", "session", "authentication", "accepted", "failed"}
		for _, keyword := range loginKeywords {
			// 중복 방지
			found := false
			for _, existing := range keywords {
				if strings.ToLower(existing) == strings.ToLower(keyword) {
					found = true
					break
				}
			}
			if !found {
				keywords = append(keywords, keyword)
			}
		}
		fmt.Printf("🔍 Added login keywords: %s\n", strings.Join(loginKeywords, ", "))
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

	// 슬랙 설정
	slackConfig := &SlackConfig{
		WebhookURL: *slackWebhook,
		Channel:    *slackChannel,
		Username:   *slackUsername,
		Enabled:    *slackWebhook != "",
	}

	if slackConfig.Enabled {
		fmt.Printf("💬 Slack alerts enabled\n")
		fmt.Printf("    📡 Webhook: %s\n", slackConfig.WebhookURL[:50]+"...")
		if slackConfig.Channel != "" {
			fmt.Printf("    📺 Channel: %s\n", slackConfig.Channel)
		}
		fmt.Printf("    🤖 Bot Name: %s\n", slackConfig.Username)
	} else {
		fmt.Printf("💬 Slack alerts disabled. Use -slack-webhook to enable.\n")
	}

	if *loginWatch {
		fmt.Printf("👁️  Login monitoring enabled (SSH, sudo, web login detection)\n")
	}
	
	// AI 분석 상태 메시지
	if *aiEnabled {
		fmt.Printf("🤖 AI log analysis enabled\n")
		fmt.Printf("    🔍 Anomaly detection and prediction\n")
		fmt.Printf("    📊 Pattern recognition and threat assessment\n")
		fmt.Printf("    🎯 Supported log types: apache, nginx, mysql, postgresql, application\n")
	} else {
		fmt.Printf("🤖 AI analysis disabled. Use -ai-analysis to enable.\n")
	}
	
	// 시스템 모니터링 상태 메시지
	if *systemEnabled {
		fmt.Printf("🖥️  System monitoring enabled\n")
		fmt.Printf("    📈 CPU, memory, disk, temperature monitoring\n")
		fmt.Printf("    ⚠️  Real-time alerts for system thresholds\n")
		fmt.Printf("    🔄 5-minute monitoring interval\n")
	} else {
		fmt.Printf("🖥️  System monitoring disabled. Use -system-monitor to enable.\n")
	}

	// 테스트 슬랙 전송
	if *testSlack {
		if !slackConfig.Enabled {
			fmt.Println("Error: Slack webhook URL required for test")
			fmt.Println("Please provide -slack-webhook")
			os.Exit(1)
		}

		fmt.Println("Sending test Slack message...")
		
		monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled)
		
		testMsg := SlackMessage{
			Text:      "🧪 *Test Message from Syslog Monitor*",
			IconEmoji: ":test_tube:",
			Username:  slackConfig.Username,
			Attachments: []SlackAttachment{
				{
					Color: "good",
					Title: "Syslog Monitor Test",
					Fields: []SlackField{
						{Title: "Status", Value: "✅ Working", Short: true},
						{Title: "Time", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
						{Title: "Features", Value: "Email alerts, Login monitoring, Error detection", Short: false},
					},
					Timestamp: time.Now().Unix(),
				},
			},
		}

		if err := monitor.sendSlackMessage(testMsg); err != nil {
			fmt.Printf("Test Slack message failed: %v\n", err)
			fmt.Println("\nTroubleshooting:")
			fmt.Println("1. Check your Slack webhook URL")
			fmt.Println("2. Verify webhook permissions")
			fmt.Println("3. Test webhook manually")
			os.Exit(1)
		}

		fmt.Printf("✅ Test Slack message sent successfully!\n")
		return
	}

	// 테스트 이메일 전송
	if *testEmail {
		if !emailConfig.Enabled {
			fmt.Println("Error: Email configuration required for test email")
			fmt.Println("Please provide -email-to and SMTP credentials")
			os.Exit(1)
		}

		fmt.Println("Sending test email...")
		
		monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled)
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
	monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled)
	
	if err := monitor.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
} 