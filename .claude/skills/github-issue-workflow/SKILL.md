---
name: gh-implement-issue
description: "End-to-end implementation workflow for a GitHub issue from planning through PR creation. Use when starting work on an issue from scratch."
category: github
---

# Implement GitHub Issue

Strict Test-Driven Development (TDD) workflow for implementing a GitHub issue from start to finish.

IMPORTANT: This workflow is sequential. You MUST complete each step fully before moving to the next.
You MUST NOT skip steps, reorder steps, or combine steps. Each step has a gate condition that
must be satisfied before proceeding.

## When to Use

- Starting work on a new issue
- Need structured workflow from branch to PR
- Want to follow best practices end-to-end
- Working on assigned GitHub issue

## Workflow

Follow these steps **in exact order**. Do NOT proceed to the next step until the current step is complete.

### Step 1: Read issue context

- Run `gh issue view <issue> --comments` to understand requirements and prior context
- Summarize the issue and your understanding to the user
- **Gate**: You have a clear understanding of what needs to be done

### Step 2: Explore the codebase

- Search the codebase to understand the relevant code, patterns, and conventions
- Identify all files that will need to change
- **Gate**: You can explain the current behavior and what needs to change

### Step 3: Create branch

- `git checkout -b feature/<issue>-<description>`
- **Gate**: You are on a new feature branch

### Step 4: Write tests FIRST (TDD - Red phase)

This is the TDD "Red" phase. You MUST write tests before writing any implementation code.

- Determine which tests are needed (unit tests, integration tests, or both)
- If the issue involves Kafka interaction, write integration tests (test name must end with `Integration`)
- Write the test code that validates the expected behavior described in the issue
- Make sure the docker is set up (`docker compose up -d`) to run integration tests
- Run the written tests using `go test` to confirm they **fail** (DO NOT skip integration tests) (Red phase). 
- **Gate**: Tests are written and it is confirmed that they fail. Do NOT write any implementation code yet.

### Step 5: Human approval (MANDATORY STOP)

You MUST stop here and wait for explicit human approval before continuing.

- Present a summary of all changes (files modified, what changed, why)
- Ask: "Please review the changes. Should I proceed with implementation?"
- Do NOT proceed until the user explicitly approves (e.g., "yes", "looks good", "proceed")
- If the user requests changes, go back to the appropriate step, make the changes, and return to this step

### Step 6: Implement code (TDD - Green phase)

This is the TDD "Green" phase. Now and only now do you write implementation code.

- Write the minimum code necessary to make the tests pass
- Run `make test` to verify unit tests pass
- **Gate**: All unit tests pass

### Step 7: Refactor (TDD - Refactor phase)

- Review the implementation for code quality, duplication, and clarity
- Refactor if needed while keeping tests green
- **Gate**: Code is clean and all tests still pass

### Step 8: Quality checks

- Run `make fmt` to format code
- Run `make lint` to run the linter
- Fix any issues found
- **Gate**: Both `make fmt` and `make lint` pass with no errors or warnings

### Step 9: Human approval (MANDATORY STOP)

You MUST stop here and wait for explicit human approval before continuing.

- Present a summary of all changes (files modified, what changed, why)
- Show the diff summary
- Ask: "Please review the changes. Should I proceed with committing and creating a PR?"
- Do NOT proceed until the user explicitly approves (e.g., "yes", "looks good", "proceed")
- If the user requests changes, go back to the appropriate step, make the changes, and return to this step

### Step 10: Commit

- Only after human approval
- `git add . && git commit -m "type(scope): description\n\nCloses #<issue>"`
- Follow conventional commit format
- **Gate**: Commit is created successfully

### Step 11: Push and create PR

- `git push -u origin <branch>`
- `gh pr create --title "..." --body "..." `
- Link the PR to the issue
- Return the PR URL to the user

## Quick Reference

```bash
# 1. Fetch issue and create branch
gh issue view <issue>
git checkout -b feature/<issue>-<description>

# 2. TDD cycle
# - Write tests FIRST (Red)
# - Implement code (Green)
# - Refactor
# - Run tests: make test

# 3. Quality checks
make fmt
make lint

# 4. Wait for human approval

# 5. Commit and PR
git add . && git commit -m "feat: description

Closes #<issue>"
git push -u origin <branch>
gh pr create --issue <issue>
```

## Branch Naming Convention

Format: `feature/<issue-number>-<description>`

Examples:

- `feature/42-add-tensor-ops`
- `feature/73-fix-memory-leak`
- `feature/105-update-docs`

## Commit Message Format

Follow conventional commits:

```text
type(scope): Brief description

Detailed explanation of changes.

Closes #<issue-number>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

## Code Quality Checklist

Before requesting human approval:

- [ ] Tests written BEFORE implementation (TDD Red phase)
- [ ] All tests pass (TDD Green phase)
- [ ] Code refactored for clarity (TDD Refactor phase)
- [ ] Issue requirements met
- [ ] Code formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] No warnings
- [ ] CHANGELOG.md updated if applicable

## References

- See CLAUDE.md for testing conventions
