package predicates

import (
	"context"
	"errors"
	"io/fs"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
)

// FS contains replacements for some built-in predicates that use the io/fs interface instead of the OS.
type FS struct {
	fsys fs.FS
	i    *prolog.Interpreter
}

func NewFS(fsys fs.FS, i *prolog.Interpreter) FS {
	return FS{
		fsys: fsys,
		i:    i,
	}
}

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

func (ff FS) DirectoryFiles(directory, files engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var dir string
	switch directory := env.Resolve(directory).(type) {
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	case engine.Atom:
		dir = string(directory)
	default:
		return engine.Error(engine.TypeErrorAtom(directory))
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		var paths []engine.Term
		err := fs.WalkDir(ff.fsys, dir, func(path string, d fs.DirEntry, err error) error {
			// don't include root
			if dir == path {
				return nil
			}

			paths = append(paths, engine.Atom(path))

			if d.IsDir() {
				// no recursion in subdirectories
				return fs.SkipDir
			}
			return nil
		})
		if err != nil {
			return engine.Error(err)
		}
		return engine.Unify(files, engine.List(paths...), k, env)
	})
}

func (ff FS) DirectoryExists(directory engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var dir string
	switch directory := env.Resolve(directory).(type) {
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	case engine.Atom:
		dir = string(directory)
	default:
		return engine.Error(engine.TypeErrorAtom(directory))
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

func (ff FS) FileExists(file engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	var f string
	switch file := env.Resolve(file).(type) {
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	case engine.Atom:
		f = string(file)
	default:
		return engine.Error(engine.TypeErrorAtom(file))
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

// consult/1.
func (ff FS) Consult(files engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	switch f := env.Resolve(files).(type) {
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
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
		return engine.DomainError("source_sink", file)
	default:
		return engine.TypeError("atom", file)
	}
}
