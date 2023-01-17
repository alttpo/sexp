package sexp

import (
	"bytes"
	"encoding/hex"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Parser interface {
	ParseNode(s io.RuneScanner) (n *Node, err error)
	ParseList(s io.RuneScanner) (n *Node, err error)
	ParseToken(s io.RuneScanner) (n *Node, err error)
	ParseHexadecimal(s io.RuneScanner, h LengthHint) (n *Node, err error)
}

type LengthHint struct {
	Has    bool
	Length uint64
}

type parser struct {
	disallowNewlines bool
}

var LimitedParser = parser{disallowNewlines: true}
var FullParser = parser{disallowNewlines: false}

var _ = FullParser

func Parse(s io.RuneScanner) (n *Node, err error) {
	return LimitedParser.ParseNode(s)
}
func (e parser) ParseNode(s io.RuneScanner) (n *Node, err error) {
	var listEnd bool
	n, listEnd, err = e.parseNode(s)
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

func (e parser) shouldDiscard(r rune) (discard bool, err error) {
	// error on unacceptable chars:
	if r > unicode.MaxASCII {
		discard, err = false, ErrNotASCII
		return
	}
	if r == '\r' || r == '\n' {
		if e.disallowNewlines {
			discard, err = false, ErrParseUnacceptableWhitespace
		} else {
			discard = true
		}
		return
	}

	// ignore acceptable whitespace chars:
	if r <= ' ' || r == '\t' || r == '\v' || r == '\f' {
		discard = true
		return
	}

	return
}

func (e parser) parseNode(s io.RuneScanner) (n *Node, listEnd bool, err error) {
	var r rune
	for {
		r, _, err = s.ReadRune()
		if err != nil {
			return
		}

		// skip whitespace or error on bad char:
		var discard bool
		discard, err = e.shouldDiscard(r)
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
			n, err = e.ParseList(s)
			return
		}

		// tokens may not start with leading decimal:
		if isTokenStart(r) {
			err = s.UnreadRune()
			if err != nil {
				return
			}
			n, err = e.ParseToken(s)
			return
		}

		// parse optional leading decimal indicating size:
		var h LengthHint
		if r >= '0' && r <= '9' {
			err = s.UnreadRune()
			if err != nil {
				return
			}

			h.Length, err = e.ParseDecimal(s)
			if err != nil {
				return
			}
			h.Has = true

			r, _, err = s.ReadRune()
			if err != nil {
				return
			}
		}

		if r == '#' {
			n, err = e.ParseHexadecimal(s, h)
			return
		}

		err = ErrUnexpectedChar
		return
	}
}

func (e parser) ParseList(s io.RuneScanner) (n *Node, err error) {
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
		discard, err = e.shouldDiscard(r)
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
		child, listEnd, err = e.parseNode(s)
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

func (e parser) ParseDecimal(s io.RuneScanner) (v uint64, err error) {
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

func (e parser) ParseToken(s io.RuneScanner) (n *Node, err error) {
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

func isHexadecimalRemainder(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'A' && r <= 'F' {
		return true
	}
	if r >= 'a' && r <= 'f' {
		return true
	}
	return false
}

func (e parser) ParseHexadecimal(s io.RuneScanner, h LengthHint) (n *Node, err error) {
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

		var discard bool
		discard, err = e.shouldDiscard(r)
		if err != nil {
			return
		}
		if discard {
			continue
		}

		if r == '#' {
			break
		}

		if !isHexadecimalRemainder(r) {
			err = ErrUnexpectedChar
			return
		}

		sb.WriteRune(r)
	}

	if !eof {
		err = s.UnreadRune()
		if err != nil {
			return
		}
	} else {
		err = io.ErrUnexpectedEOF
		return
	}

	var dst []byte
	if h.Has {
		dst = make([]byte, h.Length)
	} else {
		dst = make([]byte, hex.DecodedLen(sb.Len()))
	}

	var dn int
	dn, err = hex.Decode(dst, sb.Bytes())
	if err != nil {
		return
	}
	if h.Has && dn != len(dst) {
		err = ErrInvalidLengthPrefix
		return
	}

	n = &Node{
		Kind:        KindHexadecimal,
		OctetString: dst[:dn],
		List:        nil,
	}
	return
}
