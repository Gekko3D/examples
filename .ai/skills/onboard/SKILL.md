---
name: onboard-project
description: Bootstrap and populate AI code assist documentation from an existing codebase. Analyzes source code, package manifests, CI/CD configs, and infrastructure files to populate architecture, dependency, deployment, and convention docs. Generates tool-specific rules file.
---

# Skill: Onboard Project to AI Code Assist

## Description
Bootstrap a project for AI-assisted development. Analyzes the codebase and populates the `.ai/` documentation structure so AI agents and engineers can understand the project immediately.

## When to Invoke
- Engineer says "onboard this project" or "bootstrap AI docs" or "initialize AI code assist"
- Engineer says "set up .ai/ for this project"
- The `.ai/context.md` file contains only template placeholders

---

## Instructions

You are onboarding this project for AI-assisted development. Follow every step below. Do NOT use shell scripts. Do everything by reading files, analyzing code, and writing documentation directly.

### Phase 1: Detect Environment

1. Determine which AI tool you are running inside:
   - If `.clinerules` is being read → you are **Cline**
   - If `AGENTS.md` is being read → you are **Codex**
   - If neither is clear, ask the engineer which tool they are using

2. Note the tool — you will generate the correct rules file at the end.

### Phase 2: Analyze the Codebase

Read and analyze the following (skip any that don't exist):

**Package manifests** — determine language, framework, and dependencies:
- `package.json`, `package-lock.json`
- `pom.xml`, `build.gradle`
- `requirements.txt`, `pyproject.toml`, `setup.py`
- `go.mod`, `Cargo.toml`, `Gemfile`

**Source code structure** — understand the architecture:
- Scan the top-level directory structure
- Identify the main source directories (src/, lib/, app/, cmd/)
- Read 4-5 representative source files to extract coding patterns
- Identify the entry point (main.js, Application.java, main.py, etc.)
- Map the layers: controllers/routes, services/business logic, data access, models

**Configuration files** — understand the runtime:
- Application configs (application.yml, .env.example, config/)
- Environment variable references in code
- Feature flags, secrets references (note locations, never read actual secrets)

**Infrastructure and deployment** — understand how it ships:
- CI/CD pipelines (Jenkinsfile, .github/workflows/, .gitlab-ci.yml, buildspec.yml)
- Container configs (Dockerfile, docker-compose.yml)
- Kubernetes manifests (Helm charts, k8s/, deploy/)
- IaC (Terraform, Pulumi, CloudFormation)

**Existing documentation** — use as input:
- README.md (primary source of truth)
- Any docs/ directory
- CHANGELOG.md
- Existing architecture diagrams or wiki links

**Tests** — understand testing approach:
- Test directory structure
- Test framework being used
- Test naming conventions

### Phase 3: Populate Documentation

Using what you found, fill in each template file. Replace ALL placeholder text. For anything you cannot determine, use `[NEEDS VERIFICATION — could not determine from codebase]`.

#### 3.1: `.ai/docs/PROJECT.md`
Fill in from README.md and package manifests:
- Project name, one-liner, description
- Business context (if available in README, otherwise mark for human input)
- Tech stack table (from package manifests)
- Key users and stakeholders (mark for human input if not in README)

#### 3.2: `.ai/docs/ARCHITECTURE.md`
Fill in from source code analysis:
- System overview paragraph
- Mermaid diagram showing components and their relationships
- Core components table (name, purpose, source path)
- Primary data flow description
- Security model (from auth middleware, config references)
- Scalability approach (from deployment configs)

#### 3.3: `.ai/docs/DEPENDENCIES.md`
Fill in from package manifests, source code, and configs:
- Mermaid dependency map (upstream/downstream)
- Upstream dependencies table (services this project calls)
- Downstream dependencies table (what consumes this project)
- Infrastructure dependencies (databases, caches, queues — from configs and code)
- Critical libraries table (only architecture-impacting ones)
- Environment configuration keys (from config files and code)

#### 3.4: `.ai/docs/DEPLOYMENT.md`
Fill in from CI/CD and infrastructure configs:
- Deployment target and strategy
- Environment promotion flow (with Mermaid diagram if possible)
- Pipeline stages
- Kubernetes resources (if applicable)
- Secrets and configuration table
- Health check endpoints (from code)

#### 3.5: `.ai/context.md`
Synthesize from all the docs you just wrote:
- Condensed project summary (under 1500 words)
- Fill in Key Entry Points table with actual file paths
- Current state summary

#### 3.6: `.ai/conventions.md`
Extract from the source files you analyzed:
- Actual language, framework, linter, formatter in use
- Naming conventions as used in the code (not theoretical)
- Patterns: API layer, service layer, data access, error handling, logging
- Specific examples from real files
- "Things AI agents should AVOID" based on observed patterns

#### 3.7: `.ai/known-issues.md`
- Mark the entire file as `[NEEDS HUMAN INPUT — only the team knows the gotchas]`
- If you found any TODO/FIXME/HACK comments in the code, list them as starting points

### Phase 3.8: Detect Testing Setup
Analyze the project's testing configuration:
- Check package manifest for test scripts and test frameworks (jest, mocha, pytest, junit, go test)
- Look for test directories (test/, tests/, __tests__, src/test/)
- Check for test config files (jest.config.js, pytest.ini, etc.)
- Note: if no test framework exists, record this — the RULES.md Testing section will say "no test framework configured yet"

### Phase 3.9: Detect MCP / Ticket Tracker Integration
Check if the project environment has a ticket tracker MCP server:
- Look for MCP config references (mcp-atlassian, mcp-linear, etc.)
- Check for Jira/Linear/GitHub Issues references in existing docs or configs
- Ask the engineer: "Do you have a Jira/ticket tracker MCP server configured?"
- Record the answer — this affects how the rules file handles ticket intake

### Phase 4: Generate Tool-Specific Rules File

Read `.ai/RULES.md` (the behavioral rules source of truth).

**Before generating, fill in the placeholders in RULES.md:**
- `[TO BE FILLED DURING ONBOARDING]` in Testing section → fill with actual test framework, test directory, and test command discovered in Phase 3.8
- `[TEST_COMMAND]` → fill with actual command (e.g. `npm test`, `pytest`, `mvn test`)
- If no test framework found, replace with: "No test framework configured yet — document manual testing in feature file"

**Then generate the tool-specific file:**

**If the engineer is using Cline:**
- Create `.clinerules` at the project root
- Content: `# Project: [actual project name]` + project overview at top, then the merged rules content:
  - Base: `.ai/RULES.md` (with placeholders filled)
  - Override: `.ai/RULES.local.md` if present (local wins on conflict, local additions appended)

**If the engineer is using Codex:**
- Create `AGENTS.md` at the project root
- Content structure (in this exact order):
  1. `# Project: [actual project name]` + project overview
  2. The Quick Commands skill routing table (see below)
  3. The merged rules content:
     - Base: `.ai/RULES.md` (with placeholders filled)
     - Override: `.ai/RULES.local.md` if present (local wins on conflict, local additions appended)

The Quick Commands section MUST be at the top of AGENTS.md, immediately after the project name:

## Quick Commands (READ THE SKILL FILE BEFORE ACTING)

| Say This | FIRST Read This File |
|---|---|
| "work on [ticket]" | .ai/skills/start-feature/SKILL.md |
| "review PR" | .ai/skills/pr-review/SKILL.md |
| "APEX review" | .ai/skills/apex-code-review/SKILL.md |
| "PLSQL review" | .ai/skills/plsql-code-review/SKILL.md |
| "Generates PLSQL unit test" | .ai/skills/utplsql-builder/SKILL.md |
| "check docs" | .ai/skills/check-docs/SKILL.md |
| "complete task" | .ai/skills/complete-task/SKILL.md |
| "continue work" | .ai/skills/resume/SKILL.md |
| "rebase" | .ai/skills/rebase/SKILL.md |
| "onboard" | .ai/skills/onboard/SKILL.md |
| "init AI code assist" | .ai/skills/init/SKILL.md |
| "summarize git diff" | .ai/skills/git-diff-summary/SKILL.md |
| "prepare MR summary" | .ai/skills/git-mr-jira-summary/SKILL.md |
| "summarize jira ticket" | .ai/skills/jira-ticket-summary/SKILL.md |
| "add jira workflow comment" | .ai/skills/jira-workflow-comment/SKILL.md |
| "upgrade AI code assist" | .ai/skills/upgrade/SKILL.md |

IMPORTANT: Always read the skill file BEFORE starting the task.
Do not improvise. The skill file contains the exact steps to follow.

**If the engineer wants both (or is unsure):**
- Create both `.clinerules` and `AGENTS.md` with the same content
- The `.ai/` directory and its contents are shared — they work with any tool

**If a ticket tracker MCP was detected:**
- Ensure the rules file Step 0 includes: "If given a ticket ID, fetch it via [tracker name] MCP first"
- Update Common Tasks to reference the specific MCP server name

### Phase 4.5: Create Local Override Files

Create these files if they do not exist:

**`.ai/RULES.local.md`:**
```markdown
# Project-Specific Rule Overrides
# These take precedence over .ai/RULES.md when conflicts exist.
# Add project-specific rules, override org defaults, or extend sections here.
# This file is NEVER touched by the upgrade skill.
```

**`.ai/conventions.local.md`:**
```markdown
# Project-Specific Convention Overrides
# These take precedence over .ai/conventions.md when conflicts exist.
# Add project-specific patterns, override org defaults, or extend sections here.
# This file is NEVER touched by the upgrade skill.
```

### Phase 5: Update README

If a `README_TEMPLATE.md` exists, use it to update or create the project `README.md`:
- Fill in from project analysis
- Add the AI-Assisted Development section
- Add the Project Documentation table

If a `README.md` already exists, do NOT overwrite it. Instead, add the AI-Assisted Development section and Project Documentation table to the existing README.

### Phase 5.5: Generate Skills Index

Create `.ai/SKILLS.md` by scanning all skill directories:

1. Read every `.ai/skills/*/SKILL.md` — extract `name` and `description` from YAML frontmatter
2. Read every `.ai.local/skills/*/SKILL.md` — extract `name` and `description` from YAML frontmatter
3. Map each skill name to its trigger phrase from the Skill Routing table in RULES.md
4. Write the combined table to `.ai/SKILLS.md`

This file is human-readable and committed to the repo so developers can browse available skills without asking the AI.

### Phase 6: Report

Summarize what you did:

```
## Onboarding Complete

### Populated:
- [x] PROJECT.md — [brief summary of what was found]
- [x] ARCHITECTURE.md — [brief summary, note Mermaid diagram]
- [x] DEPENDENCIES.md — [number of dependencies found]
- [x] DEPLOYMENT.md — [brief summary of CI/CD found]
- [x] context.md — [word count]
- [x] conventions.md — [patterns extracted from N files]
- [x] Tool rules file: [.clinerules / AGENTS.md / both]

### Needs Human Review:
- [ ] known-issues.md — needs team input
- [ ] [list any sections marked NEEDS VERIFICATION]
- [ ] [any business context that couldn't be determined]

### Validation:
To verify the setup works, try:
1. Ask me: "What is this project?"
2. Say: "I'm starting FEAT-TEST-001 to add input validation"
3. Check that I create a feature context file automatically
```

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its additional instructions. Local extensions run AFTER the base workflow above and can add project-specific steps, checks, or overrides.
