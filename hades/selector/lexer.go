package selector

import "fmt"

type Lexer struct {
	input  string
	pos    int
	tokens []Token
}

func Lex(input string) ([]Token, error) {
	l := &Lexer{input: input}
	if err := l.lex(); err != nil {
		return nil, err
	}
	return l.tokens, nil
}

func (l *Lexer) lex() error {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.pos++
			continue
		}

		switch ch {
		case '(':
			l.emit(TokenLParen, "(")
		case ')':
			l.emit(TokenRParen, ")")
		case '=':
			if err := l.lexTwoChar('=', TokenEq, '~', TokenMatch); err != nil {
				return err
			}
		case '!':
			if l.pos+1 < len(l.input) {
				next := l.input[l.pos+1]
				if next == '=' {
					l.pos++
					l.emit(TokenNeq, "!=")
					continue
				}
				if next == '~' {
					l.pos++
					l.emit(TokenNotMatch, "!~")
					continue
				}
			}
			l.emit(TokenNot, "!")
		case '&':
			if err := l.expect('&', "&&"); err != nil {
				return err
			}
			l.emit(TokenAnd, "&&")
		case '|':
			if err := l.expect('|', "||"); err != nil {
				return err
			}
			l.emit(TokenOr, "||")
		case '"':
			if err := l.lexString(); err != nil {
				return err
			}
		default:
			if isIdentStart(ch) {
				l.lexIdent()
			} else {
				return fmt.Errorf("unexpected character %q at position %d", ch, l.pos)
			}
		}
	}

	l.tokens = append(l.tokens, Token{Kind: TokenEOF, Pos: l.pos})
	return nil
}

func (l *Lexer) emit(kind TokenKind, val string) {
	l.tokens = append(l.tokens, Token{Kind: kind, Val: val, Pos: l.pos})
	l.pos++
}

func (l *Lexer) lexTwoChar(eq byte, eqKind TokenKind, tilde byte, tildeKind TokenKind) error {
	start := l.pos
	if l.pos+1 >= len(l.input) {
		return fmt.Errorf("unexpected end of input at position %d, expected '=' or '~' after '='", start)
	}
	next := l.input[l.pos+1]
	if next == eq {
		l.tokens = append(l.tokens, Token{Kind: eqKind, Val: "==", Pos: start})
		l.pos += 2
		return nil
	}
	if next == tilde {
		l.tokens = append(l.tokens, Token{Kind: tildeKind, Val: "=~", Pos: start})
		l.pos += 2
		return nil
	}
	return fmt.Errorf("unexpected character %q at position %d, expected '=' or '~'", next, l.pos+1)
}

func (l *Lexer) expect(ch byte, op string) error {
	if l.pos+1 >= len(l.input) || l.input[l.pos+1] != ch {
		return fmt.Errorf("unexpected character at position %d, expected %q", l.pos, op)
	}
	l.pos++
	return nil
}

func (l *Lexer) lexString() error {
	start := l.pos
	l.pos++ // skip opening quote
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\\' {
			l.pos += 2
			continue
		}
		if ch == '"' {
			val := l.input[start+1 : l.pos]
			l.tokens = append(l.tokens, Token{Kind: TokenString, Val: val, Pos: start})
			l.pos++
			return nil
		}
		l.pos++
	}
	return fmt.Errorf("unterminated string starting at position %d", start)
}

func (l *Lexer) lexIdent() {
	start := l.pos
	for l.pos < len(l.input) && isIdentPart(l.input[l.pos]) {
		l.pos++
	}
	val := l.input[start:l.pos]
	l.tokens = append(l.tokens, Token{Kind: TokenIdent, Val: val, Pos: start})
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '-' || ch == '.' || ch == '/'
}
