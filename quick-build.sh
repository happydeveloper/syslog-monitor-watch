#!/bin/bash

# Lambda-X AI Security Monitor - Quick Build Script
# ================================================
# 
# 빠른 빌드 및 테스트용 스크립트

set -e

# 색상 정의
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Lambda-X AI Security Monitor - Quick Build${NC}"
echo "==============================================="

# 1. 의존성 확인
echo -e "${BLUE}📦 의존성 확인 중...${NC}"
go mod tidy
go mod download

# 2. 빌드
echo -e "${BLUE}🔨 빌드 중...${NC}"
go build -ldflags="-s -w" -o syslog-monitor

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 빌드 성공!${NC}"
    
    # 바이너리 정보
    BINARY_SIZE=$(du -h ./syslog-monitor | cut -f1)
    echo -e "${BLUE}📊 바이너리 크기: $BINARY_SIZE${NC}"
    
    # 테스트 실행
    echo -e "${BLUE}🧪 테스트 실행 중...${NC}"
    ./syslog-monitor --help | head -10
    
    echo -e "${GREEN}🎉 Quick Build 완료!${NC}"
    echo -e "${BLUE}💡 사용법: ./syslog-monitor -login-watch -system-monitor${NC}"
else
    echo -e "${RED}❌ 빌드 실패!${NC}"
    exit 1
fi 