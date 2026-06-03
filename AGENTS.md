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
