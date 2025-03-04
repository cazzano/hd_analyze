package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Step 1: List available disks
	disks, err := listDisks()
	if err != nil {
		fmt.Println("Error listing disks:", err)
		return
	}

	// Step 2: Ask the user to select a disk
	selectedDisk, err := selectDisk(disks)
	if err != nil {
		fmt.Println("Error selecting disk:", err)
		return
	}

	// Step 3: Run SMART test on the selected disk
	fmt.Printf("Running SMART test on %s...\n", selectedDisk)
	smartOutput, err := runSmartTest(selectedDisk)
	if err != nil {
		fmt.Println("Error running SMART test:", err)
		return
	}

	// Step 4: Parse SMART test results and determine performance
	performancePercentage := calculatePerformancePercentage(smartOutput)

	// Step 5: Display retro animation with performance percentage and progress bar
	displayRetroAnimation(performancePercentage)
}

func listDisks() ([]string, error) {
	cmd := exec.Command("lsblk", "-d", "-o", "NAME")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	disks := strings.Split(strings.TrimSpace(string(output)), "\n")[1:]
	return disks, nil
}

func selectDisk(disks []string) (string, error) {
	fmt.Println("Available disks:")
	for i, disk := range disks {
		fmt.Printf("%d: %s\n", i+1, disk)
	}

	fmt.Print("Select a disk by number: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(disks) {
		return "", fmt.Errorf("invalid selection")
	}

	return "/dev/" + disks[selection-1], nil
}

func runSmartTest(disk string) (string, error) {
	cmd := exec.Command("smartctl", "-a", disk)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func calculatePerformancePercentage(smartOutput string) int {
	// Parse SMART attributes to calculate performance percentage
	lines := strings.Split(smartOutput, "\n")
	var throughput, seekTime, latency int
	var count int

	for _, line := range lines {
		if strings.Contains(line, "Throughput Performance") {
			throughput = extractAttributeValue(line)
			if throughput > 0 {
				count++
			}
		} else if strings.Contains(line, "Seek Time Performance") {
			seekTime = extractAttributeValue(line)
			if seekTime > 0 {
				count++
			}
		} else if strings.Contains(line, "Command Timeout") {
			latency = extractAttributeValue(line)
			if latency > 0 {
				count++
			}
		}
	}

	// Calculate performance percentage only if valid attributes are found
	if count == 0 {
		return 100 // Default to 100% if no attributes are found
	}

	// Normalize values and calculate performance percentage
	throughput = normalizeValue(throughput)
	seekTime = normalizeValue(seekTime)
	latency = normalizeValue(latency)

	performancePercentage := (throughput + (100 - seekTime) + (100 - latency)) / 3
	if performancePercentage < 0 {
		performancePercentage = 0
	} else if performancePercentage > 100 {
		performancePercentage = 100
	}

	return performancePercentage
}

func extractAttributeValue(line string) int {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		value, err := strconv.Atoi(matches[1])
		if err == nil {
			return value
		}
	}
	return 0
}

func normalizeValue(value int) int {
	if value < 0 {
		return 0
	} else if value > 100 {
		return 100
	}
	return value
}

func displayRetroAnimation(performancePercentage int) {
	frames := []string{
		"  _______  \n /       \\ \n|         |\n \\_______/ ",
		"  _______  \n /       \\ \n|  O   O  |\n \\_______/ ",
		"  _______  \n /       \\ \n|  -   -  |\n \\_______/ ",
		"  _______  \n /       \\ \n|  ^   ^  |\n \\_______/ ",
	}

	color := "\033[34m" // Blue for performance
	fmt.Printf("%sDisk Performance: %d%%\n", color, performancePercentage)

	for i := 0; i < 10; i++ {
		for _, frame := range frames {
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Printf("%s%s\n", color, frame)
			drawProgressBar(performancePercentage)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func drawProgressBar(percentage int) {
	const barLength = 30
	filled := (percentage * barLength) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLength-filled)
	fmt.Printf("\n[%s] %d%%\n", bar, percentage)
}
