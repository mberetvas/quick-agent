# Clipboard TUI - Project Plan Index

**Project**: AI-Powered Clipboard Supercharger  
**Goal**: Cross-platform TUI that monitors clipboard and processes text via LLM on hotkey press.

---

## Plan Files

1. **[01-Overview.md](01-Overview.md)** - Architecture, core components, design decisions
2. **[02-Implementation.md](02-Implementation.md)** - File structure, code organization, phases
3. **[03-Configuration.md](03-Configuration.md)** - Config schema, storage, API keys
4. **[04-Platforms.md](04-Platforms.md)** - Windows, macOS, Linux specifics
5. **[05-Testing.md](05-Testing.md)** - Unit, integration, manual testing
6. **[06-Release.md](06-Release.md)** - CI/CD, packaging, distribution

---

## Quick Summary

```
User Flow:
  Copy Text → Hotkey → TUI appears → Select Action → Stream Result → Copy

Architecture:
  [Daemon: Poll Clipboard + Listen Hotkey]
           │ (stdin pipe)
           ▼
  [TUI: Display Text + Options + Result]
           │
           ▼
  [LLM Client: Ollama or OpenRouter]
```

**v1 Duration**: 4-6 weeks  
**Lines of Code**: ~1500-2000 Go  
**Dependencies**: 10 external packages

---

## Open Questions (Resolve First)

1. **Project name** - Module path, binary name, repo name
2. **Repository location** - GitHub/GitLab, personal/org
3. **License** - MIT, Apache 2.0, GPL
4. **Default hotkey** - Ctrl+Alt+V (Win/Linux), Cmd+Option+V (macOS)
5. **Default models** - Ollama: llama3:8b, OpenRouter: mistralai/mistral-7b-instruct

---

## Next Steps

1. Resolve open questions above
2. Review all plan files
3. Approve plan
4. Begin implementation with Phase 1
