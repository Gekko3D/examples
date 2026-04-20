# Available Skills

> Auto-generated during onboarding/upgrade. Reflects all skills in `.ai/skills/` and `.ai.local/skills/`.

## Org Skills

| Skill | Trigger | Description |
|---|---|---|
| onboard-project | "onboard", "initialize", "set up AI" | Bootstrap and populate AI code assist docs from codebase |
| start-feature | "work on [ticket]", "start feature" | Fetch Jira ticket, create branch, write plan, await approval |
| complete-task | "complete task", "wrap up", "finish" | Run tests, changelog, commit, push, transition ticket |
| resume-work | "continue work on [ticket]", "resume" | Restore context from feature file, pick up from Next Steps |
| pr-review | "review PR", "code review" | Structured code review with severity and merge recommendation |
| apex-security-scan | "scan apex", "security check", "check for XSS", "check for SQL injection", "scan before PR" | Static security scan of APEX split SQL export. Detects SQLi, XSS, URL tampering, misconfig. No database required. |
| rebase-docs | "rebase [ticket]", "merge docs" | Apply feature context deltas into main docs after PR merge |
| check-doc-tracking | "check docs", "verify tracking" | Compare git diff against feature context, flag gaps |
| upgrade-ai-codeassist | "upgrade AI", "update framework" | Update framework files, preserve project content |
| init-ai-codeassist | "initialize AI from scratch" | Create entire .ai/ structure and onboard |
| plsql-logger-instrumentation | "instrument this package with logger", "add logger calls", "add trace logging" | Add portable Logger instrumentation to Oracle PL/SQL while adapting to local repo conventions and hard-blocking secret logging. |
| jira-add-ticket-to-sprint | "add [ticket] to sprint", "put [ticket] in sprint", "assign [ticket] to sprint" | Add a Jira issue to an open sprint by discovering scrum boards and open sprints, then prompting for selection when multiple candidates exist. |

## Project Skills

| Skill | Trigger | Description |
|---|---|---|
| _(none yet — add project-specific skills to `.ai.local/skills/`)_ | | |

## How to Use

Say the trigger phrase in your AI tool (Cline or Codex) and the skill will execute.

For the full skill instructions, read the SKILL.md file in the corresponding directory.
