package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed templates/config.yaml
var configTemplate string

//go:embed templates/claude/prompt.md
var claudePrompt string

//go:embed templates/copilot/prompt.md
var copilotPrompt string

//go:embed templates/skills/prd-generator.md
var prdGeneratorSkill string

//go:embed templates/skills/prd-converter.md
var prdConverterSkill string

type Config struct {
	Tool          string              `yaml:"tool"`
	MaxIterations int                 `yaml:"max_iterations"`
	AutoArchive   bool                `yaml:"auto_archive"`
	PromptFile    string              `yaml:"prompt_file"`
	ToolArgs      map[string][]string `yaml:"tool_args"`
}

type PRD struct {
	Project     string      `yaml:"project"`
	BranchName  string      `yaml:"branchName"`
	Description string      `yaml:"description"`
	UserStories []UserStory `yaml:"userStories"`
}

type UserStory struct {
	ID                 string   `yaml:"id"`
	Title              string   `yaml:"title"`
	Description        string   `yaml:"description"`
	AcceptanceCriteria []string `yaml:"acceptanceCriteria"`
	Priority           int      `yaml:"priority"`
	Passes             bool     `yaml:"passes"`
	Notes              string   `yaml:"notes"`
}

func main() {
	initMode := flag.Bool("init", false, "Initialize ralph directory with config and templates")
	tool := flag.String("tool", "", "Tool to use (required for --init): claude or copilot")
	maxIterations := flag.Int("max-iterations", 0, "Maximum number of iterations (overrides config)")
	flag.Parse()

	// Handle positional argument for max iterations (backwards compatibility)
	if flag.NArg() > 0 {
		if n, err := strconv.Atoi(flag.Arg(0)); err == nil {
			*maxIterations = n
		}
	}

	// Init mode
	if *initMode {
		if *tool == "" {
			fmt.Fprintf(os.Stderr, "Error: --tool flag is required for --init\n")
			fmt.Fprintf(os.Stderr, "Usage: go-ralph --init --tool=<claude|copilot>\n")
			os.Exit(1)
		}
		if *tool != "claude" && *tool != "copilot" {
			fmt.Fprintf(os.Stderr, "Error: Invalid tool '%s'. Must be 'claude' or 'copilot'.\n", *tool)
			os.Exit(1)
		}
		runInit(*tool)
		return
	}

	// Run mode - load config
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	ralphDir := filepath.Join(workDir, ".ralph")
	configFile := filepath.Join(ralphDir, "config.yaml")

	// Load config
	if !fileExists(configFile) {
		fmt.Fprintf(os.Stderr, "Error: .ralph/config.yaml not found\n")
		fmt.Fprintf(os.Stderr, "Run 'go-ralph --init --tool=<claude|copilot>' first to initialize\n")
		os.Exit(1)
	}

	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override max iterations if provided
	if *maxIterations > 0 {
		config.MaxIterations = *maxIterations
	}

	prdFile := filepath.Join(ralphDir, "prd.yaml")
	progressFile := filepath.Join(ralphDir, "progress.txt")
	archiveDir := filepath.Join(ralphDir, "archive")
	lastBranchFile := filepath.Join(ralphDir, ".last-branch")

	// Archive previous run if branch changed
	if fileExists(prdFile) && fileExists(lastBranchFile) {
		currentBranch := getBranchFromPRD(prdFile)
		lastBranch := readFile(lastBranchFile)

		if currentBranch != "" && lastBranch != "" && currentBranch != lastBranch {
			// Archive the previous run
			date := time.Now().Format("2006-01-02")
			folderName := strings.TrimPrefix(lastBranch, ".ralph/")
			archiveFolder := filepath.Join(archiveDir, date+"-"+folderName)

			fmt.Printf("Archiving previous run: %s\n", lastBranch)
			if err := os.MkdirAll(archiveFolder, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create archive folder: %v\n", err)
			} else {
				copyFile(prdFile, filepath.Join(archiveFolder, "prd.yaml"))
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

	fmt.Printf("Starting Ralph - Tool: %s - Max iterations: %d\n", config.Tool, config.MaxIterations)

	// Run iterations
	for i := 1; i <= config.MaxIterations; i++ {
		fmt.Println()
		fmt.Println("===============================================================")
		fmt.Printf("  Ralph Iteration %d of %d (%s)\n", i, config.MaxIterations, config.Tool)
		fmt.Println("===============================================================")

		// Get tool args from config
		args := config.ToolArgs[config.Tool]
		if args == nil {
			args = []string{}
		}

		// Run the selected tool with the ralph prompt
		output, err := runToolWithInput(ralphDir, config.Tool, args, config.PromptFile)

		// Continue even on error (|| true behavior)
		if err != nil {
			// Error already shown via tee to stderr
		}

		// Check for completion signal
		if strings.Contains(output, "<promise>COMPLETE</promise>") {
			fmt.Println()
			fmt.Println("Ralph completed all tasks!")
			fmt.Printf("Completed at iteration %d of %d\n", i, config.MaxIterations)
			os.Exit(0)
		}

		fmt.Printf("Iteration %d complete. Continuing...\n", i)
		time.Sleep(2 * time.Second)
	}

	fmt.Println()
	fmt.Printf("Ralph reached max iterations (%d) without completing all tasks.\n", config.MaxIterations)
	fmt.Printf("Check %s for status.\n", progressFile)
	os.Exit(1)
}

func runInit(tool string) {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	ralphDir := filepath.Join(workDir, ".ralph")
	fmt.Printf("Initializing Ralph for tool: %s\n\n", tool)

	// Create .ralph directory
	if err := os.MkdirAll(ralphDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .ralph directory: %v\n", err)
		os.Exit(1)
	}

	// Write config.yaml
	configPath := filepath.Join(ralphDir, "config.yaml")
	if !promptOverwrite(configPath) {
		fmt.Println("Skipped config.yaml")
	} else {
		configContent := strings.Replace(configTemplate, "{{.Tool}}", tool, 1)
		if err := writeFile(configPath, configContent); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config.yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Created .ralph/config.yaml")
	}

	// Write prompt.md
	promptPath := filepath.Join(ralphDir, "prompt.md")
	if !promptOverwrite(promptPath) {
		fmt.Println("Skipped prompt.md")
	} else {
		var promptContent string
		switch tool {
		case "claude":
			promptContent = claudePrompt
		case "copilot":
			promptContent = copilotPrompt
		}
		if err := writeFile(promptPath, promptContent); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing prompt.md: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Created .ralph/prompt.md")
	}

	// Create skills based on tool
	var skillsBaseDir string
	switch tool {
	case "copilot":
		skillsBaseDir = filepath.Join(workDir, ".github", "skills")
	case "claude":
		skillsBaseDir = filepath.Join(workDir, ".claude", "skills")
	}

	// Create prd-generator skill
	prdGenDir := filepath.Join(skillsBaseDir, "prd-generator")
	if err := os.MkdirAll(prdGenDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating prd-generator skill directory: %v\n", err)
		os.Exit(1)
	}

	prdGenPath := filepath.Join(prdGenDir, "SKILL.md")
	if !promptOverwrite(prdGenPath) {
		fmt.Println("Skipped prd-generator/SKILL.md")
	} else {
		if err := writeFile(prdGenPath, prdGeneratorSkill); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing prd-generator/SKILL.md: %v\n", err)
			os.Exit(1)
		}
		relPath := strings.TrimPrefix(prdGenPath, workDir+"/")
		fmt.Printf("✓ Created %s\n", relPath)
	}

	// Create prd-converter skill
	prdConvDir := filepath.Join(skillsBaseDir, "prd-converter")
	if err := os.MkdirAll(prdConvDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating prd-converter skill directory: %v\n", err)
		os.Exit(1)
	}

	prdConvPath := filepath.Join(prdConvDir, "SKILL.md")
	if !promptOverwrite(prdConvPath) {
		fmt.Println("Skipped prd-converter/SKILL.md")
	} else {
		if err := writeFile(prdConvPath, prdConverterSkill); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing prd-converter/SKILL.md: %v\n", err)
			os.Exit(1)
		}
		relPath := strings.TrimPrefix(prdConvPath, workDir+"/")
		fmt.Printf("✓ Created %s\n", relPath)
	}

	fmt.Println("\n✅ Ralph initialization complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Create your PRD in .ralph/prd.yaml")
	fmt.Println("2. Run: go-ralph")
}

func promptOverwrite(path string) bool {
	if !fileExists(path) {
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s already exists. Overwrite? (y/n): ", filepath.Base(path))
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
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
	if err := yaml.Unmarshal(data, &prd); err != nil {
		return ""
	}

	return prd.BranchName
}

func initProgressFile(path string) {
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC1123))
	writeFile(path, content)
}

func runToolWithInput(ralphDir, tool string, args []string, inputFile string) (string, error) {
	inputPath := filepath.Join(ralphDir, inputFile)

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
