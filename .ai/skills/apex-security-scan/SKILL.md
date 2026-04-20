---
name: apex-security-scan
description: Static security scan of an APEX split SQL export. Auto-discovers the export directory, then runs apex_scan.py to detect SQL injection, XSS, URL tampering, and misconfiguration issues before a PR is merged. No database required.
---

# Skill: APEX Security Scan

## When to Invoke
- Engineer says "scan my APEX app", "security check", "check for XSS", "check for SQL injection"
- Engineer says "is this safe to push", "scan before PR", "any vulnerabilities?"
- Any file under an APEX export directory (`application/pages/`, `application/shared_components/`) is in the diff
- Automatically as part of the `pr-review` skill when APEX export files are changed

---

## Instructions

### Step 1: Auto-discover the export directory

Run this to find all valid APEX export directories in the repo:

```bash
find . -maxdepth 3 -name "create_application.sql" 2>/dev/null \
  | sed 's|/application/create_application.sql||' \
  | sed 's|^\./||'
```

This finds any directory containing `application/create_application.sql` — the
canonical marker of an APEX split export — regardless of what the folder is named.

**If one result:** use it automatically.
**If multiple results:** show the list to the engineer and ask which to scan.
**If no results:** tell the engineer "No APEX split export found. Make sure the
export directory is committed to the repo and contains `application/create_application.sql`."

Store the result as `EXPORT_DIR` for the steps below.

### Step 2: Run the scanner

Full scan:
```bash
python .ai/tools/apex_scan.py $EXPORT_DIR --format text
```

Domain-specific scan:
```bash
python .ai/tools/apex_scan.py $EXPORT_DIR --only sqli --format text
python .ai/tools/apex_scan.py $EXPORT_DIR --only xss --format text
python .ai/tools/apex_scan.py $EXPORT_DIR --only urlprotect --format text
python .ai/tools/apex_scan.py $EXPORT_DIR --only config --format text
```

PR gate (SARIF for CI):
```bash
python .ai/tools/apex_scan.py $EXPORT_DIR --threshold HIGH --format sarif --output results.sarif
```

If `python` is not found, try `python3`.

### Step 3: Interpret and report findings

For each finding explain to the engineer:
1. **Which file** is affected — exact path relative to repo root
2. **What the vulnerability means** — plain English, no jargon
3. **Where to fix it in APEX Builder** — the `Fix` field gives the exact Builder navigation path
4. **Whether it blocks the PR** — CRITICAL or HIGH = must fix before merge

### Step 4: Severity gate

| Severity | Action |
|----------|--------|
| CRITICAL | "This must be fixed before merging — do not deploy." |
| HIGH     | "Fix before merging." |
| MEDIUM   | "Fix before production release — not a merge blocker." |
| LOW      | Note as best practice improvement. |

If no findings: report "apex-scan: no security issues found in `$EXPORT_DIR`."

---

## What the Scanner Checks

| Rule ID    | Severity | Domain     | What |
|------------|----------|------------|------|
| SQLI-001   | CRITICAL | sqli       | SQL regions using string concat || or &ITEM. substitution |
| XSS-001    | HIGH     | xss        | Regions with p_escape_on_http_output => 'N' |
| URL-001    | HIGH     | urlprotect | Pages with p_protection_level => 'D' (unrestricted) |
| URL-003    | HIGH     | urlprotect | Hidden items with no URL protection |
| CONFIG-001 | MEDIUM   | config     | App p_build_status => 'RUN_AND_BUILD' |
| CONFIG-002 | MEDIUM   | config     | Browser frame protection disabled |
| URL-002    | MEDIUM   | urlprotect | Input items with no URL protection |

---

## Integration with pr-review Skill

When running the `pr-review` skill and the diff includes files under any APEX
export directory (contains `application/pages/` path):
1. Run the auto-discovery command from Step 1 to find the export root
2. Run apex-security-scan automatically as part of Step 4 (Security section)
3. Map severity into the review: CRITICAL/HIGH → Severe, MEDIUM → Moderate
4. Include findings in the **4.2 Security** section of the review output

---

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and follow its
additional instructions.
