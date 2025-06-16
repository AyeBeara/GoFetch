package main

import (
	"fmt"
	"log"
	"os"
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

func get_cpu_usage() (string, int, error) {
	cpu_load, err := cpu.Percent(time.Second, false)
	if err != nil {
		return "", 0, err
	}
	cpu_info, err := cpu.Info()
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%s @ %0.2f Ghz (%0.2f%%)", strings.TrimSpace(cpu_info[0].ModelName), cpu_info[0].Mhz/1000, cpu_load[0]), int(cpu_load[0]), nil
}

func get_disk_usage() (string, int, error) {
	path := ""
	if runtime.GOOS == "windows" {
		path = "C:"
	} else if runtime.GOOS == "linux" {
		path = "/"
	} else {
		return "", 0, fmt.Errorf("unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}
	disk, err := disk.Usage(path)
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}
	free := float64(disk.Free) / (1024 * 1024 * 1024)
	total := float64(disk.Total) / (1024 * 1024 * 1024)
	used := float64(disk.Used) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free), int(disk.UsedPercent), nil
}

func get_mem_usage() (string, int, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return "", 0, err
	}
	free := float64(memory.Available) / (1024 * 1024 * 1024)
	total := float64(memory.Total) / (1024 * 1024 * 1024)
	used := float64(memory.Used) / (1024 * 1024 * 1024)
	memory_usage := fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free)

	swp, err := mem.SwapMemory()
	if err != nil {
		return "", 0, err
	}
	free = float64(swp.Free) / (1024 * 1024 * 1024)
	total = float64(swp.Total) / (1024 * 1024 * 1024)
	used = float64(swp.Used) / (1024 * 1024 * 1024)
	swp_usage := fmt.Sprintf("[%0.2fGB / %0.2fGB (%0.2fGB free) SWAP]", used, total, free)

	return memory_usage + " " + swp_usage, int(memory.UsedPercent), nil
}

func get_os_version() (string, error) {
	var ver []byte
	var err error
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "ver")
		ver, err = cmd.Output()
		if err != nil {
			return "", err
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("uname", "-rv")
		ver, err = cmd.Output()
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}

	return strings.TrimSpace(string(ver)), nil
}

func render(area *pterm.AreaPrinter) {
	cpu_usage, cpu_load, err := get_cpu_usage()
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}

	disk_usage, disk_load, err := get_disk_usage()
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}

	memory_usage, memory_load, err := get_mem_usage()
	if err != nil {
		log.Fatalf("Error getting memory usage: %v", err)
	}

	user, _ := user.Current()

	os_version, err := get_os_version()
	if err != nil {
		log.Fatalf("Error getting OS version: %v", err)
	}

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
			Text:      memory_usage,
			TextStyle: pterm.NewStyle(pterm.FgLightYellow),
			Bullet:    "🧠  ",
		},
	}

	info, err := pterm.DefaultBulletList.WithItems(resources).Srender()
	if err != nil {
		log.Fatalf("Error rendering bullet list: %v", err)
	}

	bars := []pterm.Bar{
		{Value: cpu_load},
		{Value: disk_load},
		{Value: memory_load},
	}

	barchart, err := pterm.DefaultBarChart.WithBars(bars).WithHorizontal().WithWidth(50).Srender()
	if err != nil {
		log.Fatalf("Error rendering bar chart: %v", err)
	}

	panels := pterm.Panels{
		{
			{Data: "\n" + barchart},
			{Data: info},
		},
	}

	render, err := pterm.DefaultPanel.WithPanels(panels).Srender()
	if err != nil {
		log.Fatalf("Error rendering panel: %v", err)
	}

	area.Update(render)
}

func main() {

	// Display

	area, _ := pterm.DefaultArea.Start()
	defer area.Stop()

	if len(os.Args) > 1 && (os.Args[1] == "-l" || os.Args[1] == "--live") {
		for {
			render(area)
			time.Sleep(5 * time.Second)
		}
	} else if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Usage: GoFetch [options]")
		fmt.Println("Options:")
		fmt.Println("  -l, --live    Display system information in live mode (updates every 5 seconds)")
		fmt.Println("  -h, --help    Display this help message")
		fmt.Println("Displays system information in a terminal-friendly format.")
		fmt.Println("Press Ctrl+C to exit.")
		return
	}

	render(area)
}
