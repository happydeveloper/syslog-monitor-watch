/*
AI-Powered Syslog Monitor
========================

고급 시스템 로그 모니터링 및 분석 도구

주요 기능:
- 실시간 syslog 모니터링
- AI 기반 이상 탐지 및 예측
- 시스템 리소스 모니터링
- 로그인 패턴 감지
- 이메일/Slack 알림
- 다양한 로그 포맷 지원 (Apache, Nginx, MySQL, PostgreSQL)

작성자: Lambda-X AI Team
버전: 2.0.0
*/
package main

import (
	"flag"     // 명령줄 인수 파싱
	"fmt"      // 형식화된 I/O
	"os"       // 운영체제 인터페이스
	"os/exec"  // 외부 명령 실행
	"os/signal" // 시그널 처리
	"path/filepath" // 파일 경로 처리
	"regexp"   // 정규식
	"runtime"  // Go 런타임 정보
	"strconv"  // 문자열-숫자 변환
	"strings"  // 문자열 처리
	"syscall"  // 시스템 호출
	"time"     // 시간 처리

	"github.com/hpcloud/tail"     // 파일 tail 기능
	"github.com/sirupsen/logrus"  // 구조화된 로깅
)

// 전역 변수들
var (
	// 설정 서비스
	configService *ConfigService
	geminiService *GeminiService
)

// EmailConfig 이메일 서비스 설정 구조체
// Gmail SMTP 서버 설정 및 다중 수신자 지원
type EmailConfig struct {
	SMTPServer   string   // SMTP 서버 주소 (예: smtp.gmail.com)
	SMTPPort     string   // SMTP 포트 번호 (587: STARTTLS, 465: SSL/TLS)
	Username     string   // SMTP 인증 사용자명 (Gmail의 경우 이메일 주소)
	Password     string   // SMTP 인증 비밀번호 (Gmail의 경우 앱 패스워드)
	To           []string // 수신자 이메일 주소 목록 (여러 명에게 동시 전송 가능)
	From         string   // 발신자 이메일 주소
	Enabled      bool     // 이메일 서비스 활성화 여부
}

// SlackConfig Slack 웹훅 서비스 설정 구조체
// Slack Incoming Webhooks API를 통한 메시지 전송 설정
type SlackConfig struct {
	WebhookURL string // Slack Incoming Webhook URL (https://hooks.slack.com/...)
	Channel    string // 메시지를 전송할 Slack 채널명 (예: #alerts, #security)
	Username   string // 봇의 표시 이름 (Slack에서 보이는 발신자명)
	Enabled    bool   // Slack 서비스 활성화 여부
}

// SlackMessage Slack API 메시지 구조체
// Slack Incoming Webhooks API 스펙에 맞는 메시지 포맷
type SlackMessage struct {
	Channel     string             `json:"channel,omitempty"`     // 대상 채널 (#general, @username)
	Username    string             `json:"username,omitempty"`    // 봇 사용자명
	Text        string             `json:"text,omitempty"`        // 메인 메시지 텍스트
	IconEmoji   string             `json:"icon_emoji,omitempty"`  // 봇 아이콘 이모지 (:warning:, :robot_face:)
	Attachments []SlackAttachment  `json:"attachments,omitempty"` // 첨부된 상세 정보 블록들
}

// SlackAttachment Slack 메시지의 첨부 블록 구조체
// 메시지에 색상, 필드, 타임스탬프 등의 상세 정보를 추가
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`     // 좌측 세로 바 색상 (good, warning, danger, #hex)
	Title     string       `json:"title,omitempty"`     // 첨부 블록의 제목
	Text      string       `json:"text,omitempty"`      // 첨부 블록의 본문 텍스트
	Fields    []SlackField `json:"fields,omitempty"`    // 구조화된 필드 목록 (키-값 쌍)
	Timestamp int64        `json:"ts,omitempty"`        // Unix 타임스탬프 (메시지 하단에 시간 표시)
}

// SlackField Slack 첨부 블록 내의 개별 필드 구조체
// 키-값 쌍으로 구조화된 정보를 표시
type SlackField struct {
	Title string `json:"title"` // 필드 제목/키 (예: "사용자", "IP 주소")
	Value string `json:"value"` // 필드 값 (예: "admin", "192.168.1.100")
	Short bool   `json:"short"` // 한 줄에 여러 필드 표시 여부 (true: 2열, false: 1열)
}

// SyslogMonitor 메인 시스템 로그 모니터링 구조체
// 실시간 로그 감시, AI 분석, 알림 전송 등의 모든 기능을 통합 관리
type SyslogMonitor struct {
	logFile       string            // 모니터링할 로그 파일 경로 (/var/log/syslog, /var/log/system.log 등)
	filters       []string          // 제외할 로그 패턴의 정규식 목록 (노이즈 필터링용)
	keywords      []string          // 포함할 키워드 목록 (특정 패턴만 감시)
	outputFile    string            // 필터링된 로그 출력 파일 경로 (빈 문자열이면 stdout)
	logger        *logrus.Logger    // 구조화된 로깅을 위한 logrus 인스턴스
	emailService  *EmailService     // 이메일 알림 서비스 (Gmail SMTP 지원)
	slackService  *SlackService     // Slack 웹훅 알림 서비스
	loginDetector *LoginDetector    // SSH/sudo 등 로그인 패턴 감지 서비스
	aiAnalyzer    *AIAnalyzer       // AI 기반 이상 탐지 및 예측 분석 엔진
	systemMonitor *SystemMonitor    // CPU/메모리/디스크 등 시스템 리소스 모니터링
	logParser     *LogParserManager // 다양한 로그 포맷 파싱 (Apache, Nginx, MySQL 등)
	aiEnabled     bool              // AI 분석 기능 활성화 여부
	systemEnabled bool              // 시스템 모니터링 기능 활성화 여부
	loginWatch    bool              // 로그인 감지 기능 활성화 여부
	
	// 주기적 보고서 관련 필드
	periodicReport   bool          // 주기적 보고서 기능 활성화 여부
	reportInterval   time.Duration // 보고서 전송 간격
	lastReportTime   time.Time     // 마지막 보고서 전송 시간
	geoMapper        *GeoMapper    // 지리정보 매핑 서비스
}

// NewSyslogMonitor SyslogMonitor 인스턴스 생성자
// 모든 서비스 컴포넌트를 초기화하고 설정에 따라 기능을 활성화/비활성화
//
// 매개변수:
//   - logFile: 모니터링할 로그 파일 경로
//   - outputFile: 필터링된 로그 출력 파일 경로 (""이면 stdout)
//   - filters: 제외할 로그 패턴 정규식 배열
//   - keywords: 포함할 키워드 배열
//   - emailConfig: 이메일 알림 설정 (nil이면 비활성화)
//   - slackConfig: Slack 알림 설정 (nil이면 비활성화)
//   - aiEnabled: AI 분석 기능 활성화 여부
//   - systemEnabled: 시스템 모니터링 활성화 여부
//   - loginWatch: 로그인 감지 기능 활성화 여부
//
// 반환값:
//   - *SyslogMonitor: 초기화된 모니터 인스턴스
func NewSyslogMonitor(logFile, outputFile string, filters, keywords []string, emailConfig *EmailConfig, slackConfig *SlackConfig, aiEnabled, systemEnabled, loginWatch bool, alertInterval, reportInterval int, periodicReport bool) *SyslogMonitor {
	// 구조화된 로깅 설정
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,                   // 전체 타임스탬프 표시
		TimestampFormat: "2006-01-02 15:04:05", // 한국 표준 시간 포맷
	})

	// 로그 출력 파일 설정 (지정된 경우)
	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file) // 파일로 로그 출력 리다이렉션
		}
	}

	// 각 서비스 컴포넌트 조건부 초기화
	var emailService *EmailService   // 이메일 알림 서비스
	var slackService *SlackService   // Slack 웹훅 서비스
	var loginDetector *LoginDetector // 로그인 패턴 감지 서비스
	var aiAnalyzer *AIAnalyzer       // AI 이상 탐지 분석기
	var systemMonitor *SystemMonitor // 시스템 리소스 모니터

	// 이메일 서비스 초기화 (설정이 존재하고 활성화된 경우)
	if emailConfig != nil && emailConfig.Enabled {
		emailService = NewEmailService(emailConfig, logger)
	}

	// Slack 서비스 초기화 (설정이 존재하고 활성화된 경우)
	if slackConfig != nil && slackConfig.Enabled {
		slackService = NewSlackService(slackConfig, logger)
	}

	// 로그인 감지 서비스 초기화 (loginWatch 플래그가 true인 경우)
	if loginWatch {
		loginDetector = NewLoginDetector(logger)
	}

	// AI 분석 엔진 초기화 (aiEnabled 플래그가 true인 경우)
	if aiEnabled {
		aiAnalyzer = NewAIAnalyzer()
	}

	// 시스템 모니터링 서비스 초기화 (systemEnabled 플래그가 true인 경우)
	if systemEnabled {
		// 정기 보고서 간격 계산
		reportIntervalDuration := time.Duration(reportInterval) * time.Minute
		
		// 알림 서비스가 포함된 시스템 모니터 생성
		systemMonitor = NewSystemMonitorWithNotifications(
			DefaultMonitoringInterval, // 5분 간격 모니터링
			periodicReport,            // 정기 보고서 활성화 여부
			reportIntervalDuration,    // 보고서 간격
			emailService,              // 이메일 서비스
			slackService,              // Slack 서비스
		)
	}

	// 지리정보 매핑 서비스 초기화
	geoMapper := NewGeoMapper(logger)

	// 로그인 감지기에 시스템 모니터 연결 (리소스 정보 수집용)
	if loginDetector != nil && systemMonitor != nil {
		loginDetector.SetSystemMonitor(systemMonitor)
	}
	
	// 알림 간격 설정 적용
	if loginDetector != nil {
		alertDuration := time.Duration(alertInterval) * time.Minute
		loginDetector.SetAlertInterval(alertDuration)
		logger.Infof("📝 Login alert interval set to: %d minutes", alertInterval)
	}

	// SyslogMonitor 인스턴스 생성 및 반환
	return &SyslogMonitor{
		logFile:       logFile,                   // 모니터링 대상 로그 파일
		filters:       filters,                   // 필터링 패턴 목록
		keywords:      keywords,                  // 키워드 목록
		outputFile:    outputFile,                // 출력 파일 경로
		logger:        logger,                    // 로깅 인스턴스
		emailService:  emailService,              // 이메일 서비스 (nil 가능)
		slackService:  slackService,              // Slack 서비스 (nil 가능)
		loginDetector: loginDetector,             // 로그인 감지 서비스 (nil 가능)
		aiAnalyzer:    aiAnalyzer,                // AI 분석 엔진 (nil 가능)
		systemMonitor: systemMonitor,             // 시스템 모니터 (nil 가능)
		logParser:     NewLogParserManager(),     // 다중 로그 파서 관리자
		aiEnabled:     aiEnabled,                 // AI 기능 활성화 플래그
		systemEnabled: systemEnabled,             // 시스템 모니터링 활성화 플래그
		loginWatch:    loginWatch,                // 로그인 감지 활성화 플래그
		periodicReport: periodicReport,       // 주기적 보고서 활성화 플래그
		reportInterval: time.Duration(reportInterval) * time.Minute, // 보고서 간격
		lastReportTime: time.Now(),                // 마지막 보고서 시간
		geoMapper:     geoMapper,                  // 지리정보 매핑 서비스
	}
}

// shouldFilter 로그 라인이 필터링 패턴에 매치되는지 확인
// 설정된 정규식 필터 목록과 비교하여 제외할 로그인지 판단
//
// 매개변수:
//   - line: 검사할 로그 라인 문자열
//
// 반환값:
//   - bool: true이면 필터링 대상 (제외), false이면 통과
//
// 동작 원리:
//   1. 필터가 설정되지 않은 경우 모든 로그 통과
//   2. 각 필터 패턴을 순차적으로 검사
//   3. 하나라도 매치되면 즉시 true 반환 (필터링)
func (sm *SyslogMonitor) shouldFilter(line string) bool {
	if len(sm.filters) == 0 {
		return false // 필터가 없으면 모든 로그 통과
	}

	// 각 필터 패턴과 비교
	for _, filter := range sm.filters {
		matched, _ := regexp.MatchString(filter, line)
		if matched {
			return true // 필터 패턴에 매치되면 제외
		}
	}
	return false // 어떤 필터에도 매치되지 않으면 통과
}

// containsKeyword 로그 라인에 지정된 키워드가 포함되어 있는지 확인
// 대소문자를 구분하지 않고 키워드 매칭을 수행
//
// 매개변수:
//   - line: 검사할 로그 라인 문자열
//
// 반환값:
//   - bool: true이면 키워드 포함 (감시 대상), false이면 제외
//
// 동작 원리:
//   1. 키워드가 설정되지 않은 경우 모든 로그 포함
//   2. 로그 라인과 키워드를 소문자로 변환하여 비교
//   3. 하나라도 포함되면 즉시 true 반환
func (sm *SyslogMonitor) containsKeyword(line string) bool {
	if len(sm.keywords) == 0 {
		return true // 키워드가 없으면 모든 라인을 포함
	}

	lowLine := strings.ToLower(line) // 대소문자 무관 비교를 위한 소문자 변환
	for _, keyword := range sm.keywords {
		if strings.Contains(lowLine, strings.ToLower(keyword)) {
			return true // 키워드가 포함되면 감시 대상
		}
	}
	return false // 어떤 키워드도 포함되지 않으면 제외
}

// parseSyslogLine syslog 포맷의 로그 라인을 파싱하여 구조화된 데이터로 변환
// 표준 syslog 형식 (month day time host service: message)을 파싱
//
// 매개변수:
//   - line: 파싱할 원본 로그 라인
//
// 반환값:
//   - map[string]string: 파싱된 필드들의 키-값 맵
//     - "raw": 원본 로그 라인
//     - "timestamp": 현재 타임스탬프
//     - "month": 월 정보 (Jan, Feb 등)
//     - "day": 일 정보
//     - "time": 시간 정보 (HH:MM:SS)
//     - "host": 호스트명
//     - "service": 서비스명
//     - "message": 메시지 내용
//
// 예시 입력: "Jan 15 10:30:45 myserver sshd[1234]: Connection accepted"
// 예시 출력: {"month": "Jan", "day": "15", "time": "10:30:45", "host": "myserver", "service": "sshd[1234]:", "message": "Connection accepted"}
func (sm *SyslogMonitor) parseSyslogLine(line string) map[string]string {
	result := make(map[string]string)
	result["raw"] = line                                         // 원본 로그 보존
	result["timestamp"] = time.Now().Format("2006-01-02 15:04:05") // 처리 시점 타임스탬프

	// 기본적인 syslog 파싱 (공백으로 분리된 필드들)
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		result["month"] = parts[0] // 월 (Jan, Feb, ...)
		result["day"] = parts[1]   // 일 (1-31)
		result["time"] = parts[2]  // 시간 (HH:MM:SS)
		
		if len(parts) >= 4 {
			result["host"] = parts[3] // 호스트명
			
			if len(parts) >= 5 {
				result["service"] = parts[4]                    // 서비스명 (예: sshd[1234]:)
				result["message"] = strings.Join(parts[5:], " ") // 나머지를 메시지로 결합
			}
		}
	}

	return result
}

// 이메일 전송 기능은 EmailService로 이동됨

// Slack 전송 기능은 SlackService로 이동됨

// 로그인 감지 기능은 LoginDetector로 이동됨

// 모든 이메일 관련 함수들은 EmailService로 이동됨

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

	// 로그인 패턴 감지 (LoginDetector 서비스 사용)
	if sm.loginWatch && sm.loginDetector != nil {
		if isLogin, loginInfo := sm.loginDetector.DetectLoginPattern(line); isLogin {
			// 기본 로그 (항상 기록)
			sm.logger.WithFields(logrus.Fields{
				"level":        "LOGIN",
				"user":         loginInfo.User,
				"host":         parsed["host"],
				"status":       loginInfo.Status,
				"ip":           loginInfo.IP,
				"cpu_usage":    fmt.Sprintf("%.1f%%", loginInfo.SystemInfo.CPU.UsagePercent),
				"memory_usage": fmt.Sprintf("%.1f%%", loginInfo.SystemInfo.Memory.UsagePercent),
				"should_alert": loginInfo.ShouldAlert,
			}).Infof("🔐 User activity detected: %s from %s (Alert: %t)", 
				loginInfo.Status, loginInfo.IP, loginInfo.ShouldAlert)

			// 10분 간격 제한에 따른 선택적 알림 전송
			if loginInfo.ShouldAlert {
				// 이메일 로그인 알림 전송 (EmailService 사용)
				if sm.emailService != nil {
					sm.logger.Infof("📧 Sending login alert email (interval check passed)")
					sm.sendLoginEmailAlert(loginInfo, parsed)
				}

				// Slack 로그인 알림 전송 (SlackService 사용)
				if sm.slackService != nil {
					slackMsg := sm.slackService.CreateLoginAlert(loginInfo.ToMap(), parsed)
					sm.logger.Infof("💬 Sending login notification to Slack: %s (interval check passed)", loginInfo.User)
					go func() {
						if err := sm.slackService.SendMessage(slackMsg); err != nil {
							sm.logger.Errorf("❌ Failed to send Slack login notification: %v", err)
						} else {
							sm.logger.Infof("✅ Slack login notification sent successfully")
						}
					}()
				}
			} else {
				// 알림 제한된 경우 로그만 기록
				sm.logger.Infof("⏰ Login alert skipped due to interval limit (10min rule)")
			}
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
		
		// 에러 발생 시 이메일 알림 전송 (EmailService 사용)
		if sm.emailService != nil {
			subject := fmt.Sprintf("[%s ERROR] %s - %s", AppName, parsed["host"], parsed["service"])
			body := fmt.Sprintf("시간: %s\n호스트: %s\n서비스: %s\n메시지: %s\n원본 로그: %s", 
				parsed["timestamp"], parsed["host"], parsed["service"], parsed["message"], line)
			
			sm.logger.Infof("📧 Sending ERROR alert to: %s", sm.emailService.GetRecipientsList())
			go func() {
				if err := sm.emailService.SendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send email alert: %v", err)
				}
			}()
		}

		// 에러 시 Slack 알림도 전송 (SlackService 사용)
		if sm.slackService != nil {
			slackMsg := SlackMessage{
				Text:      fmt.Sprintf("🔴 *ERROR Alert*"),
				IconEmoji: ":rotating_light:",
				Username:  DefaultSlackUsername,
				Attachments: []SlackAttachment{
					{
						Color: SlackColorDanger,
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
				if err := sm.slackService.SendMessage(slackMsg); err != nil {
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
		
		// 크리티컬 에러 발생 시 이메일 알림 전송 (EmailService 사용)
		if sm.emailService != nil {
			subject := fmt.Sprintf("[%s CRITICAL] %s - %s", AppName, parsed["host"], parsed["service"])
			body := fmt.Sprintf("🚨 CRITICAL ALERT 🚨\n\n시간: %s\n호스트: %s\n서비스: %s\n메시지: %s\n원본 로그: %s", 
				parsed["timestamp"], parsed["host"], parsed["service"], parsed["message"], line)
			
			sm.logger.Warnf("🚨 Sending CRITICAL alert to: %s", sm.emailService.GetRecipientsList())
			go func() {
				if err := sm.emailService.SendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send critical email alert: %v", err)
				}
			}()
		}

		// 크리티컬 에러 시 Slack 긴급 알림 (SlackService 사용)
		if sm.slackService != nil {
			slackMsg := SlackMessage{
				Text:      fmt.Sprintf("🚨 *CRITICAL ALERT* 🚨"),
				IconEmoji: DefaultSlackIcon,
				Username:  DefaultSlackUsername,
				Attachments: []SlackAttachment{
					{
						Color: SlackColorDanger,
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
				if err := sm.slackService.SendMessage(slackMsg); err != nil {
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

	// 주기적 시스템 상태 보고서 시작
	if sm.periodicReport && sm.systemMonitor != nil {
		sm.logger.Infof("📊 주기적 시스템 상태 보고서가 활성화되었습니다 (간격: %v)", sm.reportInterval)
		go sm.sendPeriodicSystemReports()
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

// sendLoginEmailAlert 로그인 알림 이메일 전송 (시스템 리소스 정보 포함)
func (sm *SyslogMonitor) sendLoginEmailAlert(loginInfo *LoginInfo, parsed map[string]string) {
	// 이메일 제목 생성 (상태별 구분)
	var subject string
	var statusEmoji string
	
	switch loginInfo.Status {
	case "accepted":
		statusEmoji = "✅"
		subject = fmt.Sprintf("[%s LOGIN SUCCESS] %s logged in from %s", AppName, loginInfo.User, loginInfo.IP)
	case "failed":
		statusEmoji = "❌"
		subject = fmt.Sprintf("[%s LOGIN FAILED] Failed login attempt for %s from %s", AppName, loginInfo.User, loginInfo.IP)
	case "sudo":
		statusEmoji = "⚡"
		subject = fmt.Sprintf("[%s SUDO COMMAND] %s executed sudo command", AppName, loginInfo.User)
	case "web_login":
		statusEmoji = "🌐"
		subject = fmt.Sprintf("[%s WEB LOGIN] %s logged in via web from %s", AppName, loginInfo.User, loginInfo.IP)
	default:
		statusEmoji = "🔐"
		subject = fmt.Sprintf("[%s LOGIN ACTIVITY] User activity detected: %s", AppName, loginInfo.Status)
	}

	// 이메일 본문 생성
	body := fmt.Sprintf(`%s 로그인 활동 감지 알림
==============================

🕐 감지 시간: %s
👤 사용자: %s
📍 상태: %s %s
🌐 IP 주소: %s
🔑 인증 방법: %s
🖥️  호스트: %s

🖥️  시스템 리소스 정보 (로그인 시점):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💻 CPU 사용률: %.1f%% (코어: %d개)
  ├ 사용자: %.1f%%
  ├ 시스템: %.1f%%
  └ 대기: %.1f%%

🧠 메모리 사용률: %.1f%%
  ├ 총 메모리: %.1f GB
  ├ 사용 중: %.1f GB
  ├ 사용 가능: %.1f GB
  └ 스왑 사용: %.1f MB

🌡️  시스템 온도: %.1f°C
⚖️  로드 평균: %.2f (1분), %.2f (5분), %.2f (15분)
`,
		statusEmoji,
		loginInfo.Timestamp.Format("2006-01-02 15:04:05"),
		loginInfo.User,
		loginInfo.Status,
		statusEmoji,
		loginInfo.IP,
		loginInfo.Method,
		parsed["host"],
		loginInfo.SystemInfo.CPU.UsagePercent,
		loginInfo.SystemInfo.CPU.Cores,
		loginInfo.SystemInfo.CPU.UserPercent,
		loginInfo.SystemInfo.CPU.SystemPercent,
		loginInfo.SystemInfo.CPU.IdlePercent,
		loginInfo.SystemInfo.Memory.UsagePercent,
		loginInfo.SystemInfo.Memory.TotalMB/1024,
		loginInfo.SystemInfo.Memory.UsedMB/1024,
		loginInfo.SystemInfo.Memory.AvailableMB/1024,
		loginInfo.SystemInfo.Memory.SwapUsedMB,
		loginInfo.SystemInfo.Temperature.CPUTemp,
		loginInfo.SystemInfo.LoadAverage.Load1Min,
		loginInfo.SystemInfo.LoadAverage.Load5Min,
		loginInfo.SystemInfo.LoadAverage.Load15Min,
	)

	// IP 위치 정보 추가
	if loginInfo.IPDetails != nil {
		body += fmt.Sprintf(`
🌍 IP 위치 정보:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📍 IP 주소: %s
🏴 국가: %s
🏙️  도시: %s, %s
🏢 조직/ISP: %s
🔢 ASN: %s
🔒 IP 유형: %s
⚠️  위험도: %s
`,
			loginInfo.IPDetails.IP,
			loginInfo.IPDetails.Country,
			loginInfo.IPDetails.City,
			loginInfo.IPDetails.Region,
			loginInfo.IPDetails.Organization,
			loginInfo.IPDetails.ASN,
			func() string { if loginInfo.IPDetails.IsPrivate { return "사설 IP" } else { return "공인 IP" } }(),
			loginInfo.IPDetails.Threat,
		)
	}

	// Sudo 명령어 정보 추가
	if loginInfo.Command != "" {
		body += fmt.Sprintf(`
⚡ 실행된 명령어:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s
`, loginInfo.Command)
	}

	// 디스크 사용량 정보 추가 (모든 주요 디스크)
	if len(loginInfo.SystemInfo.Disk) > 0 {
		body += `
💾 디스크 사용량 상세정보:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`
		var totalUsed, totalSize float64
		for _, disk := range loginInfo.SystemInfo.Disk {
			// 모든 실제 디스크 표시 (tmpfs, proc 등 가상 파일시스템 제외)
			if disk.TotalGB > 0 && !strings.Contains(disk.Device, "tmpfs") && 
			   !strings.Contains(disk.Device, "proc") && !strings.Contains(disk.Device, "sys") {
				
				// 사용률에 따른 상태 이모지
				var statusEmoji string
				if disk.UsagePercent >= 90 {
					statusEmoji = "🔴" // 위험
				} else if disk.UsagePercent >= 75 {
					statusEmoji = "🟡" // 경고
				} else {
					statusEmoji = "🟢" // 정상
				}
				
				body += fmt.Sprintf("  %s 📁 %s (%s)\n", statusEmoji, disk.MountPoint, disk.Device)
				body += fmt.Sprintf("     ├ 사용률: %.1f%% (%.1fGB / %.1fGB)\n", 
					disk.UsagePercent, disk.UsedGB, disk.TotalGB)
				body += fmt.Sprintf("     ├ 남은공간: %.1f GB (%.1f%%)\n", 
					disk.FreeGB, 100-disk.UsagePercent)
				if disk.InodeUsagePercent > 0 {
					body += fmt.Sprintf("     └ inode 사용률: %.1f%%\n", disk.InodeUsagePercent)
				} else {
					body += fmt.Sprintf("     └ 여유공간: %.1f GB\n", disk.FreeGB)
				}
				body += "\n"
				
				totalUsed += disk.UsedGB
				totalSize += disk.TotalGB
			}
		}
		
		// 전체 디스크 요약
		if totalSize > 0 {
			totalFree := totalSize - totalUsed
			totalUsagePercent := (totalUsed / totalSize) * 100
			body += fmt.Sprintf("📊 전체 디스크 요약:\n")
			body += fmt.Sprintf("   ├ 총 용량: %.1f GB\n", totalSize)
			body += fmt.Sprintf("   ├ 사용량: %.1f GB (%.1f%%)\n", totalUsed, totalUsagePercent)
			body += fmt.Sprintf("   └ 여유공간: %.1f GB (%.1f%%)\n", totalFree, 100-totalUsagePercent)
		}
	}

	// 보안 권장사항
	body += `
🛡️  보안 권장사항:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• 알 수 없는 IP에서의 로그인 시도인지 확인하세요
• 시스템 리소스 사용량이 평소보다 높은지 확인하세요
• 비정상적인 시간대 로그인은 주의가 필요합니다
• 실패한 로그인 시도가 반복되면 IP 차단을 고려하세요
• 정기적으로 로그인 기록을 검토하세요

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🤖 AI-Powered Syslog Monitor v2.0.0
Lambda-X AI Security Team
`

	// 이메일 전송 (비동기)
	sm.logger.Infof("📧 Sending login alert email to: %s", sm.emailService.GetRecipientsList())
	go func() {
		if err := sm.emailService.SendEmail(subject, body); err != nil {
			sm.logger.Errorf("❌ Failed to send login alert email: %v", err)
		} else {
			sm.logger.Infof("✅ Login alert email sent successfully")
		}
	}()
}

// sendAIAlert AI 분석 결과 알림 전송 (리팩토링된 버전)
func (sm *SyslogMonitor) sendAIAlert(aiResult *AIAnalysisResult, parsedLog *ParsedLog) {
	// 이메일 알림 (EmailService 사용)
	if sm.emailService != nil {
		subject := fmt.Sprintf("[%s %s] %s", AppName, aiResult.ThreatLevel, "이상 징후 감지")
		
		body := fmt.Sprintf(`🚨 보안 이상 탐지 알람
======================
⚠️  위협 레벨: %s
📊 이상 점수: %.1f/%.0f
🕐 탐지 시간: %s

🖥️  시스템 정보:
  📍 컴퓨터명: %s
  🏠 내부 IP: %s
  🌐 외부 IP: %s

`,
			aiResult.ThreatLevel,
			aiResult.AnomalyScore,
			MaxAnomalyScore,
			aiResult.Timestamp.Format("2006-01-02 15:04:05"),
			aiResult.SystemInfo.ComputerName,
			strings.Join(aiResult.SystemInfo.InternalIPs, ", "),
			strings.Join(aiResult.SystemInfo.ExternalIPs, ", "),
		)

		// ASN 정보 추가
		if len(aiResult.SystemInfo.ASNData) > 0 {
			body += "🔍 ASN 정보:\n"
			for _, asn := range aiResult.SystemInfo.ASNData {
				body += fmt.Sprintf("  📍 %s\n", asn.IP)
				body += fmt.Sprintf("    🏢 조직: %s\n", asn.Organization)
				body += fmt.Sprintf("    🌍 국가: %s, %s, %s\n", asn.Country, asn.Region, asn.City)
				body += fmt.Sprintf("    🔢 ASN: %s\n", asn.ASN)
				body += "\n"
			}
		}

		// 로그 정보
		if parsedLog != nil {
			body += fmt.Sprintf(`
📋 로그 정보:
  📝 레벨: %s
  🏷️  타입: %s
  💬 메시지: %s
  📄 원본: %s

`,
				parsedLog.Level,
				parsedLog.LogType,
				parsedLog.Message,
				parsedLog.RawLog,
			)
		}

		// 예측 결과
		if len(aiResult.Predictions) > 0 {
			body += "🔮 위험 예측:\n"
			for _, prediction := range aiResult.Predictions {
				body += fmt.Sprintf("  ⚡ %s (확률: %.0f%%, %s)\n", 
					prediction.Event, prediction.Probability*100, prediction.TimeFrame)
				body += fmt.Sprintf("    💥 영향: %s\n", prediction.Impact)
			}
			body += "\n"
		}

		// 권장사항
		if len(aiResult.Recommendations) > 0 {
			body += "💡 권장사항:\n"
			for _, recommendation := range aiResult.Recommendations {
				body += fmt.Sprintf("  • %s\n", recommendation)
			}
			body += "\n"
		}

		// 영향받는 시스템
		if len(aiResult.AffectedSystems) > 0 {
			body += fmt.Sprintf("🎯 영향받는 시스템: %s\n", 
				strings.Join(aiResult.AffectedSystems, ", "))
		}

		body += fmt.Sprintf("🎯 신뢰도: %.0f%%\n", aiResult.Confidence*100)
		
		// 전문가 진단 정보 추가
		body += fmt.Sprintf(`
👨‍💼 전문가 진단 결과
====================
🏥 전체 시스템 건강도: %s
📊 성능 점수: %.1f/100

🖥️  서버 전문가 진단:
  🏥 서버 건강도: %s
  📊 성능 점수: %.1f/100
  🔒 보안 상태: %s
  🌐 네트워크 건강도: %s
  ⚠️  위험도: %s

💻 컴퓨터 전문가 진단:
  🔧 하드웨어 건강도: %s
  💾 소프트웨어 상태: %s
  ⚖️  시스템 안정성: %s
  📈 리소스 사용량: %s
  🔧 유지보수 필요: %s

🚨 긴급 이슈:
%s

🔧 유지보수 팁:
%s
`,
			aiResult.ExpertDiagnosis.OverallHealth,
			aiResult.ExpertDiagnosis.PerformanceScore,
			aiResult.ExpertDiagnosis.ServerExpert.ServerHealth,
			aiResult.ExpertDiagnosis.ServerExpert.PerformanceScore,
			aiResult.ExpertDiagnosis.ServerExpert.SecurityStatus,
			aiResult.ExpertDiagnosis.ServerExpert.NetworkHealth,
			aiResult.ExpertDiagnosis.ServerExpert.RiskLevel,
			aiResult.ExpertDiagnosis.ComputerExpert.HardwareHealth,
			aiResult.ExpertDiagnosis.ComputerExpert.SoftwareStatus,
			aiResult.ExpertDiagnosis.ComputerExpert.SystemStability,
			aiResult.ExpertDiagnosis.ComputerExpert.ResourceUsage,
			formatMaintenanceNeeded(aiResult.ExpertDiagnosis.ComputerExpert.MaintenanceNeeded),
			formatCriticalIssues(aiResult.ExpertDiagnosis.CriticalIssues),
			formatMaintenanceTips(aiResult.ExpertDiagnosis.MaintenanceTips),
		)
		
		sm.logger.Infof("🚨 Sending AI alert to: %s", sm.emailService.GetRecipientsList())
		go func() {
			if err := sm.emailService.SendEmail(subject, body); err != nil {
				sm.logger.Errorf("❌ Failed to send AI alert email: %v", err)
			}
		}()
	}
	
	// Slack 알림 (SlackService 사용)
	if sm.slackService != nil {
		slackMsg := sm.slackService.CreateAIAlert(aiResult)
		
		go func() {
			if err := sm.slackService.SendMessage(slackMsg); err != nil {
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
		
		// 이메일 알림 (EmailService 사용)
		if sm.emailService != nil {
			subject := fmt.Sprintf("[%s SYSTEM ALERT] %s", AppName, alert.Type)
			
			body := fmt.Sprintf(`🖥️  시스템 알림

심각도: %s
메트릭: %s
메시지: %s
현재 값: %.2f
임계값: %.2f
시간: %s

시스템에서 임계값을 초과한 상황이 감지되었습니다.`,
				alert.Level,
				alert.Type,
				alert.Message,
				alert.Value,
				alert.Threshold,
				alert.Timestamp.Format("2006-01-02 15:04:05"),
			)
			
			sm.logger.Infof("🖥️  Sending system alert to: %s", sm.emailService.GetRecipientsList())
			go func() {
				if err := sm.emailService.SendEmail(subject, body); err != nil {
					sm.logger.Errorf("❌ Failed to send system alert email: %v", err)
				}
			}()
		}
		
		// Slack 알림 (SlackService 사용)
		if sm.slackService != nil {
			slackMsg := sm.slackService.CreateSystemAlert(alert)
			
			go func() {
				if err := sm.slackService.SendMessage(slackMsg); err != nil {
					sm.logger.Errorf("❌ Failed to send system alert to Slack: %v", err)
				}
			}()
		}
	}
}

// sendPeriodicSystemReports 주기적 시스템 상태 보고서 전송
func (sm *SyslogMonitor) sendPeriodicSystemReports() {
	ticker := time.NewTicker(sm.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.sendSystemStatusReport()
		}
	}
}

// sendSystemStatusReport 시스템 상태 보고서 전송
func (sm *SyslogMonitor) sendSystemStatusReport() {
	if sm.systemMonitor == nil {
		return
	}

	metrics := sm.systemMonitor.GetCurrentMetrics()
	
	// 이메일 보고서 전송
	if sm.emailService != nil {
		sm.sendSystemStatusEmail(metrics)
	}
	
	// Slack 보고서 전송
	if sm.slackService != nil {
		sm.sendSystemStatusSlack(metrics)
	}
	
	sm.logger.Infof("📊 시스템 상태 보고서 전송 완료 (CPU: %.1f%%, 메모리: %.1f%%)", 
		metrics.CPU.UsagePercent, metrics.Memory.UsagePercent)
}

// sendSystemStatusEmail 시스템 상태 이메일 보고서 전송
func (sm *SyslogMonitor) sendSystemStatusEmail(metrics SystemMetrics) {
	subject := fmt.Sprintf("[%s] 📊 시스템 상태 보고서 - %s", AppName, time.Now().Format("2006-01-02 15:04"))
	
	body := sm.generateSystemStatusEmailBody(metrics)
	
	go func() {
		if err := sm.emailService.SendEmail(subject, body); err != nil {
			sm.logger.Errorf("❌ Failed to send system status email: %v", err)
		}
	}()
}

// sendSystemStatusSlack 시스템 상태 Slack 보고서 전송
func (sm *SyslogMonitor) sendSystemStatusSlack(metrics SystemMetrics) {
	slackMsg := sm.generateSystemStatusSlackMessage(metrics)
	
	go func() {
		if err := sm.slackService.SendMessage(slackMsg); err != nil {
			sm.logger.Errorf("❌ Failed to send system status to Slack: %v", err)
		}
	}()
}

// generateSystemStatusEmailBody 시스템 상태 이메일 본문 생성
func (sm *SyslogMonitor) generateSystemStatusEmailBody(metrics SystemMetrics) string {
	hostname, _ := os.Hostname()
	
	return fmt.Sprintf(`🖥️  시스템 상태 보고서

📅 보고서 시간: %s
🖥️  호스트명: %s

🌐 네트워크 정보:
   사설 IP: %s
   공인 IP: %s

📊 CPU 상태:
   사용률: %.1f%%
   사용자: %.1f%%
   시스템: %.1f%%
   유휴: %.1f%%
   코어 수: %d

💾 메모리 상태:
   총 메모리: %.1f MB
   사용 중: %.1f MB (%.1f%%)
   사용 가능: %.1f MB
   스왑 사용: %.1f MB (%.1f%%)

💿 디스크 상태:
%s

🌡️  온도 정보:
   CPU 온도: %.1f°C
   GPU 온도: %.1f°C

📈 시스템 부하:
   1분 평균: %.2f
   5분 평균: %.2f
   15분 평균: %.2f

🔄 프로세스 상태:
   총 프로세스: %d
   실행 중: %d
   대기 중: %d

---
📊 이 보고서는 %v마다 자동으로 전송됩니다.
🤖 AI-Powered Syslog Monitor v2.1`,
		time.Now().Format("2006-01-02 15:04:05"),
		hostname,
		formatIPList(metrics.IPInfo.PrivateIPs),
		formatIPList(metrics.IPInfo.PublicIPs),
		metrics.CPU.UsagePercent,
		metrics.CPU.UserPercent,
		metrics.CPU.SystemPercent,
		metrics.CPU.IdlePercent,
		metrics.CPU.Cores,
		metrics.Memory.TotalMB,
		metrics.Memory.UsedMB,
		metrics.Memory.UsagePercent,
		metrics.Memory.AvailableMB,
		metrics.Memory.SwapUsedMB,
		metrics.Memory.SwapFreePercent,
		sm.generateDiskStatusText(metrics.Disk),
		metrics.Temperature.CPUTemp,
		metrics.Temperature.GPUTemp,
		metrics.LoadAverage.Load1Min,
		metrics.LoadAverage.Load5Min,
		metrics.LoadAverage.Load15Min,
		metrics.ProcessCount.Total,
		metrics.ProcessCount.Running,
		metrics.ProcessCount.Sleeping,
		sm.reportInterval)
}

// generateDiskStatusText 디스크 상태 텍스트 생성
func (sm *SyslogMonitor) generateDiskStatusText(disks []DiskMetrics) string {
	if len(disks) == 0 {
		return "   정보 없음"
	}
	
	var result strings.Builder
	for _, disk := range disks {
		result.WriteString(fmt.Sprintf("   %s (%s): %.1f GB / %.1f GB (%.1f%%)\n",
			disk.Device, disk.MountPoint, disk.UsedGB, disk.TotalGB, disk.UsagePercent))
	}
	return result.String()
}

// formatIPList IP 목록을 문자열로 포맷팅
func formatIPList(ips []string) string {
	if len(ips) == 0 {
		return "없음"
	}
	return strings.Join(ips, ", ")
}

// formatMaintenanceNeeded 유지보수 필요성 포맷팅
func formatMaintenanceNeeded(needed bool) string {
	if needed {
		return "예"
	}
	return "아니오"
}

// formatCriticalIssues 긴급 이슈 포맷팅
func formatCriticalIssues(issues []string) string {
	if len(issues) == 0 {
		return "없음"
	}
	var result strings.Builder
	for _, issue := range issues {
		result.WriteString(fmt.Sprintf("  • %s\n", issue))
	}
	return result.String()
}

// formatMaintenanceTips 유지보수 팁 포맷팅
func formatMaintenanceTips(tips []string) string {
	if len(tips) == 0 {
		return "없음"
	}
	var result strings.Builder
	for _, tip := range tips {
		result.WriteString(fmt.Sprintf("  • %s\n", tip))
	}
	return result.String()
}

// generateSystemStatusSlackMessage 시스템 상태 Slack 메시지 생성
func (sm *SyslogMonitor) generateSystemStatusSlackMessage(metrics SystemMetrics) SlackMessage {
	hostname, _ := os.Hostname()
	
	// 상태에 따른 색상 결정
	color := "good"
	if metrics.CPU.UsagePercent > 80 || metrics.Memory.UsagePercent > 85 {
		color = "warning"
	}
	if metrics.CPU.UsagePercent > 90 || metrics.Memory.UsagePercent > 95 {
		color = "danger"
	}
	
	return SlackMessage{
		Text:      fmt.Sprintf("📊 시스템 상태 보고서 - %s", hostname),
		IconEmoji: ":bar_chart:",
		Attachments: []SlackAttachment{
			{
				Color: color,
				Title: "🖥️  시스템 리소스 상태",
				Fields: []SlackField{
					{Title: "CPU 사용률", Value: fmt.Sprintf("%.1f%%", metrics.CPU.UsagePercent), Short: true},
					{Title: "메모리 사용률", Value: fmt.Sprintf("%.1f%%", metrics.Memory.UsagePercent), Short: true},
					{Title: "디스크 사용률", Value: sm.getDiskUsageSummary(metrics.Disk), Short: true},
					{Title: "시스템 부하", Value: fmt.Sprintf("%.2f", metrics.LoadAverage.Load5Min), Short: true},
					{Title: "온도", Value: fmt.Sprintf("CPU: %.1f°C", metrics.Temperature.CPUTemp), Short: true},
					{Title: "프로세스", Value: fmt.Sprintf("%d 실행 중", metrics.ProcessCount.Running), Short: true},
				},
				Timestamp: metrics.Timestamp.Unix(),
			},
		},
	}
}

// getDiskUsageSummary 디스크 사용률 요약 생성
func (sm *SyslogMonitor) getDiskUsageSummary(disks []DiskMetrics) string {
	if len(disks) == 0 {
		return "N/A"
	}
	
	// 가장 사용률이 높은 디스크 반환
	maxUsage := 0.0
	for _, disk := range disks {
		if disk.UsagePercent > maxUsage {
			maxUsage = disk.UsagePercent
		}
	}
	return fmt.Sprintf("%.1f%%", maxUsage)
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
	// 설정 서비스 초기화
	configPath := os.Getenv("SYSLOG_CONFIG_PATH")
	if configPath == "" {
		configPath = "~/.syslog-monitor/config.json"
	}
	
	configService = NewConfigService(configPath)
	if err := configService.LoadConfig(); err != nil {
		fmt.Printf("❌ 설정 파일 로드 실패: %v\n", err)
		fmt.Println("💡 기본 설정으로 시작합니다.")
	}
	
	// Gemini 서비스 초기화
	geminiConfig := configService.GetGeminiConfig()
	geminiService = NewGeminiService(geminiConfig)
	
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
		
		// 새로운 알림 관련 플래그
		alertIntervalFlag   = flag.Int("alert-interval", 10, "Login alert interval in minutes (default: 10)")
		periodicReportFlag  = flag.Bool("periodic-report", false, "Enable periodic system status reports")
		reportIntervalFlag  = flag.Int("report-interval", 60, "Report interval in minutes (default: 60)")
		
		// Gemini API 관련 플래그
		geminiAPIKey = flag.String("gemini-api-key", "", "Gemini API key for advanced AI analysis")
		showConfig   = flag.Bool("show-config", false, "Show current configuration")
		
		// 백그라운드 서비스 관련 플래그
		daemonMode     = flag.Bool("daemon", false, "Run as background daemon service")
		pidFile        = flag.String("pid-file", "/usr/local/var/run/syslog-monitor.pid", "PID file path for daemon mode")
		logDir         = flag.String("log-dir", "/usr/local/var/log", "Log directory for daemon mode")
		installService = flag.Bool("install-service", false, "Install as macOS LaunchAgent service")
		removeService  = flag.Bool("remove-service", false, "Remove macOS LaunchAgent service")
		startService   = flag.Bool("start-service", false, "Start the installed service")
		stopService    = flag.Bool("stop-service", false, "Stop the running service")
		statusService  = flag.Bool("status-service", false, "Show service status")
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

	// Gemini API 키 설정
	if *geminiAPIKey != "" {
		if err := configService.SetGeminiAPIKey(*geminiAPIKey); err != nil {
			fmt.Printf("❌ Gemini API 키 설정 실패: %v\n", err)
		} else {
			fmt.Printf("✅ Gemini API 키가 설정되었습니다: %s\n", configService.getMaskedAPIKey())
		}
	}

	// 설정 정보 표시
	if *showConfig {
		configService.ShowConfigInfo()
		return
	}
	
	// 서비스 관리 명령어 처리
	if *installService {
		installLaunchAgent()
		return
	}
	
	if *removeService {
		removeLaunchAgent()
		return
	}
	
	if *startService {
		startLaunchAgent()
		return
	}
	
	if *stopService {
		stopLaunchAgent()
		return
	}
	
	if *statusService {
		showServiceStatus()
		return
	}
	
	// Daemon 모드 설정
	if *daemonMode {
		setupDaemonMode()
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
		
		monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled, *loginWatch, *alertIntervalFlag, *reportIntervalFlag, *periodicReportFlag)
		
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

		if err := monitor.slackService.SendMessage(testMsg); err != nil {
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
		
		monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled, *loginWatch, *alertIntervalFlag, *reportIntervalFlag, *periodicReportFlag)
		subject := "[TEST] Syslog Monitor Email Test"
		body := fmt.Sprintf(`이것은 syslog 모니터의 테스트 이메일입니다.

테스트 시간: %s
SMTP 서버: %s:%s
발신자: %s
수신자: %s

이 이메일을 받으셨다면 이메일 설정이 올바르게 구성되었습니다.

Syslog Monitor
`, time.Now().Format("2006-01-02 15:04:05"), *smtpServer, *smtpPort, *emailFrom, strings.Join(emailConfig.To, ", "))

		if err := monitor.emailService.SendEmail(subject, body); err != nil {
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
	monitor := NewSyslogMonitor(*logFile, *outputFile, filters, keywords, emailConfig, slackConfig, *aiEnabled, *systemEnabled, *loginWatch, *alertIntervalFlag, *reportIntervalFlag, *periodicReportFlag)
	
	if err := monitor.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// setupDaemonMode daemon 모드 설정
func setupDaemonMode() {
	fmt.Println("🔧 Setting up daemon mode...")
	
	// 로그 디렉토리 생성
	if err := os.MkdirAll(*logDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	
	// PID 파일 디렉토리 생성
	pidDir := filepath.Dir(*pidFile)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create PID directory: %v\n", err)
		os.Exit(1)
	}
	
	// 이미 실행 중인지 확인
	if isRunning() {
		fmt.Println("⚠️  Daemon is already running")
		os.Exit(1)
	}
	
	// PID 파일 생성
	pid := os.Getpid()
	if err := os.WriteFile(*pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		fmt.Printf("❌ Failed to write PID file: %v\n", err)
		os.Exit(1)
	}
	
	// 프로세스 종료 시 PID 파일 삭제
	defer func() {
		os.Remove(*pidFile)
	}()
	
	// 로그 파일 설정
	logFile := filepath.Join(*logDir, "syslog-monitor.log")
	logOut, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("❌ Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logOut.Close()
	
	// 표준 출력을 로그 파일로 리다이렉션
	os.Stdout = logOut
	os.Stderr = logOut
	
	fmt.Printf("🚀 Daemon started (PID: %d)\n", pid)
	fmt.Printf("📝 Log file: %s\n", logFile)
	fmt.Printf("📋 PID file: %s\n", *pidFile)
}

// isRunning 프로세스가 실행 중인지 확인
func isRunning() bool {
	if _, err := os.Stat(*pidFile); os.IsNotExist(err) {
		return false
	}
	
	pidBytes, err := os.ReadFile(*pidFile)
	if err != nil {
		return false
	}
	
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return false
	}
	
	// 프로세스가 실제로 실행 중인지 확인
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// macOS에서 프로세스 존재 확인
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// installLaunchAgent macOS LaunchAgent 서비스 설치
func installLaunchAgent() {
	fmt.Println("📦 Installing macOS LaunchAgent service...")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	
	// LaunchAgents 디렉토리 생성
	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create LaunchAgents directory: %v\n", err)
		os.Exit(1)
	}
	
	// plist 파일 경로
	plistFile := filepath.Join(launchAgentsDir, "com.lambda-x.syslog-monitor.plist")
	
	// 현재 디렉토리의 plist 파일을 복사
	srcPlist := "com.lambda-x.syslog-monitor.plist"
	if _, err := os.Stat(srcPlist); os.IsNotExist(err) {
		fmt.Printf("❌ plist file not found: %s\n", srcPlist)
		fmt.Println("💡 Please run this command from the project directory")
		os.Exit(1)
	}
	
	// plist 파일 복사
	plistData, err := os.ReadFile(srcPlist)
	if err != nil {
		fmt.Printf("❌ Failed to read plist file: %v\n", err)
		os.Exit(1)
	}
	
	if err := os.WriteFile(plistFile, plistData, 0644); err != nil {
		fmt.Printf("❌ Failed to write plist file: %v\n", err)
		os.Exit(1)
	}
	
	// 로그 디렉토리 생성
	if err := os.MkdirAll("/usr/local/var/log", 0755); err != nil {
		fmt.Printf("⚠️  Warning: Could not create log directory: %v\n", err)
	}
	
	fmt.Printf("✅ Service installed successfully\n")
	fmt.Printf("📄 plist file: %s\n", plistFile)
	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Printf("   Load service:   syslog-monitor -start-service\n")
	fmt.Printf("   Check status:   syslog-monitor -status-service\n")
	fmt.Printf("   View logs:      tail -f /usr/local/var/log/syslog-monitor.out.log\n")
}

// removeLaunchAgent macOS LaunchAgent 서비스 제거
func removeLaunchAgent() {
	fmt.Println("🗑️  Removing macOS LaunchAgent service...")
	
	// 먼저 서비스 중지
	stopLaunchAgent()
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	
	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.lambda-x.syslog-monitor.plist")
	
	if err := os.Remove(plistFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("⚠️  Service is not installed")
		} else {
			fmt.Printf("❌ Failed to remove plist file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("✅ Service removed successfully")
	}
}

// startLaunchAgent macOS LaunchAgent 서비스 시작
func startLaunchAgent() {
	fmt.Println("🚀 Starting macOS LaunchAgent service...")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	
	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.lambda-x.syslog-monitor.plist")
	
	// plist 파일 존재 확인
	if _, err := os.Stat(plistFile); os.IsNotExist(err) {
		fmt.Println("❌ Service is not installed. Run with -install-service first.")
		os.Exit(1)
	}
	
	// launchctl load 명령 실행
	cmd := exec.Command("launchctl", "load", plistFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ Failed to start service: %v\n", err)
		fmt.Printf("Output: %s\n", output)
		os.Exit(1)
	}
	
	fmt.Println("✅ Service started successfully")
	fmt.Printf("📋 View status: syslog-monitor -status-service\n")
	fmt.Printf("📄 View logs:   tail -f /usr/local/var/log/syslog-monitor.out.log\n")
}

// stopLaunchAgent macOS LaunchAgent 서비스 중지
func stopLaunchAgent() {
	fmt.Println("⏹️  Stopping macOS LaunchAgent service...")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	
	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.lambda-x.syslog-monitor.plist")
	
	// launchctl unload 명령 실행
	cmd := exec.Command("launchctl", "unload", plistFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		// unload 실패는 이미 중지된 상태일 수 있으므로 경고만 표시
		fmt.Printf("⚠️  Warning: %v\n", err)
		fmt.Printf("Output: %s\n", output)
	} else {
		fmt.Println("✅ Service stopped successfully")
	}
}

// showServiceStatus 서비스 상태 표시
func showServiceStatus() {
	fmt.Println("📊 Service Status")
	fmt.Println("=================")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	
	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.lambda-x.syslog-monitor.plist")
	
	// 설치 상태 확인
	if _, err := os.Stat(plistFile); os.IsNotExist(err) {
		fmt.Println("❌ Service is not installed")
		fmt.Println("💡 Install with: syslog-monitor -install-service")
		return
	}
	
	fmt.Println("✅ Service is installed")
	fmt.Printf("📄 plist file: %s\n", plistFile)
	
	// launchctl list로 실행 상태 확인
	cmd := exec.Command("launchctl", "list", "com.lambda-x.syslog-monitor")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("⏹️  Service is not running")
		fmt.Println("💡 Start with: syslog-monitor -start-service")
	} else {
		fmt.Println("🟢 Service is running")
		fmt.Printf("Details:\n%s\n", output)
	}
	
	// 로그 파일 상태 확인
	logFiles := []string{
		"/usr/local/var/log/syslog-monitor.out.log",
		"/usr/local/var/log/syslog-monitor.err.log",
	}
	
	fmt.Println("\n📄 Log Files:")
	for _, logFile := range logFiles {
		if stat, err := os.Stat(logFile); err == nil {
			fmt.Printf("  ✅ %s (size: %d bytes, modified: %s)\n", 
				logFile, stat.Size(), stat.ModTime().Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  ❌ %s (not found)\n", logFile)
		}
	}
	
	fmt.Println("\n🔧 Commands:")
	fmt.Println("  Start:   syslog-monitor -start-service")
	fmt.Println("  Stop:    syslog-monitor -stop-service")
	fmt.Println("  Remove:  syslog-monitor -remove-service")
	fmt.Println("  Logs:    tail -f /usr/local/var/log/syslog-monitor.out.log")
} 