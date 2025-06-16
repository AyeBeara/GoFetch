package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

func main() {
	sec, _ := time.ParseDuration("1s")

	cpuUsage, err := cpu.Percent(sec, false)

	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}

	fmt.Printf("CPU Usage: %v%%", cpuUsage[0])
}
