package dynamodb

import (
	"encoding/base64"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/predicates/internal"
	"github.com/ichiban/prolog/engine"
)

func av2prolog(av *dynamodb.AttributeValue) engine.Term {
	switch {
	case av.B != nil:
		enc := base64.StdEncoding.EncodeToString(av.B)
		return engine.Atom("b").Apply(engine.Atom(enc))
	case av.BS != nil:
		list := make([]engine.Term, 0, len(av.L))
		for _, v := range av.BS {
			enc := base64.StdEncoding.EncodeToString(v)
			list = append(list, engine.Atom("s").Apply(engine.Atom(enc)))
		}
		return engine.Atom("ss").Apply(engine.List(list...))
	case av.BOOL != nil:
		bool := engine.Atom("false")
		if *av.BOOL {
			bool = engine.Atom("true")
		}
		return engine.Atom("bool").Apply(bool)
	case av.L != nil:
		list := make([]engine.Term, 0, len(av.L))
		for _, v := range av.L {
			list = append(list, av2prolog(v))
		}
		return engine.Atom("l").Apply(engine.List(list...))
	case av.M != nil:
		list := make([]engine.Term, 0, len(av.M))
		for k, v := range av.M {
			item := engine.Atom("-").Apply(engine.Atom(k), av2prolog(v))
			list = append(list, item)
		}
		sortTerms(list)
		return engine.Atom("m").Apply(engine.List(list...))
	case av.N != nil:
		return engine.Atom("n").Apply(engine.Atom(*av.N))
	case av.NS != nil:
		list := make([]engine.Term, 0, len(av.L))
		for _, v := range av.NS {
			list = append(list, engine.Atom("n").Apply(engine.Atom(*v)))
		}
		return engine.Atom("ns").Apply(engine.List(list...))
	case av.NULL != nil:
		return engine.Atom("null")
	case av.S != nil:
		return engine.Atom("s").Apply(engine.Atom(*av.S))
	case av.SS != nil:
		list := make([]engine.Term, 0, len(av.L))
		for _, v := range av.SS {
			list = append(list, engine.Atom("s").Apply(engine.Atom(*v)))
		}
		return engine.Atom("ss").Apply(engine.List(list...))
	}
	return nil
}

func sortTerms(list []engine.Term) {
	sort.Slice(list, func(i, j int) bool {
		return list[i].Compare(list[j], nil) < 0
	})
}

func item2prolog(item map[string]*dynamodb.AttributeValue) engine.Term {
	t := av2prolog(&dynamodb.AttributeValue{M: item}).(*engine.Compound)
	return t.Args[0]
}

func splitkey(t engine.Term, env *engine.Env) (key string, value engine.Term, err error) {
	switch t := env.Resolve(t).(type) {
	case engine.Variable:
		return "", nil, engine.ErrInstantiation
	case *engine.Compound:
		if t.Functor != "-" || len(t.Args) != 2 {
			return "", nil, internal.TypeErrorPair(t)
		}

		switch keyArg := env.Resolve(t.Args[0]).(type) {
		case engine.Atom:
			key = string(keyArg)
		case engine.Variable:
			return "", nil, engine.ErrInstantiation
		default:
			return "", nil, internal.TypeErrorAtom(keyArg)
		}

		return key, t.Args[1], nil
	}
	return "", nil, internal.TypeErrorPair(t)
}

func splitkeys(t engine.Term, env *engine.Env) (pk, rk engine.Term, err error) {
	switch t := env.Resolve(t).(type) {
	case engine.Variable:
		return nil, nil, engine.ErrInstantiation
	case *engine.Compound:
		if t.Functor == "-" && len(t.Args) == 2 {
			return t, nil, nil
		}

		switch {
		case t.Functor == "-" && len(t.Args) == 2:
			return t, nil, nil
		case t.Functor == "-&-" && len(t.Args) == 2:
			return t.Args[0], t.Args[1], nil
		case t.Functor == "key" && len(t.Args) == 2:
			return t.Args[0], t.Args[1], nil
		}

		// TODO: better error
		return nil, nil, internal.TypeErrorCompound(t)
	}
	return nil, nil, internal.TypeErrorPair(t)
}

func parsekey(t engine.Term, env *engine.Env) (string, *dynamodb.AttributeValue, error) {
	key, val, err := splitkey(env.Resolve(t), env)
	if err != nil {
		return "", nil, err
	}
	av, err := prolog2av(val, env)
	if err != nil {
		return "", nil, err
	}
	return key, av, nil
}

func prolog2av(v engine.Term, env *engine.Env) (*dynamodb.AttributeValue, error) {
	switch v := env.Resolve(v).(type) {
	case engine.Variable:
		return nil, engine.ErrInstantiation
	case engine.Atom:
		if v == "null" {
			return &dynamodb.AttributeValue{NULL: aws.Bool(true)}, nil
		}
		return nil, internal.TypeErrorCompound(v)
	case *engine.Compound:
		arg := env.Resolve(v.Args[0])
		switch v.Functor {
		case "b":
			if a, ok := arg.(engine.Atom); ok {
				b, err := base64.StdEncoding.DecodeString(string(a))
				if err != nil {
					return nil, err
				}
				return &dynamodb.AttributeValue{B: b}, nil
			}
			return nil, internal.TypeErrorAtom(arg)
		case "bs":
			av := &dynamodb.AttributeValue{BS: [][]byte{}}
			err := engine.EachList(arg, func(elem engine.Term) error {
				switch elem := env.Resolve(elem).(type) {
				case engine.Atom:
					b, err := base64.StdEncoding.DecodeString(string(elem))
					if err != nil {
						return err
					}
					av.BS = append(av.BS, b)
					return nil
				}
				return internal.TypeErrorAtom(elem)
			}, env)
			return av, err
		case "bool":
			// TODO: check for invalid values
			if a, ok := arg.(engine.Atom); ok {
				return &dynamodb.AttributeValue{BOOL: aws.Bool(a == "true")}, nil
			}
			return nil, internal.TypeErrorAtom(arg)
		case "l":
			av := &dynamodb.AttributeValue{L: []*dynamodb.AttributeValue{}}
			err := engine.EachList(arg, func(elem engine.Term) error {
				item, err := prolog2av(env.Resolve(elem), env)
				if err != nil {
					return err
				}
				av.L = append(av.L, item)
				return nil
			}, env)
			return av, err
		case "m":
			av := &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{}}
			err := engine.EachList(arg, func(elem engine.Term) error {
				key, val, err := splitkey(env.Resolve(elem), env)
				if err != nil {
					return err
				}
				av, err := prolog2av(val, env)
				if err != nil {
					return err
				}
				av.M[key] = av
				return nil
			}, env)
			return av, err
		case "n":
			if a, ok := arg.(engine.Atom); ok {
				return &dynamodb.AttributeValue{N: aws.String(string(a))}, nil
			}
			return nil, internal.TypeErrorAtom(arg)
		case "ns":
			av := &dynamodb.AttributeValue{NS: []*string{}}
			err := engine.EachList(arg, func(elem engine.Term) error {
				switch elem := env.Resolve(elem).(type) {
				case engine.Atom:
					av.NS = append(av.NS, aws.String(string(elem)))
					return nil
				}
				return internal.TypeErrorAtom(elem)
			}, env)
			return av, err
		// case "null"
		case "s":
			if a, ok := arg.(engine.Atom); ok {
				return &dynamodb.AttributeValue{S: aws.String(string(a))}, nil
			}
			return nil, internal.TypeErrorAtom(arg)
		case "ss":
			av := &dynamodb.AttributeValue{SS: []*string{}}
			err := engine.EachList(arg, func(elem engine.Term) error {
				switch elem := env.Resolve(elem).(type) {
				case engine.Atom:
					av.SS = append(av.SS, aws.String(string(elem)))
					return nil
				}
				return internal.TypeErrorAtom(elem)
			}, env)
			return av, err
		}
	}
	return nil, internal.TypeErrorCompound(v)
}

func tableName(table engine.Term) (string, *engine.Exception) {
	switch table := table.(type) {
	case engine.Atom:
		return string(table), nil
	}
	return "", engine.ErrInstantiation
}
