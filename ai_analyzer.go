/*
AI-Powered Log Analysis Engine
=============================

고급 AI 기반 로그 분석 및 이상 탐지 엔진

주요 기능:
- 실시간 로그 패턴 분석
- 머신러닝 기반 이상 탐지
- 보안 위협 예측 및 분석
- 시스템 정보 수집 및 ASN 조회
- 동적 기준선 학습 및 적응

분석 항목:
- SQL 인젝션 공격 시도
- 무차별 대입 공격
- 권한 상승 시도
- 메모리 누수 패턴
- 데이터베이스 연결 이슈
- 비정상적인 트래픽 급증
*/
package main

import (
	"fmt"           // 형식화된 I/O
	"math"          // 수학 함수
	"regexp"        // 정규식 처리
	"sort"          // 정렬 알고리즘
	"strconv"       // 문자열-숫자 변환
	"strings"       // 문자열 처리
	"time"          // 시간 처리
	"os"            // 운영체제 인터페이스
	"net"           // 네트워크 처리
	"net/http"      // HTTP 클라이언트
	"encoding/json" // JSON 인코딩/디코딩
	"io"            // I/O 원시 기능
)

// AIAnalyzer AI 기반 로그 분석 및 이상 탐지 엔진
// 실시간으로 로그를 분석하여 보안 위협과 시스템 이상을 감지
type AIAnalyzer struct {
	patterns        []AnomalyPattern // 사전 정의된 이상 패턴 목록 (SQL 인젝션, 브루트포스 등)
	timeWindow      time.Duration    // 분석 시간 윈도우 (기본 5분, 최근 로그만 분석)
	logBuffer       []LogEntry       // 순환 버퍼로 최근 로그 항목들을 메모리에 보관
	maxBufferSize   int              // 버퍼 최대 크기 (메모리 사용량 제한, 기본 1000개)
	alertThreshold  float64          // 알림 임계값 (이상 점수가 이 값 이상이면 알림 발송)
	baselineMetrics BaselineMetrics  // 동적으로 학습되는 정상 상태 기준선 메트릭
}

// LogEntry 개별 로그 항목을 나타내는 구조체
// 원본 로그와 분석된 메타데이터를 함께 저장
type LogEntry struct {
	Timestamp time.Time   // 로그 발생 시각
	Level     string      // 로그 레벨 (DEBUG, INFO, WARNING, ERROR, CRITICAL)
	Service   string      // 로그를 생성한 서비스명 (sshd, nginx, mysql 등)
	Host      string      // 로그를 생성한 호스트명
	Message   string      // 로그 메시지 본문
	Raw       string      // 원본 로그 라인 (파싱 전 상태)
	Features  LogFeatures // 추출된 로그 특성 정보 (AI 분석용)
}

// LogFeatures 로그에서 추출한 다양한 특성들을 저장하는 구조체
// AI 분석을 위한 피처 엔지니어링 결과물
type LogFeatures struct {
	ErrorCount      int       // 에러 관련 키워드 출현 빈도
	WarningCount    int       // 경고 관련 키워드 출현 빈도
	CriticalCount   int       // 치명적 오류 관련 키워드 출현 빈도
	IPAddresses     []string  // 로그에서 추출된 IP 주소 목록
	UniqueUsers     []string  // 로그에서 추출된 사용자명 목록
	ServiceCalls    []string  // 서비스 호출 정보 목록
	ResponseTimes   []float64 // HTTP 응답 시간 목록 (밀리초 단위)
	HTTPStatusCodes []int     // HTTP 상태 코드 목록 (200, 404, 500 등)
	SQLQueries      []string  // 추출된 SQL 쿼리 목록 (보안 분석용)
	Severity        float64   // 계산된 심각도 점수 (0-10 스케일)
	Frequency       float64   // 로그 발생 빈도 (분당 횟수)
	SystemInfo      SystemInfo // 시스템 및 네트워크 정보 (IP 지리정보 포함)
}

// AnomalyPattern 이상 패턴 정의
type AnomalyPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    float64
	Description string
	Category    string
	Action      string
}

// BaselineMetrics 기준선 메트릭
type BaselineMetrics struct {
	AvgErrorRate      float64
	AvgResponseTime   float64
	TypicalLogVolume  float64
	NormalUserCount   int
	BaselineUpdatedAt time.Time
}

// AIAnalysisResult AI 분석 결과
type AIAnalysisResult struct {
	AnomalyScore    float64
	ThreatLevel     string
	Predictions     []Prediction
	Recommendations []string
	AffectedSystems []string
	Confidence      float64
	Timestamp       time.Time
	SystemInfo      SystemInfo  // 시스템 정보 추가
	ExpertDiagnosis ExpertDiagnosis // 전문가 진단 결과
}

// Prediction 예측 결과
type Prediction struct {
	Event       string
	Probability float64
	TimeFrame   string
	Impact      string
}

// ASNInfo ASN 정보 구조체
type ASNInfo struct {
	IP           string `json:"ip"`
	ASN          string `json:"asn"`
	Organization string `json:"org"`
	Country      string `json:"country"`
	Region       string `json:"region"`
	City         string `json:"city"`
}

// SystemInfo 시스템 정보 구조체
type SystemInfo struct {
	ComputerName string
	InternalIPs  []string
	ExternalIPs  []string
	ASNData      []ASNInfo
}

// ServerExpertDiagnosis 서버 전문가 진단 결과
type ServerExpertDiagnosis struct {
	ServerHealth     string   // 서버 건강도 (Excellent/Good/Fair/Poor/Critical)
	PerformanceScore float64  // 성능 점수 (0-100)
	SecurityStatus   string   // 보안 상태
	NetworkHealth    string   // 네트워크 건강도
	Issues           []string // 발견된 이슈들
	Recommendations  []string // 서버 전문가 권장사항
	RiskLevel        string   // 위험도 (Low/Medium/High/Critical)
}

// ComputerExpertDiagnosis 컴퓨터 전문가 진단 결과
type ComputerExpertDiagnosis struct {
	HardwareHealth   string   // 하드웨어 건강도
	SoftwareStatus   string   // 소프트웨어 상태
	SystemStability  string   // 시스템 안정성
	ResourceUsage    string   // 리소스 사용량 상태
	Issues           []string // 발견된 이슈들
	Recommendations  []string // 컴퓨터 전문가 권장사항
	MaintenanceNeeded bool    // 유지보수 필요 여부
}

// ExpertDiagnosis 전문가 진단 결과
type ExpertDiagnosis struct {
	ServerExpert    ServerExpertDiagnosis    // 서버 전문가 진단
	ComputerExpert  ComputerExpertDiagnosis  // 컴퓨터 전문가 진단
	OverallHealth   string                   // 전체 시스템 건강도
	CriticalIssues  []string                 // 긴급 이슈 목록
	MaintenanceTips []string                 // 유지보수 팁
	PerformanceScore float64                 // 성능 점수 (0-100)
}

// NewAIAnalyzer AI 분석기 생성
func NewAIAnalyzer() *AIAnalyzer {
	patterns := []AnomalyPattern{
		{
			Name:        "SQL_Injection_Attempt",
			Pattern:     regexp.MustCompile(`(?i)(union\s+select|or\s+1=1|drop\s+table|insert\s+into|delete\s+from|\'\s+or\s+\'\w+=\'\w+)`),
			Severity:    9.0,
			Description: "SQL 인젝션 공격 시도 감지",
			Category:    "Security",
			Action:      "immediate_block",
		},
		{
			Name:        "Brute_Force_Login",
			Pattern:     regexp.MustCompile(`(?i)(failed\s+login|authentication\s+failed|invalid\s+password)`),
			Severity:    7.5,
			Description: "무차별 대입 공격 감지",
			Category:    "Security",
			Action:      "rate_limit",
		},
		{
			Name:        "Memory_Leak_Pattern",
			Pattern:     regexp.MustCompile(`(?i)(out\s+of\s+memory|memory\s+allocation\s+failed|heap\s+exhausted)`),
			Severity:    8.0,
			Description: "메모리 누수 패턴 감지",
			Category:    "Performance",
			Action:      "investigate",
		},
		{
			Name:        "Database_Connection_Issue",
			Pattern:     regexp.MustCompile(`(?i)(connection\s+timeout|database\s+unreachable|connection\s+pool\s+exhausted)`),
			Severity:    8.5,
			Description: "데이터베이스 연결 문제",
			Category:    "Database",
			Action:      "restart_db_pool",
		},
		{
			Name:        "Unusual_Traffic_Spike",
			Pattern:     regexp.MustCompile(`(?i)(rate\s+limit\s+exceeded|too\s+many\s+requests|ddos)`),
			Severity:    8.0,
			Description: "비정상적인 트래픽 급증",
			Category:    "Network",
			Action:      "activate_ddos_protection",
		},
		{
			Name:        "File_System_Error",
			Pattern:     regexp.MustCompile(`(?i)(disk\s+full|no\s+space\s+left|file\s+system\s+error|permission\s+denied)`),
			Severity:    7.0,
			Description: "파일 시스템 오류",
			Category:    "System",
			Action:      "cleanup_logs",
		},
		{
			Name:        "Privilege_Escalation",
			Pattern:     regexp.MustCompile(`(?i)(sudo\s+su|privilege\s+escalation|unauthorized\s+access|root\s+login)`),
			Severity:    9.5,
			Description: "권한 상승 시도",
			Category:    "Security",
			Action:      "immediate_alert",
		},
	}

	return &AIAnalyzer{
		patterns:       patterns,
		timeWindow:     time.Minute * 5,
		maxBufferSize:  1000,
		alertThreshold: 7.0,
		logBuffer:      make([]LogEntry, 0),
		baselineMetrics: BaselineMetrics{
			AvgErrorRate:      0.05,
			AvgResponseTime:   500.0,
			TypicalLogVolume:  100.0,
			NormalUserCount:   50,
			BaselineUpdatedAt: time.Now(),
		},
	}
}

// AnalyzeLog 로그 분석 수행
func (ai *AIAnalyzer) AnalyzeLog(logLine string, parsed map[string]string) *AIAnalysisResult {
	// 로그 항목 생성
	entry := ai.createLogEntry(logLine, parsed)
	
	// 버퍼에 추가
	ai.addToBuffer(entry)
	
	// 특성 추출
	features := ai.extractFeatures(entry)
	entry.Features = features
	
	// 이상 패턴 감지
	anomalyScore := ai.detectAnomalies(entry)
	
	// 예측 수행
	predictions := ai.makePredictions(entry, features)
	
	// 추천사항 생성
	recommendations := ai.generateRecommendations(entry, anomalyScore)
	
	// 위협 레벨 결정
	threatLevel := ai.calculateThreatLevel(anomalyScore)
	
	// 전문가 진단 수행 (시스템 메트릭이 없는 경우 nil 전달)
	expertDiagnosis := ai.PerformExpertDiagnosis(entry, features, nil)
	
	return &AIAnalysisResult{
		AnomalyScore:    anomalyScore,
		ThreatLevel:     threatLevel,
		Predictions:     predictions,
		Recommendations: recommendations,
		AffectedSystems: ai.identifyAffectedSystems(entry),
		Confidence:      ai.calculateConfidence(anomalyScore, features),
		Timestamp:       time.Now(),
		SystemInfo:      features.SystemInfo,
		ExpertDiagnosis: expertDiagnosis,
	}
}

// createLogEntry 로그 항목 생성
func (ai *AIAnalyzer) createLogEntry(logLine string, parsed map[string]string) LogEntry {
	return LogEntry{
		Timestamp: time.Now(),
		Level:     ai.extractLevel(logLine),
		Service:   parsed["service"],
		Host:      parsed["host"],
		Message:   parsed["message"],
		Raw:       logLine,
	}
}

// extractLevel 로그 레벨 추출
func (ai *AIAnalyzer) extractLevel(logLine string) string {
	lowLine := strings.ToLower(logLine)
	if strings.Contains(lowLine, "critical") || strings.Contains(lowLine, "fatal") {
		return "CRITICAL"
	} else if strings.Contains(lowLine, "error") || strings.Contains(lowLine, "err") {
		return "ERROR"
	} else if strings.Contains(lowLine, "warn") {
		return "WARNING"
	} else if strings.Contains(lowLine, "info") {
		return "INFO"
	}
	return "UNKNOWN"
}

// addToBuffer 버퍼에 로그 항목 추가
func (ai *AIAnalyzer) addToBuffer(entry LogEntry) {
	ai.logBuffer = append(ai.logBuffer, entry)
	
	// 버퍼 크기 제한
	if len(ai.logBuffer) > ai.maxBufferSize {
		ai.logBuffer = ai.logBuffer[1:]
	}
	
	// 오래된 항목 제거 (시간 윈도우 기준)
	cutoff := time.Now().Add(-ai.timeWindow)
	for i, entry := range ai.logBuffer {
		if entry.Timestamp.After(cutoff) {
			ai.logBuffer = ai.logBuffer[i:]
			break
		}
	}
}

// extractFeatures 로그 특성 추출
func (ai *AIAnalyzer) extractFeatures(entry LogEntry) LogFeatures {
	features := LogFeatures{}
	
	// IP 주소 추출
	ipPattern := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	features.IPAddresses = ipPattern.FindAllString(entry.Raw, -1)
	
	// HTTP 상태 코드 추출
	statusPattern := regexp.MustCompile(`\b[1-5]\d{2}\b`)
	statusMatches := statusPattern.FindAllString(entry.Raw, -1)
	for _, status := range statusMatches {
		if code, err := strconv.Atoi(status); err == nil {
			features.HTTPStatusCodes = append(features.HTTPStatusCodes, code)
		}
	}
	
	// 응답 시간 추출
	responsePattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*ms`)
	responseMatches := responsePattern.FindAllStringSubmatch(entry.Raw, -1)
	for _, match := range responseMatches {
		if len(match) > 1 {
			if time, err := strconv.ParseFloat(match[1], 64); err == nil {
				features.ResponseTimes = append(features.ResponseTimes, time)
			}
		}
	}
	
	// 사용자 추출
	userPattern := regexp.MustCompile(`(?i)user[:\s=]+(\w+)`)
	userMatches := userPattern.FindAllStringSubmatch(entry.Raw, -1)
	for _, match := range userMatches {
		if len(match) > 1 {
			features.UniqueUsers = append(features.UniqueUsers, match[1])
		}
	}
	
	// SQL 쿼리 감지
	sqlPattern := regexp.MustCompile(`(?i)(select|insert|update|delete|create|drop)\s+`)
	features.SQLQueries = sqlPattern.FindAllString(entry.Raw, -1)
	
	// 심각도 계산
	features.Severity = ai.calculateSeverity(entry)
	
	// 시스템 정보 수집
	features.SystemInfo = ai.collectSystemInfo(features.IPAddresses)
	
	return features
}

// detectAnomalies 이상 패턴 감지
func (ai *AIAnalyzer) detectAnomalies(entry LogEntry) float64 {
	var maxScore float64 = 0.0
	
	// 패턴 매칭
	for _, pattern := range ai.patterns {
		if pattern.Pattern.MatchString(entry.Raw) {
			if pattern.Severity > maxScore {
				maxScore = pattern.Severity
			}
		}
	}
	
	// 빈도 기반 이상 감지
	frequencyScore := ai.analyzeFrequency(entry)
	
	// 시간 기반 이상 감지
	timeScore := ai.analyzeTimePatterns(entry)
	
	// 종합 점수 계산
	finalScore := math.Max(maxScore, math.Max(frequencyScore, timeScore))
	
	return finalScore
}

// analyzeFrequency 빈도 기반 분석
func (ai *AIAnalyzer) analyzeFrequency(entry LogEntry) float64 {
	if len(ai.logBuffer) < 10 {
		return 0.0
	}
	
	// 최근 로그에서 유사한 메시지 빈도 계산
	recentCount := 0
	for _, bufferedEntry := range ai.logBuffer {
		if time.Since(bufferedEntry.Timestamp) <= time.Minute*5 {
			if ai.calculateSimilarity(entry.Message, bufferedEntry.Message) > 0.8 {
				recentCount++
			}
		}
	}
	
	// 비정상적으로 높은 빈도면 점수 증가
	if recentCount > 10 {
		return 6.0 + float64(recentCount-10)*0.1
	}
	
	return 0.0
}

// analyzeTimePatterns 시간 패턴 분석
func (ai *AIAnalyzer) analyzeTimePatterns(entry LogEntry) float64 {
	now := time.Now()
	hour := now.Hour()
	
	// 업무 시간 외 활동 (밤 11시 ~ 오전 6시)
	if hour >= 23 || hour <= 6 {
		if entry.Level == "ERROR" || entry.Level == "CRITICAL" {
			return 5.0 // 야간 시간대 에러는 의심스러움
		}
	}
	
	// 주말 활동
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		if strings.Contains(strings.ToLower(entry.Message), "login") ||
		   strings.Contains(strings.ToLower(entry.Message), "access") {
			return 4.0 // 주말 로그인은 주의 필요
		}
	}
	
	return 0.0
}

// makePredictions 예측 수행
func (ai *AIAnalyzer) makePredictions(entry LogEntry, features LogFeatures) []Prediction {
	predictions := []Prediction{}
	
	// 메모리 관련 예측
	if strings.Contains(strings.ToLower(entry.Message), "memory") {
		predictions = append(predictions, Prediction{
			Event:       "시스템 메모리 부족",
			Probability: 0.75,
			TimeFrame:   "30분 이내",
			Impact:      "서비스 중단 가능성",
		})
	}
	
	// 로그인 실패 패턴 예측
	failedLogins := 0
	for _, bufferedEntry := range ai.logBuffer {
		if strings.Contains(strings.ToLower(bufferedEntry.Message), "failed") &&
		   strings.Contains(strings.ToLower(bufferedEntry.Message), "login") {
			failedLogins++
		}
	}
	
	if failedLogins > 5 {
		predictions = append(predictions, Prediction{
			Event:       "보안 위협 - 무차별 대입 공격",
			Probability: 0.85,
			TimeFrame:   "진행 중",
			Impact:      "계정 탈취 위험",
		})
	}
	
	// 데이터베이스 관련 예측
	if strings.Contains(strings.ToLower(entry.Message), "database") ||
	   strings.Contains(strings.ToLower(entry.Message), "connection") {
		predictions = append(predictions, Prediction{
			Event:       "데이터베이스 성능 저하",
			Probability: 0.60,
			TimeFrame:   "1시간 이내",
			Impact:      "응답 시간 증가",
		})
	}
	
	return predictions
}

// generateRecommendations 추천사항 생성
func (ai *AIAnalyzer) generateRecommendations(entry LogEntry, anomalyScore float64) []string {
	recommendations := []string{}
	
	if anomalyScore >= 8.0 {
		recommendations = append(recommendations, "🚨 즉시 보안팀에 알림")
		recommendations = append(recommendations, "🔒 해당 IP 주소 차단 검토")
		recommendations = append(recommendations, "📊 시스템 리소스 사용량 확인")
	} else if anomalyScore >= 6.0 {
		recommendations = append(recommendations, "⚠️ 모니터링 강화 필요")
		recommendations = append(recommendations, "📈 관련 로그 패턴 분석")
	}
	
	// 서비스별 추천사항
	if strings.Contains(strings.ToLower(entry.Service), "database") {
		recommendations = append(recommendations, "🗄️ 데이터베이스 연결 풀 상태 확인")
		recommendations = append(recommendations, "🔍 슬로우 쿼리 로그 분석")
	}
	
	if strings.Contains(strings.ToLower(entry.Service), "web") {
		recommendations = append(recommendations, "🌐 웹서버 부하 상태 점검")
		recommendations = append(recommendations, "🚀 캐시 상태 확인")
	}
	
	return recommendations
}

// calculateThreatLevel 위협 레벨 계산
func (ai *AIAnalyzer) calculateThreatLevel(anomalyScore float64) string {
	if anomalyScore >= 9.0 {
		return "🔴 CRITICAL"
	} else if anomalyScore >= 7.0 {
		return "🟠 HIGH"
	} else if anomalyScore >= 5.0 {
		return "🟡 MEDIUM"
	} else if anomalyScore >= 3.0 {
		return "🟢 LOW"
	}
	return "⚪ NORMAL"
}

// identifyAffectedSystems 영향받는 시스템 식별
func (ai *AIAnalyzer) identifyAffectedSystems(entry LogEntry) []string {
	systems := []string{}
	
	if entry.Host != "" {
		systems = append(systems, entry.Host)
	}
	
	if entry.Service != "" {
		systems = append(systems, entry.Service)
	}
	
	// IP 주소에서 시스템 추정
	for _, ip := range entry.Features.IPAddresses {
		systems = append(systems, "Host-"+ip)
	}
	
	return systems
}

// calculateConfidence 신뢰도 계산
func (ai *AIAnalyzer) calculateConfidence(anomalyScore float64, features LogFeatures) float64 {
	confidence := 0.5 // 기본 신뢰도
	
	// 패턴 매칭이 확실한 경우
	if anomalyScore >= 8.0 {
		confidence += 0.3
	}
	
	// 여러 특성이 감지된 경우
	if len(features.IPAddresses) > 0 {
		confidence += 0.1
	}
	if len(features.HTTPStatusCodes) > 0 {
		confidence += 0.1
	}
	
	return math.Min(confidence, 1.0)
}

// calculateSeverity 심각도 계산
func (ai *AIAnalyzer) calculateSeverity(entry LogEntry) float64 {
	severity := 1.0
	
	switch entry.Level {
	case "CRITICAL":
		severity = 9.0
	case "ERROR":
		severity = 7.0
	case "WARNING":
		severity = 5.0
	case "INFO":
		severity = 3.0
	}
	
	return severity
}

// calculateSimilarity 유사도 계산 (간단한 문자열 비교)
func (ai *AIAnalyzer) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// 단순 문자열 포함 기반 유사도
	words1 := strings.Fields(strings.ToLower(s1))
	words2 := strings.Fields(strings.ToLower(s2))
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	common := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				common++
				break
			}
		}
	}
	
	return float64(common) / float64(math.Max(float64(len(words1)), float64(len(words2))))
}

// UpdateBaseline 기준선 업데이트
func (ai *AIAnalyzer) UpdateBaseline() {
	if len(ai.logBuffer) < 50 {
		return
	}
	
	// 에러율 계산
	errorCount := 0
	totalCount := len(ai.logBuffer)
	var responseTimes []float64
	uniqueUsers := make(map[string]bool)
	
	for _, entry := range ai.logBuffer {
		if entry.Level == "ERROR" || entry.Level == "CRITICAL" {
			errorCount++
		}
		
		for _, rt := range entry.Features.ResponseTimes {
			responseTimes = append(responseTimes, rt)
		}
		
		for _, user := range entry.Features.UniqueUsers {
			uniqueUsers[user] = true
		}
	}
	
	ai.baselineMetrics.AvgErrorRate = float64(errorCount) / float64(totalCount)
	ai.baselineMetrics.TypicalLogVolume = float64(totalCount)
	ai.baselineMetrics.NormalUserCount = len(uniqueUsers)
	
	if len(responseTimes) > 0 {
		sort.Float64s(responseTimes)
		ai.baselineMetrics.AvgResponseTime = responseTimes[len(responseTimes)/2] // 중간값
	}
	
	ai.baselineMetrics.BaselineUpdatedAt = time.Now()
}

// GetAnalysisReport 분석 보고서 생성
func (ai *AIAnalyzer) GetAnalysisReport() string {
	report := fmt.Sprintf(`
🤖 AI 로그 분석 보고서
===================
📊 기준선 메트릭:
  - 평균 에러율: %.2f%%
  - 평균 응답시간: %.0fms
  - 일반적인 로그 볼륨: %.0f entries/5min
  - 정상 사용자 수: %d명
  - 마지막 업데이트: %s

📈 현재 버퍼:
  - 로그 항목 수: %d
  - 시간 윈도우: %v
  - 알림 임계값: %.1f

🔍 감지 패턴 수: %d개
`,
		ai.baselineMetrics.AvgErrorRate*100,
		ai.baselineMetrics.AvgResponseTime,
		ai.baselineMetrics.TypicalLogVolume,
		ai.baselineMetrics.NormalUserCount,
		ai.baselineMetrics.BaselineUpdatedAt.Format("2006-01-02 15:04:05"),
		len(ai.logBuffer),
		ai.timeWindow,
		ai.alertThreshold,
		len(ai.patterns),
	)
	
	return report
}

// collectSystemInfo 시스템 정보 수집
func (ai *AIAnalyzer) collectSystemInfo(ipAddresses []string) SystemInfo {
	systemInfo := SystemInfo{}
	
	// 컴퓨터 이름 가져오기
	systemInfo.ComputerName = ai.getComputerName()
	
	// IP 주소 분류
	systemInfo.InternalIPs, systemInfo.ExternalIPs = ai.classifyIPs(ipAddresses)
	
	// ASN 정보 수집 (외부 IP에 대해서만)
	systemInfo.ASNData = ai.getASNInfo(systemInfo.ExternalIPs)
	
	return systemInfo
}

// getComputerName 컴퓨터 이름 가져오기
func (ai *AIAnalyzer) getComputerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return hostname
}

// isPrivateIP IP가 사설 IP인지 확인
func (ai *AIAnalyzer) isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	// RFC 1918 사설 IP 범위
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",  // 루프백
		"169.254.0.0/16", // 링크 로컬
	}
	
	for _, rangeStr := range privateRanges {
		_, cidr, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}
	
	return false
}

// classifyIPs IP 주소를 내부/외부로 분류
func (ai *AIAnalyzer) classifyIPs(ipAddresses []string) ([]string, []string) {
	var internalIPs, externalIPs []string
	
	for _, ip := range ipAddresses {
		if ai.isPrivateIP(ip) {
			internalIPs = append(internalIPs, ip)
		} else {
			externalIPs = append(externalIPs, ip)
		}
	}
	
	return internalIPs, externalIPs
}

// getASNInfo ASN 정보 조회 (외부 API 사용)
func (ai *AIAnalyzer) getASNInfo(externalIPs []string) []ASNInfo {
	var asnData []ASNInfo
	
	for _, ip := range externalIPs {
		if ip == "" {
			continue
		}
		
		// ip-api.com을 사용한 ASN 정보 조회
		asnInfo := ai.queryASNInfo(ip)
		if asnInfo.IP != "" {
			asnData = append(asnData, asnInfo)
		}
	}
	
	return asnData
}

// queryASNInfo 단일 IP에 대한 ASN 정보 조회
func (ai *AIAnalyzer) queryASNInfo(ip string) ASNInfo {
	// 무료 API 사용: ip-api.com
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,regionName,city,org,as,query", ip)
	
	resp, err := http.Get(url)
	if err != nil {
		return ASNInfo{IP: ip, ASN: "Unknown", Organization: "Query Failed"}
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ASNInfo{IP: ip, ASN: "Unknown", Organization: "Read Failed"}
	}
	
	var result struct {
		Status      string `json:"status"`
		Message     string `json:"message"`
		Country     string `json:"country"`
		RegionName  string `json:"regionName"`
		City        string `json:"city"`
		Org         string `json:"org"`
		AS          string `json:"as"`
		Query       string `json:"query"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return ASNInfo{IP: ip, ASN: "Unknown", Organization: "Parse Failed"}
	}
	
	if result.Status != "success" {
		return ASNInfo{IP: ip, ASN: "Unknown", Organization: result.Message}
	}
	
	return ASNInfo{
		IP:           result.Query,
		ASN:          result.AS,
		Organization: result.Org,
		Country:      result.Country,
		Region:       result.RegionName,
		City:         result.City,
	}
}

// GenerateDetailedAlert 상세한 알람 메시지 생성
func (ai *AIAnalyzer) GenerateDetailedAlert(result *AIAnalysisResult, entry LogEntry) string {
	alert := fmt.Sprintf(`
🚨 보안 이상 탐지 알람
======================
⚠️  위협 레벨: %s
📊 이상 점수: %.1f/10.0
🕐 탐지 시간: %s

🖥️  시스템 정보:
  📍 컴퓨터명: %s
  🏠 내부 IP: %s
  🌐 외부 IP: %s

`, 
		result.ThreatLevel,
		result.AnomalyScore,
		result.Timestamp.Format("2006-01-02 15:04:05"),
		result.SystemInfo.ComputerName,
		strings.Join(result.SystemInfo.InternalIPs, ", "),
		strings.Join(result.SystemInfo.ExternalIPs, ", "),
	)

	// ASN 정보 추가
	if len(result.SystemInfo.ASNData) > 0 {
		alert += "🔍 ASN 정보:\n"
		for _, asn := range result.SystemInfo.ASNData {
			alert += fmt.Sprintf("  📍 %s\n", asn.IP)
			alert += fmt.Sprintf("    🏢 조직: %s\n", asn.Organization)
			alert += fmt.Sprintf("    🌍 국가: %s, %s, %s\n", asn.Country, asn.Region, asn.City)
			alert += fmt.Sprintf("    🔢 ASN: %s\n", asn.ASN)
			alert += "\n"
		}
	}

	// 로그 정보
	alert += fmt.Sprintf(`
📋 로그 정보:
  📝 레벨: %s
  🏷️  서비스: %s
  🖥️  호스트: %s
  💬 메시지: %s

`, 
		entry.Level,
		entry.Service,
		entry.Host,
		entry.Message,
	)

	// 예측 정보
	if len(result.Predictions) > 0 {
		alert += "🔮 위험 예측:\n"
		for _, pred := range result.Predictions {
			alert += fmt.Sprintf("  ⚡ %s (확률: %.0f%%, %s)\n", 
				pred.Event, pred.Probability*100, pred.TimeFrame)
			alert += fmt.Sprintf("    💥 영향: %s\n", pred.Impact)
		}
		alert += "\n"
	}

	// 권장사항
	if len(result.Recommendations) > 0 {
		alert += "💡 권장사항:\n"
		for _, rec := range result.Recommendations {
			alert += fmt.Sprintf("  • %s\n", rec)
		}
		alert += "\n"
	}

	// 영향받는 시스템
	if len(result.AffectedSystems) > 0 {
		alert += fmt.Sprintf("🎯 영향받는 시스템: %s\n", 
			strings.Join(result.AffectedSystems, ", "))
	}

	alert += fmt.Sprintf("🎯 신뢰도: %.0f%%\n", result.Confidence*100)

	return alert
}

// PerformExpertDiagnosis 전문가 진단 수행
func (ai *AIAnalyzer) PerformExpertDiagnosis(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) ExpertDiagnosis {
	serverDiagnosis := ai.performServerExpertDiagnosis(entry, features, systemMetrics)
	computerDiagnosis := ai.performComputerExpertDiagnosis(entry, features, systemMetrics)
	
	overallHealth := ai.calculateOverallHealth(serverDiagnosis, computerDiagnosis)
	criticalIssues := ai.identifyCriticalIssues(serverDiagnosis, computerDiagnosis)
	maintenanceTips := ai.generateMaintenanceTips(serverDiagnosis, computerDiagnosis)
	performanceScore := ai.calculatePerformanceScore(serverDiagnosis, computerDiagnosis)
	
	return ExpertDiagnosis{
		ServerExpert:    serverDiagnosis,
		ComputerExpert:  computerDiagnosis,
		OverallHealth:   overallHealth,
		CriticalIssues:  criticalIssues,
		MaintenanceTips: maintenanceTips,
		PerformanceScore: performanceScore,
	}
}

// performServerExpertDiagnosis 서버 전문가 진단 수행
func (ai *AIAnalyzer) performServerExpertDiagnosis(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) ServerExpertDiagnosis {
	// 서버 성능 분석
	performanceScore := ai.analyzeServerPerformance(features, systemMetrics)
	serverHealth := ai.determineServerHealth(performanceScore)
	
	// 보안 상태 분석
	securityStatus := ai.analyzeSecurityStatus(entry, features)
	
	// 네트워크 건강도 분석
	networkHealth := ai.analyzeNetworkHealth(features, systemMetrics)
	
	// 이슈 식별
	issues := ai.identifyServerIssues(entry, features, systemMetrics)
	
	// 권장사항 생성
	recommendations := ai.generateServerRecommendations(entry, features, systemMetrics)
	
	// 위험도 평가
	riskLevel := ai.calculateServerRiskLevel(entry, features, systemMetrics)
	
	return ServerExpertDiagnosis{
		ServerHealth:     serverHealth,
		PerformanceScore: performanceScore,
		SecurityStatus:   securityStatus,
		NetworkHealth:    networkHealth,
		Issues:           issues,
		Recommendations:  recommendations,
		RiskLevel:        riskLevel,
	}
}

// performComputerExpertDiagnosis 컴퓨터 전문가 진단 수행
func (ai *AIAnalyzer) performComputerExpertDiagnosis(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) ComputerExpertDiagnosis {
	// 하드웨어 건강도 분석
	hardwareHealth := ai.analyzeHardwareHealth(systemMetrics)
	
	// 소프트웨어 상태 분석
	softwareStatus := ai.analyzeSoftwareStatus(entry, features)
	
	// 시스템 안정성 분석
	systemStability := ai.analyzeSystemStability(entry, features, systemMetrics)
	
	// 리소스 사용량 분석
	resourceUsage := ai.analyzeResourceUsage(systemMetrics)
	
	// 이슈 식별
	issues := ai.identifyComputerIssues(entry, features, systemMetrics)
	
	// 권장사항 생성
	recommendations := ai.generateComputerRecommendations(entry, features, systemMetrics)
	
	// 유지보수 필요성 평가
	maintenanceNeeded := ai.evaluateMaintenanceNeeds(entry, features, systemMetrics)
	
	return ComputerExpertDiagnosis{
		HardwareHealth:   hardwareHealth,
		SoftwareStatus:   softwareStatus,
		SystemStability:  systemStability,
		ResourceUsage:    resourceUsage,
		Issues:           issues,
		Recommendations:  recommendations,
		MaintenanceNeeded: maintenanceNeeded,
	}
}

// analyzeServerPerformance 서버 성능 분석
func (ai *AIAnalyzer) analyzeServerPerformance(features LogFeatures, systemMetrics *SystemMetrics) float64 {
	score := 100.0
	
	// CPU 사용률 기반 점수 조정
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 80 {
		score -= 30
	} else if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 60 {
		score -= 15
	}
	
	// 메모리 사용률 기반 점수 조정
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 90 {
		score -= 25
	} else if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 75 {
		score -= 10
	}
	
	// 에러율 기반 점수 조정
	if features.ErrorCount > 10 {
		score -= 20
	} else if features.ErrorCount > 5 {
		score -= 10
	}
	
	// 응답시간 기반 점수 조정
	if len(features.ResponseTimes) > 0 {
		avgResponseTime := 0.0
		for _, rt := range features.ResponseTimes {
			avgResponseTime += rt
		}
		avgResponseTime /= float64(len(features.ResponseTimes))
		
		if avgResponseTime > 2000 {
			score -= 20
		} else if avgResponseTime > 1000 {
			score -= 10
		}
	}
	
	return math.Max(0, score)
}

// determineServerHealth 서버 건강도 결정
func (ai *AIAnalyzer) determineServerHealth(performanceScore float64) string {
	if performanceScore >= 90 {
		return "Excellent"
	} else if performanceScore >= 75 {
		return "Good"
	} else if performanceScore >= 60 {
		return "Fair"
	} else if performanceScore >= 40 {
		return "Poor"
	} else {
		return "Critical"
	}
}

// analyzeSecurityStatus 보안 상태 분석
func (ai *AIAnalyzer) analyzeSecurityStatus(entry LogEntry, features LogFeatures) string {
	// 보안 관련 키워드 검사
	securityKeywords := []string{"failed", "unauthorized", "denied", "attack", "injection", "brute"}
	securityScore := 0
	
	for _, keyword := range securityKeywords {
		if strings.Contains(strings.ToLower(entry.Message), keyword) {
			securityScore++
		}
	}
	
	if securityScore >= 3 {
		return "High Risk"
	} else if securityScore >= 1 {
		return "Medium Risk"
	} else {
		return "Secure"
	}
}

// analyzeNetworkHealth 네트워크 건강도 분석
func (ai *AIAnalyzer) analyzeNetworkHealth(features LogFeatures, systemMetrics *SystemMetrics) string {
	// 네트워크 관련 이슈 검사
	networkIssues := 0
	
	if len(features.IPAddresses) > 10 {
		networkIssues++
	}
	
	if features.Frequency > 100 {
		networkIssues++
	}
	
	if networkIssues >= 2 {
		return "Poor"
	} else if networkIssues >= 1 {
		return "Fair"
	} else {
		return "Good"
	}
}

// identifyServerIssues 서버 이슈 식별
func (ai *AIAnalyzer) identifyServerIssues(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) []string {
	var issues []string
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 80 {
		issues = append(issues, "높은 CPU 사용률")
	}
	
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 90 {
		issues = append(issues, "메모리 부족")
	}
	
	if features.ErrorCount > 10 {
		issues = append(issues, "과도한 에러 발생")
	}
	
	if strings.Contains(strings.ToLower(entry.Message), "timeout") {
		issues = append(issues, "서비스 응답 지연")
	}
	
	return issues
}

// generateServerRecommendations 서버 권장사항 생성
func (ai *AIAnalyzer) generateServerRecommendations(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) []string {
	var recommendations []string
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 80 {
		recommendations = append(recommendations, "CPU 사용률이 높습니다. 불필요한 프로세스를 종료하거나 서버 리소스를 확장하세요.")
	}
	
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 90 {
		recommendations = append(recommendations, "메모리 사용률이 높습니다. 메모리 정리 또는 확장을 고려하세요.")
	}
	
	if features.ErrorCount > 10 {
		recommendations = append(recommendations, "에러 로그가 많습니다. 애플리케이션 로그를 확인하고 문제를 해결하세요.")
	}
	
	if len(features.IPAddresses) > 10 {
		recommendations = append(recommendations, "다양한 IP에서 접근이 감지됩니다. 보안 설정을 검토하세요.")
	}
	
	return recommendations
}

// calculateServerRiskLevel 서버 위험도 계산
func (ai *AIAnalyzer) calculateServerRiskLevel(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) string {
	riskScore := 0
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 90 {
		riskScore += 3
	}
	
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 95 {
		riskScore += 3
	}
	
	if features.ErrorCount > 20 {
		riskScore += 2
	}
	
	if strings.Contains(strings.ToLower(entry.Message), "attack") {
		riskScore += 4
	}
	
	if riskScore >= 6 {
		return "Critical"
	} else if riskScore >= 4 {
		return "High"
	} else if riskScore >= 2 {
		return "Medium"
	} else {
		return "Low"
	}
}

// analyzeHardwareHealth 하드웨어 건강도 분석
func (ai *AIAnalyzer) analyzeHardwareHealth(systemMetrics *SystemMetrics) string {
	if systemMetrics == nil {
		return "Unknown"
	}
	
	// CPU 온도 체크
	if systemMetrics.Temperature.CPUTemp > 80 {
		return "Critical"
	} else if systemMetrics.Temperature.CPUTemp > 70 {
		return "Poor"
	} else if systemMetrics.Temperature.CPUTemp > 60 {
		return "Fair"
	} else {
		return "Good"
	}
}

// analyzeSoftwareStatus 소프트웨어 상태 분석
func (ai *AIAnalyzer) analyzeSoftwareStatus(entry LogEntry, features LogFeatures) string {
	if features.CriticalCount > 5 {
		return "Critical"
	} else if features.ErrorCount > 10 {
		return "Poor"
	} else if features.WarningCount > 5 {
		return "Fair"
	} else {
		return "Good"
	}
}

// analyzeSystemStability 시스템 안정성 분석
func (ai *AIAnalyzer) analyzeSystemStability(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) string {
	stabilityScore := 0
	
	if systemMetrics != nil && systemMetrics.LoadAverage.Load1Min > 10 {
		stabilityScore += 2
	}
	
	if features.CriticalCount > 3 {
		stabilityScore += 3
	}
	
	if features.ErrorCount > 15 {
		stabilityScore += 2
	}
	
	if stabilityScore >= 5 {
		return "Unstable"
	} else if stabilityScore >= 3 {
		return "Fair"
	} else {
		return "Stable"
	}
}

// analyzeResourceUsage 리소스 사용량 분석
func (ai *AIAnalyzer) analyzeResourceUsage(systemMetrics *SystemMetrics) string {
	if systemMetrics == nil {
		return "Unknown"
	}
	
	if systemMetrics.CPU.UsagePercent > 90 || systemMetrics.Memory.UsagePercent > 95 {
		return "Critical"
	} else if systemMetrics.CPU.UsagePercent > 80 || systemMetrics.Memory.UsagePercent > 85 {
		return "High"
	} else if systemMetrics.CPU.UsagePercent > 60 || systemMetrics.Memory.UsagePercent > 70 {
		return "Moderate"
	} else {
		return "Normal"
	}
}

// identifyComputerIssues 컴퓨터 이슈 식별
func (ai *AIAnalyzer) identifyComputerIssues(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) []string {
	var issues []string
	
	if systemMetrics != nil && systemMetrics.Temperature.CPUTemp > 75 {
		issues = append(issues, "CPU 온도가 높습니다")
	}
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 90 {
		issues = append(issues, "CPU 사용률이 매우 높습니다")
	}
	
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 95 {
		issues = append(issues, "메모리 사용률이 매우 높습니다")
	}
	
	if features.CriticalCount > 3 {
		issues = append(issues, "치명적 오류가 발생하고 있습니다")
	}
	
	return issues
}

// generateComputerRecommendations 컴퓨터 권장사항 생성
func (ai *AIAnalyzer) generateComputerRecommendations(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) []string {
	var recommendations []string
	
	if systemMetrics != nil && systemMetrics.Temperature.CPUTemp > 75 {
		recommendations = append(recommendations, "CPU 온도가 높습니다. 쿨링 시스템을 점검하고 먼지를 청소하세요.")
	}
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 90 {
		recommendations = append(recommendations, "CPU 사용률이 매우 높습니다. 불필요한 프로그램을 종료하세요.")
	}
	
	if systemMetrics != nil && systemMetrics.Memory.UsagePercent > 95 {
		recommendations = append(recommendations, "메모리 사용률이 매우 높습니다. 메모리 정리를 수행하세요.")
	}
	
	if features.CriticalCount > 3 {
		recommendations = append(recommendations, "치명적 오류가 발생하고 있습니다. 시스템 로그를 확인하고 문제를 해결하세요.")
	}
	
	return recommendations
}

// evaluateMaintenanceNeeds 유지보수 필요성 평가
func (ai *AIAnalyzer) evaluateMaintenanceNeeds(entry LogEntry, features LogFeatures, systemMetrics *SystemMetrics) bool {
	if systemMetrics != nil && systemMetrics.Temperature.CPUTemp > 80 {
		return true
	}
	
	if systemMetrics != nil && systemMetrics.CPU.UsagePercent > 95 {
		return true
	}
	
	if features.CriticalCount > 5 {
		return true
	}
	
	return false
}

// calculateOverallHealth 전체 시스템 건강도 계산
func (ai *AIAnalyzer) calculateOverallHealth(server ServerExpertDiagnosis, computer ComputerExpertDiagnosis) string {
	serverScore := 0
	computerScore := 0
	
	// 서버 점수 계산
	switch server.ServerHealth {
	case "Excellent":
		serverScore = 5
	case "Good":
		serverScore = 4
	case "Fair":
		serverScore = 3
	case "Poor":
		serverScore = 2
	case "Critical":
		serverScore = 1
	}
	
	// 컴퓨터 점수 계산
	switch computer.HardwareHealth {
	case "Good":
		computerScore = 5
	case "Fair":
		computerScore = 3
	case "Poor":
		computerScore = 2
	case "Critical":
		computerScore = 1
	}
	
	totalScore := float64(serverScore + computerScore) / 2.0
	
	if totalScore >= 4.5 {
		return "Excellent"
	} else if totalScore >= 3.5 {
		return "Good"
	} else if totalScore >= 2.5 {
		return "Fair"
	} else if totalScore >= 1.5 {
		return "Poor"
	} else {
		return "Critical"
	}
}

// identifyCriticalIssues 긴급 이슈 식별
func (ai *AIAnalyzer) identifyCriticalIssues(server ServerExpertDiagnosis, computer ComputerExpertDiagnosis) []string {
	var issues []string
	
	if server.RiskLevel == "Critical" {
		issues = append(issues, "서버 위험도가 Critical입니다")
	}
	
	if computer.HardwareHealth == "Critical" {
		issues = append(issues, "하드웨어 상태가 Critical입니다")
	}
	
	if server.ServerHealth == "Critical" {
		issues = append(issues, "서버 건강도가 Critical입니다")
	}
	
	return issues
}

// generateMaintenanceTips 유지보수 팁 생성
func (ai *AIAnalyzer) generateMaintenanceTips(server ServerExpertDiagnosis, computer ComputerExpertDiagnosis) []string {
	var tips []string
	
	if computer.MaintenanceNeeded {
		tips = append(tips, "즉시 유지보수가 필요합니다")
	}
	
	if server.RiskLevel == "High" || server.RiskLevel == "Critical" {
		tips = append(tips, "서버 보안 점검이 필요합니다")
	}
	
	if computer.HardwareHealth == "Poor" || computer.HardwareHealth == "Critical" {
		tips = append(tips, "하드웨어 점검이 필요합니다")
	}
	
	return tips
}

// calculatePerformanceScore 성능 점수 계산
func (ai *AIAnalyzer) calculatePerformanceScore(server ServerExpertDiagnosis, computer ComputerExpertDiagnosis) float64 {
	return (server.PerformanceScore + 80.0) / 2 // 컴퓨터 점수는 기본 80점으로 가정
} 