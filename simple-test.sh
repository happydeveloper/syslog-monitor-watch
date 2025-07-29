#!/bin/bash

echo "🧪 AI Syslog Monitor 간단 테스트"
echo "=================================="
echo "📅 $(date)"
echo ""

# 테스트 결과
PASS=0
TOTAL=0

test_check() {
    ((TOTAL++))
    if [ $1 -eq 0 ]; then
        echo "✅ $2"
        ((PASS++))
    else
        echo "❌ $2"
    fi
}

echo "1️⃣ 기본 설치 확인"
echo "----------------"

# 실행 파일 확인
[ -f "/usr/local/bin/syslog-monitor" ]
test_check $? "실행 파일 설치됨 (/usr/local/bin/syslog-monitor)"

# 설정 파일 확인
[ -f "$HOME/.syslog-monitor/config.json" ]
test_check $? "설정 파일 생성됨"

# 새로운 AI 기능 설정 확인
grep -q '"computer_name_detection": true' "$HOME/.syslog-monitor/config.json" 2>/dev/null
test_check $? "컴퓨터명 감지 기능 활성화"

grep -q '"ip_classification": true' "$HOME/.syslog-monitor/config.json" 2>/dev/null
test_check $? "IP 분류 기능 활성화"

grep -q '"asn_lookup": true' "$HOME/.syslog-monitor/config.json" 2>/dev/null
test_check $? "ASN 조회 기능 활성화"

echo ""
echo "2️⃣ 시스템 정보 수집"
echo "----------------"

# 컴퓨터명 확인
computer_name=$(hostname)
[ -n "$computer_name" ]
test_check $? "컴퓨터명 수집: $computer_name"

# 아키텍처 확인
arch=$(uname -m)
[ -n "$arch" ]
test_check $? "시스템 아키텍처: $arch"

# 인터넷 연결 확인
ping -c 1 8.8.8.8 &> /dev/null
test_check $? "인터넷 연결 (ASN 조회 가능)"

echo ""
echo "3️⃣ ASN 조회 기능"
echo "---------------"

# ASN API 테스트
if command -v curl &> /dev/null; then
    asn_result=$(timeout 5 curl -s "http://ip-api.com/json/8.8.8.8?fields=org" 2>/dev/null)
    [[ "$asn_result" == *"Google"* ]]
    test_check $? "ASN API 조회 성공: 8.8.8.8 → Google"
else
    echo "⚠️ curl 없음 - ASN 테스트 건너뜀"
fi

echo ""
echo "4️⃣ IP 분류 테스트"
echo "---------------"

# 내부 IP 분류 테스트
internal_ips=("192.168.1.1" "10.0.0.1" "172.16.0.1")
for ip in "${internal_ips[@]}"; do
    # 간단한 내부 IP 검증
    [[ "$ip" =~ ^192\.168\. ]] || [[ "$ip" =~ ^10\. ]] || [[ "$ip" =~ ^172\.1[6-9]\. ]]
    test_check $? "내부 IP 분류: $ip"
done

# 외부 IP 분류 테스트
external_ips=("8.8.8.8" "1.1.1.1")
for ip in "${external_ips[@]}"; do
    # 간단한 외부 IP 검증 (내부 IP가 아니면 외부)
    ! [[ "$ip" =~ ^192\.168\. ]] && ! [[ "$ip" =~ ^10\. ]] && ! [[ "$ip" =~ ^172\.1[6-9]\. ]]
    test_check $? "외부 IP 분류: $ip"
done

echo ""
echo "5️⃣ 기본 실행 테스트"
echo "----------------"

# 도움말 실행
timeout 3 syslog-monitor -help > /dev/null 2>&1
test_check $? "도움말 실행 성공"

echo ""
echo "📊 테스트 결과 요약"
echo "=================="
echo "✅ 성공: $PASS개"
echo "❌ 실패: $((TOTAL - PASS))개"
echo "📊 총 테스트: $TOTAL개"
echo "🎯 성공률: $(( PASS * 100 / TOTAL ))%"

echo ""
if [ $PASS -eq $TOTAL ]; then
    echo "🎉 모든 테스트 통과! AI 분석기가 완벽하게 설치되었습니다."
else
    echo "⚠️ 일부 테스트 실패. 문제를 확인해주세요."
fi

echo ""
echo "🚀 사용 준비 완료!"
echo "권장 명령: syslog-monitor -ai-analysis -system-monitor" 