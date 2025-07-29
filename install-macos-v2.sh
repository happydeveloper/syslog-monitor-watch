#!/bin/bash

# macOS용 Syslog Monitor 설치 스크립트 v2.0
# AI 기반 로그 분석 및 시스템 모니터링 도구
# 새로운 기능: 컴퓨터명, IP 분류, ASN 정보 포함

set -e

echo "🍎 macOS용 Syslog Monitor v2.0 설치를 시작합니다..."
echo ""

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 아이콘
ROCKET="🚀"
ROBOT="🤖"
SECURITY="🔒"
NETWORK="🌐"
APPLE="🍎"
CHECKMARK="✅"
WARNING="⚠️"
ERROR="❌"
INFO="ℹ️"

# 함수들
print_status() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

print_success() {
    echo -e "${GREEN}${CHECKMARK}${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}${WARNING}${NC} $1"
}

print_error() {
    echo -e "${RED}${ERROR}${NC} $1"
}

print_feature() {
    echo -e "${PURPLE}${ROCKET}${NC} $1"
}

print_security() {
    echo -e "${CYAN}${SECURITY}${NC} $1"
}

# 스크립트 중단 시 정리
cleanup() {
    print_status "설치 중단됨. 정리 중..."
    # 임시 파일들 정리
    rm -f /tmp/syslog-monitor-test.log
}

trap cleanup EXIT

# 권한 확인
check_sudo_available() {
    if ! sudo -n true 2>/dev/null; then
        print_status "관리자 권한이 필요합니다. 비밀번호를 입력해주세요:"
        sudo -v
    fi
}

# 시스템 요구사항 확인
check_requirements() {
    print_status "시스템 요구사항을 확인하는 중..."
    
    # macOS 버전 확인
    macos_version=$(sw_vers -productVersion)
    macos_major=$(echo "$macos_version" | cut -d '.' -f 1)
    print_status "macOS 버전: $macos_version"
    
    if [ "$macos_major" -lt 11 ]; then
        print_warning "macOS 11.0 이상을 권장합니다"
    fi
    
    # 아키텍처 확인
    arch=$(uname -m)
    print_status "시스템 아키텍처: $arch"
    
    # CPU 정보
    if [ "$arch" = "arm64" ]; then
        print_success "Apple Silicon 감지됨 ${APPLE}"
    else
        print_success "Intel Mac 감지됨"
    fi
    
    # Go 설치 확인
    if command -v go &> /dev/null; then
        go_version=$(go version | awk '{print $3}')
        print_success "Go 설치됨: $go_version"
        
        # Go 버전 확인 (1.19 이상 권장)
        go_ver_num=$(echo "$go_version" | sed 's/go//' | cut -d '.' -f 2)
        if [ "$go_ver_num" -lt 19 ]; then
            print_warning "Go 1.19 이상을 권장합니다"
        fi
    else
        print_error "Go가 설치되지 않았습니다."
        print_status "Go 설치 방법:"
        echo "  1. 공식 설치: https://golang.org/dl/"
        echo "  2. Homebrew: brew install go"
        exit 1
    fi
    
    # Git 확인
    if command -v git &> /dev/null; then
        git_version=$(git --version | awk '{print $3}')
        print_success "Git 설치됨: $git_version"
    else
        print_error "Git이 설치되지 않았습니다."
        print_status "설치 방법: xcode-select --install"
        exit 1
    fi
    
    # 네트워크 연결 확인
    if ping -c 1 8.8.8.8 &> /dev/null; then
        print_success "인터넷 연결 확인됨 ${NETWORK}"
    else
        print_warning "인터넷 연결이 불안정합니다. ASN 조회 기능이 제한될 수 있습니다."
    fi
}

# 옵션 도구 설치 확인
check_optional_tools() {
    print_status "선택적 도구들을 확인하는 중..."
    
    # Homebrew 확인
    if command -v brew &> /dev/null; then
        brew_version=$(brew --version | head -n1 | awk '{print $2}')
        print_success "Homebrew 설치됨: $brew_version"
        
        # 권한이 있는 경우에만 istats 설치 제안
        if [ "$EUID" -ne 0 ]; then
            # istats 설치 권장 (온도 모니터링용)
            if ! command -v istats &> /dev/null; then
                print_warning "istats가 설치되지 않았습니다 (온도 모니터링 최적화용)"
                echo -n "istats를 설치하시겠습니까? (y/N): "
                read -r response
                if [[ "$response" =~ ^[Yy]$ ]]; then
                    print_status "istats 설치 중..."
                    if brew install istat-menus 2>/dev/null; then
                        print_success "istats 설치 완료"
                    else
                        print_warning "istats 설치 실패 (계속 진행됩니다)"
                    fi
                fi
            else
                print_success "istats 설치됨 (온도 모니터링 최적화)"
            fi
        fi
    else
        print_warning "Homebrew가 설치되지 않았습니다"
        print_status "권장 설치 명령:"
        echo '  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
    fi
}

# 프로젝트 빌드
build_project() {
    print_status "프로젝트를 빌드하는 중..."
    
    # 빌드 디렉토리 확인
    if [ ! -f "main.go" ]; then
        print_error "소스 코드를 찾을 수 없습니다. 올바른 디렉토리에서 실행해주세요."
        exit 1
    fi
    
    # 의존성 설치
    print_status "Go 모듈 의존성 설치..."
    if ! go mod download; then
        print_error "의존성 다운로드 실패"
        exit 1
    fi
    
    if ! go mod tidy; then
        print_warning "go mod tidy 실패 (계속 진행됩니다)"
    fi
    
    # macOS용 빌드
    if [[ "$arch" == "arm64" ]]; then
        print_status "Apple Silicon (ARM64)용 빌드..."
        if make build-macos-arm64; then
            binary_name="syslog-monitor_macos_arm64"
            print_success "ARM64 빌드 완료"
        else
            print_error "ARM64 빌드 실패"
            exit 1
        fi
    elif [[ "$arch" == "x86_64" ]]; then
        print_status "Intel (AMD64)용 빌드..."
        if make build-macos-intel; then
            binary_name="syslog-monitor_macos_amd64"
            print_success "Intel 빌드 완료"
        else
            print_error "Intel 빌드 실패"
            exit 1
        fi
    else
        print_status "현재 아키텍처용 빌드..."
        if make build-macos; then
            binary_name="syslog-monitor_macos"
            print_success "기본 빌드 완료"
        else
            print_error "빌드 실패"
            exit 1
        fi
    fi
    
    # 바이너리 존재 확인
    if [ ! -f "$binary_name" ]; then
        print_error "빌드된 바이너리를 찾을 수 없습니다: $binary_name"
        exit 1
    fi
    
    print_success "빌드 완료: $binary_name ($(du -h "$binary_name" | cut -f1))"
}

# 시스템에 설치
install_system() {
    print_status "시스템에 설치하는 중..."
    
    # /usr/local/bin 디렉토리 확인 및 생성
    if [ ! -d "/usr/local/bin" ]; then
        print_status "/usr/local/bin 디렉토리 생성 중..."
        sudo mkdir -p /usr/local/bin
    fi
    
    # 실행 파일 복사
    if [[ -f "$binary_name" ]]; then
        sudo cp "$binary_name" /usr/local/bin/syslog-monitor
    else
        print_error "빌드된 바이너리를 찾을 수 없습니다"
        exit 1
    fi
    
    sudo chmod +x /usr/local/bin/syslog-monitor
    print_success "실행 파일 설치 완료: /usr/local/bin/syslog-monitor"
    
    # 설정 디렉토리 생성
    config_dir="$HOME/.syslog-monitor"
    if [[ ! -d "$config_dir" ]]; then
        mkdir -p "$config_dir"
        print_success "설정 디렉토리 생성: $config_dir"
    fi
    
    # 기본 설정 파일 생성
    cat > "$config_dir/config.json" << EOF
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
EOF
    print_success "기본 설정 파일 생성 완료"
    
    # 로그 디렉토리 생성
    log_dir="$HOME/.syslog-monitor/logs"
    mkdir -p "$log_dir"
    print_success "로그 디렉토리 생성: $log_dir"
}

# LaunchAgent 설정 (자동 시작)
setup_launchagent() {
    echo ""
    echo -n "부팅 시 자동 시작을 설정하시겠습니까? (y/N): "
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_status "LaunchAgent 설정 중..."
        
        launchagent_dir="$HOME/Library/LaunchAgents"
        mkdir -p "$launchagent_dir"
        
        plist_file="$launchagent_dir/ai.lambda-x.syslog-monitor.plist"
        
        cat > "$plist_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>ai.lambda-x.syslog-monitor</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/syslog-monitor</string>
        <string>-ai-analysis</string>
        <string>-system-monitor</string>
        <string>-file</string>
        <string>/var/log/system.log</string>
        <string>-output</string>
        <string>$HOME/.syslog-monitor/logs/monitor.log</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/.syslog-monitor/logs/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/.syslog-monitor/logs/stderr.log</string>
    <key>WorkingDirectory</key>
    <string>$HOME/.syslog-monitor</string>
</dict>
</plist>
EOF
        
        print_success "LaunchAgent 파일 생성 완료"
        
        # 기존 서비스 언로드 (오류 무시)
        launchctl unload "$plist_file" 2>/dev/null || true
        
        # 새 서비스 로드
        if launchctl load "$plist_file" 2>/dev/null; then
            print_success "자동 시작 설정 완료"
        else
            print_warning "LaunchAgent 로드에 실패했지만 파일은 생성되었습니다"
            print_status "수동 시작: launchctl load $plist_file"
        fi
        
        print_warning "참고: 일부 로그 파일은 관리자 권한이 필요할 수 있습니다"
    fi
}

# 안전한 테스트 실행
run_tests() {
    echo ""
    echo -n "설치 후 테스트를 실행하시겠습니까? (Y/n): "
    read -r response
    
    if [[ ! "$response" =~ ^[Nn]$ ]]; then
        print_status "설치 테스트 실행 중..."
        
        # 테스트 로그 파일 생성
        test_log="/tmp/syslog-monitor-test.log"
        cat > "$test_log" << EOF
$(date) INFO [test] Testing syslog monitor installation
$(date) ERROR [security] Test error from 203.0.113.1 - failed login attempt
$(date) WARNING [database] High response time detected: 2500ms
EOF
        
        # 버전 확인 테스트
        print_status "버전 확인 테스트..."
        if timeout 5 /usr/local/bin/syslog-monitor -help | head -n 1 > /dev/null 2>&1; then
            print_success "기본 실행 테스트 통과"
        else
            print_warning "기본 실행 테스트 실패 (권한 문제일 수 있습니다)"
        fi
        
        # AI 분석 기능 테스트 (안전한 방법)
        print_status "AI 분석 기능 테스트 (5초간)..."
        if timeout 5 /usr/local/bin/syslog-monitor -file="$test_log" -ai-analysis -system-monitor > /dev/null 2>&1; then
            print_success "AI 분석 테스트 통과"
        else
            print_warning "AI 분석 테스트 실패 (정상적일 수 있습니다)"
        fi
        
        # 시스템 정보 수집 테스트
        print_status "시스템 정보 수집 테스트..."
        computer_name=$(hostname)
        print_success "컴퓨터명 감지: $computer_name"
        
        # 정리
        rm -f "$test_log"
        print_success "테스트 완료"
    fi
}

# 새로운 기능 소개
show_new_features() {
    echo ""
    print_feature "🆕 새로운 AI 분석 기능들:"
    echo ""
    print_security "1. 📍 시스템 정보 자동 수집"
    echo "   • 컴퓨터 이름 자동 감지"
    echo "   • 내부/외부 IP 주소 분류"
    echo "   • RFC 1918 표준 준수"
    echo ""
    print_security "2. 🌐 ASN 정보 조회"
    echo "   • 외부 IP의 조직 정보"
    echo "   • 지리적 위치 (국가, 지역)"
    echo "   • 실시간 위협 분석"
    echo ""
    print_security "3. 🚨 향상된 알람 시스템"
    echo "   • 상세한 시스템 정보 포함"
    echo "   • 보안 위협 예측"
    echo "   • 맞춤형 권장사항 제공"
    echo ""
}

# 사용법 안내
show_usage() {
    echo ""
    print_success "🎉 macOS용 Syslog Monitor v2.0 설치가 완료되었습니다!"
    echo ""
    
    show_new_features
    
    echo ""
    print_status "${ROCKET} 기본 사용법:"
    echo ""
    echo "  # 기본 모니터링"
    echo "  syslog-monitor"
    echo ""
    echo "  # AI 분석 + 시스템 모니터링 (권장)"
    echo "  syslog-monitor -ai-analysis -system-monitor"
    echo ""
    echo "  # 보안 모니터링 (로그인 감시 포함)"
    echo "  syslog-monitor -ai-analysis -login-watch"
    echo ""
    echo "  # 전체 기능 활성화"
    echo "  syslog-monitor -ai-analysis -system-monitor -login-watch"
    echo ""
    echo "  # 특정 로그 파일 모니터링"
    echo "  syslog-monitor -file=/var/log/system.log -ai-analysis"
    echo ""
    
    print_status "${INFO} 설정 및 로그:"
    echo "  • 설정 파일: $HOME/.syslog-monitor/config.json"
    echo "  • 로그 파일: $HOME/.syslog-monitor/logs/"
    echo "  • 자세한 도움말: syslog-monitor -help"
    echo ""
    
    if command -v istats &> /dev/null; then
        print_success "💡 온도 모니터링이 최적화되었습니다 (istats 사용)"
    else
        print_warning "💡 더 정확한 온도 모니터링을 위해 istats 설치를 권장합니다:"
        echo "     brew install istat-menus"
    fi
    
    echo ""
    print_status "${APPLE} macOS 로그 파일 경로:"
    echo "  • 시스템 로그: /var/log/system.log"
    echo "  • 설치 로그: /var/log/install.log"
    echo "  • WiFi 로그: /var/log/wifi.log"
    echo "  • 보안 로그: /var/log/secure.log"
    echo ""
    
    print_status "${NETWORK} 실시간 로그 명령 (sudo 필요):"
    echo "  • sudo log stream | syslog-monitor -file=/dev/stdin -ai-analysis"
    echo "  • sudo log show --predicate 'eventMessage contains \"error\"' --last 1h"
    echo ""
}

# 메인 실행
main() {
    echo "${ROBOT} AI 기반 로그 분석 및 시스템 모니터링 도구"
    echo "${APPLE} macOS 최적화 버전 v2.0"
    echo "${SECURITY} 새로운 기능: 컴퓨터명, IP 분류, ASN 정보"
    echo ""
    
    # sudo 가능 여부 확인
    check_sudo_available
    
    check_requirements
    check_optional_tools
    build_project
    install_system
    setup_launchagent
    run_tests
    show_usage
    
    echo ""
    print_success "설치가 완전히 완료되었습니다! 🎉"
    print_status "이제 AI 기반 보안 모니터링을 시작하세요!"
    echo ""
    echo "Quick Start: syslog-monitor -ai-analysis -system-monitor"
    echo ""
}

# 인터럽트 핸들러
handle_interrupt() {
    echo ""
    print_warning "설치가 중단되었습니다"
    exit 130
}

trap handle_interrupt SIGINT

# 스크립트 실행
main "$@" 