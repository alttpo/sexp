package sexp

import (
	"encoding/hex"
	"errors"
	"strings"
)

// author: jsd1982
// date:   2023-01-07

// this go package implements a subset of S-Expressions as described in
// https://people.csail.mit.edu/rivest/Sexp.txt (copied in this folder as `rivest-sexp.txt`)

// the implemented subset adheres to the following restrictions:
//   1. newline-related whitespace characters ('\r', '\n') MUST NOT appear in the
//      serialized form of an S-expression
//   2. disallow the use of raw octet string encoding (since it would violate restriction 1)
//   3. disallow the brace form of base-64 encoding of S-expressions (unneeded complexity)
//   4. disallow the use of display hints (unneeded complexity)

// thus, we must remove '\r' and '\n' from the acceptable whitespace-char set as well as
// reduce the acceptable formats of octet-strings.

// supported octet-string encodings:
//   1. token			(abc)
//   2. hexadecimal		(#616263#)
//   3. base-64			(|YWJj|)

// unsupported octet-string encodings:
//   1. verbatim (aka raw)
//   2. quoted
// these encodings are unsupported because their encodings could contain restricted
// newline-related whitespace characters.

// this implementation only supports ASCII encoding natively. token octet-strings may not
// contain non-ASCII characters. unicode data may of course be exchanged in a Unicode
// Transformation Format but must be done with either hexadecimal or base-64 encoded
// octet-strings, NOT in token octet-strings.

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
