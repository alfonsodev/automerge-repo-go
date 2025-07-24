# Gemini Agent Development Guidelines

This document outlines the best practices for Gemini agents contributing to the `automerge-repo-go` project. Adhering to these guidelines is crucial for maintaining code quality, consistency, and ensuring a smooth development process between sessions.

## 1. The Prime Directive: Context First

Before taking any action, an agent's first step is to build a complete understanding of the project's current state.

- **Always start by reading `AGENTS.md` and `ROADMAP_PROGRESS.md`.** These files are the single source of truth for the project's status and immediate goals.

## 2. The Development Workflow

Follow this workflow for every task:

1.  **Analyze Existing Code:** Before writing or modifying code, thoroughly examine the relevant Go files (`*.go`) and their corresponding tests (`*_test.go`). The goal is to match the existing style and patterns perfectly.

2.  **Write Idiomatic Go:** All code must follow standard Go idioms, including error handling (`if err != nil`), package structure, and naming conventions.

3.  **Testing is Mandatory:**
    -   Any bug fix must be accompanied by a new test that would have failed before the fix.
    -   Any new feature must have corresponding unit tests.
    -   All tests must follow the style and structure of existing tests in the package.

4.  **Verify All Changes:** After any modification, run the full test suite to ensure no regressions have been introduced:
    ```shell
    go test ./...
    ```
    This command must pass before the task is considered complete.

## 3. Closing the Loop: Update The Record

After successfully implementing and verifying a change, the final and most critical step is to document the work.

1.  **Update `ROADMAP_PROGRESS.md`:** Add a new entry summarizing the changes you made. Be specific about what was implemented or fixed.
2.  **Update `AGENTS.md`:** Mark the completed task in the checklist and ensure the "Next Steps" section clearly indicates the next priority.

## 4. Agent Sign-Off

Upon successfully completing a task and updating the project documentation, conclude your session with a unique, fun, or clever sign-off sentence. This confirms you have read and adhered to these guidelines. **You must include an emoji.**

*Example sign-offs:*
*   "Job's done, and this repo is sparkling a little brighter now. ✨"
*   "Another task bites the dust. On to the next adventure! 🚀"
*   "Code's compiled, tests are green, and the docs are clean. Mission accomplished. 🤓"

By following this "Context, Code, Test, Document, Sign-off" cycle, we ensure that any agent can effectively contribute to the project at any time.
