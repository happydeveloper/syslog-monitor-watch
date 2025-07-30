#!/bin/bash

# Lambda-X Syslog Monitor Log Rotation Script
# ============================================
# 로그 파일 로테이션 및 압축 스크립트

set -e

# 설정
LOG_DIR="/usr/local/var/log"
MAX_DAYS=30        # 30일 이상 된 로그 파일 삭제
MAX_SIZE="100M"    # 100MB 이상일 때 로테이션
COMPRESS_DAYS=1    # 1일 이상 된 로그는 압축

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 로그 함수
log_info() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] $1"
}

log_warning() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [WARNING] $1"
}

log_error() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ERROR] $1"
}

# 파일 크기 확인 함수
get_file_size() {
    local file="$1"
    if [[ -f "$file" ]]; then
        stat -f%z "$file" 2>/dev/null || echo "0"
    else
        echo "0"
    fi
}

# 크기를 바이트로 변환
size_to_bytes() {
    local size="$1"
    case "${size: -1}" in
        K|k) echo $((${size%?} * 1024)) ;;
        M|m) echo $((${size%?} * 1024 * 1024)) ;;
        G|g) echo $((${size%?} * 1024 * 1024 * 1024)) ;;
        *) echo "${size}" ;;
    esac
}

# 로그 파일 로테이트
rotate_log() {
    local logfile="$1"
    local max_size_bytes=$(size_to_bytes "$MAX_SIZE")
    
    if [[ ! -f "$logfile" ]]; then
        return 0
    fi
    
    local file_size=$(get_file_size "$logfile")
    local basename="${logfile%.*}"
    local extension="${logfile##*.}"
    
    log_info "Checking $logfile (size: $file_size bytes)"
    
    # 크기 확인
    if [[ $file_size -gt $max_size_bytes ]]; then
        log_info "Rotating $logfile (size: $file_size > $max_size_bytes)"
        
        # 기존 로테이트된 파일들 이동
        for i in {9..1}; do
            local old_file="${basename}.${i}.${extension}"
            local new_file="${basename}.$((i+1)).${extension}"
            
            if [[ -f "$old_file" ]]; then
                mv "$old_file" "$new_file"
                log_info "Moved $old_file to $new_file"
            fi
            
            # 압축 파일도 처리
            if [[ -f "${old_file}.gz" ]]; then
                mv "${old_file}.gz" "${new_file}.gz"
                log_info "Moved ${old_file}.gz to ${new_file}.gz"
            fi
        done
        
        # 현재 로그 파일 로테이트
        if [[ -f "$logfile" ]]; then
            cp "$logfile" "${basename}.1.${extension}"
            > "$logfile"  # 원본 파일 내용 비우기
            log_info "Rotated $logfile to ${basename}.1.${extension}"
            
            # 서비스에 SIGHUP 전송 (로그 파일 재오픈)
            local pid_file="/usr/local/var/run/syslog-monitor.pid"
            if [[ -f "$pid_file" ]]; then
                local pid=$(cat "$pid_file")
                if kill -0 "$pid" 2>/dev/null; then
                    kill -HUP "$pid"
                    log_info "Sent SIGHUP to process $pid"
                fi
            fi
        fi
    fi
}

# 오래된 로그 파일 압축
compress_old_logs() {
    log_info "Compressing logs older than $COMPRESS_DAYS days"
    
    find "$LOG_DIR" -name "*.log.*" -type f ! -name "*.gz" -mtime +$COMPRESS_DAYS | while read -r file; do
        if [[ -f "$file" ]]; then
            log_info "Compressing $file"
            gzip "$file"
            if [[ $? -eq 0 ]]; then
                log_info "Successfully compressed $file"
            else
                log_error "Failed to compress $file"
            fi
        fi
    done
}

# 오래된 로그 파일 삭제
cleanup_old_logs() {
    log_info "Cleaning up logs older than $MAX_DAYS days"
    
    # 압축된 로그 파일 삭제
    find "$LOG_DIR" -name "*.log.*.gz" -type f -mtime +$MAX_DAYS | while read -r file; do
        if [[ -f "$file" ]]; then
            log_info "Removing old compressed log: $file"
            rm -f "$file"
        fi
    done
    
    # 일반 로그 파일 삭제 (매우 오래된 것들)
    find "$LOG_DIR" -name "*.log.*" -type f ! -name "*.gz" -mtime +$((MAX_DAYS + 7)) | while read -r file; do
        if [[ -f "$file" ]]; then
            log_info "Removing very old log: $file"
            rm -f "$file"
        fi
    done
}

# 디스크 사용량 체크
check_disk_usage() {
    local usage=$(df "$LOG_DIR" | tail -1 | awk '{print $5}' | sed 's/%//')
    log_info "Disk usage for $LOG_DIR: ${usage}%"
    
    if [[ $usage -gt 90 ]]; then
        log_warning "Disk usage is high (${usage}%). Consider cleaning up more aggressively."
        
        # 긴급 정리: 더 오래된 파일들 삭제
        find "$LOG_DIR" -name "*.log.*.gz" -type f -mtime +$((MAX_DAYS / 2)) | while read -r file; do
            log_warning "Emergency cleanup: removing $file"
            rm -f "$file"
        done
    fi
}

# 통계 수집
collect_stats() {
    log_info "=== Log Rotation Statistics ==="
    
    # 로그 파일 개수
    local log_count=$(find "$LOG_DIR" -name "*.log*" -type f | wc -l)
    log_info "Total log files: $log_count"
    
    # 압축된 파일 개수
    local compressed_count=$(find "$LOG_DIR" -name "*.gz" -type f | wc -l)
    log_info "Compressed files: $compressed_count"
    
    # 총 로그 디렉토리 크기
    local total_size=$(du -sh "$LOG_DIR" | cut -f1)
    log_info "Total log directory size: $total_size"
    
    # 가장 큰 파일들
    log_info "Top 5 largest log files:"
    find "$LOG_DIR" -name "*.log*" -type f -exec ls -lh {} \; | sort -k5 -hr | head -5 | while read -r line; do
        log_info "  $line"
    done
}

# 메인 함수
main() {
    log_info "Starting log rotation for Lambda-X Syslog Monitor"
    
    # 로그 디렉토리 존재 확인
    if [[ ! -d "$LOG_DIR" ]]; then
        log_error "Log directory does not exist: $LOG_DIR"
        exit 1
    fi
    
    cd "$LOG_DIR"
    
    # 주요 로그 파일들 로테이트
    rotate_log "syslog-monitor.out.log"
    rotate_log "syslog-monitor.err.log"
    rotate_log "syslog-monitor.log"
    rotate_log "logrotate.out.log"
    rotate_log "logrotate.err.log"
    
    # 오래된 로그 압축
    compress_old_logs
    
    # 오래된 로그 정리
    cleanup_old_logs
    
    # 디스크 사용량 체크
    check_disk_usage
    
    # 통계 수집
    collect_stats
    
    log_info "Log rotation completed successfully"
}

# 스크립트 실행
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi