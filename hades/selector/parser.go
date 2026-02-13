package selector

import "fmt"

type Parser struct {
	tokens []Token
	pos    int
}

func Parse(tokens []Token) (Node, error) {
	p := &Parser{tokens: tokens}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.current().Kind != TokenEOF {
		return nil, fmt.Errorf("unexpected token %s at position %d", p.current().Kind, p.current().Pos)
	}
	return node, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TokenEOF, Pos: -1}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.current()
	p.pos++
	return tok
}

// expr → or_expr
func (p *Parser) parseExpr() (Node, error) {
	return p.parseOr()
}

// or_expr → and_expr ("||" and_expr)*
func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.current().Kind == TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: TokenOr, Left: left, Right: right}
	}
	return left, nil
}

// and_expr → unary ("&&" unary)*
func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.current().Kind == TokenAnd {
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: TokenAnd, Left: left, Right: right}
	}
	return left, nil
}

// unary → "!" unary | primary
func (p *Parser) parseUnary() (Node, error) {
	if p.current().Kind == TokenNot {
		p.advance()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &NotOp{Expr: expr}, nil
	}
	return p.parsePrimary()
}

// primary → comparison | "(" expr ")"
func (p *Parser) parsePrimary() (Node, error) {
	if p.current().Kind == TokenLParen {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.current().Kind != TokenRParen {
			return nil, fmt.Errorf("expected ')' at position %d, got %s", p.current().Pos, p.current().Kind)
		}
		p.advance()
		return expr, nil
	}

	return p.parseComparison()
}

// comparison → IDENT op STRING
func (p *Parser) parseComparison() (Node, error) {
	if p.current().Kind != TokenIdent {
		return nil, fmt.Errorf("expected identifier at position %d, got %s", p.current().Pos, p.current().Kind)
	}
	ident := p.advance()

	op := p.current().Kind
	if op != TokenEq && op != TokenNeq && op != TokenMatch && op != TokenNotMatch {
		return nil, fmt.Errorf("expected operator (==, !=, =~, !~) at position %d, got %s", p.current().Pos, p.current().Kind)
	}
	p.advance()

	if p.current().Kind != TokenString {
		return nil, fmt.Errorf("expected string at position %d, got %s", p.current().Pos, p.current().Kind)
	}
	val := p.advance()

	return &Comparison{Key: ident.Val, Op: op, Val: val.Val}, nil
}
