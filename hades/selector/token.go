package selector

type TokenKind int

const (
	TokenIdent  TokenKind = iota // tag name
	TokenString                  // "quoted value"
	TokenEq                      // ==
	TokenNeq                     // !=
	TokenMatch                   // =~
	TokenNotMatch                // !~
	TokenAnd                     // &&
	TokenOr                      // ||
	TokenNot                     // !
	TokenLParen                  // (
	TokenRParen                  // )
	TokenEOF
)

type Token struct {
	Kind TokenKind
	Val  string
	Pos  int
}

func (k TokenKind) String() string {
	switch k {
	case TokenIdent:
		return "identifier"
	case TokenString:
		return "string"
	case TokenEq:
		return "=="
	case TokenNeq:
		return "!="
	case TokenMatch:
		return "=~"
	case TokenNotMatch:
		return "!~"
	case TokenAnd:
		return "&&"
	case TokenOr:
		return "||"
	case TokenNot:
		return "!"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenEOF:
		return "EOF"
	default:
		return "unknown"
	}
}
