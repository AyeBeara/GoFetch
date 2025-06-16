package main

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v4/cpu"
)

func main() {

	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}

	user, _ := user.Current()

	cpu_info, err := cpu.Info()
	if err != nil {
		log.Fatalf("Error getting CPU info: %v", err)
	}

	var ver []byte
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "ver")
		ver, err = cmd.Output()
		if err != nil {
			log.Fatalf("Error getting OS version: %v", err)
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("uname", "-rv")
		ver, err = cmd.Output()
		if err != nil {
			log.Fatalf("Error getting OS version: %v", err)
		}
	} else {
		log.Fatalf("Unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}

	os_version := strings.TrimSpace(string(ver))

	resources := []pterm.BulletListItem{
		{
			Level:     0,
			Text:      fmt.Sprintf("User: %s", user.Username),
			TextStyle: pterm.NewStyle(pterm.FgLightBlue),
			Bullet:    "👤  ",
		},
		{
			Level:     0,
			Text:      fmt.Sprintf("OS: %s", os_version),
			TextStyle: pterm.NewStyle(pterm.FgCyan),
			Bullet:    "🖥️   ",
		},
		{
			Level:     0,
			Text:      fmt.Sprintf("CPU Model: %s", cpu_info[0].ModelName),
			TextStyle: pterm.NewStyle(pterm.FgYellow),
			Bullet:    "⚙️   ",
		},
		{
			Level:     0,
			Text:      fmt.Sprintf("CPU Usage: %0.2f%%", cpuUsage[0]),
			TextStyle: pterm.NewStyle(pterm.FgGreen),
			Bullet:    "📈  ",
		},
	}

	pterm.DefaultBulletList.WithItems(resources).Render()
}
