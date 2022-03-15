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

	sols, err := p.Interpreter.Query("list_tables(T), once(scan(T, Item)), write(Item).")
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
}
