package main

import "time"

// Application constants
const (
	AppName    = "AI-Powered Syslog Monitor"
	AppVersion = "2.0.0"
)

// SMTP/Email constants
const (
	DefaultSMTPServer = "smtp.gmail.com"
	DefaultSMTPPort   = "587"
	SMTPPortSSL       = "465"
	SMTPPortTLS       = "587"
)

// Default email recipients
var DefaultEmailRecipients = []string{
	"robot@lambda-x.ai",
	"enfn2001@gmail.com",
}

// Default SMTP credentials
const (
	DefaultSMTPUser     = "enfn2001@gmail.com"
	DefaultSMTPPassword = "kwev eavp nrbi mtrj" // App password for Gmail
)

// Time intervals
const (
	DefaultMonitoringInterval = time.Minute * 5
	DefaultTimeWindow         = time.Minute * 5
	DefaultLogBufferSize      = 1000
)

// AI Analysis thresholds
const (
	DefaultAlertThreshold   = 7.0
	HighThreatThreshold     = 8.0
	CriticalThreatThreshold = 9.0
	MaxAnomalyScore         = 10.0
)

// System monitoring thresholds
const (
	DefaultCPUThreshold    = 80.0
	DefaultMemoryThreshold = 85.0
	DefaultDiskThreshold   = 90.0
	DefaultLoadThreshold   = 2.0
	DefaultTempThreshold   = 70.0
)

// Log file paths by OS
const (
	LinuxSyslogPath   = "/var/log/syslog"
	LinuxMessagesPath = "/var/log/messages"
	LinuxAuthLogPath  = "/var/log/auth.log"
	MacOSSystemPath   = "/var/log/system.log"
	MacOSInstallPath  = "/var/log/install.log"
	MacOSWiFiPath     = "/var/log/wifi.log"
)

// IP address ranges for classification
var PrivateIPRanges = []string{
	"192.168.0.0/16",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"127.0.0.0/8",
	"169.254.0.0/16",
}

// ASN lookup settings
const (
	ASNLookupURL     = "http://ip-api.com/json/"
	ASNTimeout       = 5 * time.Second
	ASNRequestFields = "?fields=org,country,region,city,as"
)

// Error messages
const (
	ErrEmailSendFailed   = "failed to send email alert"
	ErrSlackSendFailed   = "failed to send slack alert"
	ErrFileNotFound      = "log file not found"
	ErrPermissionDenied  = "permission denied accessing log file"
	ErrSMTPAuth          = "SMTP authentication failed"
	ErrInvalidConfig     = "invalid configuration"
)

// Slack settings
const (
	DefaultSlackUsername = "AI Security Monitor"
	DefaultSlackIcon     = ":warning:"
	SlackColorGood       = "good"
	SlackColorWarning    = "warning"
	SlackColorDanger     = "danger"
)

// Regular expressions patterns
const (
	IPRegexPattern        = `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	EmailRegexPattern     = `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	SQLInjectionPattern   = `(?i)(union\s+select|or\s+1\s*=\s*1|drop\s+table|insert\s+into|delete\s+from)`
	BruteForcePattern     = `(?i)(failed\s+login|authentication\s+failed|invalid\s+password)`
	PrivilegeEscPattern   = `(?i)(sudo\s+su|unauthorized\s+access|privilege\s+escalation)`
)

// Log levels
const (
	LogLevelCritical = "CRITICAL"
	LogLevelError    = "ERROR"
	LogLevelWarning  = "WARNING"
	LogLevelInfo     = "INFO"
	LogLevelDebug    = "DEBUG"
)

// Threat levels
const (
	ThreatLevelLow      = "🟢 LOW"
	ThreatLevelMedium   = "🟡 MEDIUM"
	ThreatLevelHigh     = "🟠 HIGH"
	ThreatLevelCritical = "🔴 CRITICAL"
)

// Configuration file settings
const (
	DefaultConfigDir  = ".syslog-monitor"
	DefaultConfigFile = "config.json"
	ConfigPermissions = 0755
) 