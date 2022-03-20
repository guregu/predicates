package dynamodb

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
)

var (
	truth = []map[string]engine.Term{{}}
	okay  = []map[string]engine.Term{{"OK": engine.Atom("true")}}
	fail  = []map[string]engine.Term(nil)
)

func newDB() *dynamo.DB {
	db := dynamo.New(session.New(), &aws.Config{
		Region:   aws.String("test"),
		Endpoint: aws.String("http://localhost:8880"),
		// LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
	})
	return db
}

// TODO: proper tests...

func TestListTables(t *testing.T) {
	p := internal.NewTestProlog()
	ddb := New(newDB())
	ddb.Register(p.Interpreter)

	// t.Run("list_tables/1", p.Expect(nil, "list_tables(X)."))

	// t.Run("test", p.Expect(nil, "list_tables(T), once(scan(T, Item)), write(Item)."))
}

func TestScan(t *testing.T) {
	p := internal.NewTestProlog()
	ddb := New(newDB())
	ddb.Register(p.Interpreter)

	sols, err := p.Interpreter.Query("list_tables(T), once(scan(T, Item)), write(Item), nl.")
	if err != nil {
		t.Fatal(err)
	}
	for sols.Next() {
		m := make(map[string]engine.Term)
		if err := sols.Scan(m); err != nil {
			t.Fatal(err)
		}
		fmt.Println(m)
	}
	if err := sols.Err(); err != nil {
		t.Error(err)
	}
}

func TestPut(t *testing.T) {
	p := internal.NewTestProlog()
	ddb := New(newDB())
	ddb.Register(p.Interpreter)

	sols, err := p.Interpreter.Query("put_item('TestDB', ['UserID'-n(4002), 'Time'-s('2001')]), X = ok.")
	if err != nil {
		t.Fatal(err)
	}
	for sols.Next() {
		m := make(map[string]engine.Term)
		if err := sols.Scan(m); err != nil {
			t.Fatal(err)
		}
		fmt.Println(m)
	}
	if err := sols.Err(); err != nil {
		t.Error(err)
	}
}

func TestAttributeValue(t *testing.T) {
	p := internal.NewTestProlog()
	ddb := New(newDB())
	ddb.Register(p.Interpreter)

	t.Run("value is variable", func(t *testing.T) {
		t.Run("number int", p.Expect([]map[string]engine.Term{
			{"V": engine.Integer(42)},
		}, "attribute_value(n(42), V)."))

		t.Run("number atom int", p.Expect([]map[string]engine.Term{
			{"V": engine.Integer(42)},
		}, "attribute_value(n('42'), V)."))

		t.Run("number atom float", p.Expect([]map[string]engine.Term{
			{"V": engine.Float(42)},
		}, "attribute_value(n('42.0'), V)."))

		t.Run("string", p.Expect([]map[string]engine.Term{
			{"V": engine.Atom("foo")},
		}, "attribute_value(s(foo), V)."))

		t.Run("map", p.Expect([]map[string]engine.Term{
			{"V": engine.List(
				engine.Atom("-").Apply(engine.Atom("a"), engine.Atom("b")),
				engine.Atom("-").Apply(engine.Atom("c"), engine.Atom("d")),
			)},
		}, "attribute_value(m([a-s(b), c-s(d)]), V)."))

		t.Run("list", p.Expect([]map[string]engine.Term{
			{"V": engine.List(engine.Integer(1), engine.Integer(2), engine.Integer(3))},
		}, "attribute_value(l([n(1), n(2), n(3)]), V)."))
	})

	t.Run("attr is variable", func(t *testing.T) {
		t.Run("map", p.Expect([]map[string]engine.Term{
			{"V": engine.Atom("m").Apply(engine.List(
				engine.Atom("-").Apply(engine.Atom("a"), engine.Atom("s").Apply(engine.Atom("b"))),
				engine.Atom("-").Apply(engine.Atom("c"), engine.Atom("s").Apply(engine.Atom("d"))),
			))},
		}, "attribute_value(V, [a-b, c-d])."))

		t.Run("list", p.Expect([]map[string]engine.Term{
			{"V": engine.Atom("l").Apply(engine.List(
				engine.Atom("n").Apply(engine.Atom("1")),
				engine.Atom("n").Apply(engine.Atom("2")),
				engine.Atom("n").Apply(engine.Atom("3")),
			))},
		}, "attribute_value(V, [1, 2, 3])."))
	})

	t.Run("both ground and equal", p.Expect(okay, "attribute_value(n(1), 1), attribute_value(n('1'), 1), OK = true."))

	t.Run("both ground and not equal", p.Expect(fail, "attribute_value(n(1), 2), OK = true."))
}
