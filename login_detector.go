package main

import (
	"regexp"
	"strings"
)

// LoginDetector 로그인 패턴 감지 서비스
type LoginDetector struct {
	logger Logger
}

// LoginInfo 로그인 정보 구조체
type LoginInfo struct {
	Status  string
	User    string
	IP      string
	Method  string
	Command string
	Success bool
}

// NewLoginDetector 새로운 로그인 감지 서비스 생성
func NewLoginDetector(logger Logger) *LoginDetector {
	return &LoginDetector{
		logger: logger,
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

// ConvertToMap LoginInfo를 map으로 변환 (기존 코드 호환성)
func (li *LoginInfo) ToMap() map[string]string {
	return map[string]string{
		"status":  li.Status,
		"user":    li.User,
		"ip":      li.IP,
		"method":  li.Method,
		"command": li.Command,
	}
} 