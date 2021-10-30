package Global

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	load2 "github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var WsData *WsMessage

var NetTxRx TxRx

var once sync.Once

type WsMessage struct {
	Uptime      uint64  `json:"uptime"`
	Load        float64 `json:"load"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	HddTotal    uint64  `json:"hdd_total"`
	HddUsed     uint64  `json:"hdd_used"`
	CPU         float64 `json:"cpu"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	NetworkIn   uint64  `json:"network_in"`
	NetworkOut  uint64  `json:"network_out"`
}

type TxRx struct {
	Tx      []uint64
	Rx      []uint64
	TxIndex uint64
	RxIndex uint64
	TxLen   uint64
	RxLen   uint64
}

func GetUptime() uint64 {
	uptime, _ := host.BootTime()
	uptime = uint64(time.Now().Unix()) - uptime
	return uptime
}

func GetMemory() (uint64, uint64, uint64, uint64) {
	v, _ := mem.VirtualMemory()
	s, _ := mem.SwapMemory()
	return v.Total / 1024, v.Used / 1024, s.Total / 1024, s.Used / 1024
}

func GetLoad() float64 {
	switch runtime.GOOS {
	case "windows":
		return -1.0
	default:
		load, _ := load2.Avg()
		return load.Load1
	}
}

func GetCPU(IntervalTime uint16) float64 {
	u, _ := cpu.Percent(time.Second*time.Duration(IntervalTime), false)
	v := float64(int64(u[0]*float64(100))) / 100.0
	return v
}

func GetNetInfo(IntervalTime uint16) (uint64, uint64, uint64, uint64) {
	once.Do(func() {
		NetTxRx.Tx = make([]uint64, 10)
		NetTxRx.Rx = make([]uint64, 10)
		NetTxRx.RxIndex = 0
		NetTxRx.TxIndex = 0
		NetTxRx.RxLen = 0
		NetTxRx.TxLen = 0
	})
	var (
		avgRx  uint64
		avgTx  uint64
		NetIn  uint64
		NetOut uint64
	)
	avgRx, NetIn, NetOut, avgTx = 0, 0, 0, 0
	info, _ := net.IOCounters(true)
	for _, v := range info {
		if find := strings.Contains(v.Name, "lo"); find {
			continue
		}
		if find := strings.Contains(v.Name, "tun"); find {
			continue
		}
		NetIn += v.BytesRecv
		NetOut += v.BytesSent
	}
	NetTxRx.Rx[NetTxRx.RxIndex] = NetIn
	NetTxRx.Tx[NetTxRx.TxIndex] = NetOut
	NetTxRx.RxIndex++
	NetTxRx.TxIndex++
	if NetTxRx.RxIndex >= 10 {
		NetTxRx.RxIndex = 0
		avgRx = (NetTxRx.Rx[9] - NetTxRx.Rx[0]) / 10 / uint64(IntervalTime)
	} else {
		avgRx = (NetTxRx.Rx[NetTxRx.RxIndex-1] - NetTxRx.Rx[NetTxRx.RxIndex]) / 10 / uint64(IntervalTime)
	}
	if NetTxRx.TxIndex >= 10 {
		NetTxRx.TxIndex = 0
		avgTx = (NetTxRx.Tx[9] - NetTxRx.Tx[0]) / 10 / uint64(IntervalTime)
	} else {
		avgTx = (NetTxRx.Tx[NetTxRx.TxIndex-1] - NetTxRx.Tx[NetTxRx.TxIndex]) / 10 / uint64(IntervalTime)
	}
	return NetIn, NetOut, avgRx, avgTx
}

func GetDiskInfo() (uint64, uint64) {
	validFs := []string{"ext4", "ext3", "ext2", "reiserfs", "jfs", "btrfs", "fuseblk", "zfs", "simfs", "ntfs", "fat32", "exfat", "xfs"}
	parts, _ := disk.Partitions(true)
	flag := false
	diskTotal := uint64(0)
	diskUsed := uint64(0)
	for _, part := range parts {
		diskInfo, _ := disk.Usage(part.Mountpoint)
		for _, item := range validFs {
			if item == strings.ToLower(part.Fstype) {
				flag = true
				break
			}
		}
		if flag {
			diskTotal += diskInfo.Total
			diskUsed += diskInfo.Used
		}
	}
	return uint64(float64(diskTotal) / 1024.0 / 1024.0), uint64(float64(diskUsed) / 1024.0 / 1024.0)
}

func GenWebsocketMessage(Interval uint16) *WsMessage {
	memoryTotal, memoryUsed, swapTotal, swapUsed := GetMemory()
	netIn, netOut, netRx, netTx := GetNetInfo(Interval)
	diskTotal, diskFree := GetDiskInfo()
	WsData = &WsMessage{
		Uptime:      GetUptime(),
		Load:        GetLoad(),
		MemoryTotal: memoryTotal,
		MemoryUsed:  memoryUsed,
		SwapTotal:   swapTotal,
		SwapUsed:    swapUsed,
		HddTotal:    diskTotal,
		HddUsed:     diskFree,
		CPU:         GetCPU(Interval),
		NetworkRx:   netRx,
		NetworkTx:   netTx,
		NetworkIn:   netIn,
		NetworkOut:  netOut,
	}
	return WsData
}
