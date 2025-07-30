# 🚨 시스템 모니터링 및 알림 설정 가이드

## 📋 새로 추가된 기능

### 1. 정기적 시스템 상태 보고서
- **기능**: 설정된 간격으로 시스템 상태를 이메일/Slack으로 자동 전송
- **옵션**: `-periodic-report -report-interval=N` (N은 분 단위)

### 2. 시스템 다운/이상 신호 감지
- **기능**: 시스템 응답 중단, 과부하 상태 자동 감지
- **알림**: 즉시 이메일 + Slack 긴급 알림 전송

### 3. 하트비트 모니터링
- **기능**: 5분 간격으로 시스템 생존 신호 체크
- **감지**: 10분 이상 응답 없으면 다운으로 판단

### 4. 위험 상황 자동 감지
- **CPU 과부하**: 95% 이상
- **메모리 부족**: 98% 이상  
- **디스크 부족**: 98% 이상
- **시스템 로드**: CPU 코어 수 × 3배 이상

## 🔧 사용 방법

### 기본 설정
```bash
# 시스템 모니터링만 (알림 없음)
syslog-monitor -system-monitor

# 시스템 모니터링 + 이메일 알림
syslog-monitor -system-monitor -email-to="admin@company.com"

# 전체 기능 (AI 분석 + 시스템 모니터링)
syslog-monitor -ai-analysis -system-monitor
```

### 정기 보고서 설정
```bash
# 매 시간마다 시스템 상태 보고서 전송
syslog-monitor -system-monitor -periodic-report -report-interval=60

# 30분마다 보고서 전송
syslog-monitor -system-monitor -periodic-report -report-interval=30

# 5분마다 보고서 전송 (테스트용)
syslog-monitor -system-monitor -periodic-report -report-interval=5
```

### 다중 채널 알림 설정
```bash
# 이메일 + Slack 동시 알림
syslog-monitor -system-monitor -periodic-report -report-interval=60 \
  -email-to="admin@company.com,ops@company.com" \
  -slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -slack-channel="#alerts"
```

### 고급 모니터링 설정
```bash
# AI 분석 + 시스템 모니터링 + 정기 보고서 + 로그인 감시
export GEMINI_API_KEY="your-api-key"
syslog-monitor -ai-analysis -system-monitor -login-watch \
  -periodic-report -report-interval=30 \
  -email-to="security@company.com,admin@company.com"
```

## 📧 알림 예시

### 정기 시스템 상태 보고서
```
제목: [시스템 상태 보고서] beakerui-MacBookPro.local - 2025-01-31 00:04

🤖 AI 전문가 시스템 진단 보고서
================================
⏰ 진단 시간: 2025-01-31 00:04:48
🔍 진단 대상: beakerui-MacBookPro.local

🌐 네트워크 정보:
  - 호스트명: beakerui-MacBookPro.local
  - 사설 IP: 192.0.0.2
  - 공인 IP: 118.235.3.80

💻 CPU 정보:
  - 사용률: 29.1% (임계값: 80.0%)
  - 사용자: 12.2%, 시스템: 16.9%, 대기: 70.9%
  - 코어 수: 8개

🧠 메모리 정보:
  - 사용률: 93.8% (임계값: 85.0%)
  - 총 메모리: 16.0 GB
  - 사용 중: 15.0 GB
  - 사용 가능: 0.2 GB

... (상세한 진단 내용)
```

### 시스템 다운 감지 알림
```
제목: 🚨 시스템 다운 감지

🚨 시스템 다운 감지
=================
호스트: beakerui-MacBookPro.local
시간: 2025-01-31 00:15:30
상태: 시스템이 응답하지 않습니다

마지막 하트비트: 2025-01-31 00:05:30
경과 시간: 10m0s

즉시 시스템 상태를 확인해주세요!
```

### 위험 상황 감지 알림
```
제목: 🚨 CRITICAL_MEMORY

🚨 위험 상황 감지
=================
유형: CRITICAL_MEMORY
호스트: beakerui-MacBookPro.local
시간: 2025-01-31 00:10:15

메시지: 메모리 사용률이 위험 수준입니다: 98.5%

즉시 조치가 필요합니다!
```

### Slack 요약 메시지
```
📊 시스템 상태 보고서
🖥️  beakerui-MacBookPro.local
⏰ 2025-01-31 00:04:48

💻 CPU: 29.1% | 🧠 메모리: 93.8% | 🌡️ 온도: 45.0°C
⚖️ 로드: 2.67 | 🔄 프로세스: 784개

상세 정보는 이메일을 확인하세요.
```

## ⚙️ 환경변수 설정

```bash
# 이메일 설정
export SYSLOG_EMAIL_TO="admin@company.com,ops@company.com"
export SYSLOG_SMTP_USER="your-email@gmail.com"
export SYSLOG_SMTP_PASSWORD="your-app-password"

# Slack 설정
export SYSLOG_SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
export SYSLOG_SLACK_CHANNEL="#system-alerts"

# Gemini AI 설정
export GEMINI_API_KEY="your-gemini-api-key"

# 실행
syslog-monitor -system-monitor -periodic-report -report-interval=60
```

## 🔒 보안 고려사항

1. **Gmail 앱 패스워드**: 일반 비밀번호 대신 앱 전용 패스워드 사용
2. **Slack 웹훅**: 안전한 채널에만 알림 전송
3. **API 키**: 환경변수로 관리, Git에 커밋하지 않음
4. **권한 관리**: 시스템 모니터링을 위한 적절한 권한 설정

## 🎯 추천 설정

### 개발/테스트 환경
```bash
# 10분마다 간단한 보고서
syslog-monitor -system-monitor -periodic-report -report-interval=10
```

### 프로덕션 환경
```bash
# 1시간마다 정기 보고서 + AI 분석 + 로그인 감시
syslog-monitor -ai-analysis -system-monitor -login-watch \
  -periodic-report -report-interval=60
```

### 중요 서버
```bash
# 30분마다 보고서 + 다중 채널 알림
syslog-monitor -ai-analysis -system-monitor -login-watch \
  -periodic-report -report-interval=30 \
  -email-to="admin@company.com,security@company.com,ops@company.com" \
  -slack-webhook="YOUR_WEBHOOK" -slack-channel="#critical-alerts"
```

## 📊 모니터링 지표

### 자동 감지되는 위험 상황
1. **CPU 과부하**: 95% 이상 → 즉시 알림
2. **메모리 부족**: 98% 이상 → 즉시 알림
3. **디스크 부족**: 98% 이상 → 즉시 알림
4. **시스템 로드**: 코어수×3 이상 → 즉시 알림
5. **시스템 다운**: 10분 이상 무응답 → 즉시 알림

### 정기 보고서 포함 내용
- CPU, 메모리, 디스크 사용률
- 시스템 온도 및 로드
- 네트워크 정보 (IP, 호스트명)
- AI 전문가 진단 및 권장사항
- 즉시 실행 가능한 명령어
- 성능 최적화 팁

---

**🔧 설정 확인**: `syslog-monitor -show-config`  
**📚 전체 기능**: [FEATURES.md](./FEATURES.md)  
**📖 사용 가이드**: [README.md](./README.md)