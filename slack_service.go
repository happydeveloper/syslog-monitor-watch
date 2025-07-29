package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SlackService Slack 메시지 전송 서비스
type SlackService struct {
	config *SlackConfig
	logger Logger
}

// NewSlackService 새로운 Slack 서비스 생성
func NewSlackService(config *SlackConfig, logger Logger) *SlackService {
	return &SlackService{
		config: config,
		logger: logger,
	}
}

// SendMessage Slack 메시지 전송
func (ss *SlackService) SendMessage(message SlackMessage) error {
	if !ss.config.Enabled {
		return nil
	}

	// 기본값 설정
	if message.Channel == "" {
		message.Channel = ss.config.Channel
	}
	if message.Username == "" {
		message.Username = DefaultSlackUsername
	}
	if message.IconEmoji == "" {
		message.IconEmoji = DefaultSlackIcon
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %v", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", ss.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// HTTP 클라이언트로 전송
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %v", ErrSlackSendFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	ss.logger.Infof("✅ Slack message sent successfully to channel: %s", message.Channel)
	return nil
}

// CreateLoginAlert 로그인 알림 메시지 생성
func (ss *SlackService) CreateLoginAlert(loginInfo map[string]string, parsed map[string]string) SlackMessage {
	status := loginInfo["status"]
	var color, title, emoji string
	var fields []SlackField

	switch status {
	case "accepted":
		color = SlackColorGood
		title = "✅ SSH Login Successful"
		emoji = ":white_check_mark:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "IP Address", Value: loginInfo["ip"], Short: true},
			{Title: "Method", Value: loginInfo["method"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
		}
	case "failed":
		color = SlackColorDanger
		title = "❌ SSH Login Failed"
		emoji = ":x:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "IP Address", Value: loginInfo["ip"], Short: true},
			{Title: "Method", Value: loginInfo["method"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
		}
	case "sudo":
		color = SlackColorWarning
		title = "⚡ Sudo Command Executed"
		emoji = ":zap:"
		fields = []SlackField{
			{Title: "User", Value: loginInfo["user"], Short: true},
			{Title: "Host", Value: parsed["host"], Short: true},
			{Title: "Command", Value: loginInfo["command"], Short: false},
		}
	case "web_login":
		color = SlackColorGood
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
		Text:        fmt.Sprintf("%s *%s*", emoji, title),
		IconEmoji:   ":robot_face:",
		Username:    DefaultSlackUsername,
		Attachments: []SlackAttachment{attachment},
	}
}

// CreateAIAlert AI 분석 결과 알림 메시지 생성
func (ss *SlackService) CreateAIAlert(aiResult *AIAnalysisResult) SlackMessage {
	color := SlackColorWarning
	if aiResult.AnomalyScore >= HighThreatThreshold {
		color = SlackColorDanger
	}

	fields := []SlackField{
		{Title: "위협 레벨", Value: aiResult.ThreatLevel, Short: true},
		{Title: "이상 점수", Value: fmt.Sprintf("%.1f/%.0f", aiResult.AnomalyScore, MaxAnomalyScore), Short: true},
		{Title: "신뢰도", Value: fmt.Sprintf("%.0f%%", aiResult.Confidence*100), Short: true},
		{Title: "컴퓨터명", Value: aiResult.SystemInfo.ComputerName, Short: true},
	}

	// 내부 IP 정보 추가
	if len(aiResult.SystemInfo.InternalIPs) > 0 {
		fields = append(fields, SlackField{
			Title: "🏠 내부 IP",
			Value: strings.Join(aiResult.SystemInfo.InternalIPs, ", "),
			Short: true,
		})
	}

	// 외부 IP 정보 추가
	if len(aiResult.SystemInfo.ExternalIPs) > 0 {
		fields = append(fields, SlackField{
			Title: "🌐 외부 IP",
			Value: strings.Join(aiResult.SystemInfo.ExternalIPs, ", "),
			Short: true,
		})
	}

	// ASN 정보 추가
	if len(aiResult.SystemInfo.ASNData) > 0 {
		asnText := ""
		for _, asn := range aiResult.SystemInfo.ASNData {
			asnText += fmt.Sprintf("📍 %s\n🏢 %s\n🌍 %s\n🔢 %s\n\n",
				asn.IP, asn.Organization, asn.Country, asn.ASN)
		}
		fields = append(fields, SlackField{Title: "🔍 ASN 정보", Value: asnText, Short: false})
	}

	// 영향받는 시스템
	if len(aiResult.AffectedSystems) > 0 {
		fields = append(fields, SlackField{
			Title: "🎯 영향 시스템",
			Value: strings.Join(aiResult.AffectedSystems, ", "),
			Short: false,
		})
	}

	// 예측 정보
	if len(aiResult.Predictions) > 0 {
		predictionText := ""
		for _, prediction := range aiResult.Predictions {
			predictionText += fmt.Sprintf("⚡ %s (%.0f%%, %s)\n💥 %s\n\n",
				prediction.Event, prediction.Probability*100, prediction.TimeFrame, prediction.Impact)
		}
		fields = append(fields, SlackField{Title: "🔮 위험 예측", Value: predictionText, Short: false})
	}

	// 권장사항
	if len(aiResult.Recommendations) > 0 {
		recommendationText := ""
		for _, recommendation := range aiResult.Recommendations {
			recommendationText += fmt.Sprintf("• %s\n", recommendation)
		}
		fields = append(fields, SlackField{Title: "💡 권장사항", Value: recommendationText, Short: false})
	}

	slackMsg := SlackMessage{
		Text:      fmt.Sprintf("🚨 *보안 이상 탐지 알람* %s", aiResult.ThreatLevel),
		IconEmoji: DefaultSlackIcon,
		Username:  DefaultSlackUsername,
		Attachments: []SlackAttachment{
			{
				Color:     color,
				Title:     "🤖 AI 분석 결과",
				Fields:    fields,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return slackMsg
}

// CreateSystemAlert 시스템 알림 메시지 생성
func (ss *SlackService) CreateSystemAlert(alert SystemAlert) SlackMessage {
	var color string
	var emoji string

	switch alert.Level {
	case "CRITICAL":
		color = SlackColorDanger
		emoji = ":rotating_light:"
	case "WARNING":
		color = SlackColorWarning
		emoji = ":warning:"
	default:
		color = SlackColorGood
		emoji = ":information_source:"
	}

	fields := []SlackField{
		{Title: "메트릭", Value: alert.Type, Short: true},
		{Title: "현재 값", Value: fmt.Sprintf("%.2f", alert.Value), Short: true},
		{Title: "임계값", Value: fmt.Sprintf("%.2f", alert.Threshold), Short: true},
		{Title: "심각도", Value: alert.Level, Short: true},
	}

	attachment := SlackAttachment{
		Color:     color,
		Title:     fmt.Sprintf("%s 시스템 알림: %s", emoji, alert.Type),
		Text:      alert.Message,
		Fields:    fields,
		Timestamp: alert.Timestamp.Unix(),
	}

	return SlackMessage{
		Text:        fmt.Sprintf("%s *시스템 알림*: %s", emoji, alert.Type),
		IconEmoji:   ":robot_face:",
		Username:    DefaultSlackUsername,
		Attachments: []SlackAttachment{attachment},
	}
}

// SendTestMessage 테스트 메시지 전송
func (ss *SlackService) SendTestMessage() error {
	message := SlackMessage{
		Text:      fmt.Sprintf("🧪 *%s 테스트 메시지*", AppName),
		IconEmoji: ":test_tube:",
		Username:  DefaultSlackUsername,
		Attachments: []SlackAttachment{
			{
				Color: SlackColorGood,
				Title: "✅ Slack 연동 테스트",
				Text:  fmt.Sprintf("%s v%s Slack 연동이 정상적으로 작동합니다!", AppName, AppVersion),
				Fields: []SlackField{
					{Title: "채널", Value: ss.config.Channel, Short: true},
					{Title: "테스트 시간", Value: time.Now().Format("2006-01-02 15:04:05"), Short: true},
				},
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return ss.SendMessage(message)
}

// IsEnabled Slack 서비스 활성화 여부 확인
func (ss *SlackService) IsEnabled() bool {
	return ss.config.Enabled
}

// GetChannel 설정된 채널 반환
func (ss *SlackService) GetChannel() string {
	return ss.config.Channel
} 