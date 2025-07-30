/*
Configuration Service
====================

설정 파일 관리 및 Gemini API 연동 서비스

주요 기능:
- JSON 설정 파일 읽기/쓰기
- Gemini API 키 관리
- 환경변수 기반 설정
- 설정 검증 및 기본값 처리

작성자: Lambda-X AI Team
버전: 1.0.0
*/
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 전체 설정 구조체
type Config struct {
	AI struct {
		Enabled         bool    `json:"enabled"`
		GeminiAPIKey   string  `json:"gemini_api_key"`
		GeminiModel    string  `json:"gemini_model"`
		AlertThreshold  float64 `json:"alert_threshold"`
		AnalysisInterval int    `json:"analysis_interval"`
	} `json:"ai_analysis"`

	SystemMonitoring struct {
		Enabled             bool    `json:"enabled"`
		CPUThreshold        float64 `json:"cpu_threshold"`
		MemoryThreshold     float64 `json:"memory_threshold"`
		DiskThreshold       float64 `json:"disk_threshold"`
		TemperatureThreshold float64 `json:"temperature_threshold"`
		MonitoringInterval  int     `json:"monitoring_interval"`
	} `json:"system_monitoring"`

	Email struct {
		Enabled    bool     `json:"enabled"`
		SMTPServer string   `json:"smtp_server"`
		SMTPPort   int      `json:"smtp_port"`
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		To         []string `json:"to"`
		From       string   `json:"from"`
	} `json:"email"`

	Slack struct {
		Enabled     bool   `json:"enabled"`
		WebhookURL  string `json:"webhook_url"`
		Channel     string `json:"channel"`
		Username    string `json:"username"`
	} `json:"slack"`

	Logging struct {
		LogFile    string `json:"log_file"`
		OutputFile string `json:"output_file"`
		Keywords   string `json:"keywords"`
		Filters    string `json:"filters"`
	} `json:"logging"`

	Features struct {
		ComputerNameDetection bool `json:"computer_name_detection"`
		IPClassification     bool `json:"ip_classification"`
		ASNLookup           bool `json:"asn_lookup"`
		RealTimeAnalysis    bool `json:"real_time_analysis"`
		ExpertDiagnosis     bool `json:"expert_diagnosis"`
	} `json:"features"`
}

// ConfigService 설정 관리 서비스
type ConfigService struct {
	configPath string
	config     *Config
}

// NewConfigService 설정 서비스 생성자
func NewConfigService(configPath string) *ConfigService {
	return &ConfigService{
		configPath: configPath,
		config:     &Config{},
	}
}

// LoadConfig 설정 파일 로드
func (cs *ConfigService) LoadConfig() error {
	// 설정 파일이 없으면 기본 설정 생성
	if _, err := os.Stat(cs.configPath); os.IsNotExist(err) {
		return cs.createDefaultConfig()
	}

	// 설정 파일 읽기
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// JSON 파싱
	if err := json.Unmarshal(data, cs.config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 환경변수에서 API 키 읽기
	cs.loadFromEnvironment()

	return nil
}

// SaveConfig 설정 파일 저장
func (cs *ConfigService) SaveConfig() error {
	// 디렉토리 생성
	dir := filepath.Dir(cs.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// JSON 직렬화
	data, err := json.MarshalIndent(cs.config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 파일 저장
	if err := os.WriteFile(cs.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// createDefaultConfig 기본 설정 생성
func (cs *ConfigService) createDefaultConfig() error {
	cs.config = &Config{
		AI: struct {
			Enabled         bool    `json:"enabled"`
			GeminiAPIKey   string  `json:"gemini_api_key"`
			GeminiModel    string  `json:"gemini_model"`
			AlertThreshold  float64 `json:"alert_threshold"`
			AnalysisInterval int    `json:"analysis_interval"`
		}{
			Enabled:         true,
			GeminiAPIKey:   "",
			GeminiModel:    "gemini-1.5-flash",
			AlertThreshold:  7.0,
			AnalysisInterval: 30,
		},
		SystemMonitoring: struct {
			Enabled             bool    `json:"enabled"`
			CPUThreshold        float64 `json:"cpu_threshold"`
			MemoryThreshold     float64 `json:"memory_threshold"`
			DiskThreshold       float64 `json:"disk_threshold"`
			TemperatureThreshold float64 `json:"temperature_threshold"`
			MonitoringInterval  int     `json:"monitoring_interval"`
		}{
			Enabled:             true,
			CPUThreshold:        80.0,
			MemoryThreshold:     85.0,
			DiskThreshold:       90.0,
			TemperatureThreshold: 75.0,
			MonitoringInterval:  300,
		},
		Email: struct {
			Enabled    bool     `json:"enabled"`
			SMTPServer string   `json:"smtp_server"`
			SMTPPort   int      `json:"smtp_port"`
			Username   string   `json:"username"`
			Password   string   `json:"password"`
			To         []string `json:"to"`
			From       string   `json:"from"`
		}{
			Enabled:    true,
			SMTPServer: "smtp.gmail.com",
			SMTPPort:   587,
			Username:   "enfn2001@gmail.com",
			Password:   "",
			To:         []string{"robot@lambda-x.ai", "enfn2001@gmail.com"},
			From:       "security@lambda-x.ai",
		},
		Slack: struct {
			Enabled     bool   `json:"enabled"`
			WebhookURL  string `json:"webhook_url"`
			Channel     string `json:"channel"`
			Username    string `json:"username"`
		}{
			Enabled:    false,
			WebhookURL: "",
			Channel:    "#security",
			Username:   "AI Security Monitor",
		},
		Logging: struct {
			LogFile    string `json:"log_file"`
			OutputFile string `json:"output_file"`
			Keywords   string `json:"keywords"`
			Filters    string `json:"filters"`
		}{
			LogFile:    "/var/log/system.log",
			OutputFile: "",
			Keywords:   "",
			Filters:    "",
		},
		Features: struct {
			ComputerNameDetection bool `json:"computer_name_detection"`
			IPClassification     bool `json:"ip_classification"`
			ASNLookup           bool `json:"asn_lookup"`
			RealTimeAnalysis    bool `json:"real_time_analysis"`
			ExpertDiagnosis     bool `json:"expert_diagnosis"`
		}{
			ComputerNameDetection: true,
			IPClassification:     true,
			ASNLookup:           true,
			RealTimeAnalysis:    true,
			ExpertDiagnosis:     true,
		},
	}

	// 환경변수에서 API 키 읽기
	cs.loadFromEnvironment()

	return cs.SaveConfig()
}

// loadFromEnvironment 환경변수에서 설정 로드
func (cs *ConfigService) loadFromEnvironment() {
	// Gemini API 키
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		cs.config.AI.GeminiAPIKey = apiKey
	}

	// 이메일 설정
	if emailTo := os.Getenv("SYSLOG_EMAIL_TO"); emailTo != "" {
		cs.config.Email.To = []string{emailTo}
	}
	if smtpUser := os.Getenv("SYSLOG_SMTP_USER"); smtpUser != "" {
		cs.config.Email.Username = smtpUser
	}
	if smtpPassword := os.Getenv("SYSLOG_SMTP_PASSWORD"); smtpPassword != "" {
		cs.config.Email.Password = smtpPassword
	}

	// Slack 설정
	if webhookURL := os.Getenv("SYSLOG_SLACK_WEBHOOK"); webhookURL != "" {
		cs.config.Slack.WebhookURL = webhookURL
		cs.config.Slack.Enabled = true
	}
	if channel := os.Getenv("SYSLOG_SLACK_CHANNEL"); channel != "" {
		cs.config.Slack.Channel = channel
	}
}

// GetGeminiConfig Gemini 설정 반환
func (cs *ConfigService) GetGeminiConfig() *GeminiConfig {
	return &GeminiConfig{
		APIKey:     cs.config.AI.GeminiAPIKey,
		Model:      cs.config.AI.GeminiModel,
		MaxTokens:  2048,
		Temperature: 0.7,
		Enabled:    cs.config.AI.Enabled,
	}
}

// GetConfig 전체 설정 반환
func (cs *ConfigService) GetConfig() *Config {
	return cs.config
}

// SetGeminiAPIKey Gemini API 키 설정
func (cs *ConfigService) SetGeminiAPIKey(apiKey string) error {
	cs.config.AI.GeminiAPIKey = apiKey
	cs.config.AI.Enabled = true
	return cs.SaveConfig()
}

// ValidateGeminiAPIKey Gemini API 키 검증
func (cs *ConfigService) ValidateGeminiAPIKey() error {
	if cs.config.AI.GeminiAPIKey == "" {
		return fmt.Errorf("Gemini API 키가 설정되지 않았습니다")
	}

	// 간단한 API 키 형식 검증
	if len(cs.config.AI.GeminiAPIKey) < 10 {
		return fmt.Errorf("Gemini API 키 형식이 올바르지 않습니다")
	}

	return nil
}

// GetConfigPath 설정 파일 경로 반환
func (cs *ConfigService) GetConfigPath() string {
	return cs.configPath
}

// ShowConfigInfo 설정 정보 표시
func (cs *ConfigService) ShowConfigInfo() {
	fmt.Printf(`
🔧 설정 정보
============
📁 설정 파일: %s
🤖 AI 분석: %t
🔑 Gemini API 키: %s
📊 시스템 모니터링: %t
📧 이메일 알림: %t
💬 Slack 알림: %t

💡 Gemini API 키 설정 방법:
1. https://makersuite.google.com/app/apikey 에서 API 키 생성
2. 환경변수 설정: export GEMINI_API_KEY="your-api-key"
3. 또는 설정 파일 직접 편집: %s
`,
		cs.configPath,
		cs.config.AI.Enabled,
		cs.getMaskedAPIKey(),
		cs.config.SystemMonitoring.Enabled,
		cs.config.Email.Enabled,
		cs.config.Slack.Enabled,
		cs.configPath)
}

// getMaskedAPIKey 마스킹된 API 키 반환
func (cs *ConfigService) getMaskedAPIKey() string {
	apiKey := cs.config.AI.GeminiAPIKey
	if apiKey == "" {
		return "설정되지 않음"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
} 