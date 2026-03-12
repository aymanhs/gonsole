---
description: How to build the gonsole package and its examples
---
This workflow describes how to build the project. Note that `GOOS=windows` is required for all build commands.

1. Ensure you are in the project root: `/home/ayman/code/gonsole`
// turbo
2. Build the entire project:
```bash
GOOS=windows go build ./...
```

3. Build a specific example (e.g., flappy):
```bash
GOOS=windows go build ./cmd/flappy
```
