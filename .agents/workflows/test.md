---
description: How to run tests for the gonsole package
---
This workflow describes how to run tests. Note that `GOOS=windows` is required for consistency across platforms.

1. Ensure you are in the project root: `/home/ayman/code/gonsole`
// turbo
2. Run all tests:
```bash
GOOS=windows go test ./...
```
