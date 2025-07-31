#!/bin/bash

# Lambda-X Syslog Monitor Service Uninstaller
# ===========================================
# macOS LaunchAgent 완전 제거 및 정리 스크립트

set -e

# 중복 실행 방지를 위한 락 파일 설정
LOCK_FILE="/tmp/syslog-monitor-uninstall.lock"
SCRIPT_PID=$$

# 락 파일 정리 함수
cleanup_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local lock_pid=$(cat "$LOCK_FILE" 2>/dev/null || echo "")
        if [[ "$lock_pid" == "$SCRIPT_PID" ]]; then
            rm -f "$LOCK_FILE"
        fi
    fi
}

# 스크립트 종료 시 락 파일 정리
trap cleanup_lock EXIT INT TERM

# 중복 실행 확인
check_duplicate_execution() {
    if [[ -f "$LOCK_FILE" ]]; then
        local existing_pid=$(cat "$LOCK_FILE" 2>/dev/null || echo "")
        if [[ -n "$existing_pid" ]] && kill -0 "$existing_pid" 2>/dev/null; then
            print_error "다른 제거 스크립트가 이미 실행 중입니다 (PID: $existing_pid)"
            print_status "잠시 후 다시 시도하거나, 프로세스를 확인해주세요: ps -p $existing_pid"
            exit 1
        else
            # 죽은 프로세스의 락 파일 정리
            rm -f "$LOCK_FILE"
        fi
    fi
    
    # 새로운 락 파일 생성
    echo "$SCRIPT_PID" > "$LOCK_FILE"
    print_status "제거 스크립트 시작 (PID: $SCRIPT_PID)"
}

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로고 및 헤더
echo -e "${RED}"
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                    🗑️  Lambda-X Syslog Monitor                       ║"
echo "║                       Service Uninstaller                           ║"
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

# 사용자 확인
confirm_uninstall() {
    echo -e "${YELLOW}⚠️  경고: 이 스크립트는 Lambda-X Syslog Monitor 서비스를 완전히 제거합니다.${NC}"
    echo
    echo "제거될 항목들:"
    echo "  • LaunchAgent 서비스 (자동 시작 비활성화)"
    echo "  • /usr/local/bin/syslog-monitor 실행 파일"
    echo "  • /usr/local/bin/rotate-syslog-logs.sh 스크립트"
    echo "  • LaunchAgent plist 파일들"
    echo "  • 로그 파일들 (선택사항)"
    echo "  • 설정 파일들 (선택사항)"
    echo
    
    read -p "정말로 제거하시겠습니까? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "제거가 취소되었습니다."
        exit 0
    fi
}

# 서비스 상태 확인
check_service_status() {
    print_status "현재 서비스 상태 확인 중..."
    
    # LaunchAgent 상태 확인
    if launchctl list com.lambda-x.syslog-monitor &>/dev/null; then
        print_warning "서비스가 현재 실행 중입니다."
        return 0
    else
        print_status "서비스가 실행되지 않고 있습니다."
        return 1
    fi
}

# 서비스 중지
stop_service() {
    print_status "서비스 중지 중..."
    
    # 메인 서비스 중지
    homeDir=$(eval echo ~$USER)
    plistFile="$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
    
    if [[ -f "$plistFile" ]]; then
        # 먼저 서비스가 로드되어 있는지 확인
        if launchctl list com.lambda-x.syslog-monitor &>/dev/null; then
            print_status "메인 서비스 중지 중..."
            launchctl unload "$plistFile" 2>/dev/null || {
                print_warning "정상적인 서비스 중지 실패, 강제 중지 시도 중..."
                launchctl remove com.lambda-x.syslog-monitor 2>/dev/null || true
            }
            # 서비스 중지 완료 대기
            sleep 3
            if ! launchctl list com.lambda-x.syslog-monitor &>/dev/null; then
                print_success "메인 서비스가 중지되었습니다."
            else
                print_warning "메인 서비스가 완전히 중지되지 않았습니다."
            fi
        else
            print_status "메인 서비스가 이미 중지되어 있습니다."
        fi
    else
        print_status "메인 서비스 plist 파일이 존재하지 않습니다."
    fi
    
    # 로그 로테이션 서비스 중지
    logRotatePlist="$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist"
    if [[ -f "$logRotatePlist" ]]; then
        if launchctl list com.lambda-x.syslog-monitor.logrotate &>/dev/null; then
            print_status "로그 로테이션 서비스 중지 중..."
            launchctl unload "$logRotatePlist" 2>/dev/null || {
                print_warning "로그 로테이션 서비스 정상 중지 실패, 강제 중지 시도 중..."
                launchctl remove com.lambda-x.syslog-monitor.logrotate 2>/dev/null || true
            }
            sleep 2
            if ! launchctl list com.lambda-x.syslog-monitor.logrotate &>/dev/null; then
                print_success "로그 로테이션 서비스가 중지되었습니다."
            else
                print_warning "로그 로테이션 서비스가 완전히 중지되지 않았습니다."
            fi
        else
            print_status "로그 로테이션 서비스가 이미 중지되어 있습니다."
        fi
    else
        print_status "로그 로테이션 서비스 plist 파일이 존재하지 않습니다."
    fi
    
    # 실행 중인 프로세스 종료 (더 안전한 방법)
    local process_pids=$(pgrep -f "syslog-monitor" 2>/dev/null || true)
    if [[ -n "$process_pids" ]]; then
        print_status "실행 중인 syslog-monitor 프로세스 종료 중..."
        
        # 먼저 SIGTERM으로 정상 종료 시도
        for pid in $process_pids; do
            if kill -0 "$pid" 2>/dev/null; then
                print_status "프로세스 $pid 정상 종료 시도 중..."
                kill -TERM "$pid" 2>/dev/null || true
            fi
        done
        
        # 5초 대기
        sleep 5
        
        # 아직 실행 중인 프로세스가 있다면 강제 종료
        local remaining_pids=$(pgrep -f "syslog-monitor" 2>/dev/null || true)
        if [[ -n "$remaining_pids" ]]; then
            print_warning "일부 프로세스가 아직 실행 중입니다. 강제 종료 시도 중..."
            for pid in $remaining_pids; do
                if kill -0 "$pid" 2>/dev/null; then
                    print_warning "프로세스 $pid 강제 종료 중..."
                    kill -KILL "$pid" 2>/dev/null || true
                fi
            done
            sleep 2
        fi
        
        # 최종 확인
        if ! pgrep -f "syslog-monitor" >/dev/null 2>&1; then
            print_success "모든 프로세스가 종료되었습니다."
        else
            print_error "일부 프로세스가 아직 실행 중입니다. 수동으로 확인이 필요합니다."
        fi
    else
        print_status "실행 중인 syslog-monitor 프로세스가 없습니다."
    fi
}

# plist 파일 제거
remove_plist_files() {
    print_status "LaunchAgent plist 파일 제거 중..."
    
    homeDir=$(eval echo ~$USER)
    
    # 메인 서비스 plist
    plistFile="$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
    if [[ -f "$plistFile" ]]; then
        rm -f "$plistFile"
        print_success "메인 서비스 plist 파일이 제거되었습니다."
    fi
    
    # 로그 로테이션 plist
    logRotatePlist="$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist"
    if [[ -f "$logRotatePlist" ]]; then
        rm -f "$logRotatePlist"
        print_success "로그 로테이션 plist 파일이 제거되었습니다."
    fi
}

# 실행 파일 제거
remove_binaries() {
    print_status "실행 파일 제거 중..."
    
    # 메인 바이너리
    local binary_path="/usr/local/bin/syslog-monitor"
    if [[ -f "$binary_path" ]]; then
        # 파일이 실행 중인지 확인 (lsof로 체크)
        if command -v lsof >/dev/null 2>&1; then
            if lsof "$binary_path" >/dev/null 2>&1; then
                print_warning "실행 파일이 사용 중입니다. 프로세스 종료를 기다립니다..."
                sleep 3
            fi
        fi
        
        # sudo 권한 확인 및 파일 제거
        if sudo -n true 2>/dev/null; then
            if sudo rm -f "$binary_path" 2>/dev/null; then
                print_success "syslog-monitor 실행 파일이 제거되었습니다."
            else
                print_error "syslog-monitor 실행 파일 제거 실패"
            fi
        else
            print_warning "sudo 권한이 필요합니다. syslog-monitor 실행 파일 제거를 건너뜁니다."
            echo "  수동 제거: sudo rm -f $binary_path"
        fi
    else
        print_status "syslog-monitor 실행 파일이 존재하지 않습니다."
    fi
    
    # 로그 로테이션 스크립트
    local script_path="/usr/local/bin/rotate-syslog-logs.sh"
    if [[ -f "$script_path" ]]; then
        # 스크립트가 실행 중인지 확인
        if command -v lsof >/dev/null 2>&1; then
            if lsof "$script_path" >/dev/null 2>&1; then
                print_warning "로그 로테이션 스크립트가 사용 중입니다. 잠시 대기합니다..."
                sleep 2
            fi
        fi
        
        # sudo 권한 확인 및 파일 제거
        if sudo -n true 2>/dev/null; then
            if sudo rm -f "$script_path" 2>/dev/null; then
                print_success "로그 로테이션 스크립트가 제거되었습니다."
            else
                print_error "로그 로테이션 스크립트 제거 실패"
            fi
        else
            print_warning "sudo 권한이 필요합니다. 로그 로테이션 스크립트 제거를 건너뜁니다."
            echo "  수동 제거: sudo rm -f $script_path"
        fi
    else
        print_status "로그 로테이션 스크립트가 존재하지 않습니다."
    fi
    
    # 심볼릭 링크도 확인하여 제거
    local symlink_locations=(
        "/usr/bin/syslog-monitor"
        "/bin/syslog-monitor"
        "/usr/sbin/syslog-monitor"
        "/sbin/syslog-monitor"
    )
    
    for symlink in "${symlink_locations[@]}"; do
        if [[ -L "$symlink" ]] || [[ -f "$symlink" ]]; then
            print_status "심볼릭 링크/파일 발견: $symlink"
            if sudo -n true 2>/dev/null; then
                if sudo rm -f "$symlink" 2>/dev/null; then
                    print_success "심볼릭 링크가 제거되었습니다: $symlink"
                else
                    print_warning "심볼릭 링크 제거 실패: $symlink"
                fi
            else
                print_warning "sudo 권한이 필요합니다: $symlink"
            fi
        fi
    done
}

# PID 파일 제거
remove_pid_files() {
    print_status "PID 파일 제거 중..."
    
    local pidFiles=(
        "/usr/local/var/run/syslog-monitor.pid"
        "/tmp/syslog-monitor.pid"
        "/var/run/syslog-monitor.pid"
        "/var/tmp/syslog-monitor.pid"
        "$HOME/.syslog-monitor/syslog-monitor.pid"
    )
    
    local removed_count=0
    local failed_count=0
    
    for pidFile in "${pidFiles[@]}"; do
        if [[ -f "$pidFile" ]]; then
            print_status "PID 파일 발견: $pidFile"
            
            # PID 파일 내용 읽어서 프로세스가 실제로 실행 중인지 확인
            if [[ -r "$pidFile" ]]; then
                local pid_content=$(cat "$pidFile" 2>/dev/null || echo "")
                if [[ -n "$pid_content" ]] && [[ "$pid_content" =~ ^[0-9]+$ ]]; then
                    if kill -0 "$pid_content" 2>/dev/null; then
                        print_warning "PID $pid_content가 아직 실행 중입니다. 파일 제거를 잠시 대기합니다..."
                        sleep 2
                        # 다시 확인 후 여전히 실행 중이면 경고
                        if kill -0 "$pid_content" 2>/dev/null; then
                            print_warning "프로세스가 여전히 실행 중입니다. PID 파일 제거를 강행합니다."
                        fi
                    fi
                fi
            fi
            
            # 파일 제거 시도
            local need_sudo=false
            if [[ ! -w "$pidFile" ]] || [[ ! -w "$(dirname "$pidFile")" ]]; then
                need_sudo=true
            fi
            
            if [[ "$need_sudo" == "true" ]]; then
                if sudo -n true 2>/dev/null; then
                    if sudo rm -f "$pidFile" 2>/dev/null; then
                        print_success "PID 파일이 제거되었습니다: $pidFile"
                        ((removed_count++))
                    else
                        print_error "PID 파일 제거 실패 (sudo): $pidFile"
                        ((failed_count++))
                    fi
                else
                    print_warning "sudo 권한이 필요합니다: $pidFile"
                    echo "  수동 제거: sudo rm -f $pidFile"
                    ((failed_count++))
                fi
            else
                if rm -f "$pidFile" 2>/dev/null; then
                    print_success "PID 파일이 제거되었습니다: $pidFile"
                    ((removed_count++))
                else
                    print_error "PID 파일 제거 실패: $pidFile"
                    ((failed_count++))
                fi
            fi
        fi
    done
    
    # 추가 PID 파일 검색 (패턴 매칭)
    local additional_locations=(
        "/tmp"
        "/var/tmp"
        "/usr/local/var/run"
        "/var/run"
    )
    
    for location in "${additional_locations[@]}"; do
        if [[ -d "$location" ]]; then
            local found_pids=$(find "$location" -maxdepth 1 -name "*syslog-monitor*.pid" 2>/dev/null || true)
            if [[ -n "$found_pids" ]]; then
                print_status "추가 PID 파일 발견: $location"
                while IFS= read -r additional_pid; do
                    if [[ -f "$additional_pid" ]]; then
                        print_status "추가 PID 파일 제거: $additional_pid"
                        if [[ -w "$additional_pid" ]]; then
                            rm -f "$additional_pid" 2>/dev/null && ((removed_count++)) || ((failed_count++))
                        else
                            if sudo -n true 2>/dev/null; then
                                sudo rm -f "$additional_pid" 2>/dev/null && ((removed_count++)) || ((failed_count++))
                            else
                                print_warning "sudo 권한 필요: $additional_pid"
                                ((failed_count++))
                            fi
                        fi
                    fi
                done <<< "$found_pids"
            fi
        fi
    done
    
    # 결과 요약
    if [[ $removed_count -gt 0 ]]; then
        print_success "총 ${removed_count}개의 PID 파일이 제거되었습니다."
    fi
    if [[ $failed_count -gt 0 ]]; then
        print_warning "총 ${failed_count}개의 PID 파일 제거에 실패했습니다."
    fi
    if [[ $removed_count -eq 0 ]] && [[ $failed_count -eq 0 ]]; then
        print_status "제거할 PID 파일이 없습니다."
    fi
}

# 로그 파일 제거 (선택사항)
remove_log_files() {
    echo
    read -p "로그 파일도 제거하시겠습니까? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "로그 파일 제거 중..."
        
        local logFiles=(
            "/usr/local/var/log/syslog-monitor.out.log"
            "/usr/local/var/log/syslog-monitor.err.log" 
            "/usr/local/var/log/syslog-monitor.log"
            "/usr/local/var/log/logrotate.out.log"
            "/usr/local/var/log/logrotate.err.log"
        )
        
        local removed_count=0
        local failed_count=0
        
        for logFile in "${logFiles[@]}"; do
            if [[ -f "$logFile" ]]; then
                print_status "로그 파일 제거 시도: $(basename "$logFile")"
                
                # 권한 확인 및 안전한 제거
                if [[ -w "$logFile" ]] && [[ -w "$(dirname "$logFile")" ]]; then
                    # 일반 권한으로 제거 가능
                    if rm -f "$logFile" 2>/dev/null; then
                        print_success "로그 파일이 제거되었습니다: $logFile"
                        ((removed_count++))
                    else
                        print_error "로그 파일 제거 실패: $logFile"
                        ((failed_count++))
                    fi
                else
                    # sudo 권한 필요
                    if sudo -n true 2>/dev/null; then
                        if sudo rm -f "$logFile" 2>/dev/null; then
                            print_success "로그 파일이 제거되었습니다 (sudo): $logFile"
                            ((removed_count++))
                        else
                            print_error "로그 파일 제거 실패 (sudo): $logFile"
                            ((failed_count++))
                        fi
                    else
                        print_warning "sudo 권한이 필요합니다: $logFile"
                        echo "  수동 제거: sudo rm -f $logFile"
                        ((failed_count++))
                    fi
                fi
            fi
        done
        
        # 로테이트된 로그 파일들도 제거 (더 안전한 방법)
        print_status "로테이트된 로그 파일 검색 중..."
        local log_directories=(
            "/usr/local/var/log"
            "/var/log"
            "/tmp"
        )
        
        for log_dir in "${log_directories[@]}"; do
            if [[ -d "$log_dir" ]]; then
                local rotated_logs=$(find "$log_dir" -maxdepth 1 -name "syslog-monitor*.log*" -type f 2>/dev/null || true)
                if [[ -n "$rotated_logs" ]]; then
                    print_status "로테이트된 로그 파일 발견: $log_dir"
                    while IFS= read -r rotated_log; do
                        if [[ -f "$rotated_log" ]]; then
                            print_status "로테이트된 로그 제거: $(basename "$rotated_log")"
                            if [[ -w "$rotated_log" ]]; then
                                rm -f "$rotated_log" 2>/dev/null && ((removed_count++)) || ((failed_count++))
                            else
                                if sudo -n true 2>/dev/null; then
                                    sudo rm -f "$rotated_log" 2>/dev/null && ((removed_count++)) || ((failed_count++))
                                else
                                    print_warning "sudo 권한 필요: $rotated_log"
                                    ((failed_count++))
                                fi
                            fi
                        fi
                    done <<< "$rotated_logs"
                fi
            fi
        done
        
        # 빈 로그 디렉토리 제거 시도 (안전한 확인)
        if [[ -d "/usr/local/var/log" ]]; then
            local log_contents=$(ls -A "/usr/local/var/log" 2>/dev/null || true)
            if [[ -z "$log_contents" ]]; then
                print_status "빈 로그 디렉토리 제거 시도..."
                if sudo -n true 2>/dev/null; then
                    if sudo rmdir "/usr/local/var/log" 2>/dev/null; then
                        print_success "빈 로그 디렉토리가 제거되었습니다."
                    else
                        print_warning "빈 로그 디렉토리 제거 실패 (다른 파일이 있을 수 있음)"
                    fi
                else
                    print_warning "sudo 권한이 필요하여 빈 디렉토리 제거를 건너뜁니다."
                fi
            fi
        fi
        
        # 결과 요약
        if [[ $removed_count -gt 0 ]] && [[ $failed_count -eq 0 ]]; then
            print_success "모든 로그 파일이 제거되었습니다 (${removed_count}개 파일)"
        elif [[ $removed_count -gt 0 ]] && [[ $failed_count -gt 0 ]]; then
            print_warning "일부 로그 파일이 제거되었습니다 (${removed_count}개 성공, ${failed_count}개 실패)"
        elif [[ $removed_count -eq 0 ]] && [[ $failed_count -gt 0 ]]; then
            print_error "로그 파일 제거에 실패했습니다 (${failed_count}개 실패)"
        else
            print_status "제거할 로그 파일이 없습니다."
        fi
    else
        print_status "로그 파일은 보존됩니다."
    fi
}

# 설정 파일 제거 (선택사항)
remove_config_files() {
    echo
    read -p "사용자 설정 파일도 제거하시겠습니까? (~/.syslog-monitor/) (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "설정 파일 제거 중..."
        
        homeDir=$(eval echo ~$USER)
        configDir="$homeDir/.syslog-monitor"
        
        if [[ -d "$configDir" ]]; then
            rm -rf "$configDir"
            print_success "설정 디렉토리가 제거되었습니다: $configDir"
        fi
    else
        print_status "설정 파일은 보존됩니다."
    fi
}

# 프로젝트 파일 정리 (선택사항)
cleanup_project_files() {
    echo
    read -p "현재 디렉토리의 서비스 관련 파일들도 제거하시겠습니까? (plist, 스크립트 등) (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "프로젝트 파일 정리 중..."
        
        local projectFiles=(
            "com.lambda-x.syslog-monitor.plist"
            "com.lambda-x.syslog-monitor.logrotate.plist"
            "install-service.sh"
            "uninstall-service.sh"
            "rotate-syslog-logs.sh"
            "syslog-monitor"  # 로컬 빌드된 바이너리
        )
        
        for file in "${projectFiles[@]}"; do
            if [[ -f "$file" ]]; then
                rm -f "$file"
                print_success "프로젝트 파일이 제거되었습니다: $file"
            fi
        done
    else
        print_status "프로젝트 파일은 보존됩니다."
    fi
}

# 시스템 정리
cleanup_system() {
    print_status "시스템 정리 중..."
    
    # launchctl 캐시 정리
    launchctl list | grep -i lambda-x | while read -r line; do
        serviceName=$(echo "$line" | awk '{print $3}')
        if [[ -n "$serviceName" ]]; then
            print_warning "남은 서비스 발견: $serviceName"
        fi
    done
    
    # 실행 중인 관련 프로세스 재확인
    if pgrep -f "syslog-monitor\|lambda-x" >/dev/null; then
        print_warning "아직 실행 중인 관련 프로세스가 있습니다."
        ps aux | grep -E "syslog-monitor|lambda-x" | grep -v grep
    fi
}

# 제거 확인
verify_removal() {
    print_status "제거 상태 확인 중..."
    
    local issues=0
    local warnings=0
    local cleanup_commands=()
    
    # LaunchAgent 상태 상세 확인
    print_status "LaunchAgent 서비스 상태 확인..."
    local lambda_services=$(launchctl list | grep -i lambda-x 2>/dev/null || true)
    if [[ -n "$lambda_services" ]]; then
        print_error "일부 LaunchAgent 서비스가 아직 남아있습니다:"
        echo "$lambda_services" | while IFS= read -r line; do
            echo "  - $line"
            local service_name=$(echo "$line" | awk '{print $3}')
            if [[ -n "$service_name" ]]; then
                cleanup_commands+=("launchctl remove $service_name")
            fi
        done
        ((issues++))
    else
        print_success "LaunchAgent 서비스가 완전히 제거되었습니다."
    fi
    
    # 실행 파일들 전체 확인
    print_status "실행 파일 상태 확인..."
    local binary_locations=(
        "/usr/local/bin/syslog-monitor"
        "/usr/bin/syslog-monitor"
        "/bin/syslog-monitor"
        "/usr/sbin/syslog-monitor"
        "/sbin/syslog-monitor"
        "/usr/local/bin/rotate-syslog-logs.sh"
    )
    
    for binary in "${binary_locations[@]}"; do
        if [[ -f "$binary" ]] || [[ -L "$binary" ]]; then
            print_error "실행 파일/링크가 아직 남아있습니다: $binary"
            cleanup_commands+=("sudo rm -f $binary")
            ((issues++))
        fi
    done
    
    if [[ $issues -eq 0 ]]; then
        print_success "모든 실행 파일이 제거되었습니다."
    fi
    
    # plist 파일들 확인
    print_status "plist 파일 상태 확인..."
    homeDir=$(eval echo ~$USER)
    local plist_files=(
        "$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
        "$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist"
        "/Library/LaunchDaemons/com.lambda-x.syslog-monitor.plist"
        "/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
    )
    
    for plist in "${plist_files[@]}"; do
        if [[ -f "$plist" ]]; then
            print_error "plist 파일이 아직 남아있습니다: $plist"
            cleanup_commands+=("rm -f $plist")
            ((issues++))
        fi
    done
    
    if [[ ${#plist_files[@]} -eq 0 ]] || ! find "${plist_files[@]}" -type f 2>/dev/null | grep -q .; then
        print_success "모든 plist 파일이 제거되었습니다."
    fi
    
    # 실행 중인 프로세스 상세 확인
    print_status "실행 중인 프로세스 확인..."
    local running_processes=$(pgrep -fl "syslog-monitor\|lambda-x" 2>/dev/null || true)
    if [[ -n "$running_processes" ]]; then
        print_error "관련 프로세스가 아직 실행 중입니다:"
        echo "$running_processes" | while IFS= read -r process; do
            echo "  - $process"
            local pid=$(echo "$process" | awk '{print $1}')
            if [[ -n "$pid" ]]; then
                cleanup_commands+=("kill -TERM $pid")
                cleanup_commands+=("kill -KILL $pid")
            fi
        done
        ((issues++))
    else
        print_success "실행 중인 관련 프로세스가 없습니다."
    fi
    
    # PID 파일들 확인
    print_status "PID 파일 상태 확인..."
    local pid_locations=(
        "/usr/local/var/run"
        "/var/run"
        "/tmp"
        "/var/tmp"
        "$HOME/.syslog-monitor"
    )
    
    local found_pids=false
    for location in "${pid_locations[@]}"; do
        if [[ -d "$location" ]]; then
            local remaining_pids=$(find "$location" -maxdepth 1 -name "*syslog-monitor*.pid" 2>/dev/null || true)
            if [[ -n "$remaining_pids" ]]; then
                print_warning "PID 파일이 아직 남아있습니다:"
                echo "$remaining_pids" | while IFS= read -r pid_file; do
                    echo "  - $pid_file"
                    cleanup_commands+=("sudo rm -f $pid_file")
                done
                found_pids=true
                ((warnings++))
            fi
        fi
    done
    
    if [[ "$found_pids" == "false" ]]; then
        print_success "모든 PID 파일이 제거되었습니다."
    fi
    
    # 설정 파일/디렉토리 확인 (경고만)
    print_status "설정 파일 상태 확인..."
    local config_locations=(
        "$HOME/.syslog-monitor"
        "/usr/local/etc/syslog-monitor"
        "/etc/syslog-monitor"
    )
    
    for config_dir in "${config_locations[@]}"; do
        if [[ -d "$config_dir" ]]; then
            print_warning "설정 디렉토리가 아직 남아있습니다: $config_dir"
            echo "  (사용자 선택에 따라 보존될 수 있습니다)"
            ((warnings++))
        fi
    done
    
    # 로그 파일들 확인 (경고만)
    print_status "로그 파일 상태 확인..."
    local log_locations=(
        "/usr/local/var/log/syslog-monitor.*.log*"
        "/var/log/syslog-monitor.*.log*"
        "/usr/local/var/log/logrotate.*.log*"
    )
    
    for log_pattern in "${log_locations[@]}"; do
        local found_logs=$(ls $log_pattern 2>/dev/null || true)
        if [[ -n "$found_logs" ]]; then
            print_warning "로그 파일들이 아직 남아있습니다:"
            echo "$found_logs" | while IFS= read -r log_file; do
                echo "  - $log_file"
            done
            echo "  (사용자 선택에 따라 보존될 수 있습니다)"
            ((warnings++))
            break
        fi
    done
    
    # 결과 요약
    echo
    print_status "=== 제거 상태 요약 ==="
    
    if [[ $issues -eq 0 ]] && [[ $warnings -eq 0 ]]; then
        print_success "✅ 모든 구성 요소가 성공적으로 제거되었습니다!"
    elif [[ $issues -eq 0 ]] && [[ $warnings -gt 0 ]]; then
        print_success "✅ 핵심 구성 요소는 모두 제거되었습니다."
        print_warning "⚠️  $warnings개의 경고사항이 있습니다 (일반적으로 문제없음)."
    else
        print_error "❌ $issues개의 문제가 발견되었습니다."
        if [[ $warnings -gt 0 ]]; then
            print_warning "⚠️  추가로 $warnings개의 경고사항이 있습니다."
        fi
        
        if [[ ${#cleanup_commands[@]} -gt 0 ]]; then
            echo
            print_status "💡 수동 정리 명령어:"
            for cmd in "${cleanup_commands[@]}"; do
                echo "  $cmd"
            done
        fi
    fi
    
    return $issues
}

# 제거 요약 표시
show_removal_summary() {
    echo
    print_status "제거 완료! 🎉"
    echo
    echo -e "${YELLOW}📋 제거된 항목들:${NC}"
    echo "  ✅ LaunchAgent 서비스 중지 및 제거"
    echo "  ✅ 실행 파일 제거 (/usr/local/bin/syslog-monitor)"
    echo "  ✅ plist 파일 제거"
    echo "  ✅ PID 파일 정리"
    echo "  ✅ 시스템 정리"
    echo
    echo -e "${YELLOW}📄 보존된 파일들 (선택에 따라):${NC}"
    echo "  • 로그 파일: /usr/local/var/log/syslog-monitor.*"
    echo "  • 설정 파일: ~/.syslog-monitor/"
    echo "  • 프로젝트 파일: 현재 디렉토리의 관련 파일들"
    echo
    echo -e "${YELLOW}🔧 수동 정리가 필요한 경우:${NC}"
    echo "  # 남은 프로세스 확인"
    echo "  ps aux | grep syslog-monitor"
    echo
    echo "  # 남은 LaunchAgent 확인"
    echo "  launchctl list | grep lambda-x"
    echo
    echo "  # 남은 파일 검색"
    echo "  find /usr/local -name '*syslog-monitor*' 2>/dev/null"
    echo "  find ~ -name '*syslog-monitor*' 2>/dev/null"
    echo
    echo -e "${GREEN}✨ Lambda-X Syslog Monitor가 성공적으로 제거되었습니다!${NC}"
}

# 에러 처리
handle_error() {
    local exit_code=$?
    local line_number=${BASH_LINENO[0]}
    
    print_error "제거 중 오류가 발생했습니다 (라인: $line_number, 종료 코드: $exit_code)"
    print_status "오류가 발생한 위치를 확인하고 수동으로 다음 단계를 수행해주세요:"
    echo
    
    # 현재 상태 확인하여 맞춤형 복구 방법 제공
    homeDir=$(eval echo ~$USER)
    
    echo "🔧 수동 복구 단계:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # 1. 서비스 중지
    echo "1. 서비스 강제 중지:"
    if [[ -f "$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist" ]]; then
        echo "   launchctl unload $homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
    fi
    if [[ -f "$homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist" ]]; then
        echo "   launchctl unload $homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist"
    fi
    echo "   launchctl remove com.lambda-x.syslog-monitor"
    echo "   launchctl remove com.lambda-x.syslog-monitor.logrotate"
    echo
    
    # 2. 파일 제거
    echo "2. 파일 수동 제거:"
    echo "   sudo rm -f /usr/local/bin/syslog-monitor"
    echo "   sudo rm -f /usr/local/bin/rotate-syslog-logs.sh"
    echo "   rm -f $homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.plist"
    echo "   rm -f $homeDir/Library/LaunchAgents/com.lambda-x.syslog-monitor.logrotate.plist"
    echo
    
    # 3. 프로세스 종료
    echo "3. 프로세스 강제 종료:"
    echo "   pkill -f syslog-monitor"
    echo "   pkill -9 -f syslog-monitor  # 강제 종료"
    echo
    
    # 4. PID 파일 정리
    echo "4. PID 파일 정리:"
    echo "   sudo rm -f /usr/local/var/run/syslog-monitor.pid"
    echo "   sudo rm -f /tmp/syslog-monitor.pid"
    echo "   sudo rm -f /var/run/syslog-monitor.pid"
    echo
    
    # 5. 상태 확인
    echo "5. 정리 상태 확인:"
    echo "   launchctl list | grep lambda-x"
    echo "   ps aux | grep syslog-monitor"
    echo "   find /usr/local -name '*syslog-monitor*' 2>/dev/null"
    echo
    
    # 락 파일 정리
    cleanup_lock
    
    print_error "수동 정리 후 스크립트를 다시 실행해보세요."
    exit $exit_code
}

# 메인 함수
main() {
    print_status "Lambda-X Syslog Monitor 서비스 제거를 시작합니다..."
    
    # 중복 실행 확인
    check_duplicate_execution
    
    # 에러 트랩 설정
    trap handle_error ERR
    
    confirm_uninstall
    check_service_status
    stop_service
    remove_plist_files
    remove_binaries
    remove_pid_files
    remove_log_files
    remove_config_files
    cleanup_project_files
    cleanup_system
    
    # 제거 상태 확인 및 결과에 따른 처리
    if verify_removal; then
        show_removal_summary
        print_success "제거가 완료되었습니다! ✅"
    else
        print_warning "제거 과정에서 일부 문제가 발견되었습니다."
        print_status "위의 수동 정리 명령어를 참고하여 완전한 제거를 완료해주세요."
        exit 1
    fi
}

# 스크립트 실행
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi