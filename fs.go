package predicates

import (
	"context"
	"errors"
	"io/fs"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/chars"
)

// FS provides native file system predicates.
// Non-ISO predicates are intended to maintain compatibility with Scryer Prolog's library(files).
// See: https://github.com/mthom/scryer-prolog/blob/master/src/lib/files.pl
type FS struct {
	fsys fs.FS
	i    *prolog.Interpreter
}

// NewFS returns a collection of filesystem predicates tied to fsys and i.
func NewFS(fsys fs.FS, i *prolog.Interpreter) FS {
	return FS{
		fsys: fsys,
		i:    i,
	}
}

// Register is a convenience method that registers all FS predicates with their default names. This will replace the default consult/1.
// To register these with custom names, use the interpreter's Register functions and pass a method reference instead.
func (ff FS) Register() {
	ff.i.Exec(`
		:- built_in(consult/1).
		:- built_in(directory_files/2).
		:- built_in(directory_exists/1).
		:- built_in(file_exists/1).
	`)
	ff.i.Register1("consult", ff.Consult)
	ff.i.Register2("directory_files", ff.DirectoryFiles)
	ff.i.Register1("directory_exists", ff.DirectoryExists)
	ff.i.Register1("file_exists", ff.FileExists)
}

// DirectoryFiles (directory_files/2) succeeds if files is a list of strings that contains all entries (including directories) of directory, which must be a string.
// This is useful for obtaining a list of files and directories.
// Throws an error if directory is not a string.
//
// 	directory_files(+Directory, -Files).
func (ff FS) DirectoryFiles(directory, files engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var dir string
	switch directory := env.Resolve(directory).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case *engine.Compound:
		var err error
		dir, err = chars.Value[string](directory, env)
		if err != nil {
			return engine.Error(err)
		}
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, directory, env))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		var entries []engine.Term
		err := fs.WalkDir(ff.fsys, dir, func(path string, d fs.DirEntry, err error) error {
			// don't include root
			if dir == path {
				return nil
			}

			entries = append(entries, chars.String(path))

			if d.IsDir() {
				// no recursion in subdirectories
				return fs.SkipDir
			}
			return nil
		})
		if err != nil {
			return engine.Error(err)
		}
		return engine.Unify(files, engine.List(entries...), k, env)
	})
}

// DirectoryExists (directory_exists/1) succeeds if a directory exists at the path given by the string directory.
// Throws an error if directory is not a string.
//
// 	directory_exists(+Directory).
func (ff FS) DirectoryExists(directory engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var dir string
	switch directory := env.Resolve(directory).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case *engine.Compound:
		var err error
		dir, err = chars.Value[string](directory, env)
		if err != nil {
			return engine.Error(err)
		}
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, directory, env))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		stat, err := fs.Stat(ff.fsys, dir)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return engine.Bool(false)
		case err != nil:
			return engine.Error(err)
		case !stat.IsDir():
			return engine.Bool(false)
		}
		return k(env)
	})
}

// FileExists (file_exists/1) succeeds if a file exists at the path given by the string file.
// Throws an error if file is not a string.
//
// 	file_exists(+File).
func (ff FS) FileExists(file engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var f string
	switch file := env.Resolve(file).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case *engine.Compound:
		var err error
		f, err = chars.Value[string](file, env)
		if err != nil {
			return engine.Error(err)
		}
	default:
		return engine.Error(engine.TypeError(engine.ValidTypeList, file, env))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		stat, err := fs.Stat(ff.fsys, f)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return engine.Bool(false)
		case err != nil:
			return engine.Error(err)
		case stat.IsDir():
			return engine.Bool(false)
		}
		return k(env)
	})
}

// copied from ichiban/prolog and slightly modified

// Consult (consult/1) reads and executes the given file (if given an atom) or files (if given a list of atoms).
// ".pl" will be automatically appended to the file names when needed.
// Throws an error if files is not an atom or list of atoms.
//
//	consult(+FileOrList).
func (ff FS) Consult(files engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	switch f := env.Resolve(files).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case *engine.Compound:
		if f.Functor == "." && len(f.Args) == 2 {
			iter := engine.ListIterator{List: f, Env: env}
			for iter.Next() {
				if err := ff.consultOne(iter.Current(), env); err != nil {
					return engine.Error(err)
				}
			}
			if err := iter.Err(); err != nil {
				return engine.Error(err)
			}
			return k(env)
		}
		if err := ff.consultOne(f, env); err != nil {
			return engine.Error(err)
		}
		return k(env)
	default:
		if err := ff.consultOne(f, env); err != nil {
			return engine.Error(err)
		}
		return k(env)
	}
}

func (ff FS) consultOne(file engine.Term, env *engine.Env) error {
	switch f := env.Resolve(file).(type) {
	case engine.Atom:
		for _, f := range []string{string(f), string(f) + ".pl"} {
			b, err := fs.ReadFile(ff.fsys, f)
			if err != nil {
				continue
			}

			if err := ff.i.Exec(string(b)); err != nil {
				return err
			}

			return nil
		}
		return engine.DomainError(engine.ValidDomainSourceSink, file, env)
	default:
		return engine.TypeError(engine.ValidTypeAtom, file, env)
	}
}
