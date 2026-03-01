# Audit: github.com/opd-ai/violence/pkg/crafting
**Date**: 2026-03-01
**Status**: Complete

## Summary
The crafting package provides scrap-to-ammo crafting and recipe management with genre-specific material types. The package is well-designed with proper concurrency controls, excellent test coverage (97.8%), and clean separation of concerns. No critical issues found; identified items are minor improvements for API consistency and documentation.

## Issues Found
- [ ] low documentation — Scrap struct lacks godoc comment (`crafting.go:10`)
- [ ] low documentation — No package doc.go file for extended documentation
- [ ] med api-design — Global mutable state (genreRecipes, currentGenre vars) not thread-safe (`crafting.go:190-193`)
- [ ] low api-design — GetRecipes() and GetRecipe() depend on global state rather than CraftingMenu instance
- [ ] low error-handling — CraftingMenu.Craft() consumes materials even if one Remove() succeeds but subsequent fails (`crafting.go:158-162`)
- [ ] low naming — Scrap struct is exported but only used internally, consider unexported
- [ ] med concurrency — SetGenre() modifies global currentGenre without mutex protection (`crafting.go:230-235`)
- [ ] low testing — Missing concurrent test for ScrapStorage despite RWMutex usage

## Test Coverage
97.8% (target: 65%)

## Dependencies
- Standard library only: `fmt`, `sync`
- No external dependencies
- No circular imports

## Recommendations
1. Add mutex protection for genreRecipes and currentGenre global variables or refactor to instance-based design
2. Implement atomic material consumption in CraftingMenu.Craft() to prevent partial failure
3. Add doc.go file with package-level documentation and usage examples
4. Add godoc comments for Scrap struct
5. Consider making Scrap unexported if only used internally
