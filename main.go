package HG_Agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// SystemInfo 包含我们想要监控的系统信息
type SystemInfo struct {
	CPUUtilization    float64   `json:"cpu_utilization"`    // CPU 利用率
	MemoryUtilization float64   `json:"memory_utilization"` // 内存利用率
	DiskUsage         float64   `json:"disk_usage"`         // 磁盘使用率
	BandwidthUsage    Bandwidth `json:"bandwidth_usage"`    // 带宽使用情况
}

// Bandwidth 包含带宽的上下行流量信息
type Bandwidth struct {
	BytesSent uint64 `json:"bytes_sent"` // 发送的字节数
	BytesRecv uint64 `json:"bytes_recv"` // 接收的字节数
}

// collectCPUInfo 收集CPU信息
func collectCPUInfo() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return percentages[0], nil // 返回第一个CPU的使用率
}

// collectMemoryInfo 收集内存信息
func collectMemoryInfo() (float64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return v.UsedPercent, nil // 返回内存使用率
}

// collectDiskInfo 收集磁盘信息
func collectDiskInfo() (float64, error) {
	d, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}
	return d.UsedPercent, nil // 返回磁盘使用率
}

// collectBandwidthInfo 收集带宽信息
func collectBandwidthInfo() (Bandwidth, error) {
	counters, err := net.IOCounters(false)
	if err != nil {
		return Bandwidth{}, err
	}
	return Bandwidth{
		BytesSent: counters[0].BytesSent,
		BytesRecv: counters[0].BytesRecv,
	}, nil
}

// collectSystemInfo 收集所有系统信息
func collectSystemInfo() (*SystemInfo, error) {
	cpuUtilization, err := collectCPUInfo()
	if err != nil {
		return nil, err
	}

	memoryUtilization, err := collectMemoryInfo()
	if err != nil {
		return nil, err
	}

	diskUsage, err := collectDiskInfo()
	if err != nil {
		return nil, err
	}

	bandwidthUsage, err := collectBandwidthInfo()
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		CPUUtilization:    cpuUtilization,
		MemoryUtilization: memoryUtilization,
		DiskUsage:         diskUsage,
		BandwidthUsage:    bandwidthUsage,
	}, nil
}

// sendSystemInfo 将收集到的信息发送到监控端
func sendSystemInfo(info *SystemInfo) {
	jsonValue, _ := json.Marshal(info)
	http.Post("http://monitor-server/api/collect", "application/json", bytes.NewBuffer(jsonValue))
}

func main() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			info, err := collectSystemInfo()
			if err != nil {
				// 处理错误
				continue
			}
			sendSystemInfo(info)
		}
	}
}
