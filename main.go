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
	"github.com/shirou/gopsutil/v4/disk"
)

func main() {

	// CPU Usage & Info

	cpu_usage, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}

	cpu_info, err := cpu.Info()
	if err != nil {
		log.Fatalf("Error getting CPU info: %v", err)
	}

	// Disk Usage

	var disk_usage string
	if runtime.GOOS == "windows" {
		disk, err := disk.Usage("C:")
		if err != nil {
			log.Fatalf("Error getting disk usage: %v", err)
		}
		free := disk.Free / (1024 * 1024 * 1024) // Convert to GB
		total := disk.Total / (1024 * 1024 * 1024)
		used := disk.Used / (1024 * 1024 * 1024)
		disk_usage = fmt.Sprintf("%dGB / %dGB (%dGB free)", used, total, free)
	} else if runtime.GOOS == "linux" {
		disk, err := disk.Usage("/")
		if err != nil {
			log.Fatalf("Error getting disk usage: %v", err)
		}
		free := disk.Free / (1024 * 1024 * 1024) // Convert to GB
		total := disk.Total / (1024 * 1024 * 1024)
		used := disk.Used / (1024 * 1024 * 1024)
		disk_usage = fmt.Sprintf("%dGB / %dGB (%dGB free)", used, total, free)
	} else {
		log.Fatalf("Unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}

	// Memory Usage

	// Current User

	user, _ := user.Current()

	// OS Version & Kernel (Linux)

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

	// Display

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
			Text:      fmt.Sprintf("CPU Usage: %0.2f%%", cpu_usage[0]),
			TextStyle: pterm.NewStyle(pterm.FgGreen),
			Bullet:    "📈  ",
		},
		{
			Level:     0,
			Text:      disk_usage,
			TextStyle: pterm.NewStyle(pterm.FgMagenta),
			Bullet:    "💾  ",
		},
	}

	pterm.DefaultBulletList.WithItems(resources).Render()
}
