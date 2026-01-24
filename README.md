# go-ralph

Go implementation of Ralph - an autonomous AI agent loop for software development. This project was heavily inspired by [snarktank/ralph](https://github.com/snarktank/ralph)

## Overview

Ralph is a long-running autonomous agent that executes AI-driven development workflows using Claude Desktop or GitHub Copilot CLI. It works in iterations, implementing user stories from a PRD (Product Requirements Document), tracking progress, and automatically archiving completed work.

Ralph enables autonomous development by:
- Reading a structured PRD with user stories
- Implementing stories one at a time across multiple iterations
- Running quality checks (tests, linting, type checking)
- Committing changes when checks pass
- Tracking progress and learnings for future iterations
- Archiving work when switching between projects/branches

## Installation

> go-ralph requires go 1.25+ to build or install.

### Install

```bash
go install github.com/jlucaspains/go-ralph@main
```

## Quick Start

1. **Initialize Ralph for your project:**

```bash
cd /path/to/your/project
go-ralph --init --tool=claude    # Or --tool=copilot
```

This creates:
- `.ralph/config.yaml` - Ralph configuration
- `.ralph/prompt.md` - Agent instructions for the selected tool
- `.github/skills/` or `.claude/skills/` - PRD generator and converter skills

2. **Create a PRD:**

Create `.ralph/prd.json` with your project requirements (or use the prd-generator and prd-converter skills to create one from a description):

```json
{
  "project": "Add User Authentication",
  "branchName": ".ralph/add-auth",
  "description": "Implement basic user authentication with login and signup",
  "userStories": [
    {
      "id": "US-001",
      "title": "Create login form",
      "description": "Create a login form component with email and password fields",
      "acceptanceCriteria": [
        "Form renders with email and password fields",
        "Form validates input before submission",
        "Tests pass"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

3. **Run Ralph:**

```bash
go-ralph
```

Ralph will iterate through user stories, implementing them one at a time until complete or max iterations reached.

## Usage

```bash
go-ralph [--max-iterations N]

# Examples:
go-ralph                         # Use default config, 10 iterations
go-ralph --max-iterations 20     # Override to 20 iterations
go-ralph 15                      # Positional arg also works
```

## Configuration

Configuration is stored in `.ralph/config.yaml`:

```yaml
tool: claude                    # AI tool: claude or copilot
max_iterations: 10              # Maximum iterations before stopping
auto_archive: true              # Auto-archive on branch change
prompt_file: prompt.md          # Agent instructions file
tool_args:
  claude:
    - "--dangerously-skip-permissions"
    - "--print"
  copilot:
    - "--allow-all-tools"
```

### Options

- `--init` - Initialize Ralph in the current project (requires `--tool`)
- `--tool` - Select AI tool: `claude` or `copilot` (required for `--init`)
- `--max-iterations` - Maximum iterations before stopping (overrides config)

## Requirements

### AI Tools

One of the following AI tools must be installed and available in PATH:

- **Claude Desktop** (`claude`) - [Install Claude Desktop](https://claude.ai/download)
- **GitHub Copilot CLI** (`copilot`) - [Install GitHub Copilot CLI](https://docs.github.com/en/copilot/using-github-copilot/using-github-copilot-in-the-command-line)

### Project Structure

Ralph expects:
- `.ralph/config.yaml` - Configuration (created by `--init`)
- `.ralph/prompt.md` - Agent instructions (created by `--init`)
- `.ralph/prd.json` - Product requirements document (you create this)

## How It Works

Each Ralph iteration:

1. **Reads the PRD** at `.ralph/prd.json`
2. **Reads progress log** at `.ralph/progress.txt` to understand context
3. **Checks the git branch** matches the PRD's `branchName`
4. **Picks the highest priority user story** where `passes: false`
5. **Implements the story** - writes code, makes changes
6. **Runs quality checks** - tests, linting, type checking
7. **Commits changes** if checks pass with message: `[feat|chore|etc]: [Story ID] - [Story Title]`
8. **Updates the PRD** to set `passes: true` for the completed story
9. **Appends to progress.txt** with implementation details and learnings

### Completion

When all user stories have `passes: true`, Ralph detects `<promise>COMPLETE</promise>` in the agent output and exits successfully.

If max iterations is reached without completion, Ralph exits with an error code.

## Features

### 🔄 Automatic Archiving

When switching projects/branches, Ralph automatically archives the previous run:
- Detects branch changes by reading `branchName` from `prd.json`
- Archives `.ralph/prd.json` and `.ralph/progress.txt` to `.ralph/archive/YYYY-MM-DD-branch-name/`
- Creates fresh progress log for the new work

### 📝 Progress Tracking

Ralph maintains `.ralph/progress.txt` with:
- Timestamp for each iteration
- Story ID being worked on
- Files changed
- **Learnings** - patterns discovered, gotchas, useful context for future iterations

Example progress entry:
```markdown
## 2026-01-24 10:30 - US-002
- Implemented user registration API endpoint
- Files changed: api/auth.ts, types/user.ts
- **Learnings for future iterations:**
  - This codebase uses Zod for validation in all API routes
  - Auth middleware is in middleware/auth.ts
  - Database queries use Prisma client from db/client.ts
---
```

### 🎯 Completion Detection

Ralph stops early when it detects `<promise>COMPLETE</promise>` in the agent output, indicating all user stories are complete.

### 🔧 Real-time Output

Shows tool output in real-time as the agent works (tee behavior).

### 💪 Error Tolerance

Continues on tool failures (exit code is not fatal), allowing for retries across iterations.

## PRD Format

The `prd.json` file defines what Ralph should build:

```json
{
  "project": "Project Name",
  "branchName": "ralph/feature-name",
  "description": "High-level description of the project",
  "userStories": [
    {
      "id": "US-001",
      "title": "Short story title",
      "description": "Detailed description of what needs to be built",
      "acceptanceCriteria": [
        "Specific, testable criterion 1",
        "Specific, testable criterion 2"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

**Fields:**
- `project` - Project name
- `branchName` - Git branch to work on (format: `ralph/{feature-name}`)
- `description` - Project description
- `userStories` - Array of user stories to implement
  - `id` - Unique ID (US-001, US-002, etc.)
  - `title` - Short title
  - `description` - Detailed description
  - `acceptanceCriteria` - Array of specific, testable criteria
  - `priority` - Priority (1-5, where 1 is highest)
  - `passes` - Boolean flag (Ralph sets to `true` when complete)
  - `notes` - Additional notes (Ralph may add context here)

## Skills

Ralph initialization creates two skills in your project:

### PRD Generator

Located in `.github/skills/prd-generator/` or `.claude/skills/prd-generator/`

Generates a structured PRD in Markdown format from a feature description.

### PRD Converter

Located in `.github/skills/prd-converter/` or `.claude/skills/prd-converter/`

Converts a Markdown PRD to the JSON format Ralph expects.

## Files Created

- `.ralph/config.yaml` - Ralph configuration
- `.ralph/prompt.md` - Agent instructions
- `.ralph/prd.json` - Project requirements (you create)
- `.ralph/progress.txt` - Progress log
- `.ralph/archive/` - Archived runs organized by date and branch
- `.ralph/.last-branch` - Tracks last branch for archive detection

## Tips

- **Keep user stories small and focused** - Each story should be completable in one iteration
- **Write specific acceptance criteria** - Make them testable and unambiguous
- **Review progress.txt** - Learn from previous iterations' discoveries
- **Use skills** - Generate PRDs with the prd-generator skill for consistency
- **Monitor iterations** - If Ralph gets stuck, adjust the story or split it into smaller pieces