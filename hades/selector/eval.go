package selector

import (
	"fmt"
	"regexp"
	"strings"
)

type EvalError struct {
	Msg string
	Pos int
}

func (e *EvalError) Error() string {
	return e.Msg
}

type EvalErrors struct {
	Errors []EvalError
}

func (e *EvalErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Msg
	}
	return strings.Join(msgs, "; ")
}

// Eval parses and evaluates a selector expression against the given tags.
// Returns true if the tags match the selector.
func Eval(selector string, tags map[string]string) (bool, *EvalErrors) {
	tokens, err := Lex(selector)
	if err != nil {
		return false, &EvalErrors{Errors: []EvalError{{Msg: err.Error()}}}
	}

	ast, err := Parse(tokens)
	if err != nil {
		return false, &EvalErrors{Errors: []EvalError{{Msg: err.Error()}}}
	}

	result, errs := eval(ast, tags)
	if len(errs) > 0 {
		return false, &EvalErrors{Errors: errs}
	}
	return result, nil
}

func eval(node Node, tags map[string]string) (bool, []EvalError) {
	switch n := node.(type) {
	case *Comparison:
		return evalComparison(n, tags)
	case *BinaryOp:
		left, errs := eval(n.Left, tags)
		if len(errs) > 0 {
			return false, errs
		}
		// short-circuit
		if n.Op == TokenAnd && !left {
			return false, nil
		}
		if n.Op == TokenOr && left {
			return true, nil
		}
		return eval(n.Right, tags)
	case *NotOp:
		result, errs := eval(n.Expr, tags)
		if len(errs) > 0 {
			return false, errs
		}
		return !result, nil
	default:
		return false, []EvalError{{Msg: fmt.Sprintf("unknown node type %T", node)}}
	}
}

func evalComparison(c *Comparison, tags map[string]string) (bool, []EvalError) {
	tagVal, exists := tags[c.Key]

	switch c.Op {
	case TokenEq:
		return exists && tagVal == c.Val, nil
	case TokenNeq:
		return !exists || tagVal != c.Val, nil
	case TokenMatch:
		re, err := regexp.Compile(c.Val)
		if err != nil {
			return false, []EvalError{{Msg: fmt.Sprintf("invalid regex %q: %v", c.Val, err)}}
		}
		return exists && re.MatchString(tagVal), nil
	case TokenNotMatch:
		re, err := regexp.Compile(c.Val)
		if err != nil {
			return false, []EvalError{{Msg: fmt.Sprintf("invalid regex %q: %v", c.Val, err)}}
		}
		return !exists || !re.MatchString(tagVal), nil
	default:
		return false, []EvalError{{Msg: fmt.Sprintf("unknown operator %s", c.Op)}}
	}
}
