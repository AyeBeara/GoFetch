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
	"github.com/shirou/gopsutil/v4/mem"
)

func main() {

	// CPU Usage & Info

	cpu_load, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}
	cpu_info, err := cpu.Info()
	if err != nil {
		log.Fatalf("Error getting CPU info: %v", err)
	}
	cpu_usage := fmt.Sprintf("%s @ %0.2f Ghz (%0.2f%%)", strings.TrimSpace(cpu_info[0].ModelName), cpu_info[0].Mhz/1000, cpu_load[0])

	// Disk Usage

	var disk_usage string
	if runtime.GOOS == "windows" {
		disk, err := disk.Usage("C:")
		if err != nil {
			log.Fatalf("Error getting disk usage: %v", err)
		}
		free := float64(disk.Free) / (1024 * 1024 * 1024)
		total := float64(disk.Total) / (1024 * 1024 * 1024)
		used := float64(disk.Used) / (1024 * 1024 * 1024)
		disk_usage = fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free)
	} else if runtime.GOOS == "linux" {
		disk, err := disk.Usage("/")
		if err != nil {
			log.Fatalf("Error getting disk usage: %v", err)
		}
		free := float64(disk.Free / (1024 * 1024 * 1024))
		total := float64(disk.Total) / (1024 * 1024 * 1024)
		used := float64(disk.Used) / (1024 * 1024 * 1024)
		disk_usage = fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free)
	} else {
		log.Fatalf("Unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}

	// Memory Usage

	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error getting memory usage: %v", err)
	}
	free := float64(memory.Available) / (1024 * 1024 * 1024)
	total := float64(memory.Total) / (1024 * 1024 * 1024)
	used := float64(memory.Used) / (1024 * 1024 * 1024)
	memory_usage := fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free)

	swp, err := mem.SwapMemory()
	if err != nil {
		log.Fatalf("Error getting swap memory usage: %v", err)
	}
	free = float64(swp.Free) / (1024 * 1024 * 1024)
	total = float64(swp.Total) / (1024 * 1024 * 1024)
	used = float64(swp.Used) / (1024 * 1024 * 1024)
	swp_usage := fmt.Sprintf("[%0.2fGB / %0.2fGB (%0.2fGB free) SWAP]", used, total, free)

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
			Text:      user.Username,
			TextStyle: pterm.NewStyle(pterm.FgLightBlue),
			Bullet:    "👤  ",
		},
		{
			Level:     0,
			Text:      os_version,
			TextStyle: pterm.NewStyle(pterm.FgCyan),
			Bullet:    "🖥️   ",
		},
		{
			Level:     0,
			Text:      cpu_usage,
			TextStyle: pterm.NewStyle(pterm.FgYellow),
			Bullet:    "⚙️   ",
		},
		{
			Level:     0,
			Text:      disk_usage,
			TextStyle: pterm.NewStyle(pterm.FgMagenta),
			Bullet:    "💾  ",
		},
		{
			Level:     0,
			Text:      memory_usage + " " + swp_usage,
			TextStyle: pterm.NewStyle(pterm.FgLightYellow),
			Bullet:    "🧠  ",
		},
	}

	pterm.DefaultBulletList.WithItems(resources).Render()
}
