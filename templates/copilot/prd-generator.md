# PRD Generator (GitHub Copilot)

Generate a Product Requirements Document from a feature request.

## Purpose
Convert user feature ideas into structured, actionable PRDs for development.

## Input
User describes a feature, enhancement, or project they want built.

## Output Format
```markdown
# Project: [Name]
**Branch**: ralph/[feature-slug]

## Description
[Comprehensive feature description]

## User Stories

### US-001: [Concise Title]
**Description**: As a [persona], I need [capability] to [achieve outcome]...

**Acceptance Criteria**:
- [ ] Testable criterion 1
- [ ] Testable criterion 2
- [ ] Testable criterion 3

**Priority**: [1-5]

[Additional user stories...]
```

## Guidelines
- Decompose features into atomic, implementable user stories
- Each story should be completable in a single iteration
- Acceptance criteria must be concrete and verifiable
- Prioritize: 1=critical, 5=nice-to-have
- Technical clarity over business jargon
- Think like the developer who will implement it
