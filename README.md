# kit

[![CI](https://github.com/jimmysawczuk/kit/actions/workflows/ci.yml/badge.svg)](https://github.com/jimmysawczuk/kit/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/jimmysawczuk/kit.svg)](https://pkg.go.dev/github.com/jimmysawczuk/kit) [![Go Report Card](https://goreportcard.com/badge/github.com/jimmysawczuk/kit)](https://goreportcard.com/report/github.com/jimmysawczuk/kit)

**Kit** is a collection of packages that I use to build Go services and programs. A few of the design goals (determining whether kit meets these goals is an exercise for the reader):

- **Modularity:** Use as much or as little of Kit as you want. Using one package within kit shouldn't require using another one as well.
- **Few dependencies:** Kit does use some third-party packages, but the ones it uses are required to have permissive licenses and should generally be well-known, well-designed and well-supported.
- **Well-tested:** Kit's packages should have good coverage from unit and other automated tests. _This is still a work-in-progress._

**IMPORTANT:** Kit is still being actively designed and developed and its API may change at any time.

## Acknowledgements

Kit would not exist without several open source packages:

- [Chi](https://github.com/go-chi/chi) provides the underlying router for the `web` package as well as several of the middlewares.
- [github.com/pkg/errors](https://github.com/pkg/errors) is used for wrapping errors.
- [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx) provides some nice-to-have API functions for SQL-based databases.
- [github.com/stretchr/testify](https://github.com/stretchr/testify) is used for testing.
