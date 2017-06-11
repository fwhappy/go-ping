package main

import (
	"fmt"
	"time"

	"sync"

	ping "github.com/fwhappy/go-ping"
)

func main() {
	hosts := []string{"101.37.226.25", "101.37.226.27", "101.37.226.26", "93.46.8.89"}
	// hosts := []string{"101.37.226.25"}
	// hosts := []string{"93.46.8.89"}
	for {
		wg := &sync.WaitGroup{}
		wg.Add(len(hosts))
		for _, host := range hosts {
			go p(host, wg)
		}
		wg.Wait()
		time.Sleep(time.Second * 5)
	}
	time.Sleep(100 * time.Second)
}

func p(host string, wg *sync.WaitGroup) {
	defer wg.Done()
	pinger, err := ping.NewPinger(host)
	if err != nil {
		fmt.Println("Fatar error:", err.Error())
	}

	pinger.OnStart = func() {
		fmt.Sprintf("PING %s:\n", pinger.Addr())
	}
	pinger.OnRecv = func() {
		fmt.Sprintf("来自 [%v]的回复:字节=%v 时间=%vms TTL=%v\n", pinger.Addr(), pinger.Size(), pinger.Stats.Endduration, pinger.Stats.TTL)
	}
	pinger.OnTimeout = func() {
		fmt.Sprintf("对 [" + pinger.Addr() + "]" + " 的请求超时。\n")
	}

	pinger.OnFinish = func() {
		var logString = ""
		logString += fmt.Sprintf("[%v] 的 Ping 统计信息:\n", pinger.Addr())
		logString += fmt.Sprintf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", pinger.PacketsSent, pinger.PacketsRecv, pinger.PacketsSent-pinger.PacketsRecv, int((pinger.PacketsSent-pinger.PacketsRecv)*100/pinger.PacketsSent))
		logString += fmt.Sprintf("往返行程的估计时间(以毫秒为单位):\n")
		if pinger.PacketsRecv != 0 {
			logString += fmt.Sprintf("    最短 = %dms，最长 = %dms，平均 = %dms\n", pinger.Stats.ShortT, pinger.Stats.LongT, pinger.Stats.SumT/pinger.PacketsSent)
		}
		fmt.Println(logString)
	}

	pinger.Count = 1
	err = pinger.Run()
	if err != nil {
		fmt.Println("Fatar error:", err.Error())
	}
}
