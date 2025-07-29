#!/bin/bash

# macOS용 Syslog Monitor 설치 스크립트
# AI 기반 로그 분석 및 시스템 모니터링 도구

set -e

echo "🍎 macOS용 Syslog Monitor 설치를 시작합니다..."
echo ""

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 함수들
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

# 시스템 요구사항 확인
check_requirements() {
    print_status "시스템 요구사항을 확인하는 중..."
    
    # macOS 버전 확인
    macos_version=$(sw_vers -productVersion)
    print_status "macOS 버전: $macos_version"
    
    # 아키텍처 확인
    arch=$(uname -m)
    print_status "시스템 아키텍처: $arch"
    
    # Go 설치 확인
    if command -v go &> /dev/null; then
        go_version=$(go version)
        print_success "Go 설치됨: $go_version"
    else
        print_error "Go가 설치되지 않았습니다."
        print_status "Go 설치: https://golang.org/dl/"
        exit 1
    fi
    
    # Git 확인
    if command -v git &> /dev/null; then
        print_success "Git 설치됨"
    else
        print_error "Git이 설치되지 않았습니다."
        print_status "Xcode Command Line Tools 설치: xcode-select --install"
        exit 1
    fi
}

# 옵션 도구 설치 확인
check_optional_tools() {
    print_status "선택적 도구들을 확인하는 중..."
    
    # Homebrew 확인
    if command -v brew &> /dev/null; then
        print_success "Homebrew 설치됨"
        
        # istats 설치 권장 (온도 모니터링용)
        if ! command -v istats &> /dev/null; then
            print_warning "istats가 설치되지 않았습니다 (온도 모니터링 최적화용)"
            echo -n "istats를 설치하시겠습니까? (y/N): "
            read -r response
            if [[ "$response" =~ ^[Yy]$ ]]; then
                print_status "istats 설치 중..."
                brew install istat-menus
                print_success "istats 설치 완료"
            fi
        else
            print_success "istats 설치됨 (온도 모니터링 최적화)"
        fi
    else
        print_warning "Homebrew가 설치되지 않았습니다"
        print_status "Homebrew 설치: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    fi
}

# 프로젝트 빌드
build_project() {
    print_status "프로젝트를 빌드하는 중..."
    
    # 의존성 설치
    print_status "Go 모듈 의존성 설치..."
    make install
    
    # macOS용 빌드
    if [[ "$arch" == "arm64" ]]; then
        print_status "Apple Silicon (ARM64)용 빌드..."
        make build-macos-arm64
        binary_name="syslog-monitor_macos_arm64"
    elif [[ "$arch" == "x86_64" ]]; then
        print_status "Intel (AMD64)용 빌드..."
        make build-macos-intel  
        binary_name="syslog-monitor_macos_amd64"
    else
        print_status "현재 아키텍처용 빌드..."
        make build-macos
        binary_name="syslog-monitor_macos"
    fi
    
    print_success "빌드 완료: $binary_name"
}

# 시스템에 설치
install_system() {
    print_status "시스템에 설치하는 중..."
    
    # 실행 파일 복사
    if [[ -f "$binary_name" ]]; then
        sudo cp "$binary_name" /usr/local/bin/syslog-monitor
    elif [[ -f "syslog-monitor" ]]; then
        sudo cp syslog-monitor /usr/local/bin/syslog-monitor
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
    
    # 설정 파일 복사
    if [[ -f "config.json" ]]; then
        cp config.json "$config_dir/"
        print_success "설정 파일 복사 완료"
    fi
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
        
        cat > "$launchagent_dir/ai.lambda-x.syslog-monitor.plist" << EOF
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
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/syslog-monitor.out</string>
    <key>StandardErrorPath</key>
    <string>/tmp/syslog-monitor.err</string>
</dict>
</plist>
EOF
        
        print_success "LaunchAgent 파일 생성 완료"
        print_status "LaunchAgent 로드..."
        launchctl load "$launchagent_dir/ai.lambda-x.syslog-monitor.plist"
        print_success "자동 시작 설정 완료"
        
        print_warning "참고: sudo 권한이 필요한 일부 기능은 자동 시작 시 제한될 수 있습니다"
    fi
}

# 테스트 실행
run_tests() {
    echo ""
    echo -n "설치 후 테스트를 실행하시겠습니까? (Y/n): "
    read -r response
    
    if [[ ! "$response" =~ ^[Nn]$ ]]; then
        print_status "설치 테스트 실행 중..."
        
        # 기본 도움말 테스트
        if /usr/local/bin/syslog-monitor -help > /dev/null 2>&1; then
            print_success "기본 실행 테스트 통과"
        else
            print_error "기본 실행 테스트 실패"
            exit 1
        fi
        
        # AI 분석 기능 테스트 (3초간)
        print_status "AI 분석 및 시스템 모니터링 테스트 (3초간)..."
        timeout 3 /usr/local/bin/syslog-monitor -ai-analysis -system-monitor -file=/dev/null || true
        print_success "기능 테스트 완료"
    fi
}

# 사용법 안내
show_usage() {
    echo ""
    print_success "🎉 macOS용 Syslog Monitor 설치가 완료되었습니다!"
    echo ""
    echo "🚀 사용법:"
    echo "  # 기본 모니터링"
    echo "  syslog-monitor"
    echo ""
    echo "  # AI 분석 + 시스템 모니터링"
    echo "  syslog-monitor -ai-analysis -system-monitor"
    echo ""
    echo "  # 특정 로그 파일 모니터링"
    echo "  syslog-monitor -file=/var/log/system.log -ai-analysis"
    echo ""
    echo "  # 전체 기능 활성화"
    echo "  syslog-monitor -ai-analysis -system-monitor -login-watch"
    echo ""
    echo "📖 자세한 사용법:"
    echo "  syslog-monitor -help"
    echo ""
    echo "📁 설정 파일 위치:"
    echo "  $HOME/.syslog-monitor/config.json"
    echo ""
    
    if command -v istats &> /dev/null; then
        print_success "💡 온도 모니터링이 최적화되었습니다 (istats 사용)"
    else
        print_warning "💡 더 정확한 온도 모니터링을 위해 istats 설치를 권장합니다:"
        echo "     brew install istat-menus"
    fi
    
    echo ""
    print_status "로그 파일 경로 (macOS):"
    echo "  - 시스템 로그: /var/log/system.log"
    echo "  - 설치 로그: /var/log/install.log"
    echo "  - 보안 로그: /var/log/secure.log"
    echo "  - WiFi 로그: /var/log/wifi.log"
}

# 메인 실행
main() {
    echo "🤖 AI 기반 로그 분석 및 시스템 모니터링 도구"
    echo "🍎 macOS 최적화 버전"
    echo ""
    
    check_requirements
    check_optional_tools
    build_project
    install_system
    setup_launchagent
    run_tests
    show_usage
    
    echo ""
    print_success "설치가 완전히 완료되었습니다! 즐거운 모니터링하세요! 🎉"
}

# 스크립트 실행
main "$@" 