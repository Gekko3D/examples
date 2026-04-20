---
name: utplsql-builder
description: Generates utPLSQL unit test packages (specification and body) for Oracle PL/SQL objects in the unit_test schema, following project-specific conventions for naming, annotations, indentation, APEX session management (when applicable), and best practices. This skill is activated when the user requests to build or add unit tests for a specific PL/SQL object.
compatibility: Requires Oracle Database with utPLSQL installed. Oracle APEX access is required only when the target object depends on APEX APIs (e.g., apex_session, apex_collection).
metadata:
  author: edward.talcott@oracle.com
  version: "1.0"
---

## Activation Triggers
Activate this skill when the user asks to create or extend utPLSQL tests for a specific PL/SQL object; if the target is missing or ambiguous, follow the Agent Prompting Policy before proceeding.

Examples of trigger phrases:
- "Build unit tests for package <object_name>"
- "Generate utPLSQL tests for function <object_name>"
- "Add tests to the existing utPLSQL package for procedure <object_name>"
- "Create spec and body for test_<object_name>"
- "Add new tests to test_<object_name>"

## Agent Prompting Policy
Prompt the user before generation in the following cases. Ask only the minimal targeted questions needed to proceed, batch multiple clarifications into a single numbered prompt, and stop prompting once sufficient information is gathered.

1) Missing or ambiguous target
- If the object name or type is missing or unclear, ask:
  - Which PL/SQL object should I create tests for (package/function/procedure), and what is its exact name?

2) APEX dependency uncertainty
- If APEX usage cannot be determined, ask:
  - Does the target rely on APEX APIs (e.g., apex_session, apex_util, apex_collection)?

3) APEX session lifecycle preference (only when APEX is used)
- Ask the user which session lifecycle to use:
  - Maintain a single session for the entire test run (create in before_all, delete in after_all)?
  - Create and delete a new session for each test (create in before_each, delete in after_each)?
- If the user does not specify, default to: single session for the entire test run (before_all/after_all).

4) Output paths
- If repository layout or destinations are not provided, ask:
  - Should I place the tests in the `unit_test` directory under the `plsql` directory where the source package is found, or do you have specific output directories?

5) Existing tests present
- If a test package already exists, ask:
  - Should I append new tests, update existing ones, or add a new scenario-specific test procedure?

6) Required inputs and setup
- If inputs or setup are not clear, ask:
  - Provide representative input values and any required setup (tables/collections/seed data).

7) Liquibase comments
- Determine inclusion based solely on the file(s) under test:
  - If the target file(s) include Liquibase formatted SQL headers or rollback comments, include them in the generated tests.
  - If the target file(s) do not include Liquibase comments, omit them.
- If the target file(s) cannot be inspected or are ambiguous, ask:
  - The target source's Liquibase usage is unclear. Should I include Liquibase formatted SQL headers and rollback comments in the generated files?

Behavior when prompting
- Batch unresolved items into one concise, numbered message.
- If no response after one clarification, proceed with safe defaults:
  - Assume non-APEX unless APEX usage is explicitly indicated.
  - For non-APEX: suite-level --%rollback(auto); tests inherit.
  - For APEX: if confirmed but session lifecycle unspecified, use a single session for the entire run (before_all/after_all).
  - Place the test files in the `unit_test` directory under the `plsql` directory where the source package is found, unless overridden by the user.

## Agent Checklist
- Schema/name/authid: unit_test schema; authid definer; package name test_<object_name>.
- Annotations: per Instructions (use comment annotations; placement rules; blank line after suite-level block).
- Test procedures: one per public routine's happy path; do not overload tests; for overloaded targets, use disambiguated names.
- Rollback: apply per "Rollback behavior" in Instructions.
- Style: explicit END names; follow indentation per Instructions.
- AAA pattern: Arrange → Act → Assert; include loop index in messages when iterating.
- APEX handling (if needed): before_all/after_all (default single session); optionally before_each; create/delete session; use p_commit => false for session state.
- File placement: put test files in the `unit_test` directory under the `plsql` directory where the package source is found, unless overridden.

## Instructions
Use the following rules verbatim when generating utPLSQL test packages (SPEC and BODY):

- Package schema and name:
  - All test packages are created in the unit_test schema.
  - Name: test_<object_name> (object_name is the PL/SQL package/function/procedure under test; do not include schema).
  - Use authid definer.
- Indentation and style:
  - Follow the style of the code being tested; if unclear, use 4 spaces for indentation.
  - Explicitly name END statements for all procedures/functions and the package.
- Annotations:
  - Use utPLSQL annotation comments (not pragmas) for suite, suitepath, rollback, test, throws, before_* and after_* hooks.
  - Place all utPLSQL annotations for individual tests in the SPEC, not in the BODY.
  - Include exactly one blank line between suite-level annotations and the first test-level annotation in the SPEC.
- Test procedure mapping:
  - One primary ("happy path") test procedure per public routine.
  - Do NOT overload test procedures.
  - If the routine under test is overloaded, disambiguate test proc names with a clear suffix (e.g., _numlist, _varchar2, _with_flags) and clarify in --%test(...) titles.
- Expectations and asserts:
  - Every test must assert. Avoid dummy assertions like ut.expect(null) unless explicitly documenting a known gap; prefer asserting on return values or observable side effects.
  - Use utPLSQL best practices (e.g., cursor-based asserts for multi-row comparisons).
  - When asserting in loops, include a loop index or key info in the expectation message.
  - When comparing cursors, use the framework's exclude(...) advanced comparison option to ignore nondeterministic columns (e.g. timestamps or sysguids).
- Suitepath:
  - The suitepath should represent the object under test (e.g., the package name).
- Rollback behavior:
  - Non-APEX targets: use --%rollback(auto) at the suite level; tests inherit by default.
  - APEX-dependent targets:
    - Suite level: --%rollback(manual)
    - Each test: --%rollback(auto)
- Test content scope:
  - Focus on positive paths for public procedures/functions; group multiple asserts for that routine within one test procedure when appropriate.
  - Include negative/exception scenarios for specific, expected exceptions (use --%throws(-NNNNN)); skip generic WHEN OTHERS exception scenarios.
- Data and environment:
  - Arrange test data before calling routines; assume required privileges are granted.
  - For multi-row validations, prefer cursors and explicit comparisons over broad heuristics.
- APEX session handling (only when target depends on APEX):
  - SPEC: declare hooks only (no globals):
    - --%beforeeach(before_each)  [optional, recommended when shared state exists]
    - --%beforeall(before_all)
    - --%afterall(after_all)
  - BODY: keep APEX-related globals private (e.g., gc_ut_user, gc_apex_app_id, gc_apex_page_id, g_session_id). If the user does not specify a session lifecycle, default to a single session for the entire run (before_all/after_all), per Agent Prompting Policy.
	- For the default single-session lifecycle: implement before_all to create a session (apex_session.create_session) and store the session id; implement after_all to delete the session (apex_session.delete_session) using the stored session id, then clear state.
    - If the user selects a per-test lifecycle: implement before_each to create the session and after_each to delete it accordingly.
    - Use apex_collection.create_or_truncate_collection and apex_collection.add_member as needed.
    - For apex_util.set_session_state-like calls, pass p_commit => false to leverage utPLSQL auto rollback.
- Test data:
  - Avoid hardcoding IDs when possible to avoid conflicts, but if necessary use likely safe IDs such as negative numbers.
  - When testing with multiple sets of data in a single test, prefer to keep values unique and obvious across data sets to simplify debugging. e.g. the value for the column "extra" in the 2nd test row might be the string "extra - row 2".

## Usage Examples
Note: Example object names are placeholders to illustrate APEX vs. non-APEX guidance. You will update these to your actual targets.

### Example 1: APEX-Involved Package (test_example_apex_api)
```
--liquibase formatted sql
--changeset [author]:test_example_apex_api_create_spec stripComments:false endDelimiter:/ runOnChange:true
create or replace package unit_test.test_example_apex_api authid definer as

--%suite(Example APEX API)
--%suitepath(example_apex_api)
--%rollback(manual)
--%beforeall(before_all)
--%afterall(after_all)

--------------------------------------------------------------------------------
--
--^ helper methods
--
--------------------------------------------------------------------------------
procedure before_all;
procedure after_all;

--------------------------------------------------------------------------------
--
--^ test procedures
--
--------------------------------------------------------------------------------

--%test(sample_method - happy path)
--%rollback(auto)
procedure sample_method;

end test_example_apex_api;
/
--rollback drop package unit_test.test_example_apex_api;

--liquibase formatted sql
--changeset [author]:test_example_apex_api_create_body stripComments:false endDelimiter:/ runOnChange:true
create or replace package body unit_test.test_example_apex_api as

    gc_ut_user         constant varchar2(100) := 'UNIT_TEST';
    gc_apex_app_id     constant number        := 1000; -- sample app id
    gc_apex_page_id    constant number        := 1;    -- sample page id
    g_session_id       number;

--------------------------------------------------------------------------------
--
--^ helper methods
--
--------------------------------------------------------------------------------

    procedure before_all is
    begin
        apex_session.create_session(
             p_app_id   => gc_apex_app_id
            ,p_page_id  => gc_apex_page_id
            ,p_username => gc_ut_user
        );

        g_session_id := v('APP_SESSION');
        ut.expect(g_session_id, 'APEX session should be created').to_be_not_null();
    end before_all;

    procedure after_all is
    begin
		if g_session_id is not null then
			apex_session.delete_session(p_session_id => g_session_id);
		end if;
        g_session_id := null;
    end after_all;

--------------------------------------------------------------------------------
--
--^ test procedures
--
--------------------------------------------------------------------------------

    procedure sample_method is
        l_actual varchar2(100);
    begin
        -- Arrange
        apex_collection.create_or_truncate_collection(
            p_collection_name => 'UT_EXAMPLE_COLLECTION'
        );

        apex_collection.add_member(
             p_collection_name => 'UT_EXAMPLE_COLLECTION'
            ,p_c001            => 'ROW_1'
        );

        -- Act
        l_actual := example_schema.example_apex_api.sample_method(...);

        -- Assert
        ut.expect(l_actual, 'Sample assertion for APEX-session-based test').to_equal('ROW_1');
    end sample_method;

end test_example_apex_api;
/
--rollback drop package body unit_test.test_example_apex_api;
```

### Example 2: Non-APEX Package (no_apex_use_api)
```
--liquibase formatted sql
--changeset [author]:test_example_service_api_create_spec stripComments:false endDelimiter:/ runOnChange:true
create or replace package unit_test.test_example_service_api authid definer as

--%suite(Example Service API)
--%suitepath(example_service_api)

--------------------------------------------------------------------------------
--
--^ helper methods
--
--------------------------------------------------------------------------------
procedure clear_and_set_up_test_services;

--------------------------------------------------------------------------------
--
--^ test procedures
--
--------------------------------------------------------------------------------

--%test(get_owner_for_service for valid service records)
procedure get_owner_for_service;

--%test(get_release_for_service with invalid service key - should throw invalid_service_key_erno)
--%throws(example_service_xapi.invalid_service_key_erno)
procedure get_release_for_service_invalid;

--%test(create_entity - testing various exceptions)
procedure create_entity_exceptions;

end test_example_service_api;
/
--rollback drop package unit_test.test_example_service_api;

--liquibase formatted sql
--changeset [author]:test_example_service_api_create_body stripComments:false endDelimiter:/ runOnChange:true
create or replace package body unit_test.test_example_service_api as

    type t_test_data_row is record (
         service_key      example_service.service_key%type
        ,service_owner    example_service.service_owner%type
        ,service_release  example_service.service_release%type
    );

    type t_test_data_rows is table of t_test_data_row;

    g_test_rows t_test_data_rows := t_test_data_rows(
        t_test_data_row('SERVICE_1', 'OWNER_1', '1.0.1'),
        t_test_data_row('SERVICE_2', null,      '1.0.2')
    );

    gc_invalid_service_key constant example_service.service_key%type := 'INVALID_SERVICE_KEY';

--------------------------------------------------------------------------------
--
--^ helper methods
--
--------------------------------------------------------------------------------
    procedure clear_and_set_up_test_services is
    begin
        delete from example_service;

        for i in 1..g_test_rows.count loop
            insert into example_service(
                 service_key
                ,service_owner
                ,service_release
            )
            values(
                 g_test_rows(i).service_key
                ,g_test_rows(i).service_owner
                ,g_test_rows(i).service_release
            );
        end loop;
    end clear_and_set_up_test_services;

--------------------------------------------------------------------------------
--
--^ test procedures
--
--------------------------------------------------------------------------------
    procedure get_owner_for_service is
        l_actual_owner   example_service.service_owner%type;
        l_test_cnt       integer := 0;
    begin
        -- Arrange
        clear_and_set_up_test_services;

        for i in 1..g_test_rows.count loop
            l_test_cnt := l_test_cnt + 1;

            -- Act
            l_actual_owner := example_service_api.get_owner_for_service(
                p_service_key => g_test_rows(i).service_key
            );

            -- Assert
            ut.expect(l_actual_owner, 'Loop ' || l_test_cnt || ': owner should match').to_equal(
                g_test_rows(i).service_owner
            );
        end loop;

        ut.expect(l_test_cnt, 'Need at least one test record').to_be_greater_than(0);
    end get_owner_for_service;

    procedure get_release_for_service_invalid is
        l_actual_release example_service.service_release%type;
    begin
        -- Arrange
        clear_and_set_up_test_services;

        -- Act
        l_actual_release := example_service_api.get_release_for_service(
            p_service_key => gc_invalid_service_key
        );

        -- Assert fallback (primary assertion is %throws in spec)
        ut.expect(true,'We should never reach this assertion because get_release_for_service should raise an exception').to_be_false();
    end get_release_for_service_invalid;

	procedure create_entity_exceptions is
		l_test_idx integer := 0;
	begin
		-- Scenario 1: invalid key format
		l_test_idx := l_test_idx + 1;
		begin
			example_service_api.create_entity(
				 p_entity_key   => null
				,p_entity_name  => 'Test Name'
				,p_parent_id    => 100
			);

			ut.expect(true, 'Expected exception was not raised - test ' || l_test_idx).to_be_false();
		exception
			when others then
				if sqlcode = example_service_api.entity_key_invalid_erno then
					ut.expect(true, 'Raised expected exception - test ' || l_test_idx).to_be_true();
				else
					ut.expect(false,
						'Unexpected SQLCODE ' || sqlcode ||
						', expected ' || example_service_api.entity_key_invalid_erno ||
						' - test ' || l_test_idx
					).to_be_true();
				end if;
		end;

		-- Scenario 2: invalid parent id
		l_test_idx := l_test_idx + 1;
		begin
			example_service_api.create_entity(
				 p_entity_key   => 'VALID_KEY'
				,p_entity_name  => 'Test Name'
				,p_parent_id    => -99999
			);

			ut.expect(true, 'Expected exception was not raised - test ' || l_test_idx).to_be_false();
		exception
			when others then
				if sqlcode = example_service_api.parent_id_invalid_erno then
					ut.expect(true, 'Raised expected exception - test ' || l_test_idx).to_be_true();
				else
					ut.expect(false,
						'Unexpected SQLCODE ' || sqlcode ||
						', expected ' || example_service_api.parent_id_invalid_erno ||
						' - test ' || l_test_idx
					).to_be_true();
				end if;
		end;
	end create_entity_exceptions;

end test_example_service_api;
/
--rollback drop package body unit_test.test_example_service_api;
```

## SPEC vs BODY intent
For detailed rules on annotation placement and lifecycle hooks, see the "Annotations" and "APEX session handling" sections above.
- SPEC: declarations only (procedure signatures and annotations); avoid implementation details and sample data. Keep all types private unless explicitly requested to be public.
- BODY: full implementations following Arrange/Act/Assert; keep helper and APEX-related globals private.

## Edge Cases and Troubleshooting
- Overloaded routines (under test): create separate test procedures with disambiguating names; distinguish further via --%test(...) text.
- Nondeterministic behavior (timestamps, sequences, dbms_random):
  - Use fixed seeds or compare within tolerances (e.g., within 1 second).
  - Normalize values (e.g., TRUNC timestamps) where acceptable.
- Autonomous transactions or explicit commits:
  - Prefer avoiding side effects; if unavoidable, document and isolate those tests.
  - Consider suite-level manual rollback when interacting with APEX or autonomous code.
- Global state in the package under test:
  - Reset state in before_each; reinitialize collections/caches to prevent leakage between tests.
- Large result sets:
  - Use cursor-based assertions; limit to representative samples for performance.
- When testing code using clob values:
  - Always include a test that verifies values longer than 32767 bytes are supported.
- External dependencies:
  - Prepare minimal viable APEX/session context; avoid external network calls in unit tests.

## File Placement
- Default locations:
  - Find the package source file under its existing `plsql` directory.
  - Place the generated spec and body files in the sibling `plsql/unit_test/` directory for that package tree.
  - Use filenames that match the generated package objects, for example `test_<object_name>.pls` and `test_<object_name>.plb` when that tree follows the repo's spec/body split.
- If the user specifies different output paths, honor the override.

## Incremental Updates (adding to existing tests)
- Prefer surgical edits to avoid duplication:
  - Match target insertion points by procedure signature and/or the --%test(...) annotation title.
  - If a test procedure already exists for the routine/scenario, add assertions inside it rather than creating a near-duplicate.
  - For a new overload scenario, create a new test procedure with a distinct name and annotation.
- Preserve formatting:
  - Keep Liquibase headers and rollback comments intact.
  - Maintain indentation and explicit END names.
- Safety:
  - Do not remove existing tests unless explicitly instructed.
  - When unsure, append new tests at the end of SPEC declarations and BODY implementations with clear annotations.

## Output Format
The skill outputs two separate .sql files:
- Spec: declarations, annotations, types, and procedure stubs.
- Body: implementations, helpers, constants/globals (private), and test logic with ut.expect.
Place them in the `unit_test` directory under the `plsql` directory where the package source is found (or user-provided overrides).
