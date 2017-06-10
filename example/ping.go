package main

import (
	"fmt"
	"time"

	ping "github.com/fwhappy/go-ping"
)

func main() {
	// hosts := []string{"101.37.226.25", "101.37.226.27", "101.37.226.26", "93.46.8.89"}
	hosts := []string{"101.37.226.25"}
	// hosts := []string{"93.46.8.89"}
	for {
		for _, host := range hosts {
			go p(host)
		}
		time.Sleep(time.Second)
	}
	time.Sleep(100 * time.Second)
}

func p(host string) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		fmt.Println("Fatar error:", err.Error())
	}

	pinger.OnStart = func() {
		fmt.Println("host start")
	}

	pinger.OnRecv = func() {
		fmt.Println("host recv")
	}

	pinger.OnTimeout = func() {
		fmt.Println("host timeout")
	}

	pinger.OnFinish = func() {
		fmt.Println("host finish")
	}

	err = pinger.Run()
	if err != nil {
		fmt.Println("Fatar error:", err.Error())
	}
}
