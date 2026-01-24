# PRD Generator (AMP)

Generate a structured Product Requirements Document (PRD) in Markdown format.

## Input
The user provides a feature request or project description.

## Output
A comprehensive PRD.md with:
- Project name and branch name (format: `ralph/{feature-name}`)
- Description
- User stories with:
  - ID (US-001, US-002, etc.)
  - Title
  - Description
  - Acceptance criteria (specific, testable)
  - Priority (1-5)

## Example Output Structure
```markdown
# Project: Recipe Book PDF Export
**Branch**: ralph/recipe-book-pdf-export

## Description
[Feature description]

## User Stories

### US-001: Install PDF dependencies
**Description**: As a developer, I need to add required libraries...
**Acceptance Criteria**:
- [ ] Add jspdf to dependencies
- [ ] Run install successfully
- [ ] Typecheck passes

**Priority**: 1
```

Keep user stories granular and actionable.
