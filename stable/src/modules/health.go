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

	// Step 4: Parse SMART test results and determine health
	healthStatus, healthPercentage := determineHealth(smartOutput)

	// Step 5: Display retro animation with health percentage and progress bar
	displayRetroAnimation(healthStatus, healthPercentage)
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

func determineHealth(smartOutput string) (string, int) {
	// Check if the drive passed or failed the SMART test
	if strings.Contains(smartOutput, "PASSED") {
		// Extract health percentage from SMART attributes
		healthPercentage := calculateHealthPercentage(smartOutput)
		return "healthy", healthPercentage
	} else if strings.Contains(smartOutput, "FAILED") {
		return "failing", 0
	} else {
		return "unknown", 50
	}
}

func calculateHealthPercentage(smartOutput string) int {
	// Parse SMART attributes to calculate health percentage
	lines := strings.Split(smartOutput, "\n")
	var rawValue, threshold int
	for _, line := range lines {
		if strings.Contains(line, "RAW_VALUE") && strings.Contains(line, "THRESHOLD") {
			// Extract RAW_VALUE and THRESHOLD from the line
			re := regexp.MustCompile(`RAW_VALUE\s+(\d+).*THRESHOLD\s+(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				rawValue, _ = strconv.Atoi(matches[1])
				threshold, _ = strconv.Atoi(matches[2])
				break
			}
		}
	}

	if threshold == 0 {
		return 100 // Avoid division by zero
	}

	// Calculate health percentage
	healthPercentage := 100 - ((rawValue - threshold) * 100 / threshold)
	if healthPercentage < 0 {
		healthPercentage = 0
	} else if healthPercentage > 100 {
		healthPercentage = 100
	}

	return healthPercentage
}

func displayRetroAnimation(healthStatus string, healthPercentage int) {
	frames := []string{
		"  _______  \n /       \\ \n|         |\n \\_______/ ",
		"  _______  \n /       \\ \n|  O   O  |\n \\_______/ ",
		"  _______  \n /       \\ \n|  -   -  |\n \\_______/ ",
		"  _______  \n /       \\ \n|  ^   ^  |\n \\_______/ ",
	}

	colors := map[string]string{
		"healthy": "\033[32m", // Green
		"failing": "\033[31m", // Red
		"unknown": "\033[33m", // Yellow
	}

	color := colors[healthStatus]
	fmt.Printf("%sDrive Health: %s (%d%%)\n", color, strings.ToUpper(healthStatus), healthPercentage)

	for i := 0; i < 10; i++ {
		for _, frame := range frames {
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Printf("%s%s\n", color, frame)
			drawProgressBar(healthPercentage)
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
