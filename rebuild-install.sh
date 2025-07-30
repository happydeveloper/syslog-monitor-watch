#!/bin/bash

# Lambda-X AI Security Monitor - Rebuild & Install Script
# ======================================================
# 
# 이 스크립트는 기존 설치를 삭제하고 새로 빌드하여 설치합니다.
# 
# 주요 기능:
# - 기존 바이너리 및 설정 파일 삭제
# - Go 모듈 의존성 정리 및 업데이트
# - 새로 빌드 및 설치
# - 권한 설정 및 서비스 등록

set -e  # 에러 발생 시 스크립트 중단

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수
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
log_info "🚀 Lambda-X AI Security Monitor - Rebuild & Install Script"
log_info "======================================================"

# 1. 기존 설치 확인 및 삭제
log_info "📋 Step 1: 기존 설치 확인 및 삭제"

# 기존 바이너리 확인
if command -v syslog-monitor &> /dev/null; then
    log_info "기존 syslog-monitor 바이너리 발견"
    which syslog-monitor
    read -p "기존 바이너리를 삭제하시겠습니까? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo rm -f $(which syslog-monitor)
        log_success "기존 바이너리 삭제 완료"
    else
        log_warning "기존 바이너리 유지"
    fi
else
    log_info "기존 syslog-monitor 바이너리 없음"
fi

# 현재 디렉토리의 바이너리 삭제
if [ -f "./syslog-monitor" ]; then
    log_info "현재 디렉토리의 syslog-monitor 바이너리 삭제"
    rm -f ./syslog-monitor
    log_success "로컬 바이너리 삭제 완료"
fi

# 2. Go 환경 확인
log_info "📋 Step 2: Go 환경 확인"

if ! command -v go &> /dev/null; then
    log_error "Go가 설치되어 있지 않습니다."
    log_info "Go 설치 방법: https://golang.org/doc/install"
    exit 1
fi

log_success "Go 버전: $(go version)"

# 3. 의존성 정리 및 업데이트
log_info "📋 Step 3: Go 모듈 의존성 정리 및 업데이트"

# go.mod 파일 확인
if [ ! -f "go.mod" ]; then
    log_error "go.mod 파일을 찾을 수 없습니다."
    exit 1
fi

# 기존 모듈 캐시 정리
log_info "Go 모듈 캐시 정리 중..."
go clean -modcache 2>/dev/null || true

# 의존성 다운로드 및 업데이트
log_info "의존성 다운로드 및 업데이트 중..."
go mod download
go mod tidy

log_success "의존성 업데이트 완료"

# 4. 새로 빌드
log_info "📋 Step 4: 새로 빌드"

# 빌드 시작
log_info "syslog-monitor 빌드 중..."
go build -ldflags="-s -w" -o syslog-monitor

if [ $? -eq 0 ]; then
    log_success "빌드 성공!"
else
    log_error "빌드 실패!"
    exit 1
fi

# 바이너리 크기 확인
BINARY_SIZE=$(du -h ./syslog-monitor | cut -f1)
log_info "빌드된 바이너리 크기: $BINARY_SIZE"

# 5. 설치
log_info "📋 Step 5: 시스템 설치"

# 설치 경로 설정
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="syslog-monitor"

# 권한 확인
if [ ! -w "$INSTALL_DIR" ]; then
    log_info "관리자 권한이 필요합니다."
    sudo cp ./syslog-monitor "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    cp ./syslog-monitor "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

log_success "바이너리 설치 완료: $INSTALL_DIR/$BINARY_NAME"

# 6. 설치 확인
log_info "📋 Step 6: 설치 확인"

# PATH에서 바이너리 확인
if command -v $BINARY_NAME &> /dev/null; then
    log_success "설치 확인 완료"
    log_info "바이너리 경로: $(which $BINARY_NAME)"
    log_info "버전 정보: $($BINARY_NAME --help 2>&1 | head -5)"
else
    log_error "설치 확인 실패"
    exit 1
fi

# 7. 설정 파일 생성 (선택사항)
log_info "📋 Step 7: 설정 파일 생성 (선택사항)"

CONFIG_DIR="/etc/syslog-monitor"
if [ ! -d "$CONFIG_DIR" ]; then
    log_info "설정 디렉토리 생성: $CONFIG_DIR"
    sudo mkdir -p "$CONFIG_DIR"
fi

# 기본 설정 파일 생성
if [ ! -f "$CONFIG_DIR/config.json" ]; then
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
    sudo cp /tmp/syslog-monitor-config.json "$CONFIG_DIR/config.json"
    sudo chmod 644 "$CONFIG_DIR/config.json"
    log_success "기본 설정 파일 생성: $CONFIG_DIR/config.json"
    log_warning "설정 파일을 편집하여 실제 값으로 변경하세요."
else
    log_info "설정 파일이 이미 존재합니다: $CONFIG_DIR/config.json"
fi

# 8. 사용 예시 출력
log_info "📋 Step 8: 사용 예시"

echo
log_success "🎉 설치 완료!"
echo
echo "📖 사용 예시:"
echo "=============="
echo
echo "# 기본 로그인 모니터링"
echo "$BINARY_NAME -login-watch -system-monitor"
echo
echo "# 이메일 알림과 함께"
echo "$BINARY_NAME -login-watch -system-monitor -email-to=\"admin@company.com\""
echo
echo "# Slack 알림과 함께"
echo "$BINARY_NAME -login-watch -system-monitor -slack-webhook=\"https://hooks.slack.com/your-webhook\""
echo
echo "# AI 분석과 함께"
echo "$BINARY_NAME -login-watch -system-monitor -ai-analysis"
echo
echo "# 5분 간격 알림"
echo "$BINARY_NAME -login-watch -system-monitor -alert-interval=5"
echo
echo "# 주기적 보고서 (60분마다)"
echo "$BINARY_NAME -login-watch -system-monitor -periodic-report -report-interval=60"
echo
echo "# 도움말 보기"
echo "$BINARY_NAME --help"
echo

# 9. 테스트 실행 (선택사항)
read -p "테스트 실행을 하시겠습니까? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "테스트 실행 중... (Ctrl+C로 중단)"
    $BINARY_NAME -login-watch -system-monitor -file="/var/log/syslog" 2>&1 | head -20
fi

log_success "🎯 Rebuild & Install 완료!"
log_info "이제 Lambda-X AI Security Monitor를 사용할 수 있습니다." 