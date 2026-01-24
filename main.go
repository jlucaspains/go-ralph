package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PRD struct {
	BranchName string `json:"branchName"`
}

func main() {
	tool := flag.String("tool", "amp", "Tool to use: amp, claude, or copilot")
	maxIterations := flag.Int("max-iterations", 10, "Maximum number of iterations")
	flag.Parse()

	// Handle positional argument for max iterations (backwards compatibility)
	if flag.NArg() > 0 {
		if n, err := strconv.Atoi(flag.Arg(0)); err == nil {
			*maxIterations = n
		}
	}

	// Validate tool choice
	if *tool != "amp" && *tool != "claude" && *tool != "copilot" {
		fmt.Fprintf(os.Stderr, "Error: Invalid tool '%s'. Must be 'amp', 'claude', or 'copilot'.\n", *tool)
		os.Exit(1)
	}

	scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		scriptDir, _ = os.Getwd()
	}

	prdFile := filepath.Join(scriptDir, "prd.json")
	progressFile := filepath.Join(scriptDir, "progress.txt")
	archiveDir := filepath.Join(scriptDir, "archive")
	lastBranchFile := filepath.Join(scriptDir, ".last-branch")

	// Archive previous run if branch changed
	if fileExists(prdFile) && fileExists(lastBranchFile) {
		currentBranch := getBranchFromPRD(prdFile)
		lastBranch := readFile(lastBranchFile)

		if currentBranch != "" && lastBranch != "" && currentBranch != lastBranch {
			// Archive the previous run
			date := time.Now().Format("2006-01-02")
			folderName := strings.TrimPrefix(lastBranch, "ralph/")
			archiveFolder := filepath.Join(archiveDir, date+"-"+folderName)

			fmt.Printf("Archiving previous run: %s\n", lastBranch)
			if err := os.MkdirAll(archiveFolder, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create archive folder: %v\n", err)
			} else {
				copyFile(prdFile, filepath.Join(archiveFolder, "prd.json"))
				copyFile(progressFile, filepath.Join(archiveFolder, "progress.txt"))
				fmt.Printf("   Archived to: %s\n", archiveFolder)
			}

			// Reset progress file for new run
			initProgressFile(progressFile)
		}
	}

	// Track current branch
	if fileExists(prdFile) {
		currentBranch := getBranchFromPRD(prdFile)
		if currentBranch != "" {
			writeFile(lastBranchFile, currentBranch)
		}
	}

	// Initialize progress file if it doesn't exist
	if !fileExists(progressFile) {
		initProgressFile(progressFile)
	}

	fmt.Printf("Starting Ralph - Tool: %s - Max iterations: %d\n", *tool, *maxIterations)

	// Run iterations
	for i := 1; i <= *maxIterations; i++ {
		fmt.Println()
		fmt.Println("===============================================================")
		fmt.Printf("  Ralph Iteration %d of %d (%s)\n", i, *maxIterations, *tool)
		fmt.Println("===============================================================")

		// Run the selected tool with the ralph prompt
		var output string
		var err error

		switch *tool {
		case "amp":
			output, err = runToolWithInput(scriptDir, "amp", []string{"--dangerously-allow-all"}, "prompt.md")
		case "claude":
			output, err = runToolWithInput(scriptDir, "claude", []string{"--dangerously-skip-permissions", "--print"}, "CLAUDE.md")
		case "copilot":
			output, err = runToolWithInput(scriptDir, "copilot", []string{"--allow-all-tools"}, "COPILOT.md")
		}

		// Continue even on error (|| true behavior)
		if err != nil {
			// Error already shown via tee to stderr
		}

		// Check for completion signal
		if strings.Contains(output, "<promise>COMPLETE</promise>") {
			fmt.Println()
			fmt.Println("Ralph completed all tasks!")
			fmt.Printf("Completed at iteration %d of %d\n", i, *maxIterations)
			os.Exit(0)
		}

		fmt.Printf("Iteration %d complete. Continuing...\n", i)
		time.Sleep(2 * time.Second)
	}

	fmt.Println()
	fmt.Printf("Ralph reached max iterations (%d) without completing all tasks.\n", *maxIterations)
	fmt.Printf("Check %s for status.\n", progressFile)
	os.Exit(1)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func getBranchFromPRD(prdFile string) string {
	data, err := os.ReadFile(prdFile)
	if err != nil {
		return ""
	}

	var prd PRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return ""
	}

	return prd.BranchName
}

func initProgressFile(path string) {
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	writeFile(path, content)
}

func runToolWithInput(scriptDir, tool string, args []string, inputFile string) (string, error) {
	inputPath := filepath.Join(scriptDir, inputFile)

	// Read input file
	input, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", inputFile, err)
		return "", err
	}

	// Create command
	cmd := exec.Command(tool, args...)
	cmd.Stdin = bytes.NewReader(input)

	// Capture output while displaying it (tee behavior)
	var outputBuf bytes.Buffer
	multiWriter := io.MultiWriter(os.Stdout, &outputBuf)
	multiErrWriter := io.MultiWriter(os.Stderr, &outputBuf)

	cmd.Stdout = multiWriter
	cmd.Stderr = multiErrWriter

	// Run command
	err = cmd.Run()

	return outputBuf.String(), err
}
