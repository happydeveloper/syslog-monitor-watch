#!/bin/bash

# Lambda-X AI Security Monitor - macOS Install Script
# ==================================================
# 
# macOS 전용 설치 스크립트
# Homebrew, Go 환경 자동 설정 포함

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 스크립트 시작
log_info "🍎 Lambda-X AI Security Monitor - macOS Install Script"
log_info "================================================="

# 1. macOS 버전 확인
log_info "📋 Step 1: macOS 환경 확인"

MACOS_VERSION=$(sw_vers -productVersion)
log_info "macOS 버전: $MACOS_VERSION"

# 2. Homebrew 설치 확인
log_info "📋 Step 2: Homebrew 설치 확인"

if ! command -v brew &> /dev/null; then
    log_warning "Homebrew가 설치되어 있지 않습니다."
    read -p "Homebrew를 설치하시겠습니까? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Homebrew 설치 중..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        log_success "Homebrew 설치 완료"
    else
        log_error "Homebrew가 필요합니다."
        exit 1
    fi
else
    log_success "Homebrew 설치됨: $(brew --version | head -1)"
fi

# 3. Go 설치 확인
log_info "📋 Step 3: Go 설치 확인"

if ! command -v go &> /dev/null; then
    log_warning "Go가 설치되어 있지 않습니다."
    log_info "Go 설치 중..."
    brew install go
    log_success "Go 설치 완료"
else
    log_success "Go 설치됨: $(go version)"
fi

# 4. 기존 설치 정리
log_info "📋 Step 4: 기존 설치 정리"

# 기존 바이너리 삭제
if [ -f "/usr/local/bin/syslog-monitor" ]; then
    log_info "기존 바이너리 삭제 중..."
    sudo rm -f /usr/local/bin/syslog-monitor
    log_success "기존 바이너리 삭제 완료"
fi

# 현재 디렉토리 바이너리 삭제
if [ -f "./syslog-monitor" ]; then
    log_info "로컬 바이너리 삭제 중..."
    rm -f ./syslog-monitor
    log_success "로컬 바이너리 삭제 완료"
fi

# 5. 의존성 업데이트
log_info "📋 Step 5: 의존성 업데이트"

go mod tidy
go mod download
log_success "의존성 업데이트 완료"

# 6. 빌드
log_info "📋 Step 6: 빌드"

log_info "syslog-monitor 빌드 중..."
go build -ldflags="-s -w" -o syslog-monitor

if [ $? -eq 0 ]; then
    log_success "빌드 성공!"
    BINARY_SIZE=$(du -h ./syslog-monitor | cut -f1)
    log_info "바이너리 크기: $BINARY_SIZE"
else
    log_error "빌드 실패!"
    exit 1
fi

# 7. 설치
log_info "📋 Step 7: 시스템 설치"

sudo cp ./syslog-monitor /usr/local/bin/
sudo chmod +x /usr/local/bin/syslog-monitor
log_success "바이너리 설치 완료: /usr/local/bin/syslog-monitor"

# 8. 설정 디렉토리 생성
log_info "📋 Step 8: 설정 디렉토리 생성"

sudo mkdir -p /etc/syslog-monitor
sudo chmod 755 /etc/syslog-monitor

# 기본 설정 파일 생성
if [ ! -f "/etc/syslog-monitor/config.json" ]; then
    log_info "기본 설정 파일 생성 중..."
    cat > /tmp/syslog-monitor-config.json << 'EOF'
{
    "email": {
        "enabled": true,
        "smtp_server": "smtp.gmail.com",
        "smtp_port": 587,
        "username": "your-email@gmail.com",
        "password": "your-app-password",
        "to": ["admin@company.com"],
        "from": "security@company.com"
    },
    "slack": {
        "enabled": false,
        "webhook_url": "https://hooks.slack.com/your-webhook",
        "channel": "#security",
        "username": "AI Security Monitor"
    },
    "monitoring": {
        "login_watch": true,
        "ai_analysis": true,
        "system_monitor": true,
        "alert_interval": 10,
        "periodic_report": true,
        "report_interval": 60
    }
}
EOF
    sudo cp /tmp/syslog-monitor-config.json /etc/syslog-monitor/config.json
    sudo chmod 644 /etc/syslog-monitor/config.json
    log_success "기본 설정 파일 생성: /etc/syslog-monitor/config.json"
    log_warning "설정 파일을 편집하여 실제 값으로 변경하세요."
fi

# 9. 설치 확인
log_info "📋 Step 9: 설치 확인"

if command -v syslog-monitor &> /dev/null; then
    log_success "설치 확인 완료"
    log_info "바이너리 경로: $(which syslog-monitor)"
    log_info "버전 정보: $(syslog-monitor --help 2>&1 | head -5)"
else
    log_error "설치 확인 실패"
    exit 1
fi

# 10. 사용 예시
log_info "📋 Step 10: 사용 예시"

echo
log_success "🎉 macOS 설치 완료!"
echo
echo "📖 사용 예시:"
echo "=============="
echo
echo "# 기본 로그인 모니터링 (macOS)"
echo "syslog-monitor -login-watch -system-monitor -file=\"/var/log/system.log\""
echo
echo "# 이메일 알림과 함께"
echo "syslog-monitor -login-watch -system-monitor -email-to=\"admin@company.com\""
echo
echo "# Slack 알림과 함께"
echo "syslog-monitor -login-watch -system-monitor -slack-webhook=\"https://hooks.slack.com/your-webhook\""
echo
echo "# AI 분석과 함께"
echo "syslog-monitor -login-watch -system-monitor -ai-analysis"
echo
echo "# 5분 간격 알림"
echo "syslog-monitor -login-watch -system-monitor -alert-interval=5"
echo
echo "# 주기적 보고서 (60분마다)"
echo "syslog-monitor -login-watch -system-monitor -periodic-report -report-interval=60"
echo
echo "# 도움말 보기"
echo "syslog-monitor --help"
echo

# 11. 테스트 실행 (선택사항)
read -p "테스트 실행을 하시겠습니까? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "테스트 실행 중... (Ctrl+C로 중단)"
    syslog-monitor -login-watch -system-monitor -file="/var/log/system.log" 2>&1 | head -20
fi

log_success "🎯 macOS Install 완료!"
log_info "이제 Lambda-X AI Security Monitor를 사용할 수 있습니다." 