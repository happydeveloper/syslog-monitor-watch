#!/bin/bash

# Lambda-X Syslog Monitor Service Installer
# ===========================================
# macOS LaunchAgent 자동 설치 및 설정 스크립트

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로고 및 헤더
echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                    🤖 Lambda-X Syslog Monitor                        ║"
echo "║                        Service Installer                            ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# 함수 정의
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 권한 확인
check_permissions() {
    print_status "권한 확인 중..."
    
    if [[ $EUID -eq 0 ]]; then
        print_warning "루트 권한으로 실행 중입니다. 일반 사용자 권한을 권장합니다."
    fi
    
    # 홈 디렉토리 접근 확인
    if [[ ! -w "$HOME" ]]; then
        print_error "홈 디렉토리에 쓰기 권한이 없습니다."
        exit 1
    fi
}

# 의존성 확인
check_dependencies() {
    print_status "의존성 확인 중..."
    
    # Go 설치 확인
    if ! command -v go &> /dev/null; then
        print_error "Go가 설치되지 않았습니다."
        print_status "Homebrew로 Go 설치 중..."
        if command -v brew &> /dev/null; then
            brew install go
        else
            print_error "Homebrew가 설치되지 않았습니다. 수동으로 Go를 설치해주세요."
            exit 1
        fi
    else
        print_success "Go 설치 확인됨 ($(go version))"
    fi
    
    # launchctl 확인
    if ! command -v launchctl &> /dev/null; then
        print_error "launchctl이 없습니다. macOS가 아닌 시스템인 것 같습니다."
        exit 1
    fi
}

# 빌드
build_binary() {
    print_status "바이너리 빌드 중..."
    
    if [[ ! -f "main.go" ]]; then
        print_error "main.go 파일을 찾을 수 없습니다. 프로젝트 디렉토리에서 실행해주세요."
        exit 1
    fi
    
    # 모듈 다운로드
    go mod tidy
    
    # 바이너리 빌드
    go build -o syslog-monitor .
    
    if [[ ! -f "syslog-monitor" ]]; then
        print_error "빌드 실패"
        exit 1
    fi
    
    print_success "빌드 완료"
}

# 바이너리 설치
install_binary() {
    print_status "바이너리 설치 중..."
    
    # /usr/local/bin에 복사
    if sudo cp syslog-monitor /usr/local/bin/; then
        sudo chmod +x /usr/local/bin/syslog-monitor
        print_success "바이너리가 /usr/local/bin/syslog-monitor에 설치되었습니다."
    else
        print_error "바이너리 설치 실패"
        exit 1
    fi
}

# 로그 디렉토리 생성
create_directories() {
    print_status "디렉토리 생성 중..."
    
    # 로그 디렉토리
    sudo mkdir -p /usr/local/var/log
    sudo mkdir -p /usr/local/var/run
    
    # 사용자 설정 디렉토리
    mkdir -p "$HOME/.syslog-monitor"
    
    print_success "필요한 디렉토리가 생성되었습니다."
}

# 기본 설정 파일 생성
create_default_config() {
    print_status "기본 설정 파일 생성 중..."
    
    CONFIG_FILE="$HOME/.syslog-monitor/config.json"
    
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" << 'EOF'
{
  "ai": {
    "gemini_api_key": "",
    "enabled": false
  },
  "system": {
    "monitoring_enabled": true,
    "monitoring_interval": 300,
    "alert_thresholds": {
      "cpu_percent": 80.0,
      "memory_percent": 85.0,
      "disk_percent": 90.0,
      "cpu_temp": 75.0,
      "load_average": 16.0,
      "swap_percent": 50.0,
      "inode_percent": 90.0
    }
  },
  "email": {
    "enabled": true,
    "smtp_server": "smtp.gmail.com",
    "smtp_port": "587",
    "username": "enfn2001@gmail.com",
    "password": "",
    "to": ["robot@lambda-x.ai", "enfn2001@gmail.com"],
    "from": "enfn2001@gmail.com"
  },
  "slack": {
    "enabled": false,
    "webhook_url": "",
    "channel": "#alerts",
    "username": "Lambda-X Monitor"
  },
  "logging": {
    "level": "info",
    "file": "/usr/local/var/log/syslog-monitor.log",
    "max_size": 100,
    "max_backups": 5,
    "max_age": 30
  },
  "features": {
    "login_monitoring": true,
    "ai_analysis": false,
    "system_monitoring": true,
    "periodic_reports": true,
    "report_interval": 3600
  }
}
EOF
        print_success "기본 설정 파일이 생성되었습니다: $CONFIG_FILE"
    else
        print_warning "설정 파일이 이미 존재합니다: $CONFIG_FILE"
    fi
}

# 서비스 설치
install_service() {
    print_status "LaunchAgent 서비스 설치 중..."
    
    if /usr/local/bin/syslog-monitor -install-service; then
        print_success "서비스가 설치되었습니다."
    else
        print_error "서비스 설치 실패"
        exit 1
    fi
}

# 서비스 시작
start_service() {
    print_status "서비스 시작 중..."
    
    if /usr/local/bin/syslog-monitor -start-service; then
        print_success "서비스가 시작되었습니다."
    else
        print_warning "서비스 시작에 실패했습니다. 수동으로 시작해주세요."
    fi
}

# 상태 확인
check_status() {
    print_status "서비스 상태 확인 중..."
    echo
    /usr/local/bin/syslog-monitor -status-service
}

# 사용법 안내
show_usage() {
    echo
    print_status "설치 완료! 🎉"
    echo
    echo -e "${YELLOW}📋 주요 명령어:${NC}"
    echo "  syslog-monitor -status-service     # 서비스 상태 확인"
    echo "  syslog-monitor -stop-service       # 서비스 중지"
    echo "  syslog-monitor -start-service      # 서비스 시작"
    echo "  syslog-monitor -remove-service     # 서비스 제거"
    echo "  syslog-monitor -show-config        # 현재 설정 확인"
    echo
    echo -e "${YELLOW}📄 로그 파일:${NC}"
    echo "  tail -f /usr/local/var/log/syslog-monitor.out.log  # 실시간 로그"
    echo "  tail -f /usr/local/var/log/syslog-monitor.err.log  # 에러 로그"
    echo
    echo -e "${YELLOW}⚙️  설정 파일:${NC}"
    echo "  $HOME/.syslog-monitor/config.json"
    echo
    echo -e "${YELLOW}🔧 추가 설정:${NC}"
    echo "  1. Gemini API 키 설정 (선택사항):"
    echo "     export GEMINI_API_KEY=\"your-api-key\""
    echo "  2. 이메일 설정 확인 및 수정"
    echo "  3. Slack 웹훅 설정 (선택사항)"
    echo
}

# 메인 실행
main() {
    print_status "Lambda-X Syslog Monitor 서비스 설치를 시작합니다..."
    
    check_permissions
    check_dependencies
    build_binary
    install_binary
    create_directories
    create_default_config
    install_service
    start_service
    check_status
    show_usage
    
    print_success "모든 설치가 완료되었습니다! ✅"
}

# 스크립트 실행
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi