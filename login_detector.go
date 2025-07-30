/*
Login Pattern Detection Module
=============================

고급 로그인 패턴 감지 및 보안 모니터링 서비스

주요 기능:
- SSH 로그인 성공/실패 감지
- Sudo 명령 실행 감지
- 웹 로그인 패턴 인식
- 무차별 대입 공격(Brute Force) 탐지
- 비정상적인 로그인 시도 분석
- IP 주소 기반 지리적 위치 추적

감지 패턴:
- SSH 인증 성공: "Accepted password/publickey for user from IP"
- SSH 인증 실패: "Failed password for user from IP"
- Sudo 실행: "user : TTY=pts/0 ; PWD=/path ; USER=root ; COMMAND=/bin/cmd"
- 웹 로그인: "login user:username from:IP" 패턴
- 인증 실패: "authentication failure", "invalid password" 등

보안 분석:
- 동일 IP에서 반복 실패 시도 탐지
- 비정상적인 시간대 로그인 감지
- 권한 상승 시도 모니터링
*/
package main

import (
	"encoding/json" // JSON 파싱
	"fmt"           // 문자열 포맷팅
	"io"            // I/O 인터페이스
	"net"           // 네트워크 처리
	"net/http"      // HTTP 클라이언트
	"regexp"        // 정규식 패턴 매칭
	"strings"       // 문자열 처리 및 검색
	"sync"          // 동기화 (뮤텍스)
	"time"          // 시간 처리
)

// LoginDetector 로그인 패턴 감지 서비스
// 시스템 리소스 정보와 IP 위치 정보를 포함한 고급 로그인 모니터링
// 10분 간격 알림 제한 기능 포함
type LoginDetector struct {
	logger        Logger         // 로깅 인터페이스
	systemMonitor *SystemMonitor // 시스템 메트릭 수집기 (선택적)
	
	// Alert throttling 알림 제한 관련 필드
	alertHistory  map[string]time.Time // 알림 히스토리 (사용자@IP -> 마지막 알림 시간)
	alertMutex    sync.RWMutex         // 알림 히스토리 동시 접근 보호
	alertInterval time.Duration        // 알림 간격 설정 (기본 10분)
}

// LoginInfo 로그인 정보 구조체 (시스템 리소스 정보 포함)
type LoginInfo struct {
	Status       string           // 로그인 상태 (accepted, failed, sudo 등)
	User         string           // 사용자명
	IP           string           // 접속 IP 주소
	Method       string           // 인증 방법 (ssh, password, publickey 등)
	Command      string           // 실행된 명령어 (sudo의 경우)
	Success      bool             // 로그인 성공 여부
	SystemInfo   SystemMetrics    // 로그인 시점의 시스템 리소스 정보
	IPDetails    *IPLocationInfo  // IP 주소 상세 정보 (지리적 위치 등)
	Timestamp    time.Time        // 로그인 감지 시각
	ShouldAlert  bool             // 알림 전송 여부 (10분 간격 제한 적용 결과)
}

// IPLocationInfo IP 주소 위치 및 상세 정보
type IPLocationInfo struct {
	IP           string `json:"ip"`           // IP 주소
	Country      string `json:"country"`      // 국가
	Region       string `json:"region"`       // 지역/주
	City         string `json:"city"`         // 도시
	Organization string `json:"organization"` // 소속 기관/ISP
	ASN          string `json:"asn"`          // ASN 번호
	IsPrivate    bool   `json:"is_private"`   // 사설 IP 여부
	Threat       string `json:"threat"`       // 위험도 평가
}

// NewLoginDetector 새로운 로그인 감지 서비스 생성
// 10분 간격 알림 제한 기능이 포함된 고급 로그인 모니터링 서비스
func NewLoginDetector(logger Logger) *LoginDetector {
	return &LoginDetector{
		logger:        logger,
		systemMonitor: nil, // 나중에 SetSystemMonitor로 설정 가능
		alertHistory:  make(map[string]time.Time), // 알림 히스토리 초기화
		alertInterval: DefaultLoginAlertInterval,   // 기본 10분 간격
	}
}

// SetSystemMonitor 시스템 모니터 설정 (리소스 정보 수집용)
func (ld *LoginDetector) SetSystemMonitor(sm *SystemMonitor) {
	ld.systemMonitor = sm
}

// SetAlertInterval 알림 간격 설정 (기본 10분)
func (ld *LoginDetector) SetAlertInterval(interval time.Duration) {
	ld.alertMutex.Lock()
	defer ld.alertMutex.Unlock()
	ld.alertInterval = interval
}

// shouldSendAlert 알림 전송 여부 확인 (10분 간격 제한 적용)
// 동일한 사용자@IP 조합에 대해 설정된 간격 내에는 중복 알림 방지
func (ld *LoginDetector) shouldSendAlert(loginInfo *LoginInfo) bool {
	// 중요한 이벤트는 더 짧은 간격으로 알림 (실패한 로그인, sudo 등)
	var checkInterval time.Duration
	if !loginInfo.Success || loginInfo.Status == "sudo" {
		checkInterval = CriticalAlertInterval // 2분 간격
	} else {
		checkInterval = ld.alertInterval // 기본 10분 간격
	}
	
	// 사용자@IP 조합으로 고유 키 생성
	alertKey := fmt.Sprintf("%s@%s", loginInfo.User, loginInfo.IP)
	
	ld.alertMutex.RLock()
	lastAlert, exists := ld.alertHistory[alertKey]
	ld.alertMutex.RUnlock()
	
	now := time.Now()
	
	// 첫 번째 알림이거나 간격이 지난 경우 알림 전송
	if !exists || now.Sub(lastAlert) >= checkInterval {
		// 알림 히스토리 업데이트
		ld.alertMutex.Lock()
		ld.alertHistory[alertKey] = now
		ld.alertMutex.Unlock()
		
		// 주기적으로 오래된 히스토리 정리
		go ld.cleanupAlertHistory()
		
		return true
	}
	
	return false
}

// cleanupAlertHistory 오래된 알림 히스토리 정리 (메모리 사용량 최적화)
func (ld *LoginDetector) cleanupAlertHistory() {
	ld.alertMutex.Lock()
	defer ld.alertMutex.Unlock()
	
	now := time.Now()
	cutoffTime := now.Add(-AlertHistoryCleanupInterval) // 1시간 이전 항목 삭제
	
	for key, timestamp := range ld.alertHistory {
		if timestamp.Before(cutoffTime) {
			delete(ld.alertHistory, key)
		}
	}
	
	// 히스토리 크기가 최대 크기를 초과하면 가장 오래된 항목들 삭제
	if len(ld.alertHistory) > MaxAlertHistorySize {
		// 타임스탬프 기준으로 정렬하여 오래된 항목부터 삭제
		type alertEntry struct {
			key       string
			timestamp time.Time
		}
		
		var entries []alertEntry
		for key, timestamp := range ld.alertHistory {
			entries = append(entries, alertEntry{key, timestamp})
		}
		
		// 타임스탬프 순으로 정렬 (오래된 것부터)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].timestamp.After(entries[j].timestamp) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
		
		// 최대 크기를 초과하는 오래된 항목들 삭제
		deleteCount := len(entries) - MaxAlertHistorySize
		for i := 0; i < deleteCount; i++ {
			delete(ld.alertHistory, entries[i].key)
		}
	}
}

// DetectLoginPattern 로그인 패턴 감지
func (ld *LoginDetector) DetectLoginPattern(line string) (bool, *LoginInfo) {
	line = strings.TrimSpace(line)

	// SSH 로그인 성공 패턴 감지
	if loginInfo := ld.detectSSHAccepted(line); loginInfo != nil {
		return true, loginInfo
	}

	// SSH 로그인 실패 패턴 감지
	if loginInfo := ld.detectSSHFailed(line); loginInfo != nil {
		return true, loginInfo
	}

	// Sudo 명령 실행 패턴 감지
	if loginInfo := ld.detectSudoCommand(line); loginInfo != nil {
		return true, loginInfo
	}

	// 웹 로그인 패턴 감지
	if loginInfo := ld.detectWebLogin(line); loginInfo != nil {
		return true, loginInfo
	}

	// 인증 실패 패턴 감지
	if loginInfo := ld.detectAuthFailure(line); loginInfo != nil {
		return true, loginInfo
	}

	return false, nil
}

// detectSSHAccepted SSH 로그인 성공 패턴 감지
func (ld *LoginDetector) detectSSHAccepted(line string) *LoginInfo {
	patterns := []string{
		`Accepted (\w+) for (\w+) from ([\d\.]+) port \d+`,
		`session opened for user (\w+)`,
		`authentication accepted for (\w+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			loginInfo := &LoginInfo{
				Status:  "accepted",
				Success: true,
			}

			if len(matches) >= 4 { // 첫 번째 패턴 (method, user, ip)
				loginInfo.Method = matches[1]
				loginInfo.User = matches[2]
				loginInfo.IP = matches[3]
			} else { // 다른 패턴들 (user만)
				loginInfo.User = matches[1]
				loginInfo.Method = "ssh"
			}

			// 시스템 메트릭과 IP 정보 추가
			ld.enhanceLoginInfo(loginInfo)
			return loginInfo
		}
	}

	return nil
}

// detectSSHFailed SSH 로그인 실패 패턴 감지
func (ld *LoginDetector) detectSSHFailed(line string) *LoginInfo {
	patterns := []string{
		`Failed (\w+) for (\w+) from ([\d\.]+)`,
		`authentication failure.*user=(\w+).*rhost=([\d\.]+)`,
		`Invalid user (\w+) from ([\d\.]+)`,
		`Connection closed by ([\d\.]+).*\[preauth\]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			loginInfo := &LoginInfo{
				Status:  "failed",
				Success: false,
			}

			switch len(matches) {
			case 4: // method, user, ip
				loginInfo.Method = matches[1]
				loginInfo.User = matches[2]
				loginInfo.IP = matches[3]
			case 3: // user, ip 또는 ip만
				if ld.isValidIP(matches[1]) {
					loginInfo.IP = matches[1]
					loginInfo.User = "unknown"
				} else {
					loginInfo.User = matches[1]
					loginInfo.IP = matches[2]
				}
				loginInfo.Method = "ssh"
			case 2: // ip만
				loginInfo.IP = matches[1]
				loginInfo.User = "unknown"
				loginInfo.Method = "ssh"
			}

			// 시스템 메트릭과 IP 정보 추가
			ld.enhanceLoginInfo(loginInfo)
			return loginInfo
		}
	}

	return nil
}

// detectSudoCommand Sudo 명령 실행 패턴 감지
func (ld *LoginDetector) detectSudoCommand(line string) *LoginInfo {
	patterns := []string{
		`(\w+) : TTY=\w+ ; PWD=.* ; USER=\w+ ; COMMAND=(.*)`,
		`sudo: (\w+) : (.*)`,
		`su: pam_unix.*session opened for user (\w+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			loginInfo := &LoginInfo{
				Status:  "sudo",
				Success: true,
				Method:  "sudo",
			}

			if len(matches) >= 3 {
				loginInfo.User = matches[1]
				loginInfo.Command = matches[2]
			} else {
				loginInfo.User = matches[1]
			}

			// 시스템 메트릭과 IP 정보 추가
			ld.enhanceLoginInfo(loginInfo)
			return loginInfo
		}
	}

	return nil
}

// detectWebLogin 웹 로그인 패턴 감지
func (ld *LoginDetector) detectWebLogin(line string) *LoginInfo {
	patterns := []string{
		`login.*user[:\s]+(\w+).*from[:\s]+([\d\.]+)`,
		`authentication.*user[:\s]+(\w+).*ip[:\s]+([\d\.]+)`,
		`sign.*in.*user[:\s]+(\w+)`,
		`logged.*in.*user[:\s]+(\w+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern) // case insensitive
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			loginInfo := &LoginInfo{
				Status:  "web_login",
				Success: true,
				Method:  "web",
				User:    matches[1],
			}

			if len(matches) >= 3 {
				loginInfo.IP = matches[2]
			}

			// 시스템 메트릭과 IP 정보 추가
			ld.enhanceLoginInfo(loginInfo)
			return loginInfo
		}
	}

	return nil
}

// detectAuthFailure 일반 인증 실패 패턴 감지
func (ld *LoginDetector) detectAuthFailure(line string) *LoginInfo {
	patterns := []string{
		`authentication failure`,
		`login failure`,
		`invalid password`,
		`access denied`,
		`unauthorized access`,
		`permission denied`,
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(line), pattern) {
			// IP 주소 추출 시도
			ipPattern := regexp.MustCompile(`([\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3})`)
			ipMatches := ipPattern.FindStringSubmatch(line)

			// 사용자명 추출 시도
			userPattern := regexp.MustCompile(`user[:\s]+(\w+)`)
			userMatches := userPattern.FindStringSubmatch(line)

			loginInfo := &LoginInfo{
				Status:  "failed",
				Success: false,
				Method:  "unknown",
			}

			if len(ipMatches) > 0 {
				loginInfo.IP = ipMatches[1]
			}

			if len(userMatches) > 1 {
				loginInfo.User = userMatches[1]
			} else {
				loginInfo.User = "unknown"
			}

			// 시스템 메트릭과 IP 정보 추가
			ld.enhanceLoginInfo(loginInfo)
			return loginInfo
		}
	}

	return nil
}

// isValidIP IP 주소 유효성 검사
func (ld *LoginDetector) isValidIP(ip string) bool {
	ipPattern := regexp.MustCompile(`^([\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3})$`)
	return ipPattern.MatchString(ip)
}

// GetSupportedPatterns 지원하는 패턴 목록 반환
func (ld *LoginDetector) GetSupportedPatterns() []string {
	return []string{
		"SSH Login Success",
		"SSH Login Failure",
		"Sudo Command Execution",
		"Web Login",
		"Authentication Failure",
		"Invalid User",
		"Connection Closed (preauth)",
		"Permission Denied",
	}
}

// GetStatistics 감지 통계 반환 (향후 확장용)
func (ld *LoginDetector) GetStatistics() map[string]int {
	// 향후 통계 수집 기능 구현
	return map[string]int{
		"ssh_success":  0,
		"ssh_failed":   0,
		"sudo_command": 0,
		"web_login":    0,
		"auth_failure": 0,
	}
}

// collectSystemMetrics 현재 시스템 리소스 정보 수집
// 로그인 시점의 CPU, 메모리 사용량을 실시간으로 수집
func (ld *LoginDetector) collectSystemMetrics() SystemMetrics {
	// 시스템 모니터가 설정되어 있으면 해당 메트릭 사용
	if ld.systemMonitor != nil {
		return ld.systemMonitor.GetCurrentMetrics()
	}
	
	// 시스템 모니터가 없으면 임시 모니터 생성하여 즉시 수집
	tempMonitor := NewSystemMonitor(time.Second) // 즉시 수집용
	tempMonitor.collectMetrics()
	return tempMonitor.GetCurrentMetrics()
}

// getIPLocationInfo IP 주소의 지리적 위치 및 상세 정보 조회
// 무료 IP 지리정보 API를 사용하여 실시간 조회
func (ld *LoginDetector) getIPLocationInfo(ip string) *IPLocationInfo {
	if ip == "" {
		return nil
	}
	
	// 사설 IP 주소 체크
	isPrivate := ld.isPrivateIP(ip)
	
	ipInfo := &IPLocationInfo{
		IP:        ip,
		IsPrivate: isPrivate,
	}
	
	// 사설 IP는 지리정보 조회 생략
	if isPrivate {
		ipInfo.Country = "Private Network"
		ipInfo.Organization = "Private IP Range"
		ipInfo.Threat = "LOW"
		return ipInfo
	}
	
	// 외부 API로 지리정보 조회 (5초 타임아웃)
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,regionName,city,org,as,query", ip)
	
	resp, err := client.Get(url)
	if err != nil {
		ld.logger.Errorf("Failed to query IP location for %s: %v", ip, err)
		ipInfo.Threat = "UNKNOWN"
		return ipInfo
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ld.logger.Errorf("Failed to read IP location response: %v", err)
		ipInfo.Threat = "UNKNOWN"
		return ipInfo
	}
	
	var result struct {
		Status     string `json:"status"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
		Org        string `json:"org"`
		AS         string `json:"as"`
		Query      string `json:"query"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		ld.logger.Errorf("Failed to parse IP location response: %v", err)
		ipInfo.Threat = "UNKNOWN"
		return ipInfo
	}
	
	if result.Status == "success" {
		ipInfo.Country = result.Country
		ipInfo.Region = result.RegionName
		ipInfo.City = result.City
		ipInfo.Organization = result.Org
		ipInfo.ASN = result.AS
		
		// 간단한 위험도 평가
		ipInfo.Threat = ld.assessThreatLevel(result.Country, result.Org)
	} else {
		ipInfo.Threat = "UNKNOWN"
	}
	
	return ipInfo
}

// isPrivateIP IP 주소가 사설 IP인지 확인
func (ld *LoginDetector) isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	// RFC 1918 사설 IP 범위 확인
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",     // 루프백
		"169.254.0.0/16",  // 링크 로컬
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

// assessThreatLevel 국가와 조직 정보를 바탕으로 위험도 평가
func (ld *LoginDetector) assessThreatLevel(country, org string) string {
	// 한국 내부 IP는 LOW
	if country == "South Korea" || country == "Korea" {
		return "LOW"
	}
	
	// 알려진 클라우드 서비스는 MEDIUM
	cloudProviders := []string{"Amazon", "Google", "Microsoft", "Azure", "AWS"}
	orgLower := strings.ToLower(org)
	for _, provider := range cloudProviders {
		if strings.Contains(orgLower, strings.ToLower(provider)) {
			return "MEDIUM"
		}
	}
	
	// 일반적으로 의심스러운 국가들
	suspiciousCountries := []string{"China", "Russia", "North Korea"}
	for _, suspicious := range suspiciousCountries {
		if country == suspicious {
			return "HIGH"
		}
	}
	
	// 기본적으로 해외 IP는 MEDIUM
	return "MEDIUM"
}

// enhanceLoginInfo 로그인 정보에 시스템 메트릭과 IP 정보 추가
// 10분 간격 알림 제한 로직도 적용
func (ld *LoginDetector) enhanceLoginInfo(loginInfo *LoginInfo) {
	// 타임스탬프 설정
	loginInfo.Timestamp = time.Now()
	
	// 시스템 리소스 정보 수집
	loginInfo.SystemInfo = ld.collectSystemMetrics()
	
	// IP 위치 정보 조회 (비동기로 처리하지 않고 즉시 처리)
	if loginInfo.IP != "" {
		loginInfo.IPDetails = ld.getIPLocationInfo(loginInfo.IP)
	}
	
	// 알림 전송 여부 확인 (10분 간격 제한 적용)
	loginInfo.ShouldAlert = ld.shouldSendAlert(loginInfo)
}

// ConvertToMap LoginInfo를 map으로 변환 (기존 코드 호환성)
// 확장된 정보를 포함하여 더 상세한 맵 반환
func (li *LoginInfo) ToMap() map[string]string {
	result := map[string]string{
		"status":    li.Status,
		"user":      li.User,
		"ip":        li.IP,
		"method":    li.Method,
		"command":   li.Command,
		"timestamp": li.Timestamp.Format("2006-01-02 15:04:05"),
	}
	
	// 시스템 정보 추가
	result["cpu_usage"] = fmt.Sprintf("%.1f%%", li.SystemInfo.CPU.UsagePercent)
	result["memory_usage"] = fmt.Sprintf("%.1f%%", li.SystemInfo.Memory.UsagePercent)
	result["cpu_temp"] = fmt.Sprintf("%.1f°C", li.SystemInfo.Temperature.CPUTemp)
	result["load_avg"] = fmt.Sprintf("%.2f", li.SystemInfo.LoadAverage.Load1Min)
	
	// 디스크 정보 추가 (주요 마운트 포인트들)
	var diskInfo []string
	for _, disk := range li.SystemInfo.Disk {
		// 주요 마운트 포인트만 포함
		if disk.MountPoint == "/" || disk.MountPoint == "/home" || disk.MountPoint == "C:" || 
		   disk.MountPoint == "/var" || disk.MountPoint == "/tmp" {
			diskInfo = append(diskInfo, fmt.Sprintf("%s: %.1f%% (%.1f/%.1f GB)", 
				disk.MountPoint, disk.UsagePercent, disk.UsedGB, disk.TotalGB))
		}
	}
	if len(diskInfo) > 0 {
		result["disk_usage"] = strings.Join(diskInfo, ", ")
	}
	
	// IP 위치 정보 추가
	if li.IPDetails != nil {
		result["ip_country"] = li.IPDetails.Country
		result["ip_city"] = li.IPDetails.City
		result["ip_org"] = li.IPDetails.Organization
		result["ip_threat"] = li.IPDetails.Threat
		result["ip_private"] = fmt.Sprintf("%t", li.IPDetails.IsPrivate)
	}
	
	return result
} 