package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func get_cpu_usage(cpu_chan chan map[string]int) {
	cpu_load, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}
	cpu_info, err := cpu.Info()
	if err != nil {
		log.Fatalf("Error getting CPU info: %v", err)
	}
	cpu_chan <- map[string]int{fmt.Sprintf("%s @ %0.2f Ghz [%0.2f%% Utilization]", strings.TrimSpace(cpu_info[0].ModelName), cpu_info[0].Mhz/1000, cpu_load[0]): int(cpu_load[0])}
}

func get_disk_usage(disk_chan chan map[string]int) {
	path := ""
	if runtime.GOOS == "windows" {
		path = "C:"
	} else if runtime.GOOS == "linux" {
		path = "/"
	} else {
		log.Fatalf("unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}
	disk, err := disk.Usage(path)
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}
	free := float64(disk.Free) / (1024 * 1024 * 1024)
	total := float64(disk.Total) / (1024 * 1024 * 1024)
	used := float64(disk.Used) / (1024 * 1024 * 1024)
	disk_chan <- map[string]int{fmt.Sprintf("%0.2fGB / %0.2fGB (%0.2fGB free)", used, total, free): int(disk.UsedPercent)}
}

func get_mem_usage(mem_chan chan map[string]int) {
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

	mem_chan <- map[string]int{memory_usage + " " + swp_usage: int(memory.UsedPercent)}
}

func get_os_version() (string, error) {
	var ver []byte
	var err error
	if runtime.GOOS == "windows" {
		ver, err = exec.Command("cmd", "/C", "ver").Output()
		if err != nil {
			return "", err
		}
	} else if runtime.GOOS == "linux" {
		ver, err = exec.Command("uname", "-rv").Output()
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unsupported OS: %s. Please submit a pull request at github.com/AyeBeara/GoFetch", runtime.GOOS)
	}

	return strings.TrimSpace(string(ver)), nil
}

func get_gpu_usage(gpu_chan chan map[string]int) {
	var gpu_usage string
	var gpu_load int

	if runtime.GOOS == "windows" {
		out, err := exec.Command("powershell", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Name").Output()
		if err != nil {
			log.Fatalf("Error detecting GPU: %v", err)
		}
		switch {
		case strings.Contains(strings.ToLower(string(out)), "nvidia"):
			out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.used,memory.total,utilization.gpu", "--format=csv,noheader,nounits").Output()
			if err != nil {
				log.Fatalf("Error getting NVIDIA GPU usage: %v", err)
			}
			parts := strings.Split(strings.TrimSpace(string(out)), ", ")
			load, err := strconv.Atoi(parts[3])
			if err != nil {
				log.Fatalf("Error parsing GPU load: %v", err)
			}

			used, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Fatalf("Error parsing GPU used memory: %v", err)
			}
			total, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				log.Fatalf("Error parsing GPU total memory: %v", err)
			}
			used = used / 1024
			total = total / 1024
			free := (total - used)

			gpu_usage = fmt.Sprintf("%s %0.2fGB / %0.2fGB (%0.2fGB Free) [%d%% Utilization]", parts[0], used, total, free, load)
			gpu_load = load
		case strings.Contains(strings.ToLower(string(out)), "amd"):
			// TODO: Implement AMD GPU usage retrieval
		case strings.Contains(strings.ToLower(string(out)), "intel"):
			// TODO: Implement Intel GPU usage retrieval
		}
	} else if runtime.GOOS == "linux" {

		out, err := exec.Command("lspci", "-vnn", "-s", "00:02.0").Output()
		if err != nil {
			log.Fatalf("Error detecting GPU: %v", err)
		}
		switch {
		case strings.Contains(strings.ToLower(string(out)), "nvidia"):
		case strings.Contains(strings.ToLower(string(out)), "amd"):
		case strings.Contains(strings.ToLower(string(out)), "intel"):
			cmd := exec.Command("grep", "-oP", "(?<=:\\s).*(?=\\s\\()")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Fatalf("Error creating stdin pipe: %v", err)
			}

			go func() {
				defer stdin.Close()
				io.WriteString(stdin, strings.TrimSpace(string(out)))
			}()

			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("Error getting Intel GPU name: %v", err)
			}

			gpu_usage = strings.TrimSpace(string(out))

			ctx, cancel := context.WithCancel(context.Background())
			out, err = exec.CommandContext(ctx, "intel_gpu_top", "-c").Output()
			cancel()
			if err != nil {
				log.Fatalf("Error getting Intel GPU usage: %v", err)
			}
			gpu_load, err = strconv.Atoi(strings.Split(strings.Split(strings.TrimSpace(string(out)), "\n")[1], ",")[8])
			if err != nil {
				log.Fatalf("Error parsing Intel GPU load: %v", err)
			}
		}
	}
	gpu_chan <- map[string]int{gpu_usage: gpu_load}
}

func render(area *pterm.AreaPrinter, channels []chan map[string]int) {
	go get_cpu_usage(channels[0])
	go get_disk_usage(channels[1])
	go get_mem_usage(channels[2])
	go get_gpu_usage(channels[3])

	user, _ := user.Current()

	os_version, err := get_os_version()
	if err != nil {
		log.Fatalf("Error getting OS version: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(4)

	cpu_usages := <-channels[0]
	cpu_usage, cpu_load := "", 0
	go func() {
		defer wg.Done()
		for k, v := range cpu_usages {
			cpu_usage = k
			cpu_load = v
		}
	}()

	disk_usages := <-channels[1]
	disk_usage, disk_load := "", 0
	go func() {
		defer wg.Done()
		for k, v := range disk_usages {
			disk_usage = k
			disk_load = v
		}
	}()

	memory_usages := <-channels[2]
	memory_usage, memory_load := "", 0
	go func() {
		defer wg.Done()
		for k, v := range memory_usages {
			memory_usage = k
			memory_load = v
		}
	}()

	gpu_usages := <-channels[3]
	gpu_usage, gpu_load := "", 0
	go func() {
		defer wg.Done()
		for k, v := range gpu_usages {
			gpu_usage = k
			gpu_load = v
		}
	}()

	wg.Wait()

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
			Text:      gpu_usage,
			TextStyle: pterm.NewStyle(pterm.FgLightMagenta),
			Bullet:    "🖼️   ",
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
		{Value: gpu_load},
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

	cpu_channel := make(chan map[string]int, 1)
	disk_channel := make(chan map[string]int, 1)
	mem_channel := make(chan map[string]int, 1)
	gpu_channel := make(chan map[string]int, 1)
	channels := []chan map[string]int{cpu_channel, disk_channel, mem_channel, gpu_channel}

	area, _ := pterm.DefaultArea.Start()
	defer area.Stop()

	if len(os.Args) > 1 && (os.Args[1] == "-l" || os.Args[1] == "--live") {
		for {
			render(area, channels)
			time.Sleep(time.Second)
		}
	} else if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Usage: GoFetch [options]")
		fmt.Println("Options:")
		fmt.Println("  -l, --live    Display system information in live mode (updates every second)")
		fmt.Println("  -h, --help    Display this help message")
		fmt.Println("Displays system information in a terminal-friendly format.")
		fmt.Println("Press Ctrl+C to exit.")
		return
	}

	render(area, channels)
}
