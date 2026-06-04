# Go Test-Driven Development (TDD) Guidelines

Always follow a strict, vertical Red-Green-Refactor loop when writing or modifying Go code.

## 1. Red-Green-Refactor Loop

1. **RED**: Write a single failing test in a `*_test.go` file targeting the **public interface** of the package. Run `go test ./...` to verify failure.
2. **GREEN**: Write the absolute minimum production code to satisfy the failing test. Run `go test ./...` to verify it passes.
3. **REFACTOR**: Clean up, deduplicate, and optimize code only once in a Green state. Ensure tests still pass.

## 2. Go-Specific Best Practices

- **Vertical over Horizontal Slices**: Do not write all tests first, then implement. Build feature behavior one test/implementation pair at a time.
- **Table-Driven Tests**: Group test cases incrementally. Start with a single subtest and append new cases to test edge cases, errors, and boundaries.
  ```go
  func TestParse(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          want    int
          wantErr bool
      }{
          {
              name:  "valid base",
              input: "123",
              want:  123,
          },
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              got, err := Parse(tt.input)
              if (err != nil) != tt.wantErr {
                  t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                  return
              }
              if got != tt.want {
                  t.Errorf("Parse() got = %v, want %v", got, tt.want)
              }
          })
      }
  }
  ```
- **Test Behaviors, Not Implementations**: Verify expected outputs for given inputs through the public API. Avoid mocking internal components or testing private variables/functions.
- **Local Testing**: Run `go test -v ./path/to/package` to run tests instantly during cycles. Use `go test -run TestName` for precise execution.

---

# Behavioral Guidelines to Reduce Common LLM Coding Mistakes

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For tasks requiring fewer than ~10 lines of new code with no branching logic or external dependencies, these guidelines may be applied lightly.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor unrelated code that isn't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" -> "Write tests for invalid inputs, then make them pass"
- "Fix the bug" -> "Write a test that reproduces it, then make it pass"
- "Refactor X" -> "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
