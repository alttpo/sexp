package sexp

type Producer interface {
	Token(s string) (n *Node, err error)
	Hexadecimal(s []byte) (n *Node, err error)
	List(children ...*Node) (n *Node, err error)
}

type producer struct {
	disallowNewlines bool
}

var LimitedProducer = producer{disallowNewlines: true}
var FullProducer = producer{disallowNewlines: false}

var _ = FullProducer

func MustToken(s string) (n *Node) {
	var err error
	n, err = LimitedProducer.Token(s)
	if err != nil {
		panic(err)
	}
	return n
}
func (e producer) Token(s string) (n *Node, err error) {
	for i, r := range s {
		if i == 0 && !isTokenStart(r) {
			return nil, ErrInvalidTokenChar
		} else if i > 0 && !isTokenRemainder(r) {
			return nil, ErrInvalidTokenChar
		}
	}

	return &Node{
		Kind:        KindToken,
		OctetString: []byte(s),
		List:        nil,
	}, nil
}

func MustHexadecimal(s []byte) (n *Node) {
	var err error
	n, err = LimitedProducer.Hexadecimal(s)
	if err != nil {
		panic(err)
	}
	return
}
func (e producer) Hexadecimal(s []byte) (n *Node, err error) {
	return &Node{
		Kind:        KindHexadecimal,
		OctetString: s,
		List:        nil,
	}, nil
}

func MustList(children ...*Node) (n *Node) {
	var err error
	n, err = LimitedProducer.List(children...)
	if err != nil {
		panic(err)
	}
	return
}
func (e producer) List(children ...*Node) (n *Node, err error) {
	if children == nil {
		children = make([]*Node, 0, 0)
	}
	return &Node{
		Kind:        KindList,
		OctetString: nil,
		List:        children,
	}, nil
}
