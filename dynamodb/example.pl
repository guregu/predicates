peek :- list_tables(T), once(scan(T, Item)), write(Item).

:- between(1,3,X), put_item('TestDB', ['UserID'-n(X), 'Time'-s('400')]).

:- put_item('foo', [id-n(1)]).