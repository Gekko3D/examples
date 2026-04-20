---
name: plsql-logger-instrumentation
description: Add or extend `logger` instrumentation in Oracle PL/SQL packages, procedures, and functions using portable patterns such as `gc_scope_prefix`, `l_scope`, `l_params`, `logger.append_param`, and `logger.log`, while adapting to the local repo's established style.
compatibility: Requires a target Oracle Database where the `logger` framework is already installed and executable by the calling schema.
metadata:
  author: andrew.mcgugan@oracle.com
  version: "1.0"
---

# Skill: PLSQL Logger Instrumentation

## Activation Triggers
Activate this skill when the user asks to add or extend `logger` instrumentation for a PL/SQL package, procedure, function, or body implementation. If the target object, file, or schema area is ambiguous, ask one concise clarifying question before proceeding.

Examples of trigger phrases:
- "instrument this package with logger"
- "add logger calls to this package"
- "add trace logging around this procedure"
- "wire `logger.append_param` into these routines"

## Prerequisite

Use this skill only when the target database already has the `logger` framework installed and the calling schema can execute it.

Before editing, confirm one of the following:
- the repo already contains logger-instrumented PL/SQL that compiles in the same schema area
- the engineer confirms `logger` is installed and granted
- the deployment path for the target environment already provisions `logger`

If `logger` is not present or access is uncertain, stop and tell the engineer the skill assumes a working `logger` installation plus execute access for the caller.

## Workflow

### Step 1: Inspect local patterns first

Before editing, inspect the target file and a nearby logger-enabled file.

Use the local repo's existing style instead of introducing a new one:
- Some repos use a simple `lower($$plsql_unit) || '.'` prefix
- Some repos wrap that with team, module, or application tags
- Some repos already have a shared logging helper or prefix constant in the package body

Read [logger_patterns.md](references/logger_patterns.md) for portable examples.

### Step 2: Choose the right scope prefix pattern

Prefer the narrowest change that matches the file's neighborhood.

Portable default for standalone procedures/functions or simple bodies:

```sql
gc_scope_prefix   constant varchar2(31) := lower($$plsql_unit) || '.';
```

If the package or repo already uses a different prefix pattern, reuse that pattern rather than forcing the portable default.

Do not replace an established prefix style in the file unless the engineer explicitly asks for cleanup as part of the task.

### Step 3: Add routine-level logging variables

For each routine being instrumented, add:

```sql
l_scope    logger_logs.scope%type := gc_scope_prefix || '<routine_name>';
l_params   logger.tab_param;
```

Only add `l_params` when parameters or context are worth logging.

### Step 4: Log safe inputs only

Append parameters before the first `logger.log('START ...')` call.

Use:

```sql
logger.append_param(l_params, 'p_name', p_name);
```

Rules:
- Log identifiers, flags, counts, and business keys when useful
- Never log secrets, API keys, encryption keys, tokens, passwords, wallet data, or other sensitive credential material. This is a hard requirement and overrides local patterns.
- Be cautious with free-text justification or request bodies; prefer counts or IDs over raw content

### Step 5: Add start, progress, and end logs

Minimum pattern:

```sql
logger.log('START >>>>>>>>>>>>>>>>>>>>', l_scope, null, l_params);
...
logger.log('END   <<<<<<<<<<<<<<<<<<<<', l_scope);
```

Add one or two meaningful intermediate logs when they help operations:
- row counts
- created IDs
- branch/path decisions
- completion of external calls

Avoid noisy line-by-line tracing unless the engineer explicitly asks for deep diagnostics.

### Step 6: Handle exceptions without hiding them

If the routine already has an exception block, add logging in the existing pattern.

Preferred behavior:
- log the failure
- re-raise unless the surrounding code intentionally swallows the exception

Typical approach:

```sql
exception
   when others then
      logger.log_error('Unexpected error in ' || l_scope, l_scope);
      raise;
end <routine_name>;
```

If the file already uses a different error helper, preserve that local pattern.

### Step 7: Keep package specs clean

- Instrumentation belongs in package bodies or standalone implementation files
- Do not add logger declarations to specs unless the object structure already requires it
- Do not change public signatures just to pass logging context

### Step 8: Verify

After editing:
- confirm `gc_scope_prefix` is declared once in the right scope
- confirm each new `l_scope` suffix matches the routine name
- confirm every `START` has a matching normal-path `END`
- confirm `logger.append_param` names match actual parameter names
- confirm no sensitive values were added to logs
- compile or run the lightest available verification for the touched object when possible

If the engineer also wants to inspect emitted logs after instrumentation, read [logger_querying.md](references/logger_querying.md) and adapt its examples to the local Logger installation.

## Guardrails
- Match the file's existing indentation and lowercase PL/SQL style
- Prefer minimal, surgical instrumentation over broad refactors
- Do not invent a new logging wrapper when `logger` is already the local standard for that area
- Do not add `commit`
- Never log secrets, keys, passwords, tokens, wallet contents, or credential material under any circumstance

## Local Extensions

If `SKILL.local.md` exists in this skill directory, read it now and apply its additional instructions.
