# Header/Body Classification Fix - Implementation Report

**Date**: September 18, 2025
**Project**: Better-Curl (Saul)
**Issue**: Body variables appear as HTTP headers instead of JSON body
**Status**: ❌ **FAILED** - Original bug still exists

---

## Summary

**We fixed the wrong thing.**

Implemented file-based classification (headers.toml → headers, body.toml → body) but **body content still appears as HTTP headers**. The architectural change was clean and maintained all tests, but didn't solve the user problem.

---

## What We Tried

### Approach: File-Based Classification
- **Theory**: Replace type-based detection with source-file routing
- **Implementation**: Complete rewrite of BuildHTTPRequest() function
- **Changes**: Created PresetHandlers struct, LoadPresetFiles(), file-specific routing

### Code Changes
```go
// OLD: Type-based (string = headers, non-string = body)
if strValue, ok := value.(string); ok {
    config.Headers[key] = strValue
}

// NEW: File-based (headers.toml → headers, body.toml → body)
if handlers.Headers != nil {
    config.Headers[key] = handlers.Headers.GetAsString(key)
}
if handlers.Body != nil {
    config.Body = handlers.Body.ToJSON()
}
```

---

## What Didn't Work

### The Bug Persists
```bash
# Test case that still fails:
./saul test set body message="hello"
./saul call test

# Result:
Headers: "Message": "hello"  # ❌ Wrong location
Body: ""                     # ❌ Empty
```

### Root Cause Analysis
- ✅ **File separation works**: Content correctly loaded from separate files
- ✅ **JSON generation works**: `ToJSON()` produces `{"message":"hello"}`
- ❌ **HTTP request construction fails**: Body JSON becomes headers somewhere downstream

---

## What We Learned

1. **Wrong Layer**: The bug isn't in TOML classification but in HTTP request building
2. **Test Coverage Gap**: Existing tests don't validate actual HTTP request content
3. **Agent Analysis Incomplete**: Missed the real location of the bug
4. **Architecture vs Bug Fix**: Clean code doesn't always fix user problems

---

## Current State

### Working
- All existing tests pass
- File-based architecture implemented
- Headers from headers.toml work correctly
- Code compiles and runs

### Broken
- Body content still appears as HTTP headers
- Original user problem unsolved
- Edge case investigation needed

---

## Next Steps

1. **Investigate HTTP client integration** - trace resty request construction
2. **Add debugging** to see where body JSON becomes headers
3. **Consider reverting** if downstream fix is too complex
4. **Real root cause analysis** - find the actual bug location

---

## Conclusion

**Clean architecture ≠ bug fix.** We improved the codebase but failed to solve the user problem. The real issue is likely in the HTTP client integration, not the TOML handling we targeted.

**Lesson**: Validate assumptions about root cause before major architectural changes.