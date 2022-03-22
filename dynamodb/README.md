# dynamodb [![GoDoc](https://godoc.org/github.com/guregu/predicates/dynamodb?status.svg)](https://godoc.org/github.com/guregu/predicates/dynamodb)
`import "github.com/guregu/predicates/dynamodb"`

DynamoDB client for [ichiban/prolog](https://github.com/ichiban/prolog). It uses [guregu/dynamo](https://github.com/guregu/dynamo) under the hood.
Very experimental, still playing around with the API.

## Predicates
```prolog
% tables
list_tables(-Table).

% records
scan(+Table, -Item).
get_item(+Table, +Key, -Item).
put_item(+Table, +Item).

% converting between DynamoDB attribute values (Attr) and friendly Prolog values (Value).
attribute_value(+Attr, -Value).
attribute_value(-Attr, +Value).
```

## Data representation
```prolog
% attributes have their DynamoDB type as their functor
n('42').
s('hello world').
% numbers can be atoms or numbers
n(42).
% maps use key-value lists
m([key-s(value)]).

% for predicates like get_item that take a Key, the Key is in partitionkey-type(value) form
get_item(table, my_attribute-s(some_value), Item).
% use the -&- operator to specify the sort key (range key)
% such that the key is in pk-type(v)-&-sk-type(v) form
get_item(table, userid-n(42)-&-date-s('2022'), Item).
% you can also use key(pk-type(v), sk-type(v)) if you don't want to use my ugly operator
```

## TODO
- [ ] `query/3`
- [ ] `delete_item/2`
- [ ] `update_item/?`
- [ ] `describe_table/2`
- [ ] options: conditions, filters, expressions (`query/4` etc.)
- [ ] transactions