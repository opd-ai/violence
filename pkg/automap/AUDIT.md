# Audit: github.com/opd-ai/violence/pkg/automap
**Date**: 2026-03-01
**Status**: Complete

## Summary
The automap package provides an in-game auto-mapping system with exploration tracking and annotation features. Code health is excellent with 94.1% test coverage, clean API design, and no critical issues. Primary concerns are a stub rendering function and unsafe global genre state.

## Issues Found
- [ ] high stub — Render method is empty stub with no implementation (`automap.go:75`)
- [ ] med concurrency — Global `currentGenre` variable not protected by mutex, unsafe for concurrent access (`automap.go:77`)
- [ ] med api-design — SetGenre/GetCurrentGenre use package-level global state instead of map instance state (`automap.go:80-87`)
- [ ] low documentation — Missing doc.go file for package-level documentation (`automap.go:1`)
- [ ] low bounds-check — AddAnnotation does not validate x/y coordinates are within map bounds (`automap.go:49`)

## Test Coverage
94.1% (target: 65%)

## Dependencies
No external dependencies. Zero imports from standard library or third-party packages.

## Recommendations
1. Implement Render() method or remove if functionality moved elsewhere
2. Move genre state into Map struct or protect global with sync.RWMutex
3. Add bounds validation to AddAnnotation to match Reveal's behavior
4. Consider adding doc.go for package documentation
