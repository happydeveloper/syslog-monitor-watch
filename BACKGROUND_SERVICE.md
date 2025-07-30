# 🔄 백그라운드 서비스 설정 가이드

## 📋 개요

Lambda-X Syslog Monitor를 macOS에서 백그라운드 서비스로 실행하는 완벽한 가이드입니다. 시스템 부팅 시 자동 시작, 프로세스 관리, 로그 로테이션 등 모든 기능을 제공합니다.

## 🚀 자동 설치 (권장)

### 원클릭 설치
```bash
# 프로젝트 디렉토리에서 실행
chmod +x install-service.sh
./install-service.sh
```

이 스크립트는 자동으로:
- ✅ Go 빌드 및 바이너리 설치
- ✅ LaunchAgent 서비스 등록
- ✅ 로그 디렉토리 생성
- ✅ 기본 설정 파일 생성
- ✅ 서비스 시작 및 상태 확인

## 🔧 수동 설치

### 1단계: 빌드 및 설치
```bash
# 바이너리 빌드
go build -o syslog-monitor .

# 시스템에 설치
sudo cp syslog-monitor /usr/local/bin/
sudo chmod +x /usr/local/bin/syslog-monitor
```

### 2단계: 디렉토리 생성
```bash
# 로그 디렉토리
sudo mkdir -p /usr/local/var/log
sudo mkdir -p /usr/local/var/run

# 사용자 설정 디렉토리
mkdir -p ~/.syslog-monitor
```

### 3단계: 서비스 설치
```bash
# LaunchAgent 서비스 설치
syslog-monitor -install-service

# 서비스 시작
syslog-monitor -start-service

# 상태 확인
syslog-monitor -status-service
```

## 📊 서비스 관리 명령어

### 기본 명령어
```bash
# 서비스 상태 확인
syslog-monitor -status-service

# 서비스 시작
syslog-monitor -start-service

# 서비스 중지
syslog-monitor -stop-service

# 서비스 제거
syslog-monitor -remove-service

# 현재 설정 확인
syslog-monitor -show-config
```

### LaunchAgent 직접 관리
```bash
# 서비스 로드 (시작)
launchctl load ~/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist

# 서비스 언로드 (중지)
launchctl unload ~/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist

# 서비스 목록 확인
launchctl list | grep lambda-x

# 서비스 상세 정보
launchctl list com.lambda-x.syslog-monitor
```

## 📁 파일 구조

### 주요 파일 위치
```
/usr/local/bin/
├── syslog-monitor                    # 메인 실행 파일
└── rotate-syslog-logs.sh            # 로그 로테이션 스크립트

~/Library/LaunchAgents/
└── com.lambda-x.syslog-monitor.plist # LaunchAgent 설정

~/.syslog-monitor/
└── config.json                      # 사용자 설정 파일

/usr/local/var/log/
├── syslog-monitor.out.log           # 표준 출력 로그
├── syslog-monitor.err.log           # 에러 로그
├── syslog-monitor.log               # 애플리케이션 로그
├── logrotate.out.log                # 로그 로테이션 출력
└── logrotate.err.log                # 로그 로테이션 에러

/usr/local/var/run/
└── syslog-monitor.pid               # PID 파일 (daemon 모드)
```

## 🔍 로그 모니터링

### 실시간 로그 확인
```bash
# 메인 로그 (실시간)
tail -f /usr/local/var/log/syslog-monitor.out.log

# 에러 로그 (실시간)
tail -f /usr/local/var/log/syslog-monitor.err.log

# 모든 로그 동시 확인
tail -f /usr/local/var/log/syslog-monitor.*.log

# 로그 검색
grep "ERROR" /usr/local/var/log/syslog-monitor.err.log
grep "ALERT" /usr/local/var/log/syslog-monitor.out.log
```

### 로그 분석
```bash
# 오늘의 로그만 확인
grep "$(date '+%Y-%m-%d')" /usr/local/var/log/syslog-monitor.out.log

# 에러 패턴 분석
awk '/ERROR|FAIL/ {print $0}' /usr/local/var/log/syslog-monitor.err.log

# 시간대별 로그 활동
awk '{print $1, $2}' /usr/local/var/log/syslog-monitor.out.log | sort | uniq -c
```

## 🔄 로그 로테이션

### 자동 로그 로테이션 설정
```bash
# 로그 로테이션 스크립트 설치
sudo cp rotate-syslog-logs.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/rotate-syslog-logs.sh

# 로테이션 LaunchAgent 설치 (매일 자정 실행)
cp com.lambda-x.syslog-monitor.logrotate.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist
```

### 수동 로그 로테이션
```bash
# 즉시 로그 로테이션 실행
/usr/local/bin/rotate-syslog-logs.sh

# 로테이션 상태 확인
ls -la /usr/local/var/log/syslog-monitor.*
```

### 로테이션 설정
- **최대 크기**: 100MB (설정 가능)
- **보관 기간**: 30일
- **압축**: 1일 이상 된 로그 자동 압축
- **실행 주기**: 매일 자정

## ⚙️ 서비스 설정

### LaunchAgent 설정 (plist)
주요 설정 항목:
- **RunAtLoad**: 사용자 로그인 시 자동 시작
- **KeepAlive**: 프로세스 크래시 시 자동 재시작
- **ThrottleInterval**: 재시작 간격 (10초)
- **Nice**: 프로세스 우선순위 (1 = 낮은 우선순위)

### 환경변수 설정
```bash
# ~/.zshrc 또는 ~/.bash_profile에 추가
export GEMINI_API_KEY="your-api-key"
export SYSLOG_EMAIL_TO="admin@company.com,ops@company.com"
export SYSLOG_SLACK_WEBHOOK="https://hooks.slack.com/..."
```

### 설정 파일 편집
```bash
# 설정 파일 열기
nano ~/.syslog-monitor/config.json

# 설정 검증
syslog-monitor -show-config
```

## 🚨 트러블슈팅

### 서비스가 시작되지 않는 경우
```bash
# 1. 권한 확인
ls -la /usr/local/bin/syslog-monitor

# 2. 설정 파일 확인
syslog-monitor -show-config

# 3. 에러 로그 확인
tail -50 /usr/local/var/log/syslog-monitor.err.log

# 4. 수동 실행 테스트
/usr/local/bin/syslog-monitor -system-monitor
```

### 서비스가 자주 재시작되는 경우
```bash
# 1. 시스템 로그 확인
log show --predicate 'subsystem contains "com.lambda-x.syslog-monitor"' --last 1h

# 2. 메모리 사용량 확인
ps aux | grep syslog-monitor

# 3. LaunchAgent 상태 확인
launchctl list com.lambda-x.syslog-monitor
```

### 로그 파일이 생성되지 않는 경우
```bash
# 1. 로그 디렉토리 권한 확인
ls -la /usr/local/var/log/

# 2. 디렉토리 재생성
sudo mkdir -p /usr/local/var/log
sudo chown $(whoami) /usr/local/var/log

# 3. 서비스 재시작
syslog-monitor -stop-service
syslog-monitor -start-service
```

### 이메일이 전송되지 않는 경우
```bash
# 1. 이메일 테스트
syslog-monitor -test-email

# 2. SMTP 설정 확인
syslog-monitor -show-config | grep -A 10 "email"

# 3. Gmail 앱 패스워드 설정 확인
```

## 🔒 보안 고려사항

### 파일 권한
```bash
# 실행 파일 권한 확인
chmod 755 /usr/local/bin/syslog-monitor

# 설정 파일 보안 (API 키 포함)
chmod 600 ~/.syslog-monitor/config.json

# 로그 파일 권한
chmod 644 /usr/local/var/log/syslog-monitor.*
```

### 네트워크 보안
- SMTP는 TLS/STARTTLS 사용
- Slack 웹훅 URL 보안 관리
- API 키는 환경변수 또는 보안 파일에 저장

### 프로세스 격리
- 일반 사용자 권한으로 실행 (root 권한 불필요)
- LaunchAgent로 사용자 세션에서만 실행
- 리소스 제한 설정 (CPU, 메모리)

## 📈 성능 최적화

### 모니터링 간격 조정
```bash
# 높은 부하 시스템 (간격 늘리기)
syslog-monitor -system-monitor -report-interval=120

# 실시간 모니터링 (간격 줄이기)
syslog-monitor -system-monitor -report-interval=30
```

### 리소스 사용량 확인
```bash
# CPU 및 메모리 사용량
ps aux | grep syslog-monitor

# 파일 핸들러 사용량
lsof -p $(pgrep syslog-monitor)

# 네트워크 연결 확인
netstat -an | grep $(pgrep syslog-monitor)
```

## 🔄 업데이트 및 유지보수

### 서비스 업데이트
```bash
# 1. 서비스 중지
syslog-monitor -stop-service

# 2. 새 바이너리 빌드 및 설치
go build -o syslog-monitor .
sudo cp syslog-monitor /usr/local/bin/

# 3. 서비스 재시작
syslog-monitor -start-service

# 4. 상태 확인
syslog-monitor -status-service
```

### 백업 및 복원
```bash
# 설정 백업
cp ~/.syslog-monitor/config.json ~/.syslog-monitor/config.json.backup

# 로그 백업 (선택사항)
tar -czf syslog-monitor-logs-$(date +%Y%m%d).tar.gz /usr/local/var/log/syslog-monitor.*

# 설정 복원
cp ~/.syslog-monitor/config.json.backup ~/.syslog-monitor/config.json
```

## 📞 지원 및 도움말

### 설정 확인 명령어
```bash
# 전체 시스템 상태
syslog-monitor -status-service

# 설정 파일 내용
syslog-monitor -show-config

# 도움말
syslog-monitor -help
```

### 로그 수집 (지원 요청 시)
```bash
# 진단 정보 수집 스크립트
cat > collect-diagnostics.sh << 'EOF'
#!/bin/bash
echo "=== Lambda-X Syslog Monitor Diagnostics ==="
echo "Date: $(date)"
echo "System: $(uname -a)"
echo ""

echo "=== Service Status ==="
syslog-monitor -status-service

echo ""
echo "=== Configuration ==="
syslog-monitor -show-config

echo ""
echo "=== Recent Logs ==="
tail -50 /usr/local/var/log/syslog-monitor.out.log

echo ""
echo "=== Recent Errors ==="
tail -20 /usr/local/var/log/syslog-monitor.err.log

echo ""
echo "=== LaunchAgent Status ==="
launchctl list com.lambda-x.syslog-monitor

echo ""
echo "=== Process Information ==="
ps aux | grep syslog-monitor
EOF

chmod +x collect-diagnostics.sh
./collect-diagnostics.sh > diagnostics-$(date +%Y%m%d-%H%M%S).txt
```

---

**🎯 주요 명령어 요약:**
- 설치: `./install-service.sh`
- 상태: `syslog-monitor -status-service`
- 시작: `syslog-monitor -start-service`
- 중지: `syslog-monitor -stop-service`
- 로그: `tail -f /usr/local/var/log/syslog-monitor.out.log`