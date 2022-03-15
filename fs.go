package predicates

import (
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

// copied from ichiban/prolog and slightly modified

// consult/1.
func (ff FS) Consult(files engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	switch f := env.Resolve(files).(type) {
	case engine.Variable:
		return engine.Error(engine.ErrInstantiation)
	case *engine.Compound:
		if f.Functor == "." && len(f.Args) == 2 {
			if err := engine.EachList(f, func(elem engine.Term) error {
				return ff.consultOne(elem, env)
			}, env); err != nil {
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
