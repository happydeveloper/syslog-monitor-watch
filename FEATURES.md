# 🤖 Lambda-X AI Security Monitor - 기능 리스트

## 🚀 핵심 기능

### 1. 🔍 **실시간 로그 모니터링**
- **지능형 패턴 인식**: SQL 인젝션, 무차별 대입 공격, 권한 상승 등
- **다중 로그 포맷 지원**: Apache, Nginx, MySQL, PostgreSQL, 시스템 로그
- **키워드 및 정규식 필터링**: 정밀한 로그 필터링
- **실시간 분석**: 지연 없는 즉시 위험 감지
- **파일 또는 스트림 입력**: 파일 모니터링 또는 실시간 스트림 처리

### 2. 🤖 **AI 기반 위험 분석**
- **Google Gemini API 연동**: 고급 AI 기반 시스템 진단
- **실시간 AI 분석**: 로그 패턴, 보안 위협, 시스템 상태 분석
- **전문가 진단**: 자연어 기반 시스템 문제 진단 및 권장사항
- **기본 모드 지원**: API 키 없어도 기본 AI 진단 작동
- **위협 패턴 감지**:
  - 🔴 SQL 인젝션 공격 감지
  - 🟠 무차별 대입 공격 탐지
  - 🟡 메모리 누수 패턴 분석
  - 🔵 데이터베이스 연결 문제
  - 🟣 비정상적인 트래픽 급증
  - 🟤 파일 시스템 오류
  - ⚫ 권한 상승 시도 감지

### 3. 🖥️ **종합 시스템 모니터링**
- **실시간 메트릭 수집**:
  - CPU 사용률 (사용자/시스템/대기)
  - 메모리 사용률 (총/사용/가용)
  - 디스크 사용률 (마운트 포인트별)
  - 시스템 온도 (CPU/GPU)
  - 시스템 로드 (1분/5분/15분)
  - 프로세스 수
- **네트워크 정보**:
  - 호스트명 자동 감지
  - 사설 IP 주소 수집
  - 공인 IP 주소 자동 감지
  - 네트워크 패킷 통계
- **임계값 기반 알림**: 사용자 정의 알림 기준
- **주기적 시스템 상태 보고서**: 설정 가능한 간격으로 자동 보고

### 4. 🔐 **보안 감시 기능**
- **로그인 모니터링**:
  - SSH 로그인 감지
  - sudo 명령어 사용 감지
  - 웹 애플리케이션 로그인 감지
  - 로그인 실패 패턴 분석
- **지리정보 매핑**:
  - ASN (Autonomous System Number) 조회
  - IP 주소 지리적 위치 확인
  - 조직 정보 자동 수집
- **위협 예측 및 분석**:
  - 보안 위협 사전 예측
  - 위험도 점수 산출
  - 권장 대응 방안 제시

### 5. 📧 **다중 채널 알림 시스템**
- **이메일 알림**:
  - Gmail SMTP 지원
  - 다중 수신자 지원
  - 상세한 시스템 정보 포함
  - 자동 이메일 설정 (App Password)
- **Slack 통합**:
  - Incoming Webhooks 지원
  - 실시간 채널 알림
  - 구조화된 메시지 형태
  - 상태별 색상 구분
- **알림 내용**:
  - AI 분석 결과
  - 시스템 메트릭
  - 보안 위협 정보
  - 권장 대응 방안

## ⚙️ 설정 및 관리

### 1. 🔧 **설정 관리**
- **JSON 기반 설정 파일**: `~/.syslog-monitor/config.json`
- **환경변수 지원**: API 키, 이메일, Slack 설정
- **명령행 옵션**: 실시간 설정 변경
- **설정 확인**: `-show-config` 옵션

### 2. 🛠️ **명령행 옵션**
```bash
# 기본 옵션
-file string          # 모니터링할 로그 파일 경로
-output string        # 필터링된 로그 출력 파일
-keywords string      # 포함할 키워드 (쉼표 구분)
-filters string       # 제외할 패턴 (정규식, 쉼표 구분)
-help                 # 도움말 표시

# AI 분석 옵션
-ai-analysis          # AI 기반 로그 분석 활성화
-alert-threshold      # AI 알림 임계값 (기본: 7.0)
-log-type string      # 로그 타입 (auto, apache, nginx, mysql)
-gemini-api-key       # Gemini AI API 키 설정
-show-config          # 현재 설정 정보 표시

# 시스템 모니터링 옵션
-system-monitor       # 시스템 메트릭 모니터링 활성화
-cpu-threshold        # CPU 사용률 임계값 (기본: 80)
-memory-threshold     # 메모리 사용률 임계값 (기본: 85)
-disk-threshold       # 디스크 사용률 임계값 (기본: 90)

# 알림 옵션
-email-to string      # 수신자 이메일 (쉼표 구분)
-smtp-server string   # SMTP 서버 (기본: smtp.gmail.com)
-smtp-port string     # SMTP 포트 (기본: 587)
-smtp-user string     # SMTP 사용자명
-smtp-password string # SMTP 비밀번호
-slack-webhook string # Slack 웹훅 URL
-slack-channel string # Slack 채널

# 보안 옵션
-login-watch          # 로그인 모니터링 활성화 (SSH, sudo, 웹)

# 테스트 옵션
-test-email           # 이메일 설정 테스트
-test-slack           # Slack 설정 테스트

# 주기적 보고서 옵션
-periodic-report      # 주기적 시스템 상태 보고서 활성화
-report-interval int  # 보고서 간격 (분 단위, 기본: 60)
-alert-interval int   # 알림 간격 (분 단위, 기본: 10)
```

### 3. 🌍 **환경변수**
```bash
# Gemini AI 설정
GEMINI_API_KEY              # Gemini AI API 키

# 이메일 설정
SYSLOG_EMAIL_TO             # 수신자 이메일 (쉼표 구분)
SYSLOG_SMTP_USER            # SMTP 사용자명
SYSLOG_SMTP_PASSWORD        # SMTP 비밀번호/앱 비밀번호
SYSLOG_EMAIL_FROM           # 발신자 이메일
SYSLOG_SMTP_SERVER          # SMTP 서버
SYSLOG_SMTP_PORT            # SMTP 포트

# Slack 설정
SYSLOG_SLACK_WEBHOOK        # Slack 웹훅 URL
SYSLOG_SLACK_CHANNEL        # Slack 채널
SYSLOG_SLACK_USERNAME       # Slack 봇 사용자명

# 기타 설정
SYSLOG_CONFIG_PATH          # 설정 파일 경로
```

## 📊 사용 예시

### 1. 기본 사용법
```bash
# 기본 모니터링
syslog-monitor

# AI 분석 활성화
syslog-monitor -ai-analysis

# 시스템 모니터링
syslog-monitor -system-monitor

# 전체 기능 활성화
syslog-monitor -ai-analysis -system-monitor -login-watch
```

### 2. Gemini AI 연동
```bash
# API 키 설정
export GEMINI_API_KEY="your-api-key"
syslog-monitor -ai-analysis -system-monitor

# 또는 명령행에서 직접 설정
syslog-monitor -gemini-api-key="your-api-key" -ai-analysis
```

### 3. 고급 사용법
```bash
# 특정 로그 파일 모니터링
syslog-monitor -file=/var/log/auth.log -ai-analysis

# 키워드 필터링
syslog-monitor -keywords="error,failed" -ai-analysis

# 정규식 필터링
syslog-monitor -filters="systemd,kernel" -output=./filtered.log

# 주기적 시스템 보고서 (5분마다)
syslog-monitor -system-monitor -periodic-report -report-interval=5

# 다중 채널 알림
syslog-monitor -ai-analysis \
  -email-to="admin@company.com,security@company.com" \
  -slack-webhook="https://hooks.slack.com/..."
```

### 4. macOS 전용 사용법
```bash
# 실시간 시스템 로그
sudo log stream | syslog-monitor -file=/dev/stdin -ai-analysis

# 특정 로그 파일들
syslog-monitor -file=/var/log/system.log -ai-analysis
syslog-monitor -file=/var/log/install.log -keywords=error
syslog-monitor -file=/var/log/wifi.log -system-monitor
```

## 🔒 보안 기능

### 1. API 키 보호
- 설정 파일이 Git에서 자동 제외
- 환경변수를 통한 안전한 API 키 관리
- 마스킹된 API 키 표시

### 2. 로그 분석
- SQL 인젝션 패턴 감지
- 무차별 대입 공격 탐지
- 권한 상승 시도 감지
- 비정상적인 네트워크 활동 감지

### 3. 실시간 알림
- 보안 위협 즉시 알림
- 상세한 위협 정보 제공
- 권장 대응 방안 제시

## 🎯 AI 진단 예시

### 기본 모드 (API 키 없음)
```
🔬 AI 전문가 진단 결과 (기본 모드)
==================================
📊 전반적인 시스템 건강도: 🔴 POOR
⚠️  발견된 문제점:
  🔴 메모리 사용률이 매우 높습니다

💡 전문가 권장사항:
==================
• 메모리 누수 확인: `ps aux --sort=-%mem`
• 스왑 사용량 확인: `vm_stat`

💡 Gemini API 키를 설정하면 더 정교한 AI 진단을 받을 수 있습니다.
```

### Gemini AI 모드 (API 키 설정)
```
🔬 AI 전문가 진단 결과
=====================
📊 전반적인 시스템 건강도: 🔴 CRITICAL
⚠️  발견된 문제점:
  🔴 메모리 사용률이 93.8%로 매우 높습니다
  🟡 CPU 사용률이 52.7%로 높은 편입니다

💡 전문가 권장사항:
==================
• 즉시 메모리 정리 작업 수행
• 불필요한 프로세스 종료: `killall -9 [process_name]`
• 시스템 재부팅 고려
• 메모리 누수 프로세스 확인: `ps aux --sort=-%mem | head -10`

🔧 즉시 실행 가능한 명령어:
==========================
• 메모리 사용량 확인: `vm_stat`
• 높은 메모리 사용 프로세스: `ps aux --sort=-%mem | head -5`
• 시스템 부하 확인: `top -l 1`
```

## 📈 성능 최적화

### 1. 시스템 요구사항
- **운영체제**: macOS 10.14+, Linux (Ubuntu 18.04+, CentOS 7+)
- **메모리**: 최소 512MB, 권장 1GB
- **네트워크**: 인터넷 연결 (공인 IP 조회, ASN 정보용)

### 2. 최적화 기능
- 효율적인 로그 파싱
- 메모리 사용량 최적화
- 비동기 알림 처리
- 설정 가능한 모니터링 간격

### 3. 확장성
- 다중 로그 파일 동시 모니터링
- 분산 환경 지원
- 대용량 로그 처리
- 클러스터 모니터링

---

**📝 버전**: v2.2  
**🔗 GitHub**: [Lambda-X AI Security Monitor](https://github.com/your-repo)  
**📚 문서**: [README.md](./README.md)  
**⚙️ 설정**: [설정 가이드](./README.md#설정-파일)