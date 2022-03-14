package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
)

type Dynamo struct {
	db *dynamo.DB
}

func New(db *dynamo.DB) Dynamo {
	d := Dynamo{
		db: db,
	}
	return d
}

func (d Dynamo) Register(p *prolog.Interpreter) {
	d.Bootstrap(p)
	p.Register1("list_tables", d.ListTables)
	p.Register2("scan", d.Scan)
	p.Register3("get_item", d.GetItem3)
}

func (d Dynamo) Bootstrap(p *prolog.Interpreter) {
	if err := p.Exec(`:- op(501, xfx, -&-).`); err != nil {
		panic(err)
	}
}

func (d Dynamo) ListTables(name engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	tables, err := d.db.ListTables().All()
	if err != nil {
		return engine.Error(err)
	}
	ks := make([]func(context.Context) *engine.Promise, 0, len(tables))
	for _, t := range tables {
		table := engine.Atom(t)
		ks = append(ks, func(_ context.Context) *engine.Promise {
			return engine.Unify(name, table, k, env)
		})
	}
	return engine.Delay(ks...)
}

func (d Dynamo) Scan(table, item engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	from, ex := tableName(env.Resolve(table))
	if ex != nil {
		return engine.Error(ex)
	}

	iter := d.db.Table(from).Scan().Iter()
	return engine.Delay(func(context.Context) *engine.Promise {
		return engine.Repeat(func(ctx context.Context) *engine.Promise {
			var result map[string]*dynamodb.AttributeValue
			if !iter.NextWithContext(ctx, &result) {
				// done
				if err := iter.Err(); err != nil {
					return engine.Error(err)
				}
				return engine.Bool(true)
			}
			value := item2prolog(result)
			return engine.Unify(item, value, k, env)
		})
	})
}

func (d Dynamo) GetItem3(table, keys, item engine.Term, k func(*engine.Env) *engine.Promise, env *engine.Env) *engine.Promise {
	from, ex := tableName(env.Resolve(table))
	if ex != nil {
		return engine.Error(ex)
	}

	pk, rk, err := splitkeys(env.Resolve(keys), env)
	if err != nil {
		return engine.Error(err)
	}

	pkName, pkValue, err := parsekey(env.Resolve(pk), env)
	if err != nil {
		return engine.Error(err)
	}
	q := d.db.Table(from).Get(pkName, pkValue)

	if rk != nil {
		rkName, rkValue, err := parsekey(env.Resolve(rk), env)
		if err != nil {
			return engine.Error(err)
		}
		q.Range(rkName, dynamo.Equal, rkValue)
	}

	return engine.Delay(func(context.Context) *engine.Promise {
		var result map[string]*dynamodb.AttributeValue
		err := q.One(&result)
		if err != nil {
			return engine.Error(err)
		}
		it := item2prolog(result)
		return engine.Unify(it, item, k, env)
	})
}
