# 🤖 AI-Powered Syslog Monitor v2.0

**AI 기반 로그 분석 및 시스템 모니터링 도구**

리눅스와 macOS 시스템의 syslog를 실시간으로 감시하고 **차세대 AI 기반 이상 징후 분석**, 시스템 메트릭 모니터링, **향상된 다중 플랫폼 알림**을 제공하는 최신 모니터링 솔루션입니다.

## 🆕 v2.0 새로운 기능

### 📍 **시스템 정보 자동 수집**
- **컴퓨터 이름 자동 감지**: 알람 시 호스트명 자동 포함
- **내부/외부 IP 분류**: RFC 1918 표준 준수 자동 분류
- **실시간 네트워크 정보**: 시스템의 모든 IP 주소 수집

### 🌐 **ASN 정보 실시간 조회**
- **조직 정보**: 외부 IP의 소속 조직 자동 조회
- **지리적 위치**: 국가, 지역, 도시 정보 수집
- **보안 위협 분석**: ASN 기반 위험도 평가

### 🚨 **향상된 알람 시스템**
- **상세한 시스템 정보 포함**: 컴퓨터명, IP, ASN 정보
- **맞춤형 보안 권장사항**: AI 기반 대응 가이드
- **실시간 위협 예측**: 보안 사고 사전 예방

## 📋 목차
- [핵심 기능](#핵심-기능)
- [빠른 시작](#빠른-시작)
- [설치 방법](#설치-방법)
- [사용법](#사용법)
- [AI 분석 기능](#ai-분석-기능)
- [시스템 모니터링](#시스템-모니터링)
- [알림 설정](#알림-설정)
- [테스트](#테스트)
- [문제 해결](#문제-해결)

## 🚀 핵심 기능

### 🔍 **실시간 로그 모니터링**
- **지능형 패턴 인식**: SQL 인젝션, 무차별 대입 공격, 권한 상승 등
- **다중 로그 포맷 지원**: Apache, Nginx, MySQL, PostgreSQL, 시스템 로그
- **키워드 및 정규식 필터링**: 정밀한 로그 필터링
- **실시간 분석**: 지연 없는 즉시 위험 감지

### 🤖 **AI 기반 위험 분석**
```
🎯 지원하는 위험 패턴:
├── 🔴 SQL 인젝션 공격 감지
├── 🟠 무차별 대입 공격 탐지  
├── 🟡 메모리 누수 패턴 분석
├── 🔵 데이터베이스 연결 문제
├── 🟣 비정상적인 트래픽 급증
├── 🟤 파일 시스템 오류
└── ⚫ 권한 상승 시도 감지
```

### 🖥️ **종합 시스템 모니터링**
- **실시간 메트릭**: CPU, 메모리, 디스크, 온도
- **네트워크 상태**: 패킷 손실률, 연결 상태
- **프로세스 추적**: 비정상 프로세스 감지
- **임계값 알림**: 사용자 정의 알림 기준

### 📧 **다중 채널 알림**
- **이메일 알림**: Gmail SMTP 지원, 다중 수신자
- **Slack 통합**: 실시간 채널 알림
- **상세 보고서**: AI 분석 결과 포함된 알림

## ⚡ 빠른 시작

### macOS 사용자 (권장)

```bash
# 1. 저장소 클론
git clone <repository-url>
cd syslog-monitor

# 2. v2.0 자동 설치 (권장)
./install-macos-v2.sh

# 3. 즉시 사용
syslog-monitor -ai-analysis -system-monitor
```

### 수동 설치

```bash
# 1. 빌드
make build-macos         # macOS용
make build-linux         # Linux용

# 2. 설치
sudo cp syslog-monitor_* /usr/local/bin/syslog-monitor
sudo chmod +x /usr/local/bin/syslog-monitor

# 3. 실행
syslog-monitor -help
```

## 🛠️ 설치 방법

### macOS 설치 스크립트 v2.0

**새로운 설치 스크립트**는 다음 기능들을 자동으로 설정합니다:

```bash
./install-macos-v2.sh
```

#### 설치 스크립트 기능:
- ✅ **시스템 요구사항 자동 확인**
- ✅ **Apple Silicon/Intel 자동 감지**
- ✅ **의존성 자동 설치** (Go, Homebrew, istats)
- ✅ **최적화된 빌드** (아키텍처별)
- ✅ **자동 시작 설정** (LaunchAgent)
- ✅ **설정 파일 생성** (새로운 AI 기능 포함)
- ✅ **종합 테스트 실행**

#### 설치 후 확인:
```bash
# 설치 확인
which syslog-monitor

# 새로운 AI 기능 확인
cat ~/.syslog-monitor/config.json

# 컴퓨터명 확인
hostname
```

### Linux 설치

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install golang-go git

# CentOS/RHEL
sudo yum install golang git

# 빌드 및 설치
git clone <repository-url>
cd syslog-monitor
make build-linux
sudo cp syslog-monitor_linux /usr/local/bin/syslog-monitor
```

### 빌드 옵션

```bash
# 현재 플랫폼용
make build

# macOS 전용 빌드
make build-macos              # 현재 아키텍처
make build-macos-arm64        # Apple Silicon
make build-macos-intel        # Intel Mac
make build-macos-universal    # 유니버설 바이너리

# 모든 플랫폼
make build-all

# 정리
make clean
```

## 📖 사용법

### 기본 명령어

```bash
# 기본 모니터링
syslog-monitor

# AI 분석 활성화 (권장)
syslog-monitor -ai-analysis

# 전체 기능 활성화
syslog-monitor -ai-analysis -system-monitor -login-watch

# 특정 로그 파일 모니터링
syslog-monitor -file=/var/log/auth.log -ai-analysis
```

### macOS 사용자 전용

```bash
# 실시간 시스템 로그 (권한 필요)
sudo log stream | syslog-monitor -file=/dev/stdin -ai-analysis

# 특정 로그 파일들
syslog-monitor -file=/var/log/system.log -ai-analysis
syslog-monitor -file=/var/log/install.log -keywords=error
syslog-monitor -file=/var/log/wifi.log -system-monitor

# 에러 로그만 필터링
sudo log show --predicate 'eventMessage contains "error"' --last 1h
```

### 고급 사용 예시

```bash
# 보안 모니터링 (SSH, sudo 로그인 감시)
syslog-monitor -ai-analysis -login-watch

# 성능 모니터링
syslog-monitor -system-monitor -keywords="memory,cpu,disk"

# 다중 채널 알림
syslog-monitor -ai-analysis \
  -email-to="admin@company.com,security@company.com" \
  -slack-webhook="https://hooks.slack.com/..."

# 필터링 및 출력
syslog-monitor -keywords="error,failed" \
  -filters="systemd,kernel" \
  -output=./filtered.log
```

## 🤖 AI 분석 기능

### 새로운 v2.0 AI 기능

#### 1. 시스템 정보 자동 수집
```json
{
  "computer_name": "beakerui-MacBookPro.local",
  "internal_ips": ["192.168.1.100", "10.0.0.50"],
  "external_ips": ["203.0.113.42"],
  "asn_data": [
    {
      "ip": "203.0.113.42",
      "organization": "Example Corp",
      "country": "United States",
      "asn": "AS64496"
    }
  ]
}
```

#### 2. 지능형 위험 감지
- **SQL 인젝션**: `OR 1=1`, `UNION SELECT` 등 패턴 감지
- **무차별 대입 공격**: 반복 로그인 실패 패턴 분석
- **권한 상승**: `sudo su`, 비인가 접근 감지
- **메모리 누수**: 메모리 할당 실패 패턴 분석

#### 3. 예측 분석
```
🔮 AI 예측 예시:
┌─────────────────────────────────────┐
│ 시스템 메모리 부족                    │
│ 확률: 75% | 시간: 30분 이내           │
│ 영향: 서비스 중단 가능성               │
│ 권장: 메모리 정리 및 프로세스 점검      │
└─────────────────────────────────────┘
```

### AI 분석 설정

```bash
# AI 분석 임계값 조정
syslog-monitor -ai-analysis -alert-threshold=8.0

# 특정 로그 타입만 AI 분석
syslog-monitor -ai-analysis -log-type=nginx

# AI 분석 결과 로그 저장
syslog-monitor -ai-analysis -output=./ai-analysis.log
```

## 🖥️ 시스템 모니터링

### 모니터링 메트릭

| 메트릭 | 설명 | 임계값 |
|--------|------|--------|
| **CPU 사용률** | 실시간 CPU 사용량 | 80% |
| **메모리 사용률** | RAM 사용률 | 85% |
| **디스크 사용률** | 디스크 공간 사용률 | 90% |
| **로드 평균** | 시스템 부하 | 2.0 |
| **온도** | CPU/시스템 온도 | 70°C |
| **네트워크** | 패킷 손실률 | 5% |

### 시스템 모니터링 명령어

```bash
# 기본 시스템 모니터링
syslog-monitor -system-monitor

# 사용자 정의 임계값
syslog-monitor -system-monitor -cpu-threshold=70 -memory-threshold=80

# 온도 모니터링 (macOS - istats 필요)
brew install istat-menus
syslog-monitor -system-monitor

# 실시간 시스템 상태 보고서
syslog-monitor -system-monitor -report-interval=300  # 5분마다
```

## 📧 알림 설정

### 이메일 알림

#### 환경변수 설정 (권장)
```bash
export SYSLOG_EMAIL_TO="admin@company.com,security@company.com"
export SYSLOG_SMTP_USER="your@gmail.com"
export SYSLOG_SMTP_PASSWORD="your-app-password"

syslog-monitor -ai-analysis
```

#### 명령행 설정
```bash
syslog-monitor -ai-analysis \
  -email-to="admin@company.com,security@company.com" \
  -smtp-user="your@gmail.com" \
  -smtp-password="your-app-password"
```

#### Gmail 설정
1. **2단계 인증 활성화**: Google 계정에서 2단계 인증 설정
2. **앱 비밀번호 생성**: https://myaccount.google.com/apppasswords
3. **앱 비밀번호 사용**: 일반 비밀번호 대신 앱 비밀번호 사용

### Slack 알림

```bash
# Slack 웹훅 설정
syslog-monitor -ai-analysis \
  -slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -slack-channel="#security-alerts"

# 환경변수 사용
export SYSLOG_SLACK_WEBHOOK="https://hooks.slack.com/..."
export SYSLOG_SLACK_CHANNEL="#alerts"
syslog-monitor -ai-analysis
```

### 향상된 알림 내용

v2.0의 알림에는 다음 정보가 포함됩니다:

```
🚨 보안 이상 탐지 알람
======================
⚠️  위협 레벨: 🔴 CRITICAL
📊 이상 점수: 9.0/10.0
🕐 탐지 시간: 2025-07-29 15:30:00

🖥️  시스템 정보:
  📍 컴퓨터명: beakerui-MacBookPro.local
  🏠 내부 IP: 192.168.1.100, 10.0.0.50
  🌐 외부 IP: 203.0.113.42

🔍 ASN 정보:
  📍 203.0.113.42
    🏢 조직: Example Corp
    🌍 국가: United States, California, San Francisco
    🔢 ASN: AS64496

📋 로그 정보:
  📝 레벨: CRITICAL
  🏷️  서비스: database
  💬 메시지: SQL injection attempt detected

🔮 위험 예측:
  ⚡ 추가 공격 시도 (확률: 85%, 1시간 이내)
    💥 영향: 데이터 유출 위험

💡 권장사항:
  • 🚨 즉시 보안팀에 알림
  • 🔒 해당 IP 주소 차단 검토
  • 📊 시스템 리소스 사용량 확인

🎯 신뢰도: 95%
```

## 🧪 테스트

### 빠른 기본 테스트

```bash
# 간단한 기능 테스트
./simple-test.sh
```

예상 출력:
```
🧪 AI Syslog Monitor 간단 테스트
==================================
1️⃣ 기본 설치 확인
✅ 실행 파일 설치됨
✅ 설정 파일 생성됨
✅ 컴퓨터명 감지 기능 활성화
✅ IP 분류 기능 활성화
✅ ASN 조회 기능 활성화

2️⃣ 시스템 정보 수집
✅ 컴퓨터명 수집: beakerui-MacBookPro.local
✅ 시스템 아키텍처: arm64
✅ 인터넷 연결 (ASN 조회 가능)

📊 테스트 결과: 100% 통과
🎉 모든 테스트 통과!
```

### 종합 상세 테스트

```bash
# 전체 기능 테스트 (10-15분 소요)
./test-ai-features.sh
```

테스트 항목:
- ✅ 설치 상태 확인
- ✅ 기본 실행 테스트
- ✅ 시스템 정보 수집
- ✅ AI 분석 기능
- ✅ 보안 위협 시나리오 (6가지)
- ✅ ASN 정보 조회
- ✅ IP 주소 분류
- ✅ 성능 테스트 (1000줄 로그)
- ✅ 메모리 사용량
- ✅ 로그 출력 형식

### 수동 테스트

```bash
# 이메일 알림 테스트
syslog-monitor -test-email

# Slack 알림 테스트  
syslog-monitor -test-slack -slack-webhook="YOUR_WEBHOOK"

# AI 분석 테스트
echo "$(date) CRITICAL [security] SQL injection detected" | \
  syslog-monitor -file=/dev/stdin -ai-analysis
```

## ⚙️ 설정 파일

### 자동 생성된 설정 파일
위치: `~/.syslog-monitor/config.json`

```json
{
    "ai_analysis": true,
    "system_monitoring": true,
    "log_file": "/var/log/system.log",
    "alert_threshold": 7.0,
    "email_alerts": true,
    "slack_alerts": false,
    "features": {
        "computer_name_detection": true,
        "ip_classification": true,
        "asn_lookup": true,
        "real_time_analysis": true
    }
}
```

### 환경변수

| 변수명 | 설명 | 기본값 |
|--------|------|--------|
| `SYSLOG_EMAIL_TO` | 수신자 이메일 (쉼표 구분) | `robot@lambda-x.ai,enfn2001@gmail.com` |
| `SYSLOG_SMTP_USER` | SMTP 사용자명 | `enfn2001@gmail.com` |
| `SYSLOG_SMTP_PASSWORD` | SMTP 비밀번호/앱 비밀번호 | 설정됨 |
| `SYSLOG_SLACK_WEBHOOK` | Slack 웹훅 URL | - |
| `SYSLOG_SLACK_CHANNEL` | Slack 채널 | - |

## 🔧 명령행 옵션

### 기본 옵션
```bash
syslog-monitor [옵션]

주요 옵션:
  -file string          모니터링할 로그 파일 경로
  -output string        필터링된 로그 출력 파일
  -keywords string      포함할 키워드 (쉼표 구분)
  -filters string       제외할 패턴 (정규식, 쉼표 구분)
  -help                 도움말 표시
```

### AI 분석 옵션
```bash
  -ai-analysis          AI 기반 로그 분석 활성화
  -alert-threshold      AI 알림 임계값 (기본: 7.0)
  -log-type string      로그 타입 (auto, apache, nginx, mysql)
```

### 시스템 모니터링 옵션
```bash
  -system-monitor       시스템 메트릭 모니터링 활성화
  -cpu-threshold        CPU 사용률 임계값 (기본: 80)
  -memory-threshold     메모리 사용률 임계값 (기본: 85)
  -disk-threshold       디스크 사용률 임계값 (기본: 90)
```

### 알림 옵션
```bash
  -email-to string      수신자 이메일 (쉼표 구분)
  -smtp-server string   SMTP 서버 (기본: smtp.gmail.com)
  -smtp-port string     SMTP 포트 (기본: 587)
  -smtp-user string     SMTP 사용자명
  -smtp-password string SMTP 비밀번호
  -slack-webhook string Slack 웹훅 URL
  -slack-channel string Slack 채널
```

### 보안 옵션
```bash
  -login-watch          로그인 모니터링 활성화 (SSH, sudo, 웹)
```

### 테스트 옵션
```bash
  -test-email           이메일 설정 테스트
  -test-slack           Slack 설정 테스트
```

## 🔄 자동 시작 설정

### macOS LaunchAgent

설치 스크립트가 자동으로 생성하는 파일:
`~/Library/LaunchAgents/ai.lambda-x.syslog-monitor.plist`

```bash
# 수동 시작/중지
launchctl load ~/Library/LaunchAgents/ai.lambda-x.syslog-monitor.plist
launchctl unload ~/Library/LaunchAgents/ai.lambda-x.syslog-monitor.plist

# 상태 확인
launchctl list | grep syslog-monitor
```

### Linux Systemd

```bash
# 서비스 파일 생성
sudo tee /etc/systemd/system/syslog-monitor.service << EOF
[Unit]
Description=AI Syslog Monitor
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/syslog-monitor -ai-analysis -system-monitor
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 서비스 활성화
sudo systemctl enable syslog-monitor
sudo systemctl start syslog-monitor
sudo systemctl status syslog-monitor
```

## 🔍 문제 해결

### 일반적인 문제

#### 1. 권한 오류
```bash
# macOS
sudo chown $(whoami) /var/log/system.log
# 또는 sudo로 실행
sudo syslog-monitor -ai-analysis

# Linux
sudo chmod 644 /var/log/syslog
```

#### 2. 이메일 전송 실패
```bash
# Gmail 앱 비밀번호 확인
# 2단계 인증 활성화 여부 확인
# SMTP 설정 테스트
syslog-monitor -test-email
```

#### 3. AI 분석 오류
```bash
# 인터넷 연결 확인 (ASN 조회용)
ping 8.8.8.8

# 로그 파일 접근 권한 확인
ls -la /var/log/system.log

# 간단한 테스트
./simple-test.sh
```

#### 4. 메모리 사용량 높음
```bash
# 로그 버퍼 크기 조정 (기본: 1000줄)
# 임계값 조정 (기본: 7.0)
syslog-monitor -ai-analysis -alert-threshold=8.5
```

### macOS 특화 문제

#### 1. 로그 파일 접근
```bash
# macOS Big Sur/Monterey 이후
sudo log stream --predicate 'process == "kernel"' | \
  syslog-monitor -file=/dev/stdin -ai-analysis

# 권한 부여
sudo chmod +r /var/log/system.log
```

#### 2. 온도 모니터링
```bash
# istats 설치
brew install istat-menus

# 수동 온도 확인
istats temp
```

### 로그 파일 위치

#### macOS
- 시스템 로그: `/var/log/system.log`
- 설치 로그: `/var/log/install.log`
- WiFi 로그: `/var/log/wifi.log`
- 보안 로그: `/var/log/secure.log`

#### Linux
- 시스템 로그: `/var/log/syslog` (Ubuntu/Debian)
- 시스템 로그: `/var/log/messages` (CentOS/RHEL)
- 인증 로그: `/var/log/auth.log`
- 커널 로그: `/var/log/kern.log`

## 🎯 성능 최적화

### 권장 설정

```bash
# 일반 사용 (권장)
syslog-monitor -ai-analysis -system-monitor

# 고성능 서버
syslog-monitor -ai-analysis -alert-threshold=8.0 \
  -keywords="error,critical,failed"

# 보안 중심
syslog-monitor -ai-analysis -login-watch \
  -keywords="failed,unauthorized,attack"

# 경량 모니터링
syslog-monitor -keywords="error,critical" \
  -filters="debug,info"
```

### 리소스 사용량

| 구성 | CPU | 메모리 | 디스크 |
|------|-----|--------|--------|
| 기본 모니터링 | <5% | 20-50MB | 최소 |
| AI 분석 | 5-15% | 50-100MB | 낮음 |
| 시스템 모니터링 | 10-20% | 100-200MB | 보통 |
| 전체 기능 | 15-25% | 150-300MB | 보통 |

## 📚 추가 리소스

### 설정 예시 모음

```bash
# 1. 웹 서버 모니터링
syslog-monitor -file=/var/log/nginx/access.log \
  -ai-analysis -keywords="error,404,500"

# 2. 데이터베이스 모니터링  
syslog-monitor -file=/var/log/mysql/error.log \
  -ai-analysis -log-type=mysql

# 3. 보안 모니터링
syslog-monitor -file=/var/log/auth.log \
  -ai-analysis -login-watch

# 4. 개발 환경
syslog-monitor -file=./app.log \
  -keywords="error,exception" -output=./filtered.log
```

### API 연동

ASN 정보 조회에 사용되는 API:
- **ip-api.com**: 무료, 월 1000회 제한
- **ipinfo.io**: 유료, 높은 정확도
- **MaxMind GeoIP**: 로컬 데이터베이스

## 🤝 기여하기

1. **이슈 리포트**: 버그나 기능 요청
2. **코드 기여**: Pull Request 환영
3. **문서화**: README 개선사항
4. **테스트**: 새로운 환경에서의 테스트

## 📄 라이선스

MIT License - 자유롭게 사용, 수정, 배포 가능

## 🔗 링크

- **GitHub**: [프로젝트 저장소]
- **문서**: [온라인 문서]
- **이슈 트래킹**: [GitHub Issues]

---

**🎉 AI-Powered Syslog Monitor v2.0**  
**더 스마트하고, 더 안전하고, 더 강력한 로그 모니터링 솔루션**

**Made with ❤️ by Lambda-X AI Team** 