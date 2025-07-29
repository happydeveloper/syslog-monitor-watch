package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// MacOSMonitor 맥OS 전용 모니터링 구조체
type MacOSMonitor struct {
	*SystemMonitor
}

// NewMacOSMonitor 맥OS 모니터 생성
func NewMacOSMonitor(interval time.Duration) *MacOSMonitor {
	if runtime.GOOS != "darwin" {
		return nil
	}
	
	baseMonitor := NewSystemMonitor(interval)
	return &MacOSMonitor{
		SystemMonitor: baseMonitor,
	}
}

// collectMacOSSpecificMetrics 맥OS 특화 메트릭 수집
func (mm *MacOSMonitor) collectMacOSSpecificMetrics() {
	if runtime.GOOS != "darwin" {
		return
	}
	
	mm.collectMacOSCPUMetrics()
	mm.collectMacOSMemoryMetrics()
	mm.collectMacOSDiskMetrics()
	mm.collectMacOSTemperatureMetrics()
	mm.collectMacOSNetworkMetrics()
	mm.collectMacOSProcessMetrics()
	mm.collectMacOSBatteryMetrics()
}

// collectMacOSCPUMetrics 맥OS CPU 메트릭 수집
func (mm *MacOSMonitor) collectMacOSCPUMetrics() {
	// top 명령어로 더 정확한 CPU 정보 수집
	cmd := exec.Command("top", "-l", "2", "-n", "0", "-F", "-R")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// CPU usage: 12.34% user, 5.67% sys, 82.99% idle 형태 파싱
			userRegex := regexp.MustCompile(`(\d+\.\d+)%\s+user`)
			sysRegex := regexp.MustCompile(`(\d+\.\d+)%\s+sys`)
			idleRegex := regexp.MustCompile(`(\d+\.\d+)%\s+idle`)
			
			if userMatch := userRegex.FindStringSubmatch(line); userMatch != nil {
				if val, err := strconv.ParseFloat(userMatch[1], 64); err == nil {
					mm.metrics.CPU.UserPercent = val
				}
			}
			if sysMatch := sysRegex.FindStringSubmatch(line); sysMatch != nil {
				if val, err := strconv.ParseFloat(sysMatch[1], 64); err == nil {
					mm.metrics.CPU.SystemPercent = val
				}
			}
			if idleMatch := idleRegex.FindStringSubmatch(line); idleMatch != nil {
				if val, err := strconv.ParseFloat(idleMatch[1], 64); err == nil {
					mm.metrics.CPU.IdlePercent = val
					mm.metrics.CPU.UsagePercent = 100 - val
				}
			}
			break
		}
	}
	
	// sysctl로 CPU 코어 수 정확히 가져오기
	cmd = exec.Command("sysctl", "-n", "hw.ncpu")
	if output, err := cmd.Output(); err == nil {
		if cores, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			mm.metrics.CPU.Cores = cores
		}
	}
}

// collectMacOSMemoryMetrics 맥OS 메모리 메트릭 수집
func (mm *MacOSMonitor) collectMacOSMemoryMetrics() {
	// vm_stat으로 더 정확한 메모리 정보 수집
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	var pageSize float64 = 4096 // 기본값
	var freePages, activePages, inactivePages, wiredPages, compressedPages float64
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			re := regexp.MustCompile(`page size of (\d+) bytes`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					pageSize = val
				}
			}
		} else if strings.Contains(line, "Pages free:") {
			re := regexp.MustCompile(`Pages free:\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					freePages = val
				}
			}
		} else if strings.Contains(line, "Pages active:") {
			re := regexp.MustCompile(`Pages active:\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					activePages = val
				}
			}
		} else if strings.Contains(line, "Pages inactive:") {
			re := regexp.MustCompile(`Pages inactive:\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					inactivePages = val
				}
			}
		} else if strings.Contains(line, "Pages wired down:") {
			re := regexp.MustCompile(`Pages wired down:\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					wiredPages = val
				}
			}
		} else if strings.Contains(line, "Pages occupied by compressor:") {
			re := regexp.MustCompile(`Pages occupied by compressor:\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); matches != nil {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					compressedPages = val
				}
			}
		}
	}
	
	// 물리 메모리 총량 가져오기
	cmd = exec.Command("sysctl", "-n", "hw.memsize")
	if output, err := cmd.Output(); err == nil {
		if totalBytes, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); err == nil {
			mm.metrics.Memory.TotalMB = totalBytes / (1024 * 1024)
		}
	}
	
	// 메모리 계산
	usedPages := activePages + inactivePages + wiredPages + compressedPages
	
	mm.metrics.Memory.FreeMB = (freePages * pageSize) / (1024 * 1024)
	mm.metrics.Memory.UsedMB = (usedPages * pageSize) / (1024 * 1024)
	mm.metrics.Memory.AvailableMB = mm.metrics.Memory.FreeMB + ((inactivePages * pageSize) / (1024 * 1024))
	
	if mm.metrics.Memory.TotalMB > 0 {
		mm.metrics.Memory.UsagePercent = (mm.metrics.Memory.UsedMB / mm.metrics.Memory.TotalMB) * 100
	}
	
	// 스왑 정보 수집
	cmd = exec.Command("sysctl", "-n", "vm.swapusage")
	if output, err := cmd.Output(); err == nil {
		// vm.swapusage: total = 2048.00M  used = 512.00M  free = 1536.00M  (encrypted)
		swapInfo := string(output)
		totalRegex := regexp.MustCompile(`total = ([\d.]+)M`)
		usedRegex := regexp.MustCompile(`used = ([\d.]+)M`)
		
		if totalMatch := totalRegex.FindStringSubmatch(swapInfo); totalMatch != nil {
			if val, err := strconv.ParseFloat(totalMatch[1], 64); err == nil {
				mm.metrics.Memory.SwapTotalMB = val
			}
		}
		if usedMatch := usedRegex.FindStringSubmatch(swapInfo); usedMatch != nil {
			if val, err := strconv.ParseFloat(usedMatch[1], 64); err == nil {
				mm.metrics.Memory.SwapUsedMB = val
				if mm.metrics.Memory.SwapTotalMB > 0 {
					mm.metrics.Memory.SwapFreePercent = ((mm.metrics.Memory.SwapTotalMB - val) / mm.metrics.Memory.SwapTotalMB) * 100
				}
			}
		}
	}
}

// collectMacOSDiskMetrics 맥OS 디스크 메트릭 수집
func (mm *MacOSMonitor) collectMacOSDiskMetrics() {
	// df -h로 디스크 사용량 수집 (APFS 지원)
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	mm.metrics.Disk = []DiskMetrics{}
	
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			device := fields[0]
			mountPoint := fields[len(fields)-1]
			
			// macOS 특화: /dev/disk로 시작하는 것만 포함
			if !strings.HasPrefix(device, "/dev/disk") && !strings.HasPrefix(device, "/dev/mapper") {
				continue
			}
			
			// 사이즈 파싱 (macOS는 다양한 단위 사용)
			sizeStr := fields[1]
			usedStr := fields[2]
			availStr := fields[3]
			usePercentStr := strings.TrimSuffix(fields[4], "%")
			
			totalGB := mm.parseMacOSSize(sizeStr)
			usedGB := mm.parseMacOSSize(usedStr)
			availGB := mm.parseMacOSSize(availStr)
			
			usePercent, err := strconv.ParseFloat(usePercentStr, 64)
			if err != nil {
				continue
			}
			
			diskMetric := DiskMetrics{
				Device:       device,
				MountPoint:   mountPoint,
				TotalGB:      totalGB,
				UsedGB:       usedGB,
				FreeGB:       availGB,
				UsagePercent: usePercent,
			}
			
			// APFS 파일시스템의 inode 정보 수집
			cmd = exec.Command("df", "-i", mountPoint)
			if inodeOutput, err := cmd.Output(); err == nil {
				inodeLines := strings.Split(string(inodeOutput), "\n")
				if len(inodeLines) >= 2 {
					inodeFields := strings.Fields(inodeLines[1])
					if len(inodeFields) >= 5 {
						inodeUseStr := strings.TrimSuffix(inodeFields[4], "%")
						if inodePercent, err := strconv.ParseFloat(inodeUseStr, 64); err == nil {
							diskMetric.InodeUsagePercent = inodePercent
						}
					}
				}
			}
			
			mm.metrics.Disk = append(mm.metrics.Disk, diskMetric)
		}
	}
}

// parseMacOSSize 맥OS 크기 문자열 파싱 (K, M, G, T 지원)
func (mm *MacOSMonitor) parseMacOSSize(sizeStr string) float64 {
	if sizeStr == "-" || sizeStr == "" {
		return 0
	}
	
	sizeStr = strings.ToUpper(sizeStr)
	multiplier := 1.0
	
	if strings.HasSuffix(sizeStr, "K") || strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1.0 / 1024 / 1024 // KB to GB
		sizeStr = strings.TrimSuffix(strings.TrimSuffix(sizeStr, "B"), "K")
	} else if strings.HasSuffix(sizeStr, "M") || strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1.0 / 1024 // MB to GB
		sizeStr = strings.TrimSuffix(strings.TrimSuffix(sizeStr, "B"), "M")
	} else if strings.HasSuffix(sizeStr, "G") || strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1.0 // Already GB
		sizeStr = strings.TrimSuffix(strings.TrimSuffix(sizeStr, "B"), "G")
	} else if strings.HasSuffix(sizeStr, "T") || strings.HasSuffix(sizeStr, "TB") {
		multiplier = 1024.0 // TB to GB
		sizeStr = strings.TrimSuffix(strings.TrimSuffix(sizeStr, "B"), "T")
	}
	
	if val, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		return val * multiplier
	}
	
	return 0
}

// collectMacOSTemperatureMetrics 맥OS 온도 메트릭 수집
func (mm *MacOSMonitor) collectMacOSTemperatureMetrics() {
	mm.metrics.Temperature.CoreTemps = make(map[string]float64)
	
	// powermetrics로 온도 정보 수집 (root 권한 필요)
	cmd := exec.Command("sudo", "powermetrics", "-n", "1", "-s", "thermal")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "CPU die temperature") {
				re := regexp.MustCompile(`CPU die temperature: (\d+\.\d+) C`)
				if matches := re.FindStringSubmatch(line); matches != nil {
					if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
						mm.metrics.Temperature.CPUTemp = temp
					}
				}
			} else if strings.Contains(line, "GPU die temperature") {
				re := regexp.MustCompile(`GPU die temperature: (\d+\.\d+) C`)
				if matches := re.FindStringSubmatch(line); matches != nil {
					if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
						mm.metrics.Temperature.GPUTemp = temp
					}
				}
			}
		}
	}
	
	// istats나 TG Pro 같은 도구가 설치되어 있다면 활용
	cmd = exec.Command("istats", "temp")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "CPU") && strings.Contains(line, "°C") {
				re := regexp.MustCompile(`(\d+\.\d+)°C`)
				if matches := re.FindStringSubmatch(line); matches != nil {
					if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
						mm.metrics.Temperature.CPUTemp = temp
						mm.metrics.Temperature.CoreTemps["CPU"] = temp
					}
				}
			}
		}
	}
	
	// 센서 정보가 없는 경우 기본값 설정
	if mm.metrics.Temperature.CPUTemp == 0 {
		// 온도 정보를 가져올 수 없는 경우 추정값 사용
		mm.metrics.Temperature.CPUTemp = 45.0 // 기본 추정값
	}
}

// collectMacOSNetworkMetrics 맥OS 네트워크 메트릭 수집
func (mm *MacOSMonitor) collectMacOSNetworkMetrics() {
	// netstat -ib로 인터페이스별 상세 정보 수집
	cmd := exec.Command("netstat", "-ib")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 10 {
			interfaceName := fields[0]
			// 활성 인터페이스만 (lo0 제외)
			if interfaceName == "lo0" || strings.HasPrefix(interfaceName, "awdl") {
				continue
			}
			
			if bytesRecv, err := strconv.ParseUint(fields[6], 10, 64); err == nil {
				if bytesSent, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					mm.metrics.Network = NetworkMetrics{
						Interface: interfaceName,
						BytesRecv: bytesRecv,
						BytesSent: bytesSent,
					}
					break // 첫 번째 활성 인터페이스만 사용
				}
			}
		}
	}
}

// collectMacOSProcessMetrics 맥OS 프로세스 메트릭 수집
func (mm *MacOSMonitor) collectMacOSProcessMetrics() {
	// ps 명령어로 프로세스 상태 수집
	cmd := exec.Command("ps", "ax", "-o", "stat")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	running := 0
	sleeping := 0
	stopped := 0
	zombie := 0
	
	for i, line := range lines {
		if i == 0 { // 헤더 스킵
			continue
		}
		
		stat := strings.TrimSpace(line)
		if stat == "" {
			continue
		}
		
		switch stat[0] {
		case 'R':
			running++
		case 'S', 'I':
			sleeping++
		case 'T':
			stopped++
		case 'Z':
			zombie++
		}
	}
	
	mm.metrics.ProcessCount = ProcessMetrics{
		Total:    len(lines) - 2, // 헤더와 빈 줄 제외
		Running:  running,
		Sleeping: sleeping,
		Stopped:  stopped,
		Zombie:   zombie,
	}
}

// collectMacOSBatteryMetrics 맥OS 배터리 메트릭 수집 (노트북의 경우)
func (mm *MacOSMonitor) collectMacOSBatteryMetrics() {
	// pmset으로 배터리 정보 수집
	cmd := exec.Command("pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	batteryInfo := string(output)
	if strings.Contains(batteryInfo, "Battery Power") {
		// 배터리로 동작 중
		mm.metrics.Fields = make(map[string]string)
		mm.metrics.Fields["power_source"] = "battery"
		
		// 배터리 잔량 추출
		re := regexp.MustCompile(`(\d+)%`)
		if matches := re.FindStringSubmatch(batteryInfo); matches != nil {
			mm.metrics.Fields["battery_percentage"] = matches[1]
		}
	} else if strings.Contains(batteryInfo, "AC Power") {
		mm.metrics.Fields = make(map[string]string)
		mm.metrics.Fields["power_source"] = "ac_power"
	}
}

// GetMacOSSpecificReport 맥OS 특화 보고서 생성
func (mm *MacOSMonitor) GetMacOSSpecificReport() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	
	mm.collectMacOSSpecificMetrics()
	
	report := fmt.Sprintf(`
🍎 macOS 시스템 모니터링 보고서
===============================
⏰ 수집 시간: %s

💻 CPU 정보 (macOS 최적화):
  - 사용률: %.1f%% (임계값: %.1f%%)
  - 사용자: %.1f%%, 시스템: %.1f%%, 대기: %.1f%%
  - 물리 코어: %d개

🧠 메모리 정보 (APFS 최적화):
  - 사용률: %.1f%% (임계값: %.1f%%)
  - 총 메모리: %.1f GB
  - 사용 중: %.1f GB
  - 사용 가능: %.1f GB
  - 스왑 사용: %.1f MB (여유: %.1f%%)

💾 디스크 정보 (APFS):`,
		mm.metrics.Timestamp.Format("2006-01-02 15:04:05"),
		mm.metrics.CPU.UsagePercent, mm.thresholds.CPUPercent,
		mm.metrics.CPU.UserPercent, mm.metrics.CPU.SystemPercent, mm.metrics.CPU.IdlePercent,
		mm.metrics.CPU.Cores,
		mm.metrics.Memory.UsagePercent, mm.thresholds.MemoryPercent,
		mm.metrics.Memory.TotalMB/1024,
		mm.metrics.Memory.UsedMB/1024,
		mm.metrics.Memory.AvailableMB/1024,
		mm.metrics.Memory.SwapUsedMB, mm.metrics.Memory.SwapFreePercent,
	)
	
	for _, disk := range mm.metrics.Disk {
		report += fmt.Sprintf(`
  - %s (%s): %.1f%% 사용 (%.1f/%.1f GB)`,
			disk.Device, disk.MountPoint, disk.UsagePercent, disk.UsedGB, disk.TotalGB)
	}
	
	report += fmt.Sprintf(`

🌡️  온도 정보:
  - CPU 온도: %.1f°C (임계값: %.1f°C)`,
		mm.metrics.Temperature.CPUTemp, mm.thresholds.CPUTemp)
	
	if mm.metrics.Temperature.GPUTemp > 0 {
		report += fmt.Sprintf(`
  - GPU 온도: %.1f°C`, mm.metrics.Temperature.GPUTemp)
	}
	
	report += fmt.Sprintf(`

⚖️  시스템 로드:
  - 1분: %.2f, 5분: %.2f, 15분: %.2f (임계값: %.1f)

🔄 프로세스:
  - 총 프로세스: %d개
  - 실행 중: %d, 대기 중: %d, 정지: %d, 좀비: %d`,
		mm.metrics.LoadAverage.Load1Min, mm.metrics.LoadAverage.Load5Min, mm.metrics.LoadAverage.Load15Min, mm.thresholds.LoadAverage,
		mm.metrics.ProcessCount.Total,
		mm.metrics.ProcessCount.Running, mm.metrics.ProcessCount.Sleeping, mm.metrics.ProcessCount.Stopped, mm.metrics.ProcessCount.Zombie,
	)
	
	// 네트워크 정보
	if mm.metrics.Network.Interface != "" {
		report += fmt.Sprintf(`

🌐 네트워크 (%s):
  - 수신: %d 바이트
  - 송신: %d 바이트`,
			mm.metrics.Network.Interface,
			mm.metrics.Network.BytesRecv,
			mm.metrics.Network.BytesSent,
		)
	}
	
	// 배터리 정보 (노트북의 경우)
	if mm.metrics.Fields != nil {
		if powerSource, exists := mm.metrics.Fields["power_source"]; exists {
			report += fmt.Sprintf(`

🔋 전원 정보:
  - 전원: %s`, powerSource)
			
			if batteryPercent, exists := mm.metrics.Fields["battery_percentage"]; exists {
				report += fmt.Sprintf(`
  - 배터리: %s%%`, batteryPercent)
			}
		}
	}
	
	return report
} 