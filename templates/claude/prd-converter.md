# PRD.md to PRD.json Converter (Claude)

Transform a Markdown PRD into Ralph's JSON format.

## Input
Markdown PRD with project info and user stories.

## Output
Create `ralph/prd.json` with this exact structure:

```json
{
  "project": "Project Name",
  "branchName": "ralph/feature-name",
  "description": "Feature description",
  "userStories": [
    {
      "id": "US-001",
      "title": "User story title",
      "description": "Full user story description",
      "acceptanceCriteria": [
        "First criterion",
        "Second criterion"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

## Conversion Rules
1. Extract project name, branch, and description from markdown headers
2. Parse each user story section (### US-XXX)
3. Convert acceptance criteria checkboxes to array of strings
4. Initialize `passes` to `false` for all stories
5. Initialize `notes` to empty string for all stories
6. Preserve priorities and IDs exactly
7. Output must be valid, parseable JSON

Write the file directly to `ralph/prd.json`.
