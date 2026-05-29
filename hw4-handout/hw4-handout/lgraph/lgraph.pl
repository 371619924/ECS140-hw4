find_sequence(Graph1, Graph2, Start, Target, K, Sequence) :-
    valid_length(Sequence, K),
    path(Graph1, Start, Target, Sequence),
    not(path(Graph2, Start, Target, Sequence)).

valid_length(Sequence, K) :-
    var(K),
    length(Sequence, K).

valid_length(Sequence, K) :-
    not(var(K)),
    K >= 0,
    length(Sequence, K).

node(Graph, Node) :-
    edge(Graph, Node, _, _).

node(Graph, Node) :-
    edge(Graph, _, _, Node).

path(Graph, Node, Node, []) :-
    node(Graph, Node).

path(Graph, Start, Target, [Label|Rest]) :-
    edge(Graph, Start, Label, Next),
    path(Graph, Next, Target, Rest).