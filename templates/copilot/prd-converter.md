# PRD.md to PRD.json Converter (GitHub Copilot)

Convert Markdown PRDs to Ralph's JSON format.

## Input
Markdown PRD from the PRD Generator skill.

## Output
Generate `ralph/prd.json` with structure:

```json
{
  "project": "Project Name",
  "branchName": "ralph/feature-slug",
  "description": "Feature description text",
  "userStories": [
    {
      "id": "US-001",
      "title": "Story title",
      "description": "Complete story description",
      "acceptanceCriteria": [
        "Criterion text 1",
        "Criterion text 2"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
```

## Transformation Rules
1. Parse markdown headers for project, branch, description
2. Extract each user story (### US-XXX sections)
3. Convert bullet points under "Acceptance Criteria" to string array
4. Set `passes: false` for all stories (initial state)
5. Set `notes: ""` for all stories (empty initially)
6. Preserve all IDs, priorities, and content exactly
7. Ensure output is valid JSON

Write directly to `ralph/prd.json` in the repository.
