package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogParser 로그 파서 인터페이스
type LogParser interface {
	Parse(line string) (*ParsedLog, error)
	GetLogType() string
	DetectFormat(line string) bool
}

// ParsedLog 파싱된 로그 구조체
type ParsedLog struct {
	Timestamp    time.Time         `json:"timestamp"`
	LogType      string            `json:"log_type"`
	Level        string            `json:"level"`
	Source       string            `json:"source"`
	Message      string            `json:"message"`
	Fields       map[string]string `json:"fields"`
	RawLog       string            `json:"raw_log"`
	HTTPDetails  *HTTPLogDetails   `json:"http_details,omitempty"`
	DBDetails    *DBLogDetails     `json:"db_details,omitempty"`
	ErrorDetails *ErrorDetails     `json:"error_details,omitempty"`
}

// HTTPLogDetails HTTP 로그 상세 정보
type HTTPLogDetails struct {
	Method         string `json:"method"`
	URL            string `json:"url"`
	StatusCode     int    `json:"status_code"`
	ResponseSize   int64  `json:"response_size"`
	ResponseTime   int64  `json:"response_time_ms"`
	UserAgent      string `json:"user_agent"`
	Referer        string `json:"referer"`
	ClientIP       string `json:"client_ip"`
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
}

// DBLogDetails 데이터베이스 로그 상세 정보
type DBLogDetails struct {
	QueryType      string  `json:"query_type"`
	Query          string  `json:"query"`
	ExecutionTime  float64 `json:"execution_time_ms"`
	RowsAffected   int64   `json:"rows_affected"`
	Database       string  `json:"database"`
	Table          string  `json:"table"`
	Connection     string  `json:"connection"`
	ErrorCode      string  `json:"error_code"`
	SlowQuery      bool    `json:"slow_query"`
}

// ErrorDetails 에러 상세 정보
type ErrorDetails struct {
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	StackTrace   string `json:"stack_trace"`
	Module       string `json:"module"`
	Function     string `json:"function"`
	LineNumber   int    `json:"line_number"`
}

// ApacheLogParser Apache 로그 파서
type ApacheLogParser struct {
	commonLogRegex    *regexp.Regexp
	combinedLogRegex  *regexp.Regexp
	errorLogRegex     *regexp.Regexp
}

// NginxLogParser Nginx 로그 파서
type NginxLogParser struct {
	accessLogRegex *regexp.Regexp
	errorLogRegex  *regexp.Regexp
}

// MySQLLogParser MySQL 로그 파서
type MySQLLogParser struct {
	errorLogRegex     *regexp.Regexp
	slowQueryRegex    *regexp.Regexp
	generalLogRegex   *regexp.Regexp
	binlogRegex       *regexp.Regexp
}

// PostgreSQLLogParser PostgreSQL 로그 파서
type PostgreSQLLogParser struct {
	logRegex      *regexp.Regexp
	errorRegex    *regexp.Regexp
	slowQueryRegex *regexp.Regexp
}

// ApplicationLogParser 애플리케이션 로그 파서
type ApplicationLogParser struct {
	jsonLogRegex    *regexp.Regexp
	structuredRegex *regexp.Regexp
	errorRegex      *regexp.Regexp
}

// NewApacheLogParser Apache 로그 파서 생성
func NewApacheLogParser() *ApacheLogParser {
	return &ApacheLogParser{
		// Common Log Format: IP - - [timestamp] "method url protocol" status size
		commonLogRegex: regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) ([^"]*)" (\d+) (\S+)`),
		// Combined Log Format: Common + "referer" "user-agent"
		combinedLogRegex: regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) ([^"]*)" (\d+) (\S+) "([^"]*)" "([^"]*)"`),
		// Error Log: [timestamp] [level] [pid] [client IP] message
		errorLogRegex: regexp.MustCompile(`^\[([^\]]+)\] \[([^\]]+)\] \[([^\]]+)\] (.+)`),
	}
}

// Parse Apache 로그 파싱
func (p *ApacheLogParser) Parse(line string) (*ParsedLog, error) {
	parsed := &ParsedLog{
		LogType: "apache",
		RawLog:  line,
		Fields:  make(map[string]string),
	}

	// Error log 먼저 시도
	if matches := p.errorLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("Mon Jan 02 15:04:05.000000 2006", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[2])
		parsed.Fields["pid"] = matches[3]
		parsed.Message = matches[4]
		
		if strings.Contains(parsed.Level, "ERROR") || strings.Contains(parsed.Level, "CRIT") {
			parsed.ErrorDetails = &ErrorDetails{
				ErrorType: parsed.Level,
				Module:    "apache",
			}
		}
		return parsed, nil
	}

	// Combined log 시도
	if matches := p.combinedLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
		parsed.Timestamp = timestamp
		parsed.Level = "INFO"
		
		statusCode, _ := strconv.Atoi(matches[6])
		responseSize, _ := strconv.ParseInt(matches[7], 10, 64)
		
		parsed.HTTPDetails = &HTTPLogDetails{
			ClientIP:     matches[1],
			Method:       matches[3],
			URL:          matches[4],
			Protocol:     matches[5],
			StatusCode:   statusCode,
			ResponseSize: responseSize,
			Referer:      matches[8],
			UserAgent:    matches[9],
		}
		
		parsed.Fields["client_ip"] = matches[1]
		parsed.Fields["status_code"] = matches[6]
		parsed.Message = fmt.Sprintf("%s %s %s - %d", matches[3], matches[4], matches[5], statusCode)
		
		// 에러 상태 코드 체크
		if statusCode >= 400 {
			if statusCode >= 500 {
				parsed.Level = "ERROR"
			} else {
				parsed.Level = "WARNING"
			}
		}
		
		return parsed, nil
	}

	// Common log 시도
	if matches := p.commonLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
		parsed.Timestamp = timestamp
		parsed.Level = "INFO"
		
		statusCode, _ := strconv.Atoi(matches[6])
		responseSize, _ := strconv.ParseInt(matches[7], 10, 64)
		
		parsed.HTTPDetails = &HTTPLogDetails{
			ClientIP:     matches[1],
			Method:       matches[3],
			URL:          matches[4],
			Protocol:     matches[5],
			StatusCode:   statusCode,
			ResponseSize: responseSize,
		}
		
		parsed.Fields["client_ip"] = matches[1]
		parsed.Fields["status_code"] = matches[6]
		parsed.Message = fmt.Sprintf("%s %s %s - %d", matches[3], matches[4], matches[5], statusCode)
		
		if statusCode >= 400 {
			if statusCode >= 500 {
				parsed.Level = "ERROR"
			} else {
				parsed.Level = "WARNING"
			}
		}
		
		return parsed, nil
	}

	// 파싱 실패 시 기본 처리
	parsed.Timestamp = time.Now()
	parsed.Level = "INFO"
	parsed.Message = line
	return parsed, nil
}

// GetLogType 로그 타입 반환
func (p *ApacheLogParser) GetLogType() string {
	return "apache"
}

// DetectFormat 포맷 감지
func (p *ApacheLogParser) DetectFormat(line string) bool {
	return p.commonLogRegex.MatchString(line) || 
	       p.combinedLogRegex.MatchString(line) || 
	       p.errorLogRegex.MatchString(line)
}

// NewNginxLogParser Nginx 로그 파서 생성
func NewNginxLogParser() *NginxLogParser {
	return &NginxLogParser{
		// Nginx access log: IP - - [timestamp] "method url protocol" status size "referer" "user-agent" rt
		accessLogRegex: regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*) ([^"]*)" (\d+) (\S+) "([^"]*)" "([^"]*)"(?:\s+(\d+\.\d+))?`),
		// Nginx error log: timestamp [level] pid message
		errorLogRegex: regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[([^\]]+)\] (\d+)#\d+: (.+)`),
	}
}

// Parse Nginx 로그 파싱
func (p *NginxLogParser) Parse(line string) (*ParsedLog, error) {
	parsed := &ParsedLog{
		LogType: "nginx",
		RawLog:  line,
		Fields:  make(map[string]string),
	}

	// Error log 먼저 시도
	if matches := p.errorLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("2006/01/02 15:04:05", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[2])
		parsed.Fields["pid"] = matches[3]
		parsed.Message = matches[4]
		
		if strings.Contains(parsed.Level, "ERROR") || strings.Contains(parsed.Level, "CRIT") {
			parsed.ErrorDetails = &ErrorDetails{
				ErrorType: parsed.Level,
				Module:    "nginx",
			}
		}
		return parsed, nil
	}

	// Access log 시도
	if matches := p.accessLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
		parsed.Timestamp = timestamp
		parsed.Level = "INFO"
		
		statusCode, _ := strconv.Atoi(matches[6])
		responseSize, _ := strconv.ParseInt(matches[7], 10, 64)
		
		httpDetails := &HTTPLogDetails{
			ClientIP:     matches[1],
			Method:       matches[3],
			URL:          matches[4],
			Protocol:     matches[5],
			StatusCode:   statusCode,
			ResponseSize: responseSize,
			Referer:      matches[8],
			UserAgent:    matches[9],
		}
		
		// 응답 시간이 있는 경우
		if len(matches) > 10 && matches[10] != "" {
			if rt, err := strconv.ParseFloat(matches[10], 64); err == nil {
				httpDetails.ResponseTime = int64(rt * 1000) // 초를 밀리초로 변환
			}
		}
		
		parsed.HTTPDetails = httpDetails
		parsed.Fields["client_ip"] = matches[1]
		parsed.Fields["status_code"] = matches[6]
		parsed.Message = fmt.Sprintf("%s %s %s - %d", matches[3], matches[4], matches[5], statusCode)
		
		if statusCode >= 400 {
			if statusCode >= 500 {
				parsed.Level = "ERROR"
			} else {
				parsed.Level = "WARNING"
			}
		}
		
		return parsed, nil
	}

	// 파싱 실패 시 기본 처리
	parsed.Timestamp = time.Now()
	parsed.Level = "INFO"
	parsed.Message = line
	return parsed, nil
}

// GetLogType 로그 타입 반환
func (p *NginxLogParser) GetLogType() string {
	return "nginx"
}

// DetectFormat 포맷 감지
func (p *NginxLogParser) DetectFormat(line string) bool {
	return p.accessLogRegex.MatchString(line) || p.errorLogRegex.MatchString(line)
}

// NewMySQLLogParser MySQL 로그 파서 생성
func NewMySQLLogParser() *MySQLLogParser {
	return &MySQLLogParser{
		// MySQL error log: timestamp [level] message
		errorLogRegex: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z|\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \[([^\]]+)\] (.+)`),
		// Slow query log: # Time: timestamp # User@Host: user[user] @ host [IP] # Query_time: time Lock_time: time Rows_sent: num Rows_examined: num
		slowQueryRegex: regexp.MustCompile(`# Time: (.+)|# User@Host: (.+)|# Query_time: (\d+\.\d+)\s+Lock_time: (\d+\.\d+)\s+Rows_sent: (\d+)\s+Rows_examined: (\d+)|^(SELECT|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER)`),
		// General log: timestamp ID Command Argument
		generalLogRegex: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\s+(\d+)\s+(\w+)\s+(.+)`),
	}
}

// Parse MySQL 로그 파싱
func (p *MySQLLogParser) Parse(line string) (*ParsedLog, error) {
	parsed := &ParsedLog{
		LogType: "mysql",
		RawLog:  line,
		Fields:  make(map[string]string),
	}

	// Error log 시도
	if matches := p.errorLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[2])
		parsed.Message = matches[3]
		
		if strings.Contains(parsed.Level, "ERROR") {
			parsed.ErrorDetails = &ErrorDetails{
				ErrorType: parsed.Level,
				Module:    "mysql",
			}
		}
		
		// 데이터베이스 관련 정보 추출
		if strings.Contains(parsed.Message, "Query") {
			parsed.DBDetails = &DBLogDetails{
				QueryType: "UNKNOWN",
				Query:     parsed.Message,
			}
		}
		
		return parsed, nil
	}

	// General log 시도
	if matches := p.generalLogRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = "INFO"
		parsed.Fields["connection_id"] = matches[2]
		parsed.Fields["command"] = matches[3]
		parsed.Message = matches[4]
		
		command := strings.ToUpper(matches[3])
		if command == "QUERY" {
			query := matches[4]
			queryType := "SELECT"
			if strings.HasPrefix(strings.ToUpper(query), "INSERT") {
				queryType = "INSERT"
			} else if strings.HasPrefix(strings.ToUpper(query), "UPDATE") {
				queryType = "UPDATE"
			} else if strings.HasPrefix(strings.ToUpper(query), "DELETE") {
				queryType = "DELETE"
			}
			
			parsed.DBDetails = &DBLogDetails{
				QueryType:  queryType,
				Query:      query,
				Connection: matches[2],
			}
		}
		
		return parsed, nil
	}

	// Slow query log는 여러 줄에 걸쳐 있어서 별도 처리 필요
	if strings.HasPrefix(line, "# Time:") || strings.HasPrefix(line, "# User@Host:") {
		parsed.Timestamp = time.Now()
		parsed.Level = "WARNING"
		parsed.Message = line
		parsed.DBDetails = &DBLogDetails{
			SlowQuery: true,
		}
		return parsed, nil
	}

	// 파싱 실패 시 기본 처리
	parsed.Timestamp = time.Now()
	parsed.Level = "INFO"
	parsed.Message = line
	return parsed, nil
}

// GetLogType 로그 타입 반환
func (p *MySQLLogParser) GetLogType() string {
	return "mysql"
}

// DetectFormat 포맷 감지
func (p *MySQLLogParser) DetectFormat(line string) bool {
	return p.errorLogRegex.MatchString(line) || 
	       p.generalLogRegex.MatchString(line) ||
	       strings.HasPrefix(line, "# Time:") ||
	       strings.HasPrefix(line, "# User@Host:")
}

// NewPostgreSQLLogParser PostgreSQL 로그 파서 생성
func NewPostgreSQLLogParser() *PostgreSQLLogParser {
	return &PostgreSQLLogParser{
		// PostgreSQL log: timestamp [pid] level: message
		logRegex: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+) [A-Z]+ \[(\d+)\] (\w+):\s+(.+)`),
		// Error pattern
		errorRegex: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+) [A-Z]+ \[(\d+)\] (ERROR|FATAL|PANIC):\s+(.+)`),
		// Slow query detection
		slowQueryRegex: regexp.MustCompile(`duration: (\d+\.\d+) ms\s+statement: (.+)`),
	}
}

// Parse PostgreSQL 로그 파싱
func (p *PostgreSQLLogParser) Parse(line string) (*ParsedLog, error) {
	parsed := &ParsedLog{
		LogType: "postgresql",
		RawLog:  line,
		Fields:  make(map[string]string),
	}

	// Error log 시도
	if matches := p.errorRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("2006-01-02 15:04:05.000", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[3])
		parsed.Fields["pid"] = matches[2]
		parsed.Message = matches[4]
		
		parsed.ErrorDetails = &ErrorDetails{
			ErrorType: parsed.Level,
			Module:    "postgresql",
		}
		return parsed, nil
	}

	// 일반 log 시도
	if matches := p.logRegex.FindStringSubmatch(line); matches != nil {
		timestamp, _ := time.Parse("2006-01-02 15:04:05.000", matches[1])
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[3])
		parsed.Fields["pid"] = matches[2]
		parsed.Message = matches[4]
		
		// Slow query 체크
		if slowMatches := p.slowQueryRegex.FindStringSubmatch(matches[4]); slowMatches != nil {
			duration, _ := strconv.ParseFloat(slowMatches[1], 64)
			parsed.DBDetails = &DBLogDetails{
				ExecutionTime: duration,
				Query:         slowMatches[2],
				SlowQuery:     duration > 1000, // 1초 이상은 slow query
			}
			
			// Query type 추출
			queryUpper := strings.ToUpper(strings.TrimSpace(slowMatches[2]))
			if strings.HasPrefix(queryUpper, "SELECT") {
				parsed.DBDetails.QueryType = "SELECT"
			} else if strings.HasPrefix(queryUpper, "INSERT") {
				parsed.DBDetails.QueryType = "INSERT"
			} else if strings.HasPrefix(queryUpper, "UPDATE") {
				parsed.DBDetails.QueryType = "UPDATE"
			} else if strings.HasPrefix(queryUpper, "DELETE") {
				parsed.DBDetails.QueryType = "DELETE"
			}
		}
		
		return parsed, nil
	}

	// 파싱 실패 시 기본 처리
	parsed.Timestamp = time.Now()
	parsed.Level = "INFO"
	parsed.Message = line
	return parsed, nil
}

// GetLogType 로그 타입 반환
func (p *PostgreSQLLogParser) GetLogType() string {
	return "postgresql"
}

// DetectFormat 포맷 감지
func (p *PostgreSQLLogParser) DetectFormat(line string) bool {
	return p.logRegex.MatchString(line) || p.errorRegex.MatchString(line)
}

// NewApplicationLogParser 애플리케이션 로그 파서 생성
func NewApplicationLogParser() *ApplicationLogParser {
	return &ApplicationLogParser{
		// JSON log pattern
		jsonLogRegex: regexp.MustCompile(`^\{.*\}$`),
		// Structured log: timestamp [level] module: message
		structuredRegex: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?)\s+\[?(\w+)\]?\s+(?:(\w+):)?\s*(.+)`),
		// Error with stack trace
		errorRegex: regexp.MustCompile(`(Exception|Error|at\s+\w+\.\w+)`),
	}
}

// Parse 애플리케이션 로그 파싱
func (p *ApplicationLogParser) Parse(line string) (*ParsedLog, error) {
	parsed := &ParsedLog{
		LogType: "application",
		RawLog:  line,
		Fields:  make(map[string]string),
	}

	// JSON 로그 시도
	if p.jsonLogRegex.MatchString(line) {
		// JSON 파싱은 복잡하므로 기본 처리
		parsed.Timestamp = time.Now()
		parsed.Level = "INFO"
		parsed.Message = line
		return parsed, nil
	}

	// 구조화된 로그 시도
	if matches := p.structuredRegex.FindStringSubmatch(line); matches != nil {
		timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
		if err != nil {
			timestamp, _ = time.Parse("2006-01-02 15:04:05.000", matches[1])
		}
		parsed.Timestamp = timestamp
		parsed.Level = strings.ToUpper(matches[2])
		if matches[3] != "" {
			parsed.Fields["module"] = matches[3]
		}
		parsed.Message = matches[4]
		
		// 에러 패턴 체크
		if p.errorRegex.MatchString(parsed.Message) {
			if parsed.Level == "INFO" {
				parsed.Level = "ERROR"
			}
			parsed.ErrorDetails = &ErrorDetails{
				ErrorType: "APPLICATION_ERROR",
				Module:    matches[3],
				StackTrace: parsed.Message,
			}
		}
		
		return parsed, nil
	}

	// 파싱 실패 시 기본 처리
	parsed.Timestamp = time.Now()
	parsed.Level = "INFO"
	parsed.Message = line
	return parsed, nil
}

// GetLogType 로그 타입 반환
func (p *ApplicationLogParser) GetLogType() string {
	return "application"
}

// DetectFormat 포맷 감지
func (p *ApplicationLogParser) DetectFormat(line string) bool {
	return p.jsonLogRegex.MatchString(line) || p.structuredRegex.MatchString(line)
}

// LogParserManager 로그 파서 관리자
type LogParserManager struct {
	parsers []LogParser
}

// NewLogParserManager 로그 파서 관리자 생성
func NewLogParserManager() *LogParserManager {
	return &LogParserManager{
		parsers: []LogParser{
			NewApacheLogParser(),
			NewNginxLogParser(),
			NewMySQLLogParser(),
			NewPostgreSQLLogParser(),
			NewApplicationLogParser(),
		},
	}
}

// ParseLog 로그 파싱 (자동 감지)
func (lpm *LogParserManager) ParseLog(line string) *ParsedLog {
	// 각 파서로 포맷 감지 시도
	for _, parser := range lpm.parsers {
		if parser.DetectFormat(line) {
			if parsed, err := parser.Parse(line); err == nil {
				return parsed
			}
		}
	}
	
	// 모든 파서가 실패하면 기본 파싱
	return &ParsedLog{
		Timestamp: time.Now(),
		LogType:   "unknown",
		Level:     "INFO",
		Message:   line,
		RawLog:    line,
		Fields:    make(map[string]string),
	}
}

// ParseLogWithType 특정 타입으로 로그 파싱
func (lpm *LogParserManager) ParseLogWithType(line string, logType string) *ParsedLog {
	for _, parser := range lpm.parsers {
		if parser.GetLogType() == logType {
			if parsed, err := parser.Parse(line); err == nil {
				return parsed
			}
		}
	}
	
	// 해당 타입 파서가 없거나 실패 시 기본 파싱
	return &ParsedLog{
		Timestamp: time.Now(),
		LogType:   logType,
		Level:     "INFO",
		Message:   line,
		RawLog:    line,
		Fields:    make(map[string]string),
	}
}

// GetSupportedTypes 지원하는 로그 타입 반환
func (lpm *LogParserManager) GetSupportedTypes() []string {
	types := make([]string, len(lpm.parsers))
	for i, parser := range lpm.parsers {
		types[i] = parser.GetLogType()
	}
	return types
} 