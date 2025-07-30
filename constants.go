/*
Application Constants and Configuration
======================================

시스템 전반에서 사용되는 상수 및 기본 설정값 정의

포함 항목:
- 애플리케이션 메타데이터
- SMTP/이메일 서버 설정
- 시스템 모니터링 임계값
- AI 분석 매개변수
- 로그 파일 경로 (OS별)
- 네트워크 설정 (IP 범위, ASN 조회)
- 에러 메시지 상수
- Slack 메시지 포맷 설정
*/
package main

import "time" // 시간 간격 상수 정의용

// Application constants 애플리케이션 기본 정보
const (
	AppName    = "AI-Powered Syslog Monitor" // 애플리케이션 이름
	AppVersion = "2.0.0"                     // 현재 버전 (시맨틱 버저닝)
)

// SMTP/Email constants SMTP 서버 및 이메일 관련 설정
const (
	DefaultSMTPServer = "smtp.gmail.com" // Gmail SMTP 서버 주소
	DefaultSMTPPort   = "587"            // STARTTLS 포트 (권장)
	SMTPPortSSL       = "465"            // SSL/TLS 직접 연결 포트
	SMTPPortTLS       = "587"            // STARTTLS 포트 (동일)
)

// Default email recipients 기본 이메일 수신자 목록
// 긴급 알림을 받을 이메일 주소들 (여러 명에게 동시 전송)
var DefaultEmailRecipients = []string{
	"robot@lambda-x.ai", // Lambda-X AI 팀 메인 주소
	"enfn2001@gmail.com", // 개발자 개인 주소
}

// Default SMTP credentials 기본 SMTP 인증 정보
// Gmail 앱 패스워드 사용 (2단계 인증 필수)
const (
	DefaultSMTPUser     = "enfn2001@gmail.com"    // Gmail 계정
	DefaultSMTPPassword = "kwev eavp nrbi mtrj"   // Gmail 앱 패스워드 (16자리)
)

// Time intervals 시간 간격 관련 설정값
const (
	DefaultMonitoringInterval = time.Minute * 5 // 시스템 모니터링 주기 (5분마다 메트릭 수집)
	DefaultTimeWindow         = time.Minute * 5 // AI 분석 시간 윈도우 (최근 5분간 로그 분석)
	DefaultLogBufferSize      = 1000            // 로그 버퍼 최대 크기 (메모리 사용량 제한)
	
	// Login alert throttling 로그인 알림 제한 설정
	DefaultLoginAlertInterval   = time.Minute * 10 // 기본 로그인 알림 간격 (10분)
	CriticalAlertInterval       = time.Minute * 2  // 중요 알림 간격 (실패한 로그인 등, 2분)
	MaxAlertHistorySize         = 100              // 알림 히스토리 최대 크기
	AlertHistoryCleanupInterval = time.Hour * 1    // 알림 히스토리 정리 간격 (1시간)
)

// AI Analysis thresholds AI 분석 및 이상 탐지 임계값
const (
	DefaultAlertThreshold   = 7.0  // 기본 알림 임계값 (7점 이상시 알림 발송)
	HighThreatThreshold     = 8.0  // 높은 위험도 임계값 (긴급 처리 필요)
	CriticalThreatThreshold = 9.0  // 치명적 위험도 임계값 (즉시 대응 필요)
	MaxAnomalyScore         = 10.0 // 이상 점수 최대값 (정규화 기준)
)

// System monitoring thresholds 시스템 리소스 모니터링 임계값
const (
	DefaultCPUThreshold    = 80.0 // CPU 사용률 경고 임계값 (80% 이상)
	DefaultMemoryThreshold = 85.0 // 메모리 사용률 경고 임계값 (85% 이상)
	DefaultDiskThreshold   = 90.0 // 디스크 사용률 경고 임계값 (90% 이상)
	DefaultLoadThreshold   = 2.0  // 로드 평균 경고 임계값 (CPU 코어 수 * 2)
	DefaultTempThreshold   = 70.0 // CPU 온도 경고 임계값 (70°C 이상)
)

// Log file paths by OS 운영체제별 기본 로그 파일 경로
const (
	LinuxSyslogPath   = "/var/log/syslog"     // Linux 메인 시스템 로그
	LinuxMessagesPath = "/var/log/messages"   // Linux 일반 메시지 로그
	LinuxAuthLogPath  = "/var/log/auth.log"   // Linux 인증 관련 로그
	MacOSSystemPath   = "/var/log/system.log" // macOS 시스템 로그 (Monterey 이전)
	MacOSInstallPath  = "/var/log/install.log" // macOS 소프트웨어 설치 로그
	MacOSWiFiPath     = "/var/log/wifi.log"    // macOS WiFi 연결 로그
)

// IP address ranges for classification IP 주소 분류를 위한 사설 IP 대역
// RFC 1918 및 특수 용도 IP 주소 범위 정의
var PrivateIPRanges = []string{
	"192.168.0.0/16", // 클래스 C 사설 네트워크 (가정/소규모 사무실)
	"10.0.0.0/8",     // 클래스 A 사설 네트워크 (대규모 기업)
	"172.16.0.0/12",  // 클래스 B 사설 네트워크 (중규모 기업)
	"127.0.0.0/8",    // 루프백 주소 (localhost)
	"169.254.0.0/16", // APIPA 자동 사설 IP 주소
}

// ASN lookup settings ASN(Autonomous System Number) 조회 설정
// IP 주소의 지리적 위치 및 소유 기관 정보 조회
const (
	ASNLookupURL     = "http://ip-api.com/json/"              // 무료 IP 지리정보 API
	ASNTimeout       = 5 * time.Second                        // API 요청 타임아웃 (5초)
	ASNRequestFields = "?fields=org,country,region,city,as"   // 조회할 필드 목록
)

// Error messages 에러 메시지 상수 정의
// 사용자에게 표시되는 일관된 에러 메시지
const (
	ErrEmailSendFailed   = "failed to send email alert"           // 이메일 전송 실패
	ErrSlackSendFailed   = "failed to send slack alert"           // Slack 알림 전송 실패
	ErrFileNotFound      = "log file not found"                   // 로그 파일 없음
	ErrPermissionDenied  = "permission denied accessing log file" // 로그 파일 접근 권한 없음
	ErrSMTPAuth          = "SMTP authentication failed"           // SMTP 인증 실패
	ErrInvalidConfig     = "invalid configuration"                // 잘못된 설정
)

// Slack settings Slack 메시지 포맷 및 디자인 설정
const (
	DefaultSlackUsername = "AI Security Monitor" // 기본 Slack 봇 사용자명
	DefaultSlackIcon     = ":warning:"            // 기본 Slack 봇 아이콘
	SlackColorGood       = "good"                 // 정상/성공 상태 색상 (녹색)
	SlackColorWarning    = "warning"              // 경고 상태 색상 (노란색)
	SlackColorDanger     = "danger"               // 위험/에러 상태 색상 (빨간색)
)

// Regular expressions patterns 정규식 패턴 상수
// 보안 위협 및 로그 분석을 위한 사전 정의된 패턴들
const (
	IPRegexPattern        = `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`                                          // IPv4 주소 매칭
	EmailRegexPattern     = `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`                            // 이메일 주소 매칭
	SQLInjectionPattern   = `(?i)(union\s+select|or\s+1\s*=\s*1|drop\s+table|insert\s+into|delete\s+from)` // SQL 인젝션 공격 패턴
	BruteForcePattern     = `(?i)(failed\s+login|authentication\s+failed|invalid\s+password)`              // 무차별 대입 공격 패턴
	PrivilegeEscPattern   = `(?i)(sudo\s+su|unauthorized\s+access|privilege\s+escalation)`                // 권한 상승 시도 패턴
)

// Log levels 로그 레벨 표준 정의
// RFC 5424 Syslog 표준을 따른 로그 심각도 분류
const (
	LogLevelCritical = "CRITICAL" // 치명적 오류 (시스템 다운 등)
	LogLevelError    = "ERROR"    // 에러 (기능 동작 불가)
	LogLevelWarning  = "WARNING"  // 경고 (잠재적 문제)
	LogLevelInfo     = "INFO"     // 정보성 메시지
	LogLevelDebug    = "DEBUG"    // 디버그 정보
)

// Threat levels 위협 레벨 시각적 표시
// 이모지를 포함한 직관적인 위험도 표시
const (
	ThreatLevelLow      = "🟢 LOW"      // 낮은 위험도 (정상 범위)
	ThreatLevelMedium   = "🟡 MEDIUM"   // 중간 위험도 (주의 필요)
	ThreatLevelHigh     = "🟠 HIGH"     // 높은 위험도 (긴급 대응)
	ThreatLevelCritical = "🔴 CRITICAL" // 치명적 위험도 (즉시 대응)
)

// Configuration file settings 설정 파일 관련 상수
const (
	DefaultConfigDir  = ".syslog-monitor" // 설정 파일 디렉토리 (~/.syslog-monitor)
	DefaultConfigFile = "config.json"     // 설정 파일명
	ConfigPermissions = 0755              // 설정 디렉토리 권한 (rwxr-xr-x)
) 