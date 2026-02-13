package selector

// Node is an AST node in a selector expression.
type Node interface {
	node()
}

// BinaryOp represents && or ||.
type BinaryOp struct {
	Op    TokenKind // TokenAnd or TokenOr
	Left  Node
	Right Node
}

// NotOp represents !expr.
type NotOp struct {
	Expr Node
}

// Comparison represents key == "value", key != "value", key =~ "pattern", key !~ "pattern".
type Comparison struct {
	Key string
	Op  TokenKind // TokenEq, TokenNeq, TokenMatch, TokenNotMatch
	Val string
}

func (*BinaryOp) node()   {}
func (*NotOp) node()      {}
func (*Comparison) node() {}
