package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemMonitor 시스템 메트릭 모니터링 구조체
type SystemMonitor struct {
	interval       time.Duration
	alertChannel   chan SystemAlert
	metrics        *SystemMetrics
	thresholds     SystemThresholds
	history        []SystemMetrics
	maxHistorySize int
}

// SystemMetrics 시스템 메트릭 구조체
type SystemMetrics struct {
	Timestamp    time.Time            `json:"timestamp"`
	CPU          CPUMetrics           `json:"cpu"`
	Memory       MemoryMetrics        `json:"memory"`
	Disk         []DiskMetrics        `json:"disk"`
	Network      NetworkMetrics       `json:"network"`
	Temperature  TempMetrics          `json:"temperature"`
	LoadAverage  LoadMetrics          `json:"load_average"`
	ProcessCount ProcessMetrics       `json:"processes"`
	Fields       map[string]string    `json:"fields,omitempty"` // macOS 배터리 정보 등 추가 필드
}

// CPUMetrics CPU 관련 메트릭
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UserPercent  float64 `json:"user_percent"`
	SystemPercent float64 `json:"system_percent"`
	IdlePercent  float64 `json:"idle_percent"`
	IOWaitPercent float64 `json:"iowait_percent"`
	Cores        int     `json:"cores"`
}

// MemoryMetrics 메모리 관련 메트릭
type MemoryMetrics struct {
	TotalMB      float64 `json:"total_mb"`
	UsedMB       float64 `json:"used_mb"`
	FreeMB       float64 `json:"free_mb"`
	AvailableMB  float64 `json:"available_mb"`
	UsagePercent float64 `json:"usage_percent"`
	SwapTotalMB  float64 `json:"swap_total_mb"`
	SwapUsedMB   float64 `json:"swap_used_mb"`
	SwapFreePercent float64 `json:"swap_free_percent"`
}

// DiskMetrics 디스크 관련 메트릭
type DiskMetrics struct {
	Device       string  `json:"device"`
	MountPoint   string  `json:"mount_point"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	UsagePercent float64 `json:"usage_percent"`
	InodeUsagePercent float64 `json:"inode_usage_percent"`
}

// NetworkMetrics 네트워크 관련 메트릭
type NetworkMetrics struct {
	Interface    string  `json:"interface"`
	BytesRecv    uint64  `json:"bytes_recv"`
	BytesSent    uint64  `json:"bytes_sent"`
	PacketsRecv  uint64  `json:"packets_recv"`
	PacketsSent  uint64  `json:"packets_sent"`
	ErrorsRecv   uint64  `json:"errors_recv"`
	ErrorsSent   uint64  `json:"errors_sent"`
	DroppedRecv  uint64  `json:"dropped_recv"`
	DroppedSent  uint64  `json:"dropped_sent"`
}

// TempMetrics 온도 관련 메트릭
type TempMetrics struct {
	CPUTemp     float64            `json:"cpu_temp"`
	CoreTemps   map[string]float64 `json:"core_temps"`
	GPUTemp     float64            `json:"gpu_temp"`
	MotherboardTemp float64        `json:"motherboard_temp"`
}

// LoadMetrics 로드 평균 메트릭
type LoadMetrics struct {
	Load1Min   float64 `json:"load_1min"`
	Load5Min   float64 `json:"load_5min"`
	Load15Min  float64 `json:"load_15min"`
}

// ProcessMetrics 프로세스 관련 메트릭
type ProcessMetrics struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
	Stopped  int `json:"stopped"`
	Zombie   int `json:"zombie"`
}

// SystemThresholds 알림 임계값
type SystemThresholds struct {
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryPercent    float64 `json:"memory_percent"`
	DiskPercent      float64 `json:"disk_percent"`
	CPUTemp          float64 `json:"cpu_temp"`
	LoadAverage      float64 `json:"load_average"`
	SwapPercent      float64 `json:"swap_percent"`
	InodePercent     float64 `json:"inode_percent"`
}

// SystemAlert 시스템 알림 구조체
type SystemAlert struct {
	Level       string             `json:"level"`
	Type        string             `json:"type"`
	Message     string             `json:"message"`
	Value       float64            `json:"value"`
	Threshold   float64            `json:"threshold"`
	Metrics     SystemMetrics      `json:"metrics"`
	Timestamp   time.Time          `json:"timestamp"`
	Suggestions []string           `json:"suggestions"`
}

// NewSystemMonitor 시스템 모니터 생성
func NewSystemMonitor(interval time.Duration) *SystemMonitor {
	return &SystemMonitor{
		interval:       interval,
		alertChannel:   make(chan SystemAlert, 100),
		metrics:        &SystemMetrics{},
		history:        make([]SystemMetrics, 0),
		maxHistorySize: 288, // 24시간 분량 (5분 간격)
		thresholds: SystemThresholds{
			CPUPercent:    80.0,
			MemoryPercent: 85.0,
			DiskPercent:   90.0,
			CPUTemp:       75.0,
			LoadAverage:   float64(runtime.NumCPU()) * 2.0,
			SwapPercent:   50.0,
			InodePercent:  90.0,
		},
	}
}

// Start 시스템 모니터링 시작
func (sm *SystemMonitor) Start() {
	ticker := time.NewTicker(sm.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				sm.collectMetrics()
				sm.checkAlerts()
				sm.updateHistory()
			}
		}
	}()
}

// collectMetrics 시스템 메트릭 수집
func (sm *SystemMonitor) collectMetrics() {
	sm.metrics = &SystemMetrics{
		Timestamp: time.Now(),
	}

	// 각 메트릭 수집
	sm.collectCPUMetrics()
	sm.collectMemoryMetrics()
	sm.collectDiskMetrics()
	sm.collectNetworkMetrics()
	sm.collectTemperatureMetrics()
	sm.collectLoadMetrics()
	sm.collectProcessMetrics()
}

// collectCPUMetrics CPU 메트릭 수집
func (sm *SystemMonitor) collectCPUMetrics() {
	sm.metrics.CPU.Cores = runtime.NumCPU()

	if runtime.GOOS == "linux" {
		// /proc/stat 파일에서 CPU 사용률 계산
		data, err := ioutil.ReadFile("/proc/stat")
		if err != nil {
			return
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu ") {
				fields := strings.Fields(line)
				if len(fields) >= 8 {
					user, _ := strconv.ParseFloat(fields[1], 64)
					nice, _ := strconv.ParseFloat(fields[2], 64)
					system, _ := strconv.ParseFloat(fields[3], 64)
					idle, _ := strconv.ParseFloat(fields[4], 64)
					iowait, _ := strconv.ParseFloat(fields[5], 64)

					total := user + nice + system + idle + iowait
					if total > 0 {
						sm.metrics.CPU.UserPercent = (user / total) * 100
						sm.metrics.CPU.SystemPercent = (system / total) * 100
						sm.metrics.CPU.IdlePercent = (idle / total) * 100
						sm.metrics.CPU.IOWaitPercent = (iowait / total) * 100
						sm.metrics.CPU.UsagePercent = 100 - sm.metrics.CPU.IdlePercent
					}
				}
				break
			}
		}
	} else {
		// macOS/기타 OS용 top 명령어 사용
		cmd := exec.Command("top", "-l", "1", "-n", "0")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "CPU usage:") {
					// CPU usage: 12.5% user, 6.25% sys, 81.25% idle 형태 파싱
					parts := strings.Split(line, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if strings.Contains(part, "% user") {
							userStr := strings.Fields(part)[0]
							if val, err := strconv.ParseFloat(strings.TrimSuffix(userStr, "%"), 64); err == nil {
								sm.metrics.CPU.UserPercent = val
							}
						} else if strings.Contains(part, "% sys") {
							sysStr := strings.Fields(part)[0]
							if val, err := strconv.ParseFloat(strings.TrimSuffix(sysStr, "%"), 64); err == nil {
								sm.metrics.CPU.SystemPercent = val
							}
						} else if strings.Contains(part, "% idle") {
							idleStr := strings.Fields(part)[0]
							if val, err := strconv.ParseFloat(strings.TrimSuffix(idleStr, "%"), 64); err == nil {
								sm.metrics.CPU.IdlePercent = val
								sm.metrics.CPU.UsagePercent = 100 - val
							}
						}
					}
					break
				}
			}
		}
	}
}

// collectMemoryMetrics 메모리 메트릭 수집
func (sm *SystemMonitor) collectMemoryMetrics() {
	if runtime.GOOS == "linux" {
		// /proc/meminfo 파일에서 메모리 정보 수집
		data, err := ioutil.ReadFile("/proc/meminfo")
		if err != nil {
			return
		}

		memInfo := make(map[string]float64)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					key := strings.TrimSuffix(parts[0], ":")
					if val, err := strconv.ParseFloat(parts[1], 64); err == nil {
						memInfo[key] = val / 1024 // KB to MB
					}
				}
			}
		}

		sm.metrics.Memory.TotalMB = memInfo["MemTotal"]
		sm.metrics.Memory.FreeMB = memInfo["MemFree"]
		sm.metrics.Memory.AvailableMB = memInfo["MemAvailable"]
		sm.metrics.Memory.UsedMB = sm.metrics.Memory.TotalMB - sm.metrics.Memory.FreeMB
		sm.metrics.Memory.SwapTotalMB = memInfo["SwapTotal"]
		sm.metrics.Memory.SwapUsedMB = sm.metrics.Memory.SwapTotalMB - memInfo["SwapFree"]

		if sm.metrics.Memory.TotalMB > 0 {
			sm.metrics.Memory.UsagePercent = (sm.metrics.Memory.UsedMB / sm.metrics.Memory.TotalMB) * 100
		}
		if sm.metrics.Memory.SwapTotalMB > 0 {
			sm.metrics.Memory.SwapFreePercent = (memInfo["SwapFree"] / sm.metrics.Memory.SwapTotalMB) * 100
		}
	} else {
		// macOS용 vm_stat 명령어 사용
		cmd := exec.Command("vm_stat")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			var pageSize float64 = 4096 // 기본 페이지 크기
			var freePages, activePages, inactivePages, wiredPages float64

			for _, line := range lines {
				if strings.Contains(line, "page size of") {
					parts := strings.Fields(line)
					if len(parts) >= 8 {
						if val, err := strconv.ParseFloat(parts[7], 64); err == nil {
							pageSize = val
						}
					}
				} else if strings.Contains(line, "Pages free:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "."), 64); err == nil {
							freePages = val
						}
					}
				} else if strings.Contains(line, "Pages active:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "."), 64); err == nil {
							activePages = val
						}
					}
				} else if strings.Contains(line, "Pages inactive:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[2], "."), 64); err == nil {
							inactivePages = val
						}
					}
				} else if strings.Contains(line, "Pages wired down:") {
					parts := strings.Fields(line)
					if len(parts) >= 4 {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[3], "."), 64); err == nil {
							wiredPages = val
						}
					}
				}
			}

			totalPages := freePages + activePages + inactivePages + wiredPages
			usedPages := activePages + inactivePages + wiredPages

			sm.metrics.Memory.TotalMB = (totalPages * pageSize) / (1024 * 1024)
			sm.metrics.Memory.FreeMB = (freePages * pageSize) / (1024 * 1024)
			sm.metrics.Memory.UsedMB = (usedPages * pageSize) / (1024 * 1024)
			sm.metrics.Memory.AvailableMB = sm.metrics.Memory.FreeMB

			if sm.metrics.Memory.TotalMB > 0 {
				sm.metrics.Memory.UsagePercent = (sm.metrics.Memory.UsedMB / sm.metrics.Memory.TotalMB) * 100
			}
		}
	}
}

// collectDiskMetrics 디스크 메트릭 수집
func (sm *SystemMonitor) collectDiskMetrics() {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	sm.metrics.Disk = []DiskMetrics{}

	for i, line := range lines {
		if i == 0 || line == "" { // 헤더 스킵
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			device := fields[0]
			mountPoint := fields[len(fields)-1]

			// 숫자로 변환 가능한 필드만 처리
			totalStr := strings.TrimSuffix(fields[1], "G")
			usedStr := strings.TrimSuffix(fields[2], "G")
			availStr := strings.TrimSuffix(fields[3], "G")
			usePercentStr := strings.TrimSuffix(fields[4], "%")

			total, err1 := strconv.ParseFloat(totalStr, 64)
			used, err2 := strconv.ParseFloat(usedStr, 64)
			avail, err3 := strconv.ParseFloat(availStr, 64)
			usePercent, err4 := strconv.ParseFloat(usePercentStr, 64)

			if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
				diskMetric := DiskMetrics{
					Device:        device,
					MountPoint:    mountPoint,
					TotalGB:       total,
					UsedGB:        used,
					FreeGB:        avail,
					UsagePercent:  usePercent,
				}

				// inode 사용률 추가 수집
				inodeCmd := exec.Command("df", "-i", mountPoint)
				inodeOutput, err := inodeCmd.Output()
				if err == nil {
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

				sm.metrics.Disk = append(sm.metrics.Disk, diskMetric)
			}
		}
	}
}

// collectNetworkMetrics 네트워크 메트릭 수집
func (sm *SystemMonitor) collectNetworkMetrics() {
	if runtime.GOOS == "linux" {
		data, err := ioutil.ReadFile("/proc/net/dev")
		if err != nil {
			return
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if i < 2 { // 헤더 스킵
				continue
			}

			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					interfaceName := strings.TrimSpace(parts[0])
					if interfaceName == "lo" { // 루프백 인터페이스 스킵
						continue
					}

					fields := strings.Fields(parts[1])
					if len(fields) >= 16 {
						bytesRecv, _ := strconv.ParseUint(fields[0], 10, 64)
						packetsRecv, _ := strconv.ParseUint(fields[1], 10, 64)
						errorsRecv, _ := strconv.ParseUint(fields[2], 10, 64)
						droppedRecv, _ := strconv.ParseUint(fields[3], 10, 64)
						bytesSent, _ := strconv.ParseUint(fields[8], 10, 64)
						packetsSent, _ := strconv.ParseUint(fields[9], 10, 64)
						errorsSent, _ := strconv.ParseUint(fields[10], 10, 64)
						droppedSent, _ := strconv.ParseUint(fields[11], 10, 64)

						sm.metrics.Network = NetworkMetrics{
							Interface:   interfaceName,
							BytesRecv:   bytesRecv,
							BytesSent:   bytesSent,
							PacketsRecv: packetsRecv,
							PacketsSent: packetsSent,
							ErrorsRecv:  errorsRecv,
							ErrorsSent:  errorsSent,
							DroppedRecv: droppedRecv,
							DroppedSent: droppedSent,
						}
						break // 첫 번째 활성 인터페이스만 사용
					}
				}
			}
		}
	}
}

// collectTemperatureMetrics 온도 메트릭 수집
func (sm *SystemMonitor) collectTemperatureMetrics() {
	sm.metrics.Temperature.CoreTemps = make(map[string]float64)

	if runtime.GOOS == "linux" {
		// /sys/class/thermal/thermal_zone*/temp 파일들 확인
		cmd := exec.Command("find", "/sys/class/thermal", "-name", "thermal_zone*", "-type", "d")
		output, err := cmd.Output()
		if err == nil {
			thermalZones := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, zone := range thermalZones {
				if zone != "" {
					tempFile := zone + "/temp"
					if data, err := ioutil.ReadFile(tempFile); err == nil {
						tempStr := strings.TrimSpace(string(data))
						if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
							temp = temp / 1000 // 밀리도에서 도로 변환
							sm.metrics.Temperature.CoreTemps[zone] = temp
							if sm.metrics.Temperature.CPUTemp == 0 || temp > sm.metrics.Temperature.CPUTemp {
								sm.metrics.Temperature.CPUTemp = temp
							}
						}
					}
				}
			}
		}

		// sensors 명령어 시도
		if sm.metrics.Temperature.CPUTemp == 0 {
			cmd := exec.Command("sensors")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "°C") {
						parts := strings.Fields(line)
						for _, part := range parts {
							if strings.Contains(part, "°C") {
								tempStr := strings.Split(part, "°C")[0]
								if strings.HasPrefix(tempStr, "+") {
									tempStr = tempStr[1:]
								}
								if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
									if strings.Contains(strings.ToLower(line), "core") {
										sm.metrics.Temperature.CoreTemps[line] = temp
									}
									if sm.metrics.Temperature.CPUTemp == 0 || temp > sm.metrics.Temperature.CPUTemp {
										sm.metrics.Temperature.CPUTemp = temp
									}
								}
							}
						}
					}
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		// macOS에서는 osx-cpu-temp 같은 도구 필요 (없으면 스킵)
		cmd := exec.Command("osx-cpu-temp")
		output, err := cmd.Output()
		if err == nil {
			tempStr := strings.TrimSpace(string(output))
			tempStr = strings.TrimSuffix(tempStr, "°C")
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
				sm.metrics.Temperature.CPUTemp = temp
			}
		}
	}
}

// collectLoadMetrics 로드 평균 수집
func (sm *SystemMonitor) collectLoadMetrics() {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		data, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {
			// macOS나 다른 시스템에서는 uptime 명령어 사용
			cmd := exec.Command("uptime")
			output, err := cmd.Output()
			if err == nil {
				// load averages: 1.23 1.45 1.67 형태 파싱
				line := string(output)
				if strings.Contains(line, "load average") {
					parts := strings.Split(line, "load average")
					if len(parts) == 2 {
						loadParts := strings.Split(parts[1], ",")
						if len(loadParts) >= 3 {
							load1, _ := strconv.ParseFloat(strings.TrimSpace(loadParts[0]), 64)
							load5, _ := strconv.ParseFloat(strings.TrimSpace(loadParts[1]), 64)
							load15, _ := strconv.ParseFloat(strings.TrimSpace(loadParts[2]), 64)

							sm.metrics.LoadAverage = LoadMetrics{
								Load1Min:  load1,
								Load5Min:  load5,
								Load15Min: load15,
							}
						}
					}
				}
			}
			return
		}

		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			load1, _ := strconv.ParseFloat(parts[0], 64)
			load5, _ := strconv.ParseFloat(parts[1], 64)
			load15, _ := strconv.ParseFloat(parts[2], 64)

			sm.metrics.LoadAverage = LoadMetrics{
				Load1Min:  load1,
				Load5Min:  load5,
				Load15Min: load15,
			}
		}
	}
}

// collectProcessMetrics 프로세스 메트릭 수집
func (sm *SystemMonitor) collectProcessMetrics() {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	sm.metrics.ProcessCount.Total = len(lines) - 2 // 헤더와 빈 줄 제외

	// 간단한 프로세스 상태 카운트
	sm.metrics.ProcessCount.Running = sm.metrics.ProcessCount.Total
	sm.metrics.ProcessCount.Sleeping = 0
	sm.metrics.ProcessCount.Stopped = 0
	sm.metrics.ProcessCount.Zombie = 0
}

// checkAlerts 알림 확인
func (sm *SystemMonitor) checkAlerts() {
	// CPU 사용률 체크
	if sm.metrics.CPU.UsagePercent > sm.thresholds.CPUPercent {
		alert := SystemAlert{
			Level:     "HIGH",
			Type:      "CPU",
			Message:   fmt.Sprintf("CPU 사용률이 높습니다: %.1f%%", sm.metrics.CPU.UsagePercent),
			Value:     sm.metrics.CPU.UsagePercent,
			Threshold: sm.thresholds.CPUPercent,
			Metrics:   *sm.metrics,
			Timestamp: time.Now(),
			Suggestions: []string{
				"🔍 높은 CPU 사용률의 프로세스 확인: top 또는 htop 명령어 사용",
				"⏹️  불필요한 프로세스 종료 검토",
				"📈 시스템 성능 모니터링 강화",
			},
		}
		sm.sendAlert(alert)
	}

	// 메모리 사용률 체크
	if sm.metrics.Memory.UsagePercent > sm.thresholds.MemoryPercent {
		alert := SystemAlert{
			Level:     "HIGH",
			Type:      "MEMORY",
			Message:   fmt.Sprintf("메모리 사용률이 높습니다: %.1f%%", sm.metrics.Memory.UsagePercent),
			Value:     sm.metrics.Memory.UsagePercent,
			Threshold: sm.thresholds.MemoryPercent,
			Metrics:   *sm.metrics,
			Timestamp: time.Now(),
			Suggestions: []string{
				"🧹 시스템 캐시 정리: sync && echo 3 > /proc/sys/vm/drop_caches",
				"📊 메모리 사용량이 높은 프로세스 확인",
				"💾 스왑 공간 확인 및 확장 검토",
			},
		}
		sm.sendAlert(alert)
	}

	// 디스크 사용률 체크
	for _, disk := range sm.metrics.Disk {
		if disk.UsagePercent > sm.thresholds.DiskPercent {
			alert := SystemAlert{
				Level:     "CRITICAL",
				Type:      "DISK",
				Message:   fmt.Sprintf("디스크 공간이 부족합니다 (%s): %.1f%%", disk.MountPoint, disk.UsagePercent),
				Value:     disk.UsagePercent,
				Threshold: sm.thresholds.DiskPercent,
				Metrics:   *sm.metrics,
				Timestamp: time.Now(),
				Suggestions: []string{
					"🗑️  불필요한 파일 삭제",
					"📦 로그 파일 압축 또는 삭제",
					"💽 디스크 공간 확장 검토",
				},
			}
			sm.sendAlert(alert)
		}
	}

	// CPU 온도 체크
	if sm.metrics.Temperature.CPUTemp > sm.thresholds.CPUTemp {
		alert := SystemAlert{
			Level:     "HIGH",
			Type:      "TEMPERATURE",
			Message:   fmt.Sprintf("CPU 온도가 높습니다: %.1f°C", sm.metrics.Temperature.CPUTemp),
			Value:     sm.metrics.Temperature.CPUTemp,
			Threshold: sm.thresholds.CPUTemp,
			Metrics:   *sm.metrics,
			Timestamp: time.Now(),
			Suggestions: []string{
				"🌡️  시스템 쿨링 상태 확인",
				"🧹 먼지 청소 및 팬 상태 점검",
				"⚡ CPU 부하 확인 및 조정",
			},
		}
		sm.sendAlert(alert)
	}

	// 로드 평균 체크
	if sm.metrics.LoadAverage.Load1Min > sm.thresholds.LoadAverage {
		alert := SystemAlert{
			Level:     "MEDIUM",
			Type:      "LOAD",
			Message:   fmt.Sprintf("시스템 로드가 높습니다: %.2f", sm.metrics.LoadAverage.Load1Min),
			Value:     sm.metrics.LoadAverage.Load1Min,
			Threshold: sm.thresholds.LoadAverage,
			Metrics:   *sm.metrics,
			Timestamp: time.Now(),
			Suggestions: []string{
				"🔍 높은 부하를 유발하는 프로세스 확인",
				"⚖️  작업 부하 분산 검토",
				"🚀 시스템 리소스 업그레이드 고려",
			},
		}
		sm.sendAlert(alert)
	}
}

// sendAlert 알림 전송
func (sm *SystemMonitor) sendAlert(alert SystemAlert) {
	select {
	case sm.alertChannel <- alert:
	default:
		// 채널이 가득 차면 무시 (논블로킹)
	}
}

// GetAlertChannel 알림 채널 반환
func (sm *SystemMonitor) GetAlertChannel() <-chan SystemAlert {
	return sm.alertChannel
}

// updateHistory 히스토리 업데이트
func (sm *SystemMonitor) updateHistory() {
	sm.history = append(sm.history, *sm.metrics)
	if len(sm.history) > sm.maxHistorySize {
		sm.history = sm.history[1:]
	}
}

// GetCurrentMetrics 현재 메트릭 반환
func (sm *SystemMonitor) GetCurrentMetrics() SystemMetrics {
	return *sm.metrics
}

// GetMetricsHistory 메트릭 히스토리 반환
func (sm *SystemMonitor) GetMetricsHistory() []SystemMetrics {
	return sm.history
}

// GetSystemReport 시스템 보고서 생성
func (sm *SystemMonitor) GetSystemReport() string {
	metrics := sm.GetCurrentMetrics()
	
	report := fmt.Sprintf(`
🖥️  시스템 모니터링 보고서
========================
⏰ 수집 시간: %s

💻 CPU 정보:
  - 사용률: %.1f%% (임계값: %.1f%%)
  - 사용자: %.1f%%, 시스템: %.1f%%, 대기: %.1f%%
  - 코어 수: %d개

🧠 메모리 정보:
  - 사용률: %.1f%% (임계값: %.1f%%)
  - 총 메모리: %.1f GB
  - 사용 중: %.1f GB
  - 사용 가능: %.1f GB

💾 디스크 정보:`,
		metrics.Timestamp.Format("2006-01-02 15:04:05"),
		metrics.CPU.UsagePercent, sm.thresholds.CPUPercent,
		metrics.CPU.UserPercent, metrics.CPU.SystemPercent, metrics.CPU.IdlePercent,
		metrics.CPU.Cores,
		metrics.Memory.UsagePercent, sm.thresholds.MemoryPercent,
		metrics.Memory.TotalMB/1024,
		metrics.Memory.UsedMB/1024,
		metrics.Memory.AvailableMB/1024,
	)

	for _, disk := range metrics.Disk {
		report += fmt.Sprintf(`
  - %s (%s): %.1f%% 사용 (%.1f/%.1f GB)`,
			disk.Device, disk.MountPoint, disk.UsagePercent, disk.UsedGB, disk.TotalGB)
	}

	report += fmt.Sprintf(`

🌡️  온도 정보:
  - CPU 온도: %.1f°C (임계값: %.1f°C)

⚖️  시스템 로드:
  - 1분: %.2f, 5분: %.2f, 15분: %.2f (임계값: %.1f)

🔄 프로세스:
  - 총 프로세스 수: %d개
`,
		metrics.Temperature.CPUTemp, sm.thresholds.CPUTemp,
		metrics.LoadAverage.Load1Min, metrics.LoadAverage.Load5Min, metrics.LoadAverage.Load15Min, sm.thresholds.LoadAverage,
		metrics.ProcessCount.Total,
	)

	// 네트워크 정보 추가
	if metrics.Network.Interface != "" {
		report += fmt.Sprintf(`
🌐 네트워크 (%s):
  - 수신: %d 바이트, %d 패킷
  - 송신: %d 바이트, %d 패킷
  - 에러: 수신 %d, 송신 %d
`,
			metrics.Network.Interface,
			metrics.Network.BytesRecv, metrics.Network.PacketsRecv,
			metrics.Network.BytesSent, metrics.Network.PacketsSent,
			metrics.Network.ErrorsRecv, metrics.Network.ErrorsSent,
		)
	}

	return report
}

// SetThresholds 임계값 설정
func (sm *SystemMonitor) SetThresholds(thresholds SystemThresholds) {
	sm.thresholds = thresholds
}

// GetThresholds 현재 임계값 반환
func (sm *SystemMonitor) GetThresholds() SystemThresholds {
	return sm.thresholds
} 