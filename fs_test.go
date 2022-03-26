package predicates

import (
	"testing"
	"testing/fstest"

	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/chars"
	"github.com/guregu/predicates/internal"
)

func TestFS(t *testing.T) {
	p := internal.NewTestProlog()
	fsys := fstest.MapFS{
		"test.pl":    {Data: []byte("hello(world).")},
		"dir/a.pl":   {Data: []byte("path('dir/a.pl').")},
		"dir/b.pl":   {Data: []byte("path('dir/b.pl').")},
		"dir/c/1.pl": {Data: []byte("path('dir/c/1.pl').")},
	}
	ff := NewFS(fsys, p.Interpreter)
	ff.Register()

	t.Run("consult/1", func(t *testing.T) {
		p.MustExec(t, ":- consult(test).")
		t.Run("value is atom", p.Expect([]map[string]engine.Term{
			{"X": engine.Atom("world")},
		}, `hello(X).`))
	})

	t.Run("directory_files/2", func(t *testing.T) {
		t.Run("files is variable", p.Expect([]map[string]engine.Term{
			{"Files": chars.List("dir/a.pl", "dir/b.pl", "dir/c")},
		}, `directory_files("dir", Files).`))
	})

	t.Run("file_exists/1", func(t *testing.T) {
		t.Run("succeed", p.Expect(internal.TestOK, `file_exists("test.pl"), OK = true.`))
		t.Run("fail", p.Expect(internal.TestOK, `\+file_exists("dir/c"), OK = true.`))
	})

	t.Run("directory_exists/1", func(t *testing.T) {
		t.Run("succeed", p.Expect(internal.TestOK, `directory_exists("dir/c"), OK = true.`))
		t.Run("fail", p.Expect(internal.TestOK, `\+directory_exists("dir/a.pl"), OK = true.`))
	})
}
