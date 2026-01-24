# PRD.md to PRD.json Converter

Convert a Markdown PRD to the structured JSON format used by Ralph.

## Input
A PRD in Markdown format (from PRD Generator skill).

## Output
A `ralph/prd.json` file with structure:
```json
{
  "project": "Project Name",
  "branchName": "ralph/feature-name",
  "description": "Feature description",
  "userStories": [
    {
      "id": "US-001",
      "title": "Story title",
      "description": "Story description",
      "acceptanceCriteria": ["Criterion 1", "Criterion 2"],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

## Rules
- Set all `passes` fields to `false` initially
- Set all `notes` fields to empty strings
- Preserve exact IDs, priorities, and criteria from markdown
- Output valid JSON only
