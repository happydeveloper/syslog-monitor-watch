#!/bin/bash

# Syslog Monitor Email Setup Script
echo "=== Syslog Monitor Email Setup ==="
echo

# 이메일 주소 입력 (여러 명 지원)
echo "수신할 이메일 주소를 입력하세요:"
echo "- 여러 명의 경우 쉼표로 구분: user1@example.com,user2@example.com"
echo "- 기본값: robot@lambda-x.ai,enfn2001@gmail.com"
read -p "이메일 주소: " EMAIL_TO
EMAIL_TO=${EMAIL_TO:-"robot@lambda-x.ai,enfn2001@gmail.com"}

# Gmail 계정 입력
read -p "Gmail 계정을 입력하세요 (예: your@gmail.com): " SMTP_USER
if [ -z "$SMTP_USER" ]; then
    echo "Error: Gmail 계정이 필요합니다."
    exit 1
fi

# App Password 입력
echo "Gmail App Password를 입력하세요:"
echo "App Password 생성 방법:"
echo "1. Google 계정 설정 > 보안 > 2단계 인증 활성화"
echo "2. https://myaccount.google.com/apppasswords 방문"
echo "3. 앱 선택 > 메일, 기기 선택 > 기타"
echo "4. 생성된 16자리 비밀번호 사용"
echo
read -s -p "App Password: " SMTP_PASSWORD
echo

if [ -z "$SMTP_PASSWORD" ]; then
    echo "Error: App Password가 필요합니다."
    exit 1
fi

# 환경변수 파일 생성
ENV_FILE=".env"
cat > $ENV_FILE << EOF
# Syslog Monitor Email Configuration
export SYSLOG_EMAIL_TO="$EMAIL_TO"
export SYSLOG_EMAIL_FROM="$SMTP_USER"
export SYSLOG_SMTP_SERVER="smtp.gmail.com"
export SYSLOG_SMTP_PORT="587"
export SYSLOG_SMTP_USER="$SMTP_USER"
export SYSLOG_SMTP_PASSWORD="$SMTP_PASSWORD"
EOF

echo
echo "설정이 완료되었습니다!"
echo "설정 파일: $ENV_FILE"
echo
echo "사용 방법:"
echo "1. 환경변수 로드: source .env"
echo "2. 테스트: ./syslog-monitor -test-email"
echo "3. 실행: ./syslog-monitor -file=test.log -keywords=error,critical,failed"
echo
echo "보안 주의사항:"
echo "- .env 파일에는 민감한 정보가 포함되어 있습니다"
echo "- .gitignore에 .env를 추가하여 버전 관리에서 제외하세요"
echo "- 파일 권한: chmod 600 .env"

# 파일 권한 설정
chmod 600 $ENV_FILE
echo
echo "파일 권한이 600으로 설정되었습니다." 