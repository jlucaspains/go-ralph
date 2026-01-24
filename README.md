# go-ralph

Go implementation of Ralph Wiggum - a long-running AI agent loop.

## Overview

Ralph runs AI tools (amp, claude, or copilot) in a loop until completion or max iterations, tracking progress and archiving results.

## Installation

```bash
go build -o go-ralph
```

Or install directly:

```bash
go install
```

## Usage

```bash
./go-ralph [--tool amp|claude|copilot] [--max-iterations N]

# Examples:
./go-ralph                           # Use amp, 10 iterations (default)
./go-ralph --tool claude             # Use claude, 10 iterations
./go-ralph --max-iterations 20       # Use amp, 20 iterations
./go-ralph --tool copilot 15         # Use copilot, 15 iterations (positional arg)
```

## Options

- `--tool`: AI tool to use (`amp`, `claude`, or `copilot`). Default: `amp`
- `--max-iterations`: Maximum iterations before stopping. Default: `10`

## Requirements

- The selected AI tool must be installed and available in PATH:
  - `amp` for --tool amp
  - `claude` for --tool claude  
  - `copilot` for --tool copilot

- Input files in the same directory as the binary:
  - `prompt.md` for amp
  - `CLAUDE.md` for claude
  - `COPILOT.md` for copilot
  - `prd.json` (optional) for branch tracking

## Features

- **Branch-based archiving**: Automatically archives progress when switching branches
- **Progress tracking**: Maintains progress.txt with run history
- **Completion detection**: Stops early when `<promise>COMPLETE</promise>` is detected
- **Real-time output**: Shows tool output as it runs (tee behavior)
- **Error tolerance**: Continues on tool failures

## Files

- `prd.json` - Project requirements document with branch info
- `progress.txt` - Progress log for current run
- `archive/` - Archived runs organized by date and branch
- `.last-branch` - Tracks last branch for archive detection
