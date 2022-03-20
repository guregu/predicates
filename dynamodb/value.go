package dynamodb

import (
	"encoding/base64"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ichiban/prolog/engine"

	"github.com/guregu/predicates/internal"
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
		return engine.Atom("null").Apply(engine.Atom("true"))
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
		return &dynamodb.AttributeValue{S: aws.String(string(v))}, nil
	case engine.Integer:
		return &dynamodb.AttributeValue{N: aws.String(string(strconv.FormatInt(int64(v), 10)))}, nil
	case engine.Float:
		return &dynamodb.AttributeValue{N: aws.String(string(strconv.FormatFloat(float64(v), 'f', -1, 64)))}, nil
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
			if a, ok := arg.(engine.Atom); ok {
				switch a {
				case "true":
					return &dynamodb.AttributeValue{BOOL: aws.Bool(true)}, nil
				case "false":
					return &dynamodb.AttributeValue{BOOL: aws.Bool(false)}, nil
				default:
					return nil, engine.DomainError("boolean", a)
				}
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
			return makemap(arg, env)
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
			switch x := arg.(type) {
			case engine.Atom:
				return &dynamodb.AttributeValue{N: aws.String(string(x))}, nil
			case engine.Integer:
				return &dynamodb.AttributeValue{N: aws.String(string(strconv.FormatInt(int64(x), 10)))}, nil
			case engine.Float:
				return &dynamodb.AttributeValue{N: aws.String(string(strconv.FormatFloat(float64(x), 'f', 64, 64)))}, nil
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
		case "null":
			if a, ok := arg.(engine.Atom); ok && a == "true" {
				return &dynamodb.AttributeValue{NULL: aws.Bool(true)}, nil
			}
			return nil, engine.DomainError("true", arg)
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

		case ".":
			// prolog list
			// try to figure out if it's a M like [foo-bar] or L like [foo]
			// TODO: maybe this is dumb idk

			isMap := true
			engine.EachList(v, func(elem engine.Term) error {
				if !isMap {
					return nil
				}
				cmp, ok := env.Resolve(elem).(*engine.Compound)
				if !ok {
					isMap = false
					return nil
				}
				if cmp.Functor != "-" || len(cmp.Args) != 2 {
					log.Println(cmp.Functor, cmp.Args)
					isMap = false
				}
				return nil
			}, env)

			if isMap {
				return makemap(v, env)
			}
			return makelist(v, env)
		default:
			return nil, engine.DomainError("attribute_value", v)
		}
	}
	return nil, internal.TypeErrorCompound(v)
}

func makemap(arg engine.Term, env *engine.Env) (*dynamodb.AttributeValue, error) {
	av := &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{}}
	err := engine.EachList(arg, func(elem engine.Term) error {
		key, val, err := splitkey(env.Resolve(elem), env)
		if err != nil {
			return err
		}
		avv, err := prolog2av(val, env)
		if err != nil {
			return err
		}
		av.M[key] = avv
		return nil
	}, env)
	return av, err
}

func makelist(arg engine.Term, env *engine.Env) (*dynamodb.AttributeValue, error) {
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
}

func simplify(v engine.Term, env *engine.Env) (engine.Term, error) {
	switch v := env.Resolve(v).(type) {
	case engine.Variable:
		return nil, engine.ErrInstantiation
	case engine.Atom:
		return v, nil
	case engine.Integer:
		return v, nil
	case engine.Float:
		return v, nil
	case *engine.Compound:
		arg := env.Resolve(v.Args[0])
		switch v.Functor {
		case "l":
			list := make([]engine.Term, 0)
			err := engine.EachList(arg, func(elem engine.Term) error {
				val, err := simplify(elem, env)
				if err != nil {
					return err
				}
				list = append(list, val)
				return nil
			}, env)
			if err != nil {
				return nil, err
			}
			return engine.List(list...), nil
		case "m":
			list := make([]engine.Term, 0)
			err := engine.EachList(arg, func(elem engine.Term) error {
				key, val, err := splitkey(env.Resolve(elem), env)
				if err != nil {
					return err
				}
				sv, err := simplify(val, env)
				if err != nil {
					return err
				}
				member := engine.Atom("-").Apply(engine.Atom(key), sv)
				list = append(list, member)
				return nil
			}, env)
			if err != nil {
				return nil, err
			}
			return engine.List(list...), nil
		case "n":
			switch x := arg.(type) {
			case engine.Atom:
				// TODO: make this an option instead
				if strings.ContainsRune(string(x), '.') {
					f, err := strconv.ParseFloat(string(x), 64)
					if err != nil {
						return nil, err // TODO: wrap error?
					}
					return engine.Float(f), nil
				}
				n, err := strconv.ParseInt(string(x), 10, 64)
				if err != nil {
					return nil, err
				}
				return engine.Integer(n), nil
			case engine.Integer:
				return engine.Integer(x), nil
			case engine.Float:
				return engine.Float(x), nil
			}
			return nil, internal.TypeErrorAtom(arg)
		case "s":
			if a, ok := arg.(engine.Atom); ok {
				return engine.Atom(a), nil
			}
			return nil, internal.TypeErrorAtom(arg)
		default:
			return v, nil
		}
	}
	return nil, internal.TypeErrorCompound(v)
}

func list2item(list engine.Term, env *engine.Env) (map[string]*dynamodb.AttributeValue, error) {
	avs := make(map[string]*dynamodb.AttributeValue)
	err := engine.EachList(env.Resolve(list), func(elem engine.Term) error {
		key, av, err := parsekey(env.Resolve(elem), env)
		if err != nil {
			return err
		}
		avs[key] = av
		return nil
	}, env)
	return avs, err
}

func tableName(table engine.Term) (string, *engine.Exception) {
	switch table := table.(type) {
	case engine.Atom:
		return string(table), nil
	}
	return "", engine.ErrInstantiation
}
