package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configContent := `tool: claude
max_iterations: 5
auto_archive: true
prompt_file: prompt.md
tool_args:
  claude:
    - --arg1
    - --arg2
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		config, err := loadConfig(configPath)
		if err != nil {
			t.Fatalf("loadConfig failed: %v", err)
		}

		if config.Tool != "claude" {
			t.Errorf("Expected tool 'claude', got '%s'", config.Tool)
		}
		if config.MaxIterations != 5 {
			t.Errorf("Expected max_iterations 5, got %d", config.MaxIterations)
		}
		if !config.AutoArchive {
			t.Error("Expected auto_archive true, got false")
		}
		if config.PromptFile != "prompt.md" {
			t.Errorf("Expected prompt_file 'prompt.md', got '%s'", config.PromptFile)
		}
		if len(config.ToolArgs["claude"]) != 2 {
			t.Errorf("Expected 2 args for claude, got %d", len(config.ToolArgs["claude"]))
		}
	})

	t.Run("invalid config file", func(t *testing.T) {
		_, err := loadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error for non-existent config file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		invalidYaml := "invalid: yaml: content:\n  bad indentation"
		if err := os.WriteFile(configPath, []byte(invalidYaml), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		_, err := loadConfig(configPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if !fileExists(testFile) {
			t.Error("Expected fileExists to return true for existing file")
		}
	})

	t.Run("non-existing file", func(t *testing.T) {
		if fileExists("/nonexistent/file.txt") {
			t.Error("Expected fileExists to return false for non-existing file")
		}
	})
}

func TestReadFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "test content\n  with whitespace  \n"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := readFile(testFile)
		expected := "test content\n  with whitespace"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		result := readFile("/nonexistent/file.txt")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content"

	err := writeFile(testFile, content)
	if err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected '%s', got '%s'", content, string(data))
	}
}

func TestCopyFile(t *testing.T) {
	t.Run("successful copy", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "dest.txt")
		content := "test content"

		if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		err := copyFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("copyFile failed: %v", err)
		}

		data, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected '%s', got '%s'", content, string(data))
		}
	})

	t.Run("non-existent source", func(t *testing.T) {
		tmpDir := t.TempDir()
		dstFile := filepath.Join(tmpDir, "dest.txt")

		err := copyFile("/nonexistent/file.txt", dstFile)
		if err == nil {
			t.Error("Expected error when copying non-existent file")
		}
	})
}

func TestGetBranchFromPRD(t *testing.T) {
	t.Run("valid PRD file", func(t *testing.T) {
		tmpDir := t.TempDir()
		prdFile := filepath.Join(tmpDir, "prd.json")

		prd := PRD{
			Project:     "Test Project",
			BranchName:  "feature/test-branch",
			Description: "Test description",
		}

		data, err := json.Marshal(prd)
		if err != nil {
			t.Fatalf("Failed to marshal PRD: %v", err)
		}

		if err := os.WriteFile(prdFile, data, 0644); err != nil {
			t.Fatalf("Failed to create PRD file: %v", err)
		}

		branch := getBranchFromPRD(prdFile)
		if branch != "feature/test-branch" {
			t.Errorf("Expected 'feature/test-branch', got '%s'", branch)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		branch := getBranchFromPRD("/nonexistent/prd.json")
		if branch != "" {
			t.Errorf("Expected empty string, got '%s'", branch)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		prdFile := filepath.Join(tmpDir, "prd.json")

		if err := os.WriteFile(prdFile, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to create PRD file: %v", err)
		}

		branch := getBranchFromPRD(prdFile)
		if branch != "" {
			t.Errorf("Expected empty string for invalid JSON, got '%s'", branch)
		}
	})
}

func TestInitProgressFile(t *testing.T) {
	tmpDir := t.TempDir()
	progressFile := filepath.Join(tmpDir, "progress.txt")

	initProgressFile(progressFile)

	data, err := os.ReadFile(progressFile)
	if err != nil {
		t.Fatalf("Failed to read progress file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Ralph Progress Log") {
		t.Error("Expected progress file to contain header")
	}
	if !strings.Contains(content, "Started:") {
		t.Error("Expected progress file to contain start time")
	}
	if !strings.Contains(content, "---") {
		t.Error("Expected progress file to contain separator")
	}
}

func TestConfigStruct(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		config := Config{
			Tool:          "copilot",
			MaxIterations: 10,
			AutoArchive:   false,
			PromptFile:    "custom.md",
			ToolArgs: map[string][]string{
				"copilot": {"--arg1", "--arg2"},
			},
		}

		data, err := yaml.Marshal(&config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		var unmarshaled Config
		if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		if unmarshaled.Tool != config.Tool {
			t.Errorf("Expected tool '%s', got '%s'", config.Tool, unmarshaled.Tool)
		}
		if unmarshaled.MaxIterations != config.MaxIterations {
			t.Errorf("Expected max_iterations %d, got %d", config.MaxIterations, unmarshaled.MaxIterations)
		}
		if unmarshaled.AutoArchive != config.AutoArchive {
			t.Errorf("Expected auto_archive %v, got %v", config.AutoArchive, unmarshaled.AutoArchive)
		}
		if unmarshaled.PromptFile != config.PromptFile {
			t.Errorf("Expected prompt_file '%s', got '%s'", config.PromptFile, unmarshaled.PromptFile)
		}
	})
}

func TestPRDStruct(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		prd := PRD{
			Project:     "Test Project",
			BranchName:  "feature/test",
			Description: "Test description",
			UserStories: []UserStory{
				{
					ID:          "US-1",
					Title:       "Test Story",
					Description: "Test description",
					AcceptanceCriteria: []string{
						"Criteria 1",
						"Criteria 2",
					},
					Priority: 1,
					Passes:   false,
					Notes:    "Test notes",
				},
			},
		}

		data, err := json.Marshal(&prd)
		if err != nil {
			t.Fatalf("Failed to marshal PRD: %v", err)
		}

		var unmarshaled PRD
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal PRD: %v", err)
		}

		if unmarshaled.Project != prd.Project {
			t.Errorf("Expected project '%s', got '%s'", prd.Project, unmarshaled.Project)
		}
		if unmarshaled.BranchName != prd.BranchName {
			t.Errorf("Expected branch '%s', got '%s'", prd.BranchName, unmarshaled.BranchName)
		}
		if len(unmarshaled.UserStories) != 1 {
			t.Errorf("Expected 1 user story, got %d", len(unmarshaled.UserStories))
		}
		if unmarshaled.UserStories[0].ID != "US-1" {
			t.Errorf("Expected story ID 'US-1', got '%s'", unmarshaled.UserStories[0].ID)
		}
	})
}

func TestUserStoryStruct(t *testing.T) {
	story := UserStory{
		ID:          "US-123",
		Title:       "Test Title",
		Description: "Test Description",
		AcceptanceCriteria: []string{
			"AC1",
			"AC2",
			"AC3",
		},
		Priority: 2,
		Passes:   true,
		Notes:    "Test notes",
	}

	data, err := json.Marshal(&story)
	if err != nil {
		t.Fatalf("Failed to marshal UserStory: %v", err)
	}

	var unmarshaled UserStory
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal UserStory: %v", err)
	}

	if unmarshaled.ID != story.ID {
		t.Errorf("Expected ID '%s', got '%s'", story.ID, unmarshaled.ID)
	}
	if unmarshaled.Priority != story.Priority {
		t.Errorf("Expected priority %d, got %d", story.Priority, unmarshaled.Priority)
	}
	if unmarshaled.Passes != story.Passes {
		t.Errorf("Expected passes %v, got %v", story.Passes, unmarshaled.Passes)
	}
	if len(unmarshaled.AcceptanceCriteria) != 3 {
		t.Errorf("Expected 3 acceptance criteria, got %d", len(unmarshaled.AcceptanceCriteria))
	}
}

func TestEmbeddedTemplates(t *testing.T) {
	t.Run("config template exists", func(t *testing.T) {
		if configTemplate == "" {
			t.Error("Expected configTemplate to be embedded")
		}
		if !strings.Contains(configTemplate, "tool:") {
			t.Error("Expected configTemplate to contain 'tool:' field")
		}
	})

	t.Run("claude prompt exists", func(t *testing.T) {
		if claudePrompt == "" {
			t.Error("Expected claudePrompt to be embedded")
		}
	})

	t.Run("copilot prompt exists", func(t *testing.T) {
		if copilotPrompt == "" {
			t.Error("Expected copilotPrompt to be embedded")
		}
	})

	t.Run("prd generator skill exists", func(t *testing.T) {
		if prdGeneratorSkill == "" {
			t.Error("Expected prdGeneratorSkill to be embedded")
		}
	})

	t.Run("prd converter skill exists", func(t *testing.T) {
		if prdConverterSkill == "" {
			t.Error("Expected prdConverterSkill to be embedded")
		}
	})
}

func TestRunToolWithInputError(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with non-existent input file
	_, err := runToolWithInput(tmpDir, "echo", []string{}, "nonexistent.txt")
	if err == nil {
		t.Error("Expected error when input file doesn't exist")
	}
}

func TestProgressFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	progressFile := filepath.Join(tmpDir, "progress.txt")

	// Record the time before creating the file
	beforeTime := time.Now()

	initProgressFile(progressFile)

	data, err := os.ReadFile(progressFile)
	if err != nil {
		t.Fatalf("Failed to read progress file: %v", err)
	}

	content := string(data)

	// Check structure
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}

	if lines[0] != "# Ralph Progress Log" {
		t.Errorf("Expected first line to be '# Ralph Progress Log', got '%s'", lines[0])
	}

	if !strings.HasPrefix(lines[1], "Started:") {
		t.Errorf("Expected second line to start with 'Started:', got '%s'", lines[1])
	}

	if lines[2] != "---" {
		t.Errorf("Expected third line to be '---', got '%s'", lines[2])
	}

	// Verify the timestamp is reasonable (within a few seconds)
	afterTime := time.Now()
	if !beforeTime.Before(afterTime) || afterTime.Sub(beforeTime) > 5*time.Second {
		t.Error("Progress file creation took too long")
	}
}
