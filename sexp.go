package sexp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
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

const (
	KindList Kind = iota
	KindToken
	KindHexadecimal
	KindBase64
)

type Node struct {
	Kind
	OctetString []byte
	List        []*Node
}

var (
	ErrNotASCII                    = errors.New("only ASCII encoding supported")
	ErrParseUnacceptableWhitespace = errors.New("unacceptable whitespace char")
	ErrUnexpectedChar              = errors.New("unexpected character")
)

func Parse(s io.RuneScanner) (n *Node, err error) {
	var listEnd bool
	n, listEnd, err = parseNode(s)
	if listEnd {
		err = ErrUnexpectedChar
	}
	if err == io.EOF {
		// allow regular EOF errors but still fail on ErrUnexpectedEOF
		err = nil
	}
	if err != nil {
		n = nil
	}
	return
}

func shouldDiscard(r rune) (discard bool, err error) {
	// error on unacceptable chars:
	if r > unicode.MaxASCII {
		discard, err = false, ErrNotASCII
		return
	}
	if r == '\r' || r == '\n' {
		discard, err = false, ErrParseUnacceptableWhitespace
		return
	}

	// ignore acceptable whitespace chars:
	if r <= ' ' || r == '\t' || r == '\v' || r == '\f' {
		discard = true
		return
	}

	return
}

func parseNode(s io.RuneScanner) (n *Node, listEnd bool, err error) {
	var r rune
	for {
		r, _, err = s.ReadRune()
		if err != nil {
			return
		}

		// skip whitespace or error on bad char:
		var discard bool
		discard, err = shouldDiscard(r)
		if err != nil {
			return
		}
		if discard {
			continue
		}

		if r == ')' {
			return nil, true, nil
		}
		if r == '(' {
			n, err = parseList(s)
			return
		}

		// tokens may not start with leading decimal:
		if isTokenStart(r) {
			err = s.UnreadRune()
			if err != nil {
				return
			}
			n, err = parseToken(s)
			return
		}

		// parse optional leading decimal indicating size:
		var decimal uint64
		hasDecimal := false
		if r >= '0' && r <= '9' {
			decimal, err = parseDecimal(s)
			if err != nil {
				return
			}
			hasDecimal = true

			r, _, err = s.ReadRune()
			if err != nil {
				return
			}
		}

		if r == '|' {
			n, err = parseBase64(s, decimal, hasDecimal)
			return
		}
		if r == '#' {
			n, err = parseHexadecimal(s, decimal, hasDecimal)
			return
		}

		err = ErrUnexpectedChar
		return
	}
}

func parseList(s io.RuneScanner) (n *Node, err error) {
	defer func() {
		// convert regular EOF errors to ErrUnexpectedEOF
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	n = &Node{
		Kind:        KindList,
		OctetString: nil,
		List:        make([]*Node, 0, 10),
	}

	var r rune
	for {
		r, _, err = s.ReadRune()
		if err != nil {
			return
		}

		var discard bool
		discard, err = shouldDiscard(r)
		if err != nil {
			return
		}
		if discard {
			continue
		}

		err = s.UnreadRune()
		if err != nil {
			return
		}

		var child *Node
		var listEnd bool
		child, listEnd, err = parseNode(s)
		if err != nil {
			return nil, err
		}
		if listEnd {
			break
		}

		n.List = append(n.List, child)
	}

	return
}

func isAlpha(r rune) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	return false
}

func isDigit(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	return false
}

func isGraphic(r rune) bool {
	return r == '-' ||
		r == '.' ||
		r == '/' ||
		r == '_' ||
		r == ':' ||
		r == '*' ||
		r == '+' ||
		r == '='
}

func isTokenStart(r rune) bool {
	return isAlpha(r) || isGraphic(r)
}

func isTokenRemainder(r rune) bool {
	return isAlpha(r) || isDigit(r) || isGraphic(r)
}

func parseDecimal(s io.RuneScanner) (v uint64, err error) {
	var sb strings.Builder

	var r rune
	for {
		r, _, err = s.ReadRune()
		if err != nil {
			return
		}

		if !isDigit(r) {
			err = s.UnreadRune()
			if err != nil {
				return
			}

			v, err = strconv.ParseUint(sb.String(), 10, 64)
			return
		}

		sb.WriteRune(r)
	}
}

func parseToken(s io.RuneScanner) (n *Node, err error) {
	var sb bytes.Buffer

	var r rune
	eof := false
	for !eof {
		r, _, err = s.ReadRune()
		if err == io.EOF {
			eof = true
			err = nil
			break
		}
		if err != nil {
			return
		}

		if !isTokenRemainder(r) {
			break

		}

		sb.WriteRune(r)
	}

	if !eof {
		err = s.UnreadRune()
		if err != nil {
			return
		}
	}

	n = &Node{
		Kind:        KindToken,
		OctetString: sb.Bytes(),
		List:        nil,
	}
	if eof {
		err = io.EOF
	}
	return
}

func parseBase64(s io.RuneScanner, decimal uint64, hasDecimal bool) (n *Node, err error) {
	return nil, errors.New("unimplemented")
}

func parseHexadecimal(s io.RuneScanner, decimal uint64, hasDecimal bool) (n *Node, err error) {
	return nil, errors.New("unimplemented")
}
