/*
GeoMapper - IP 지리정보 및 지도 매핑 모듈
==========================================

IP 주소의 지리적 위치 정보를 수집하고 지도 데이터와 매핑하여
시각적 분석을 제공하는 모듈

주요 기능:
- IP 주소 지리정보 실시간 조회
- ASN 정보 및 조직 정보 수집
- 위험도 기반 색상 코딩
- 지도 좌표 변환 및 매핑
- 정기적 위치 보고서 생성

지원 API:
- ip-api.com (무료 IP 지리정보)
- ipinfo.io (상세 ASN 정보)
- Google Maps API (지도 시각화)
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GeoLocationInfo 지리적 위치 정보
type GeoLocationInfo struct {
	IP           string  `json:"ip"`           // IP 주소
	Country      string  `json:"country"`      // 국가
	Region       string  `json:"region"`       // 지역/주
	City         string  `json:"city"`         // 도시
	Latitude     float64 `json:"latitude"`     // 위도
	Longitude    float64 `json:"longitude"`    // 경도
	Organization string  `json:"organization"` // 소속 기관/ISP
	ASN          string  `json:"asn"`          // ASN 번호
	IsPrivate    bool    `json:"is_private"`   // 사설 IP 여부
	Threat       string  `json:"threat"`       // 위험도 평가
	Timezone     string  `json:"timezone"`     // 시간대
	ISP          string  `json:"isp"`          // 인터넷 서비스 제공업체
	LastSeen     time.Time `json:"last_seen"`  // 마지막 감지 시각
}

// MapMarker 지도 마커 정보
type MapMarker struct {
	Latitude   float64 `json:"lat"`      // 위도
	Longitude  float64 `json:"lng"`      // 경도
	Title      string  `json:"title"`    // 마커 제목
	Content    string  `json:"content"`  // 마커 내용
	Color      string  `json:"color"`    // 마커 색상
	Icon       string  `json:"icon"`     // 마커 아이콘
	Threat     string  `json:"threat"`   // 위험도
	LastSeen   string  `json:"last_seen"` // 마지막 감지 시각
}

// GeoMapper 지리정보 매핑 서비스
type GeoMapper struct {
	logger        Logger
	locationCache map[string]*GeoLocationInfo // 위치 정보 캐시
	cacheTimeout  time.Duration              // 캐시 만료 시간
	apiTimeout    time.Duration              // API 요청 타임아웃
}

// NewGeoMapper 새로운 지리정보 매핑 서비스 생성
func NewGeoMapper(logger Logger) *GeoMapper {
	return &GeoMapper{
		logger:        logger,
		locationCache: make(map[string]*GeoLocationInfo),
		cacheTimeout:  30 * time.Minute, // 30분 캐시
		apiTimeout:    10 * time.Second, // 10초 타임아웃
	}
}

// GetLocationInfo IP 주소의 지리정보 조회 (캐시 포함)
func (gm *GeoMapper) GetLocationInfo(ip string) *GeoLocationInfo {
	if ip == "" {
		return nil
	}

	// 사설 IP 체크
	if gm.isPrivateIP(ip) {
		return &GeoLocationInfo{
			IP:        ip,
			Country:   "Private Network",
			City:      "Local Network",
			IsPrivate: true,
			Threat:    "LOW",
			LastSeen:  time.Now(),
		}
	}

	// 캐시 확인
	if cached, exists := gm.locationCache[ip]; exists {
		if time.Since(cached.LastSeen) < gm.cacheTimeout {
			return cached
		}
		// 캐시 만료된 경우 삭제
		delete(gm.locationCache, ip)
	}

	// API로 지리정보 조회
	locationInfo := gm.fetchLocationFromAPI(ip)
	if locationInfo != nil {
		locationInfo.LastSeen = time.Now()
		gm.locationCache[ip] = locationInfo
	}

	return locationInfo
}

// fetchLocationFromAPI 외부 API로 지리정보 조회
func (gm *GeoMapper) fetchLocationFromAPI(ip string) *GeoLocationInfo {
	// ip-api.com 사용 (무료, 상세 정보 제공)
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,regionName,city,lat,lon,org,as,timezone,isp,query", ip)
	
	client := &http.Client{Timeout: gm.apiTimeout}
	resp, err := client.Get(url)
	if err != nil {
		gm.logger.Errorf("Failed to query IP location for %s: %v", ip, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		gm.logger.Errorf("Failed to read IP location response: %v", err)
		return nil
	}

	var result struct {
		Status     string  `json:"status"`
		Country    string  `json:"country"`
		RegionName string  `json:"regionName"`
		City       string  `json:"city"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		Org        string  `json:"org"`
		AS         string  `json:"as"`
		Timezone   string  `json:"timezone"`
		ISP        string  `json:"isp"`
		Query      string  `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		gm.logger.Errorf("Failed to parse IP location response: %v", err)
		return nil
	}

	if result.Status == "success" {
		locationInfo := &GeoLocationInfo{
			IP:           ip,
			Country:      result.Country,
			Region:       result.RegionName,
			City:         result.City,
			Latitude:     result.Lat,
			Longitude:    result.Lon,
			Organization: result.Org,
			ASN:          result.AS,
			Timezone:     result.Timezone,
			ISP:          result.ISP,
			IsPrivate:    false,
			Threat:       gm.assessThreatLevel(result.Country, result.Org),
		}
		return locationInfo
	}

	return nil
}

// isPrivateIP IP 주소가 사설 IP인지 확인
func (gm *GeoMapper) isPrivateIP(ipStr string) bool {
	// 간단한 사설 IP 체크 (더 정확한 체크는 net 패키지 사용)
	privateRanges := []string{
		"10.", "172.16.", "172.17.", "172.18.", "172.19.", "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.", "172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.", "127.", "169.254.",
	}

	for _, rangePrefix := range privateRanges {
		if strings.HasPrefix(ipStr, rangePrefix) {
			return true
		}
	}
	return false
}

// assessThreatLevel 국가와 조직 정보를 바탕으로 위험도 평가
func (gm *GeoMapper) assessThreatLevel(country, org string) string {
	// 한국 내부 IP는 LOW
	if country == "South Korea" || country == "Korea" {
		return "LOW"
	}

	// 알려진 클라우드 서비스는 MEDIUM
	cloudProviders := []string{"Amazon", "Google", "Microsoft", "Azure", "AWS", "Cloudflare"}
	orgLower := strings.ToLower(org)
	for _, provider := range cloudProviders {
		if strings.Contains(orgLower, strings.ToLower(provider)) {
			return "MEDIUM"
		}
	}

	// 일반적으로 의심스러운 국가들
	suspiciousCountries := []string{"China", "Russia", "North Korea", "Iran"}
	for _, suspicious := range suspiciousCountries {
		if country == suspicious {
			return "HIGH"
		}
	}

	// 기본적으로 해외 IP는 MEDIUM
	return "MEDIUM"
}

// CreateMapMarker 지도 마커 생성
func (gm *GeoMapper) CreateMapMarker(location *GeoLocationInfo) *MapMarker {
	if location == nil || location.IsPrivate {
		return nil
	}

	// 위험도에 따른 색상 및 아이콘 설정
	var color, icon string
	switch location.Threat {
	case "HIGH":
		color = "red"
		icon = "🔴"
	case "MEDIUM":
		color = "orange"
		icon = "🟡"
	case "LOW":
		color = "green"
		icon = "🟢"
	default:
		color = "gray"
		icon = "⚪"
	}

	content := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif;">
			<h3 style="margin: 0 0 10px 0; color: #333;">%s</h3>
			<p><strong>IP:</strong> %s</p>
			<p><strong>위치:</strong> %s, %s, %s</p>
			<p><strong>조직:</strong> %s</p>
			<p><strong>ASN:</strong> %s</p>
			<p><strong>ISP:</strong> %s</p>
			<p><strong>위험도:</strong> <span style="color: %s;">%s</span></p>
			<p><strong>마지막 감지:</strong> %s</p>
		</div>
	`, icon, location.IP, location.City, location.Region, location.Country, 
		location.Organization, location.ASN, location.ISP, color, location.Threat,
		location.LastSeen.Format("2006-01-02 15:04:05"))

	return &MapMarker{
		Latitude:   location.Latitude,
		Longitude:  location.Longitude,
		Title:      fmt.Sprintf("%s %s", icon, location.IP),
		Content:    content,
		Color:      color,
		Icon:       icon,
		Threat:     location.Threat,
		LastSeen:   location.LastSeen.Format("2006-01-02 15:04:05"),
	}
}

// GenerateMapHTML 지도 HTML 생성
func (gm *GeoMapper) GenerateMapHTML(markers []*MapMarker) string {
	if len(markers) == 0 {
		return "<p>지도 데이터가 없습니다.</p>"
	}

	// Google Maps API를 사용한 지도 HTML 생성
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>IP 위치 지도</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
			#map { height: 500px; width: 100%; border-radius: 8px; }
			.legend { margin-top: 20px; padding: 15px; background: #f5f5f5; border-radius: 5px; }
			.legend-item { display: inline-block; margin-right: 20px; }
		</style>
	</head>
	<body>
		<h1>🌍 IP 위치 지도</h1>
		<div id="map"></div>
		<div class="legend">
			<div class="legend-item">🟢 낮은 위험도</div>
			<div class="legend-item">🟡 중간 위험도</div>
			<div class="legend-item">🔴 높은 위험도</div>
		</div>
		<script>
			function initMap() {
				const map = new google.maps.Map(document.getElementById('map'), {
					zoom: 2,
					center: { lat: 0, lng: 0 }
				});

				const markers = ` + gm.markersToJSON(markers) + `;

				markers.forEach(markerData => {
					const marker = new google.maps.Marker({
						position: { lat: markerData.lat, lng: markerData.lng },
						map: map,
						title: markerData.title,
						icon: {
							url: 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(markerData.icon),
							scaledSize: new google.maps.Size(30, 30)
						}
					});

					const infowindow = new google.maps.InfoWindow({
						content: markerData.content
					});

					marker.addListener('click', () => {
						infowindow.open(map, marker);
					});
				});
			}
		</script>
		<script async defer
			src="https://maps.googleapis.com/maps/api/js?key=YOUR_API_KEY&callback=initMap">
		</script>
	</body>
	</html>`

	return html
}

// markersToJSON 마커 데이터를 JSON으로 변환
func (gm *GeoMapper) markersToJSON(markers []*MapMarker) string {
	jsonData, _ := json.Marshal(markers)
	return string(jsonData)
}

// GetCurrentSystemIP 현재 시스템의 공인 IP 조회
func (gm *GeoMapper) GetCurrentSystemIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		gm.logger.Errorf("Failed to get current system IP: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		gm.logger.Errorf("Failed to read IP response: %v", err)
		return ""
	}

	return strings.TrimSpace(string(body))
}

// GenerateLocationReport 위치 정보 보고서 생성
func (gm *GeoMapper) GenerateLocationReport() string {
	currentIP := gm.GetCurrentSystemIP()
	if currentIP == "" {
		return "현재 시스템 IP를 조회할 수 없습니다."
	}

	location := gm.GetLocationInfo(currentIP)
	if location == nil {
		return "위치 정보를 조회할 수 없습니다."
	}

	report := fmt.Sprintf(`
🌍 시스템 위치 정보 보고서
==============================

📍 현재 시스템 IP: %s
🏴 국가: %s
🏙️  도시: %s, %s
🌐 위도/경도: %.6f, %.6f
🏢 조직: %s
🔢 ASN: %s
🌐 ISP: %s
⏰ 시간대: %s
⚠️  위험도: %s
🕐 조회 시각: %s

📊 캐시된 위치 정보: %d개
`, location.IP, location.Country, location.City, location.Region,
		location.Latitude, location.Longitude, location.Organization,
		location.ASN, location.ISP, location.Timezone, location.Threat,
		location.LastSeen.Format("2006-01-02 15:04:05"),
		len(gm.locationCache))

	return report
} 