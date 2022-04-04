# predicates [![GoDoc](https://godoc.org/github.com/guregu/predicates?status.svg)](https://godoc.org/github.com/guregu/predicates)
`import "github.com/guregu/predicates"`

Native predicates for [ichiban/prolog](https://github.com/ichiban/prolog).

## Prolog

Filesystem predicates using [`io/fs.FS`](https://pkg.go.dev/io/fs). 

### Built-in replacements

- `consult/1`

### `library(files)`

These predicates are intended to be compatible with Scryer Prolog's [`library(files)`](https://github.com/mthom/scryer-prolog/blob/master/src/lib/files.pl).
These use strings (lists of characters) for filenames.

- `directory_files/2`
- `directory_exists/1`
- `file_exists/1`

### Package [`taujson`](https://godoc.org/github.com/guregu/predicates/taujson)

These predicates are intended to be compatible with Tau Prolog's [`library(js)`](http://tau-prolog.org/documentation#js).
They are peculiar predicates that use opaque native objects for JSON conversion.

- `json_atom/2`
- `json_prolog/2`

### Graduated

- [`between/3`](https://github.com/ichiban/prolog/releases/tag/v0.9.0) made it into ichiban/prolog in `v0.9.0`!

## Go

Package [`chars`](https://godoc.org/github.com/guregu/predicates/chars) provides some convenience functions for working with Prolog strings.

## License

BSD 2-clause. Includes code from ichiban/prolog (MIT license).
See LICENSE.