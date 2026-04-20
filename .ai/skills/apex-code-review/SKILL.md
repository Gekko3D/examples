---
name: apex-code-review
description: How to run a focused Oracle APEX and PL/SQL code review and produce a single Markdown report with severity summary, actionable findings, and remediation guidance.
---

# Oracle APEX Code Review

Use this guide to review Oracle APEX / PL/SQL changes consistently and safely.

## Quick Start (How to Use)

1. Open your AI tool in the repository containing the APEX changes.

2. Provide the review prompt context (PR, commit range, or changed files) and ask for an APEX security/quality review.

3. Require a single output artifact:
   - **`CodeReview.md`**

## What the AI Should Do

When invoked, the AI should:
- Review changed PL/SQL, APEX pages/processes, SQL/views, and related metadata
- Prioritize high-impact risks (security, correctness, performance, maintainability)
- Classify findings by severity (Severe / Moderate / Low)
- Provide concrete fix guidance with minimal before/after snippets
- Mark unclear cases as **Potential issue** when diff context is incomplete

## Review Focus Areas

- Security
  - SQL injection (dynamic SQL, concatenation, missing bind variables)
  - XSS in APEX output/substitutions and JavaScript contexts
  - Authorization/session enforcement in changed flows
  - Sensitive data exposure in logs/debug output
  - Item-level access protection and safe dynamic input handling (e.g., `DBMS_ASSERT`)
- APEX design quality
  - Process/branch logic correctness
  - Proper use of shared components and authorization schemes
  - Hard-coded values that should be configurable
  - Risky or deprecated API usage
- Performance and maintainability
  - Inefficient SQL patterns or repeated processing
  - Readability, cohesion, and long-term supportability

## Output Format Requirements

The output must be one Markdown document titled `CodeReview.md` with:

1. `## Summary by Severity`
   - `Severe issues: N`
   - `Moderate issues: N`
   - `Low issues: N`

2. `## Detailed Findings` grouped under:
   - `### PL/SQL Changes`
   - `### APEX Pages & Processes`
   - `### Views & SQL`
   - `### Security`
   - `### Performance`
   - `### Maintainability`

3. For each finding:
   - `#### [Severity] Short issue title`
   - **Severity:** Severe | Moderate | Low
   - **Change Location:** file/object/page/process/line range (if available)
   - **Description:** what is wrong and why it matters
   - **Impact:** security/correctness/performance/maintainability effect
   - **Recommendation:** specific actionable fix
   - **Code snippet (before):** minimal relevant example
   - **Suggested fix (after or pseudo-code):** concrete improvement

## Severity Guidance

- **Severe:** realistic risk of data corruption, exploitable security vulnerability, major logic failure, or severe performance degradation.
- **Moderate:** meaningful risk or inefficiency likely to cause recurring issues.
- **Low:** minor best-practice, readability, or maintainability improvements.

## Notes

- Do **not** flag APEX release version differences as review findings.
- Focus on signal over noise; avoid trivial comments.
- If uncertain due to missing context, label the finding as **Potential issue** and state assumptions.
