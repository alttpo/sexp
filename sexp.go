package sexp

import (
	"encoding/hex"
	"errors"
	"strings"
)

type Kind int

var (
	ErrNotASCII                    = errors.New("only ASCII encoding supported")
	ErrParseUnacceptableWhitespace = errors.New("unacceptable whitespace char")
	ErrUnexpectedChar              = errors.New("unexpected character")
	ErrInvalidLengthPrefix         = errors.New("invalid length prefix")
	ErrInvalidTokenChar            = errors.New("invalid token character")
)

const (
	KindList Kind = iota
	KindToken
	KindHexadecimal
)

type Node struct {
	Kind
	OctetString []byte
	List        []*Node
}

func (n *Node) String() string {
	var sb strings.Builder

	err := n.appendToBuilder(&sb)
	if err != nil {
		return "!!(" + err.Error() + ")!!"
	}

	return sb.String()
}

func (n *Node) appendToBuilder(sb *strings.Builder) (err error) {
	if n == nil {
		return
	}

	switch n.Kind {
	case KindList:
		sb.WriteRune('(')
		for i, c := range n.List {
			err = c.appendToBuilder(sb)
			if err != nil {
				return
			}
			if i < len(n.List)-1 {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune(')')
		return
	case KindToken:
		sb.Write(n.OctetString)
		return
	case KindHexadecimal:
		sb.WriteRune('#')
		_, err = hex.NewEncoder(sb).Write(n.OctetString)
		if err != nil {
			return
		}
		sb.WriteRune('#')
		return
	}

	return
}
