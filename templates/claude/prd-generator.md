# PRD Generator (Claude)

Create a structured Product Requirements Document from a feature request.

## Input
User provides a feature idea, enhancement, or project description.

## Output Format
Generate a PRD in Markdown with this structure:

```markdown
# Project: [Project Name]
**Branch**: ralph/[feature-slug]

## Description
[Detailed feature description]

## User Stories

### US-001: [Title]
**Description**: As a [role], I want [goal] so that [benefit]...

**Acceptance Criteria**:
- [ ] Specific criterion 1
- [ ] Specific criterion 2
- [ ] Specific criterion 3

**Priority**: [1-5]

### US-002: [Title]
...
```

## Guidelines
- Break down features into small, testable user stories
- Each story should be completable in one development iteration
- Acceptance criteria must be specific and verifiable
- Priority 1 = highest, 5 = lowest
- Use clear, technical language for developers
