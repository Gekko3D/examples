# Logger Querying

Use this reference after instrumentation when the engineer wants to inspect emitted Logger entries.

## Assumptions

These examples are intentionally generic:
- Logger is installed in the target database
- the caller can read the relevant Logger tables or views, directly or through granted synonyms
- object names may differ slightly by installation

Adapt the examples to the local installation before running them.

## What to look for first

- the table or view that stores log rows
- the column used for scope or unit name
- the column used for timestamp ordering
- the column used for message text
- whether parameter values are stored inline or in a related structure

## Common inspection flow

1. Find the most recent rows for a routine scope.
2. Confirm you see the `START` marker.
3. Confirm the expected progress message appears.
4. Confirm you see the matching `END` marker.
5. Confirm no secret, key, password, token, or wallet material was logged.

## Generic example patterns

Recent rows by scope:

```sql
select *
  from logger_logs
 where scope like '%<routine_name>%'
 order by id desc;
```

Recent rows with explicit columns:

```sql
select id
     , log_date
     , scope
     , text
  from logger_logs
 where scope = '<package_or_unit>.<routine_name>'
 order by log_date desc;
```

Recent rows for a time window:

```sql
select id
     , log_date
     , scope
     , text
  from logger_logs
 where log_date >= systimestamp - interval '15' minute
   and scope like '%<routine_name>%'
 order by log_date desc;
```

Search for suspicious sensitive terms:

```sql
select id
     , log_date
     , scope
     , text
  from logger_logs
 where lower(text) like '%password%'
    or lower(text) like '%secret%'
    or lower(text) like '%token%'
    or lower(text) like '%wallet%'
 order by log_date desc;
```

## Adaptation notes

- Some installations use a view instead of `logger_logs`
- Some use `msg_text`, `message_text`, or `text` for the message column
- Some use `time_stamp`, `created_on`, or `log_date` for the timestamp column
- Some expose helper procedures or cleanup APIs instead of direct table access

## Hard rule

Log inspection is also subject to the same security requirement:
- never expose or preserve secrets, keys, passwords, tokens, wallet material, or credential values in copied query output
