/*
Gemini AI Service
=================

Google Gemini API를 이용한 고급 AI 분석 서비스

주요 기능:
- 실시간 시스템 진단
- 로그 패턴 분석
- 보안 위협 감지
- 전문가 권장사항 생성
- 자연어 기반 시스템 분석

작성자: Lambda-X AI Team
버전: 1.0.0
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GeminiConfig Gemini API 설정 구조체
type GeminiConfig struct {
	APIKey     string `json:"api_key"`
	Model      string `json:"model"`
	MaxTokens  int    `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Enabled    bool   `json:"enabled"`
}

// GeminiRequest Gemini API 요청 구조체
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent Gemini API 콘텐츠 구조체
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart Gemini API 파트 구조체
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig Gemini API 생성 설정
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	TopK           int     `json:"topK"`
	TopP           float64 `json:"topP"`
	MaxOutputTokens int    `json:"maxOutputTokens"`
}

// GeminiResponse Gemini API 응답 구조체
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	PromptFeedback GeminiPromptFeedback `json:"promptFeedback,omitempty"`
}

// GeminiCandidate Gemini API 후보 응답
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
	FinishReason string `json:"finishReason"`
	Index int `json:"index"`
}

// GeminiPromptFeedback Gemini API 프롬프트 피드백
type GeminiPromptFeedback struct {
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings"`
}

// GeminiSafetyRating Gemini API 안전성 평가
type GeminiSafetyRating struct {
	Category string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiService Gemini AI 서비스 구조체
type GeminiService struct {
	config     *GeminiConfig
	httpClient *http.Client
	baseURL    string
}

// NewGeminiService Gemini 서비스 생성자
func NewGeminiService(config *GeminiConfig) *GeminiService {
	return &GeminiService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
	}
}

// AnalyzeSystemDiagnosis 시스템 진단 분석
func (gs *GeminiService) AnalyzeSystemDiagnosis(metrics SystemMetrics) (string, error) {
	if !gs.config.Enabled || gs.config.APIKey == "" {
		return gs.generateBasicDiagnosis(metrics), nil
	}

	prompt := gs.buildSystemDiagnosisPrompt(metrics)
	return gs.callGeminiAPI(prompt)
}

// AnalyzeLogPattern 로그 패턴 분석
func (gs *GeminiService) AnalyzeLogPattern(logLine string, context map[string]string) (string, error) {
	if !gs.config.Enabled || gs.config.APIKey == "" {
		return gs.generateBasicLogAnalysis(logLine, context), nil
	}

	prompt := gs.buildLogAnalysisPrompt(logLine, context)
	return gs.callGeminiAPI(prompt)
}

// AnalyzeSecurityThreat 보안 위협 분석
func (gs *GeminiService) AnalyzeSecurityThreat(threatData map[string]interface{}) (string, error) {
	if !gs.config.Enabled || gs.config.APIKey == "" {
		return gs.generateBasicSecurityAnalysis(threatData), nil
	}

	prompt := gs.buildSecurityAnalysisPrompt(threatData)
	return gs.callGeminiAPI(prompt)
}

// callGeminiAPI Gemini API 호출
func (gs *GeminiService) callGeminiAPI(prompt string) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", gs.baseURL, gs.config.Model, gs.config.APIKey)
	
	request := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     gs.config.Temperature,
			TopK:           40,
			TopP:           0.95,
			MaxOutputTokens: gs.config.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := gs.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(body))
	}

	var response GeminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// buildSystemDiagnosisPrompt 시스템 진단 프롬프트 생성
func (gs *GeminiService) buildSystemDiagnosisPrompt(metrics SystemMetrics) string {
	return fmt.Sprintf(`당신은 시스템 관리 전문가입니다. 다음 시스템 메트릭을 분석하고 전문적인 진단과 권장사항을 제공해주세요.

시스템 정보:
- 호스트명: %s
- 사설 IP: %s
- 공인 IP: %s

CPU 정보:
- 사용률: %.1f%%
- 사용자: %.1f%%, 시스템: %.1f%%, 대기: %.1f%%
- 코어 수: %d개

메모리 정보:
- 사용률: %.1f%%
- 총 메모리: %.1f GB
- 사용 중: %.1f GB
- 사용 가능: %.1f GB

온도 정보:
- CPU 온도: %.1f°C

프로세스 정보:
- 총 프로세스 수: %d개

다음 형식으로 전문가 진단을 제공해주세요:

🔬 AI 전문가 진단 결과
=====================
📊 전반적인 시스템 건강도: [EXCELLENT/GOOD/FAIR/POOR/CRITICAL]
⚠️  발견된 문제점:
  [구체적인 문제점들]

💡 전문가 권장사항:
==================
[구체적인 해결 방법들]

🔧 즉시 실행 가능한 명령어:
==========================
[실제 터미널 명령어들]

📈 성능 최적화 팁:
==================
[시스템 최적화 조언들]

한국어로 답변해주세요.`,
		metrics.IPInfo.Hostname,
		formatIPListForReport(metrics.IPInfo.PrivateIPs),
		formatIPListForReport(metrics.IPInfo.PublicIPs),
		metrics.CPU.UsagePercent,
		metrics.CPU.UserPercent, metrics.CPU.SystemPercent, metrics.CPU.IdlePercent,
		metrics.CPU.Cores,
		metrics.Memory.UsagePercent,
		metrics.Memory.TotalMB/1024,
		metrics.Memory.UsedMB/1024,
		metrics.Memory.AvailableMB/1024,
		metrics.Temperature.CPUTemp,
		metrics.ProcessCount.Total)
}

// buildLogAnalysisPrompt 로그 분석 프롬프트 생성
func (gs *GeminiService) buildLogAnalysisPrompt(logLine string, context map[string]string) string {
	return fmt.Sprintf(`당신은 보안 전문가입니다. 다음 로그 라인을 분석하고 보안 위협을 평가해주세요.

로그 라인: %s
컨텍스트: %v

다음 형식으로 분석해주세요:

🔍 로그 분석 결과
=================
📊 위협 레벨: [LOW/MEDIUM/HIGH/CRITICAL]
🎯 위협 유형: [구체적인 위협 유형]
💡 분석: [상세한 분석 내용]
🚨 권장사항: [대응 방안]

한국어로 답변해주세요.`,
		logLine, context)
}

// buildSecurityAnalysisPrompt 보안 분석 프롬프트 생성
func (gs *GeminiService) buildSecurityAnalysisPrompt(threatData map[string]interface{}) string {
	threatJSON, _ := json.Marshal(threatData)
	
	return fmt.Sprintf(`당신은 사이버 보안 전문가입니다. 다음 보안 위협 데이터를 분석하고 대응 방안을 제시해주세요.

위협 데이터: %s

다음 형식으로 분석해주세요:

🚨 보안 위협 분석
=================
📊 위협 등급: [LOW/MEDIUM/HIGH/CRITICAL]
🎯 공격 유형: [구체적인 공격 유형]
💥 잠재적 영향: [시스템에 미칠 수 있는 영향]
🛡️  대응 방안: [구체적인 대응 방법]
📈 예방 조치: [향후 예방을 위한 조치]

한국어로 답변해주세요.`,
		string(threatJSON))
}

// generateBasicDiagnosis 기본 진단 생성 (API 없을 때)
func (gs *GeminiService) generateBasicDiagnosis(metrics SystemMetrics) string {
	return fmt.Sprintf(`🔬 AI 전문가 진단 결과 (기본 모드)
=====================
📊 전반적인 시스템 건강도: %s
⚠️  발견된 문제점:
%s

💡 전문가 권장사항:
==================
%s

🔧 즉시 실행 가능한 명령어:
==========================
• 시스템 상태 확인: ` + "`top -l 1`" + `
• 메모리 사용량: ` + "`vm_stat`" + `
• 디스크 사용량: ` + "`df -h`" + `
• 네트워크 상태: ` + "`ifconfig`" + `
• 프로세스 확인: ` + "`ps aux --sort=-%%cpu | head -10`" + `

📈 성능 최적화 팁:
==================
• 정기적인 시스템 재부팅으로 메모리 정리
• 불필요한 시작 프로그램 비활성화
• 디스크 정리 및 최적화
• 네트워크 연결 상태 모니터링

💡 Gemini API 키를 설정하면 더 정교한 AI 진단을 받을 수 있습니다.`,
		gs.getOverallHealth(metrics),
		gs.getIssues(metrics),
		gs.getRecommendations(metrics))
}

// generateBasicLogAnalysis 기본 로그 분석 생성
func (gs *GeminiService) generateBasicLogAnalysis(logLine string, context map[string]string) string {
	return fmt.Sprintf(`🔍 로그 분석 결과 (기본 모드)
=================
📊 위협 레벨: %s
🎯 위협 유형: %s
💡 분석: 기본 패턴 매칭을 통한 분석
🚨 권장사항: 로그 모니터링 강화

💡 Gemini API 키를 설정하면 더 정교한 AI 분석을 받을 수 있습니다.`,
		gs.getThreatLevel(logLine),
		gs.getThreatType(logLine))
}

// generateBasicSecurityAnalysis 기본 보안 분석 생성
func (gs *GeminiService) generateBasicSecurityAnalysis(threatData map[string]interface{}) string {
	return fmt.Sprintf(`🚨 보안 위협 분석 (기본 모드)
=================
📊 위협 등급: MEDIUM
🎯 공격 유형: 패턴 기반 감지
💥 잠재적 영향: 시스템 보안 위험
🛡️  대응 방안: 즉시 보안팀에 알림
📈 예방 조치: 로그 모니터링 강화

💡 Gemini API 키를 설정하면 더 정교한 AI 분석을 받을 수 있습니다.`)
}

// getOverallHealth 전반적인 건강도 평가
func (gs *GeminiService) getOverallHealth(metrics SystemMetrics) string {
	if metrics.CPU.UsagePercent > 80 || metrics.Memory.UsagePercent > 90 {
		return "🔴 CRITICAL"
	} else if metrics.CPU.UsagePercent > 60 || metrics.Memory.UsagePercent > 80 {
		return "🟡 FAIR"
	} else {
		return "🟢 EXCELLENT"
	}
}

// getIssues 발견된 문제점
func (gs *GeminiService) getIssues(metrics SystemMetrics) string {
	var issues []string
	
	if metrics.CPU.UsagePercent > 80 {
		issues = append(issues, "  🔴 CPU 사용률이 매우 높습니다")
	} else if metrics.CPU.UsagePercent > 60 {
		issues = append(issues, "  🟡 CPU 사용률이 높습니다")
	}
	
	if metrics.Memory.UsagePercent > 90 {
		issues = append(issues, "  🔴 메모리 사용률이 매우 높습니다")
	} else if metrics.Memory.UsagePercent > 80 {
		issues = append(issues, "  🟡 메모리 사용률이 높습니다")
	}
	
	if len(issues) == 0 {
		return "  ✅ 특별한 문제점이 발견되지 않았습니다"
	}
	
	return strings.Join(issues, "\n")
}

// getRecommendations 권장사항
func (gs *GeminiService) getRecommendations(metrics SystemMetrics) string {
	var recommendations []string
	
	if metrics.CPU.UsagePercent > 60 {
		recommendations = append(recommendations, "• CPU 집약적 프로세스 모니터링")
	} else {
		recommendations = append(recommendations, "✅ CPU 상태 양호")
	}
	
	if metrics.Memory.UsagePercent > 80 {
		recommendations = append(recommendations, "• 메모리 누수 확인: `ps aux --sort=-%mem`")
		recommendations = append(recommendations, "• 스왑 사용량 확인: `vm_stat`")
	} else {
		recommendations = append(recommendations, "✅ 메모리 상태 양호")
	}
	
	return strings.Join(recommendations, "\n")
}

// getThreatLevel 위협 레벨 평가
func (gs *GeminiService) getThreatLevel(logLine string) string {
	lowLine := strings.ToLower(logLine)
	
	if strings.Contains(lowLine, "error") || strings.Contains(lowLine, "critical") {
		return "🔴 CRITICAL"
	} else if strings.Contains(lowLine, "warning") || strings.Contains(lowLine, "failed") {
		return "🟡 MEDIUM"
	} else {
		return "🟢 LOW"
	}
}

// getThreatType 위협 유형 평가
func (gs *GeminiService) getThreatType(logLine string) string {
	lowLine := strings.ToLower(logLine)
	
	if strings.Contains(lowLine, "sql") || strings.Contains(lowLine, "injection") {
		return "SQL 인젝션 공격"
	} else if strings.Contains(lowLine, "login") || strings.Contains(lowLine, "auth") {
		return "인증 실패"
	} else if strings.Contains(lowLine, "error") {
		return "시스템 오류"
	} else {
		return "일반 로그"
	}
} 