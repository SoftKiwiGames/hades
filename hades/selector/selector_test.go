package selector

import (
	"testing"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []TokenKind
		wantErr bool
	}{
		{
			name:  "simple comparison",
			input: `env == "dev"`,
			want:  []TokenKind{TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:  "and expression",
			input: `cluster == "db" && env == "dev"`,
			want:  []TokenKind{TokenIdent, TokenEq, TokenString, TokenAnd, TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:  "or expression",
			input: `env == "dev" || env == "staging"`,
			want:  []TokenKind{TokenIdent, TokenEq, TokenString, TokenOr, TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:  "not equal",
			input: `env != "prod"`,
			want:  []TokenKind{TokenIdent, TokenNeq, TokenString, TokenEOF},
		},
		{
			name:  "regex match",
			input: `name =~ "db-[0-9]+"`,
			want:  []TokenKind{TokenIdent, TokenMatch, TokenString, TokenEOF},
		},
		{
			name:  "regex not match",
			input: `name !~ "test.*"`,
			want:  []TokenKind{TokenIdent, TokenNotMatch, TokenString, TokenEOF},
		},
		{
			name:  "negation",
			input: `!active == "true"`,
			want:  []TokenKind{TokenNot, TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:  "parentheses",
			input: `(env == "dev")`,
			want:  []TokenKind{TokenLParen, TokenIdent, TokenEq, TokenString, TokenRParen, TokenEOF},
		},
		{
			name:  "ident with dots and dashes",
			input: `k8s.io/cluster-name == "main"`,
			want:  []TokenKind{TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:    "unterminated string",
			input:   `env == "dev`,
			wantErr: true,
		},
		{
			name:    "unexpected character",
			input:   `env @ "dev"`,
			wantErr: true,
		},
		{
			name:    "lone ampersand",
			input:   `env == "a" & env == "b"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := Lex(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.want) {
				t.Fatalf("token count: got %d, want %d", len(tokens), len(tt.want))
			}
			for i, tok := range tokens {
				if tok.Kind != tt.want[i] {
					t.Errorf("token[%d]: got %s, want %s", i, tok.Kind, tt.want[i])
				}
			}
		})
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		tags     map[string]string
		want     bool
		wantErr  bool
	}{
		{
			name:     "simple match",
			selector: `env == "dev"`,
			tags:     map[string]string{"env": "dev"},
			want:     true,
		},
		{
			name:     "simple no match",
			selector: `env == "prod"`,
			tags:     map[string]string{"env": "dev"},
			want:     false,
		},
		{
			name:     "missing tag",
			selector: `env == "dev"`,
			tags:     map[string]string{},
			want:     false,
		},
		{
			name:     "not equal match",
			selector: `env != "prod"`,
			tags:     map[string]string{"env": "dev"},
			want:     true,
		},
		{
			name:     "not equal no match",
			selector: `env != "dev"`,
			tags:     map[string]string{"env": "dev"},
			want:     false,
		},
		{
			name:     "not equal missing tag",
			selector: `env != "prod"`,
			tags:     map[string]string{},
			want:     true,
		},
		{
			name:     "and both true",
			selector: `cluster == "db" && env == "dev"`,
			tags:     map[string]string{"cluster": "db", "env": "dev"},
			want:     true,
		},
		{
			name:     "and one false",
			selector: `cluster == "db" && env == "prod"`,
			tags:     map[string]string{"cluster": "db", "env": "dev"},
			want:     false,
		},
		{
			name:     "or one true",
			selector: `env == "dev" || env == "staging"`,
			tags:     map[string]string{"env": "dev"},
			want:     true,
		},
		{
			name:     "or both false",
			selector: `env == "dev" || env == "staging"`,
			tags:     map[string]string{"env": "prod"},
			want:     false,
		},
		{
			name:     "negation",
			selector: `!(env == "prod")`,
			tags:     map[string]string{"env": "dev"},
			want:     true,
		},
		{
			name:     "negation false",
			selector: `!(env == "dev")`,
			tags:     map[string]string{"env": "dev"},
			want:     false,
		},
		{
			name:     "complex expression",
			selector: `(env == "dev" || env == "staging") && cluster == "db"`,
			tags:     map[string]string{"env": "dev", "cluster": "db"},
			want:     true,
		},
		{
			name:     "complex expression no match",
			selector: `(env == "dev" || env == "staging") && cluster == "web"`,
			tags:     map[string]string{"env": "dev", "cluster": "db"},
			want:     false,
		},
		{
			name:     "regex match",
			selector: `name =~ "db-[0-9]+"`,
			tags:     map[string]string{"name": "db-42"},
			want:     true,
		},
		{
			name:     "regex no match",
			selector: `name =~ "db-[0-9]+"`,
			tags:     map[string]string{"name": "web-1"},
			want:     false,
		},
		{
			name:     "regex not match",
			selector: `name !~ "test.*"`,
			tags:     map[string]string{"name": "production"},
			want:     true,
		},
		{
			name:     "regex not match fails",
			selector: `name !~ "test.*"`,
			tags:     map[string]string{"name": "test-server"},
			want:     false,
		},
		{
			name:     "invalid regex",
			selector: `name =~ "["`,
			tags:     map[string]string{"name": "anything"},
			wantErr:  true,
		},
		{
			name:     "precedence and binds tighter than or",
			selector: `a == "1" || b == "2" && c == "3"`,
			tags:     map[string]string{"a": "1", "b": "x", "c": "x"},
			want:     true,
		},
		{
			name:     "precedence verified",
			selector: `a == "1" || b == "2" && c == "3"`,
			tags:     map[string]string{"a": "x", "b": "2", "c": "x"},
			want:     false,
		},
		{
			name:     "nil tags",
			selector: `env == "dev"`,
			tags:     nil,
			want:     false,
		},
		{
			name:     "triple and",
			selector: `a == "1" && b == "2" && c == "3"`,
			tags:     map[string]string{"a": "1", "b": "2", "c": "3"},
			want:     true,
		},
		{
			name:     "parse error",
			selector: `env ==`,
			wantErr:  true,
		},
		{
			name:     "empty selector",
			selector: ``,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errs := Eval(tt.selector, tt.tags)
			if tt.wantErr {
				if errs == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if errs != nil {
				t.Fatalf("unexpected error: %v", errs)
			}
			if got != tt.want {
				t.Errorf("Eval(%q) = %v, want %v", tt.selector, got, tt.want)
			}
		})
	}
}
