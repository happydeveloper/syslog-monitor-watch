# AI-Powered Syslog Monitor

**🤖 AI 기반 로그 분석 및 시스템 모니터링 도구**

리눅스와 macOS 시스템의 syslog를 실시간으로 감시하고 AI 기반 이상 징후 분석, 시스템 메트릭 모니터링, **다중 플랫폼 알림**을 제공하는 차세대 모니터링 솔루션입니다.

## 목차
- [기능](#기능)
- [빠른 시작](#빠른-시작-quick-start)
- [빌드 방법](#빌드-방법)
- [상세 사용법](#상세-사용법)
- [명령행 옵션](#명령행-옵션)
- [환경변수](#환경변수)
- [Gmail 설정 방법](#gmail-설정-방법)
- [테스트 가이드](#테스트-가이드)
- [문제 해결 및 FAQ](#문제-해결-및-faq)
- [시스템 서비스로 실행](#시스템-서비스로-실행)

## 기능

### 🔍 기본 모니터링
- **실시간 로그 감시**: syslog 파일을 실시간으로 모니터링
- **키워드 필터링**: 특정 키워드가 포함된 로그만 표시
- **정규식 필터**: 정규식을 사용한 로그 제외 필터링
- **로그 레벨 분류**: ERROR, WARNING, CRITICAL, INFO 자동 분류
- **파일 출력**: 필터링된 로그를 파일로 저장

### 🤖 AI 기반 로그 분석
- **이상 패턴 감지**: SQL 인젝션, 무차별 대입 공격, 메모리 누수 등 자동 감지
- **예측 분석**: 시스템 장애 및 보안 위협 사전 예측
- **위협 레벨 평가**: 실시간 위험도 점수 계산 (0-10점)
- **다중 로그 포맷 지원**: Apache, Nginx, MySQL, PostgreSQL, 애플리케이션 로그
- **시간 패턴 분석**: 야간/주말 비정상 활동 감지
- **빈도 분석**: 비정상적인 로그 급증 패턴 감지

### 🖥️ 시스템 메트릭 모니터링
- **실시간 시스템 감시**: CPU, 메모리, 디스크, 온도 모니터링
- **임계값 알림**: 설정 가능한 임계값 초과 시 즉시 알림
- **로드 평균 추적**: 시스템 부하 상태 실시간 모니터링
- **네트워크 상태**: 패킷 손실, 에러율 추적
- **디스크 공간**: 사용률 및 inode 사용률 모니터링
- **프로세스 상태**: 실행 중인 프로세스 수 추적

### 📧 다중 알림 시스템
- **이메일 알림**: 에러/크리티컬 로그 발생 시 자동 이메일 전송
- **Slack 통합**: 실시간 Slack 채널 알림
- **Gmail 지원**: Gmail SMTP를 통한 이메일 알림
- **환경변수 설정**: 환경변수를 통한 간편한 이메일 설정
- **AI 분석 결과 알림**: 이상 징후 감지 시 상세 분석 보고서 전송
- **시스템 알림**: 시스템 임계값 초과 시 즉시 알림

### 🔐 보안 기능
- **로그인 모니터링**: SSH, sudo, 웹 로그인 패턴 감지
- **보안 위협 탐지**: 권한 상승, 무차별 대입 공격 등
- **IP 추적**: 의심스러운 IP 주소 자동 추적
- **신호 처리**: Ctrl+C로 안전한 종료

## 빌드 방법

### 필요 조건
- Go 1.21 이상
- Linux 또는 macOS

### 빌드 명령어

```bash
# 의존성 설치
make install

# 현재 플랫폼용 빌드
make build

# Linux용 빌드 (크로스 컴파일)
make build-linux

# macOS/Unix용 빌드
make build-unix

# 빌드 결과물 정리
make clean
```

## 사용법

### 기본 사용법

```bash
# 기본 syslog 파일 (/var/log/syslog) 감시
./syslog-monitor

# 특정 파일 감시
./syslog-monitor -file=/var/log/auth.log

# 도움말 보기
./syslog-monitor -help
```

### 고급 사용법

```bash
# 특정 키워드만 포함된 로그 감시
./syslog-monitor -keywords=error,failed,warning

# 특정 패턴 제외하고 감시 (정규식 사용)
./syslog-monitor -filters="systemd,kernel"

# 결과를 파일로 저장
./syslog-monitor -output=security.log -keywords=failed,unauthorized

# 복합 필터링 예제
./syslog-monitor -file=/var/log/auth.log -keywords=failed,error -output=security_alerts.log
```

## 명령행 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-file` | `/var/log/syslog` | 감시할 syslog 파일 경로 |
| `-keywords` | (없음) | 포함할 키워드 (쉼표로 구분) |
| `-filters` | (없음) | 제외할 정규식 패턴 (쉼표로 구분) |
| `-output` | stdout | 출력할 파일 경로 |
| `-email-to` | (없음) | 알림받을 이메일 주소 (쉼표로 구분) |
| `-email-from` | (자동설정) | 발신자 이메일 주소 |
| `-smtp-server` | `smtp.gmail.com` | SMTP 서버 주소 |
| `-smtp-port` | `587` | SMTP 포트 |
| `-smtp-user` | (없음) | SMTP 사용자명 |
| `-smtp-password` | (없음) | SMTP 비밀번호 |
| `-test-email` | - | 테스트 이메일 전송 후 종료 |
| `-help` | - | 도움말 표시 |

## 환경변수

| 변수명 | 설명 |
|--------|------|
| `SYSLOG_EMAIL_TO` | 알림받을 이메일 주소 (쉼표로 구분) |
| `SYSLOG_EMAIL_FROM` | 발신자 이메일 주소 |
| `SYSLOG_SMTP_SERVER` | SMTP 서버 주소 |
| `SYSLOG_SMTP_PORT` | SMTP 포트 |
| `SYSLOG_SMTP_USER` | SMTP 사용자명 |
| `SYSLOG_SMTP_PASSWORD` | SMTP 비밀번호 |

## Gmail 설정 방법

### 1. 2단계 인증 활성화
1. [Google 계정 설정](https://myaccount.google.com/) 접속
2. 보안 > 2단계 인증 활성화

### 2. App Password 생성
1. [App Passwords](https://myaccount.google.com/apppasswords) 접속
2. 앱 선택 > 메일
3. 기기 선택 > 기타 (사용자 정의 이름)
4. 생성된 16자리 비밀번호 복사

### 3. 설정 적용
```bash
# 간편 설정 스크립트 사용
./email-setup.sh

# 또는 직접 환경변수 설정
export SYSLOG_EMAIL_TO="enfn2001@gmail.com"
export SYSLOG_SMTP_USER="your@gmail.com"
export SYSLOG_SMTP_PASSWORD="generated-app-password"

# 테스트
./syslog-monitor -test-email
```

## 빠른 시작 (Quick Start)

### 🍎 macOS 사용자 (권장)
```bash
# 원클릭 설치 (AI 분석 + 시스템 모니터링 포함)
./install-macos.sh

# 또는 수동 설치
make build-macos && sudo make install-macos

# AI 기반 모니터링 시작
syslog-monitor -ai-analysis -system-monitor
```

### 🐧 Linux 사용자
```bash
# 빌드
make build-linux

# AI 기반 모니터링 시작  
./syslog-monitor_linux -ai-analysis -system-monitor
```

### 🚀 즉시 실행 (기본 설정)
```bash
# 현재 플랫폼용 빌드
make build

# 기본 설정으로 즉시 시작 (robot@lambda-x.ai, enfn2001@gmail.com에게 자동 알림)
./syslog-monitor

# AI 분석 + 시스템 모니터링 활성화
./syslog-monitor -ai-analysis -system-monitor

# 테스트 이메일 전송
./syslog-monitor -test-email
```

### 📧 이메일 알림 테스트

#### 1. 기본 설정 테스트 (2명 수신자)
```bash
./syslog-monitor -test-email
# 결과: robot@lambda-x.ai, enfn2001@gmail.com에게 전송
```

#### 2. 커스텀 여러 명 테스트
```bash
./syslog-monitor -test-email -email-to="admin@company.com,security@company.com,ops@company.com"
# 결과: 3명에게 동시 전송
```

## 상세 사용법

### 1. 🤖 AI 기반 로그 분석
```bash
# AI 이상 징후 감지 활성화
./syslog-monitor -ai-analysis

# 특정 로그 타입으로 AI 분석
./syslog-monitor -file=/var/log/nginx/access.log -log-type=nginx -ai-analysis

# 데이터베이스 로그 AI 분석
./syslog-monitor -file=/var/log/mysql/error.log -log-type=mysql -ai-analysis
```

### 2. 🖥️ 시스템 메트릭 모니터링
```bash
# 시스템 리소스 모니터링 활성화
./syslog-monitor -system-monitor

# AI 분석 + 시스템 모니터링 통합
./syslog-monitor -ai-analysis -system-monitor

# macOS 특화 모니터링 (배터리, APFS, 온도)
./syslog-monitor -ai-analysis -system-monitor -file=/var/log/system.log
```

### 3. 🔍 기본 syslog 감시
```bash
# 기본 syslog 파일 감시 (이메일 알림 포함)
./syslog-monitor

# 특정 파일 감시
./syslog-monitor -file=/var/log/auth.log

# 특정 키워드만 감시
./syslog-monitor -keywords=error,critical,failed,warning

# 완전한 보안 모니터링
./syslog-monitor -login-watch -ai-analysis -keywords=failed,error,critical
```

### 4. 📧 이메일 알림 설정

#### 방법 1: 기본 설정 사용 (추천)
```bash
# 자동으로 robot@lambda-x.ai, enfn2001@gmail.com에게 알림
./syslog-monitor -keywords=error,critical,failed

# AI 분석 결과 알림 포함
./syslog-monitor -ai-analysis -system-monitor
```

#### 방법 2: 여러 명 커스텀 설정
```bash
# 팀 전체에게 알림
./syslog-monitor -email-to="admin@company.com,security@company.com,ops@company.com,cto@company.com"

# 프로젝트 팀에게 알림
./syslog-monitor -file=/var/log/app.log -email-to="dev@lambda-x.ai,ops@lambda-x.ai,pm@lambda-x.ai"
```

#### 방법 3: 간편 설정 스크립트
```bash
./email-setup.sh  # 대화형 설정
source .env        # 환경변수 로드
./syslog-monitor -test-email  # 테스트
```

#### 방법 4: 환경변수 설정
```bash
export SYSLOG_EMAIL_TO="team1@company.com,team2@company.com,manager@company.com"
export SYSLOG_SMTP_USER="your@gmail.com"
export SYSLOG_SMTP_PASSWORD="your-app-password"
./syslog-monitor
```

### 5. 🏢 실제 운영 시나리오

#### 🛡️ 통합 보안 모니터링 (AI + 시스템)
```bash
./syslog-monitor \
  -ai-analysis \
  -system-monitor \
  -login-watch \
  -file=/var/log/auth.log \
  -keywords=failed,unauthorized,invalid,breach \
  -email-to="security@company.com,admin@company.com,ciso@company.com" \
  -slack-webhook="https://hooks.slack.com/..." \
  -output=security_alerts.log
```

#### 🍎 macOS 서버 모니터링
```bash
./syslog-monitor \
  -ai-analysis \
  -system-monitor \
  -file=/var/log/system.log \
  -keywords=error,warning,failed \
  -email-to="ops@company.com,admin@company.com"
```

#### 🌐 웹서버 성능 + 보안 모니터링
```bash
./syslog-monitor \
  -file=/var/log/nginx/access.log \
  -log-type=nginx \
  -ai-analysis \
  -keywords=error,502,503,504 \
  -email-to="webteam@company.com,ops@company.com"
```

#### 웹서버 에러 감시
```bash
./syslog-monitor \
  -file=/var/log/nginx/error.log \
  -keywords=error,502,503,504 \
  -email-to="webteam@company.com,ops@company.com"
```

#### 데이터베이스 크리티컬 감시
```bash
./syslog-monitor \
  -file=/var/log/mysql/error.log \
  -keywords=critical,error,crash,deadlock \
  -email-to="dba@company.com,ops@company.com,cto@company.com"
```

#### 로그 필터링 + 다중 알림
```bash
# systemd, kernel 로그 제외하고 감시
./syslog-monitor \
  -filters="systemd.*,kernel.*,cron.*" \
  -keywords=error,critical,failed \
  -email-to="admin@company.com,security@company.com"
```

## 실행 출력 예시

### 시작 시 출력
```bash
$ ./syslog-monitor -file=test.log -keywords=error,critical,failed

📧 Email alerts enabled with DEFAULT settings
    📨 Recipients (2): robot@lambda-x.ai, enfn2001@gmail.com
    🔑 Using built-in Gmail credentials (enfn2001@gmail.com)
    💡 To add more recipients: -email-to="user1@example.com,user2@example.com"

INFO[2025-07-26 00:12:42] Starting syslog monitor for file: test.log   
INFO[2025-07-26 00:12:42] Syslog monitor started. Press Ctrl+C to stop. 
2025/07/26 00:12:42 Seeked test.log - &{Offset:0 Whence:2}
```

### 에러 감지 및 이메일 전송
```bash
ERRO[2025-07-26 00:12:55] Error loading configuration file - file not found  fields.level=ERROR host=server01 service="app:"
INFO[2025-07-26 00:12:55] 📧 Sending ERROR alert to: robot@lambda-x.ai, enfn2001@gmail.com 
INFO[2025-07-26 00:12:58] ✅ Gmail email sent successfully to: robot@lambda-x.ai, enfn2001@gmail.com 

FATA[2025-07-26 00:13:10] Critical database failure - all connections lost  fields.level=CRITICAL host=database service="mysql:"
INFO[2025-07-26 00:13:10] 🚨 Sending CRITICAL alert to: robot@lambda-x.ai, enfn2001@gmail.com 
INFO[2025-07-26 00:13:13] ✅ Gmail email sent successfully to: robot@lambda-x.ai, enfn2001@gmail.com 
```

### 테스트 이메일 출력
```bash
$ ./syslog-monitor -test-email

📧 Email alerts enabled with DEFAULT settings
    📨 Recipients (2): robot@lambda-x.ai, enfn2001@gmail.com
    🔑 Using built-in Gmail credentials (enfn2001@gmail.com)
    💡 To add more recipients: -email-to="user1@example.com,user2@example.com"

Sending test email...
INFO[2025-07-26 00:12:10] ✅ Gmail email sent successfully to: robot@lambda-x.ai, enfn2001@gmail.com 
✅ Test email sent successfully to 2 recipients: robot@lambda-x.ai, enfn2001@gmail.com
```

### 여러 명 커스텀 설정 출력
```bash
$ ./syslog-monitor -test-email -email-to="admin@company.com,security@company.com,ops@company.com"

📧 Email alerts enabled with CUSTOM settings
    📨 Recipients (3): admin@company.com, security@company.com, ops@company.com

Sending test email...
INFO[2025-07-26 00:12:30] ✅ Gmail email sent successfully to: admin@company.com, security@company.com, ops@company.com
✅ Test email sent successfully to 3 recipients: admin@company.com, security@company.com, ops@company.com
```

## 로그 레벨 분류

- **🔴 ERROR**: "error", "err" 키워드 포함 → 이메일 알림 전송
- **⚠️ WARNING**: "warn", "warning" 키워드 포함 → 로그만 기록
- **🚨 CRITICAL**: "critical", "fail" 키워드 포함 → 긴급 이메일 알림 전송
- **ℹ️ INFO**: 기타 모든 로그 → 로그만 기록

## 시스템 서비스로 실행

### systemd 서비스 설정

`/etc/systemd/system/syslog-monitor.service` 파일 생성:

```ini
[Unit]
Description=Syslog Monitor Service
After=network.target

[Service]
Type=simple
User=syslog
ExecStart=/usr/local/bin/syslog-monitor -output=/var/log/syslog-monitor.log
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

서비스 등록 및 시작:
```bash
sudo systemctl daemon-reload
sudo systemctl enable syslog-monitor
sudo systemctl start syslog-monitor
```

## 테스트 가이드

### 🧪 기능 테스트

#### 1. 이메일 알림 테스트
```bash
# 기본 2명에게 테스트 이메일
./syslog-monitor -test-email
# 예상 결과: ✅ Test email sent successfully to 2 recipients: robot@lambda-x.ai, enfn2001@gmail.com

# 커스텀 여러 명에게 테스트
./syslog-monitor -test-email -email-to="user1@test.com,user2@test.com,user3@test.com"
# 예상 결과: ✅ Test email sent successfully to 3 recipients: user1@test.com, user2@test.com, user3@test.com
```

#### 2. 실시간 로그 감시 테스트
```bash
# 터미널 1: 감시 시작
./syslog-monitor -file=test.log -keywords=error,critical,failed

# 터미널 2: 테스트 로그 추가
echo "$(date) server01 app: Error loading configuration file" >> test.log
echo "$(date) server01 db: Critical database connection failed" >> test.log

# 예상 결과:
# INFO[시간] 📧 Sending ERROR alert to: robot@lambda-x.ai, enfn2001@gmail.com
# INFO[시간] ✅ Gmail email sent successfully to: robot@lambda-x.ai, enfn2001@gmail.com
```

#### 3. 필터링 테스트
```bash
# 특정 패턴 제외 테스트
./syslog-monitor -file=test.log -filters="systemd,kernel" -keywords=error

# 키워드 조합 테스트
./syslog-monitor -file=test.log -keywords="error,critical,failed,warning"
```

### 📊 성능 테스트

#### 대용량 로그 파일 테스트
```bash
# 대용량 파일 생성
for i in {1..1000}; do 
  echo "$(date) server01 app: Test log entry $i" >> large_test.log
done

# 감시 성능 확인
time ./syslog-monitor -file=large_test.log -keywords=test
```

#### 다중 수신자 성능 테스트
```bash
# 10명에게 동시 전송 테스트
./syslog-monitor -test-email -email-to="user1@test.com,user2@test.com,user3@test.com,user4@test.com,user5@test.com,user6@test.com,user7@test.com,user8@test.com,user9@test.com,user10@test.com"
```

## 문제 해결 및 FAQ

### ❌ 이메일 전송 실패

#### 문제: 535 5.7.8 Username and Password not accepted
```
ERRO[시간] ❌ Failed to send email alert: 535 5.7.8 Username and Password not accepted
```

**해결방법:**
1. Gmail 2단계 인증 활성화 확인
2. App Password 재생성: https://myaccount.google.com/apppasswords
3. 올바른 App Password 설정 확인

```bash
# 테스트로 확인
./syslog-monitor -test-email -smtp-user=your@gmail.com -smtp-password=correct-app-password
```

#### 문제: TLS 연결 오류
```
ERRO[시간] failed to connect to SMTP server: tls: first record does not look like a TLS handshake
```

**해결방법:**
- Gmail SMTP는 자동으로 처리됩니다. 기본 설정 사용:
```bash
./syslog-monitor -test-email  # 기본 Gmail 설정 사용
```

### ⚠️ 파일 접근 문제

#### 권한 문제
```bash
# syslog 파일은 보통 root 권한 필요
sudo ./syslog-monitor -file=/var/log/syslog

# 또는 사용자를 syslog 그룹에 추가
sudo usermod -a -G syslog $USER
```

#### 파일 경로 확인
```bash
# 시스템별 syslog 위치 확인
ls -la /var/log/syslog      # Ubuntu/Debian
ls -la /var/log/messages    # CentOS/RHEL
ls -la /var/log/system.log  # macOS
```

### 🔧 성능 최적화

#### inotify 한계 증가
```bash
# 현재 한계 확인
cat /proc/sys/fs/inotify/max_user_watches

# 한계 증가 (root 권한 필요)
echo 524288 | sudo tee /proc/sys/fs/inotify/max_user_watches
```

#### 메모리 사용량 확인
```bash
# 실행 중 메모리 사용량 모니터링
ps aux | grep syslog-monitor
top -p $(pgrep syslog-monitor)
```

### 🌐 네트워크 문제

#### SMTP 연결 테스트
```bash
# Gmail SMTP 서버 연결 확인
telnet smtp.gmail.com 587

# 방화벽 확인
sudo iptables -L | grep 587
```

### 📝 로그 레벨 이해

- **ERROR**: `error`, `err` 키워드 포함
- **WARNING**: `warn`, `warning` 키워드 포함  
- **CRITICAL**: `critical`, `fail` 키워드 포함
- **INFO**: 기타 모든 로그

### 💡 팁과 트릭

#### 환경변수 파일 사용
```bash
# .env 파일 생성
cat > .env << EOF
export SYSLOG_EMAIL_TO="team@company.com,admin@company.com"
export SYSLOG_SMTP_USER="alerts@company.com"
export SYSLOG_SMTP_PASSWORD="app-password-here"
EOF

# 보안 설정
chmod 600 .env

# 사용
source .env && ./syslog-monitor
```

#### systemd 서비스 자동 재시작
```bash
# 서비스 파일에 재시작 정책 추가
[Service]
Restart=always
RestartSec=10
```

## 라이센스

MIT License

## 기여하기

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 프로젝트 구조

```
syslog-monitor/
├── main.go                # 메인 소스코드
├── go.mod                 # Go 모듈 정의
├── Makefile              # 빌드 스크립트
├── email-setup.sh        # 이메일 설정 간편 스크립트
├── README.md             # 이 문서
├── test.log              # 테스트용 로그 파일
├── syslog-monitor        # 빌드된 실행파일
└── .env                  # 환경변수 파일 (생성 후)
```

## 주요 파일 설명

- **`main.go`**: 핵심 로직 (syslog 감시, 이메일 전송, 필터링)
- **`email-setup.sh`**: 대화형 이메일 설정 스크립트
- **`Makefile`**: 빌드, 테스트, 정리 명령어
- **`test.log`**: 기능 테스트용 샘플 로그 파일

## 라이센스

MIT License

## 기여하기

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## 연락처

이슈나 질문이 있으시면 GitHub Issues를 통해 연락해주세요.

## 버전 히스토리

- **v1.0.0**: 기본 syslog 감시 기능
- **v1.1.0**: 이메일 알림 기능 추가
- **v1.2.0**: 여러 명 동시 이메일 알림 지원
- **v1.3.0**: Gmail SMTP 최적화 및 안정성 개선 