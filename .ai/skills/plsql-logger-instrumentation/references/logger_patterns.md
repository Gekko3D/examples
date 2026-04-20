# Logger Patterns

Use these patterns as portable starting points. Always adapt them to the local repo before editing code.

## Required environment

These patterns assume:
- a `logger` package is installed in the target database
- the calling schema has execute access to `logger`
- referenced logger types such as `logger.tab_param` and `logger_logs.scope%type` are resolvable at compile time

## Portable standalone pattern

Highlights:
- local `gc_scope_prefix` inside the routine
- `l_scope` built from the routine name
- optional `l_params` for safe business parameters
- `START` and `END` markers around the main work

```sql
gc_scope_prefix     constant varchar2(31) := lower($$plsql_unit) || '.';
l_scope             logger_logs.scope%type := gc_scope_prefix || '<routine_name>';
l_params            logger.tab_param;
...
logger.append_param(l_params, 'p_id', p_id);
logger.log('START >>>>>>>>>>>>>>>>>>>>', l_scope, null, l_params);
...
logger.log('END   <<<<<<<<<<<<<<<<<<<<', l_scope);
```

## Portable procedure pattern

- same simple scope prefix
- multiple `append_param` calls before the start log
- one business-success message before the end log when useful

```sql
l_scope   logger_logs.scope%type := gc_scope_prefix || '<procedure_name>';
l_params  logger.tab_param;
...
logger.append_param(l_params, 'p_name', p_name);
logger.append_param(l_params, 'p_flag', p_flag);
logger.log('START >>>>>>>>>>>>>>>>>>>>', l_scope, null, l_params);
...
logger.log('Completed successfully', l_scope);
logger.log('END   <<<<<<<<<<<<<<<<<<<<', l_scope);
```

## Package body pattern

Highlights:
- package-level prefix is declared once and reused by each routine
- each routine declares its own `l_scope`
- `append_param` is used selectively for procedure inputs
- progress logs are used for counts and important milestones

```sql
gc_scope_prefix constant varchar2(100) := lower($$plsql_unit) || '.';
...
l_scope   logger_logs.scope%type := gc_scope_prefix || '<routine_name>';
l_params  logger.tab_param;
...
logger.append_param(l_params, 'p_group_id', p_group_id);
logger.append_param(l_params, 'p_user_id', p_user_id);
logger.log('START >>>>>>>>>>>>>>>>>>>>', l_scope, null, l_params);
...
logger.log('Completed business action for ID = ' || p_group_id, l_scope);
logger.log('END   <<<<<<<<<<<<<<<<<<<<', l_scope);
```

## Adaptation checklist

- Prefer `lower($$plsql_unit) || '.'` as the portable default
- Reuse any existing package-level prefix if the repo already defines one
- Keep `l_scope` suffixes aligned with the routine name
- Add `l_params` only when there is useful, safe context to record
- Log IDs, flags, and counts; never log secrets, keys, passwords, tokens, or wallet material
- Use `START` and `END` markers consistently
- Add one or two meaningful progress logs, not a log on every line
