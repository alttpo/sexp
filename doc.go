// Package sexp
//
// s-expressions encoder and decoder (custom format)
//
//	author: jsd1982
//	date:   2023-01-07
//
// an s-expression is recursively either an atom expression or a list of s-expressions surrounded by '(' ')'.
//
// encoding restrictions:
//  1. newline whitespace characters ('\r', '\n') MUST NOT appear in an encoded s-expression.
//  2. encoding is ASCII and 8-bit clean with exception of '\r' and '\n'.
//  3. each atom has a single encoding with no stylistic variation so as to simplify encoding and decoding logic.
//
// s-expression examples:
//
//	(test-exp abc d.e.f/gh snake_case nil true false #616263# ^3#616263#)
//	(^$a#00 01 02 03 04 05 06 07 08 09 0a# 1023 $3ff)
//	("abc\ndef\t\"123\"\x00\xff" () ^5"12345")
//	((a 1) (b 2) (c 3) (d nil) (e false))
//
// atom types:
//
//	nil
//	bool
//	integer
//	token
//	hex-string
//	quoted-string
//	list
//
// nil atom type:
//
//	literal "nil" keyword
//
// bool atom type:
//
//	literal "true" and "false" keywords
//
// integer atom type:
//
//	a base-10 or base-16 integer value of arbitrary length
//	may start with optional '-' to indicate negative value
//	integers contain only allowable digit characters depending on the base
//	no extra formatting-related ('_'), division (','), or white-space characters are allowed
//	any number of leading zeros are allowed and *do not* signify base-8
//	the default base is 10
//	base-10 integer must contain only digits '0'..'9'
//	base-16 integer must start with '$'
//	base-16 integer must contain only digits '0'..'9','a'..'f'
//	base-16 integer cannot contain upper-case 'A'..'F' so as to simplify encoding and decoding logic
//
//	examples:
//		  `-1024`  = -1024
//		`0001023`  =  1023
//		   `1023`  =  1023
//		   `$3ff`  =  1023
//
// token atom type:
//
//	alpha-numeric identifier of arbitrary length without white-space
//	may begin with one optional '@' to escape reserved keywords like "nil", "true", "false"
//	cannot start with a decimal digit
//	may contain alpha characters 'a' .. 'z', 'A' .. 'Z'
//	may contain special punctuation chars "_" | "." | "/" | "?" | "!"
//	may contain non-ASCII characters 128 <= char <= 255
//	may contain decimal digits '0' .. '9'
//
//	examples:
//	  `test_exp`
//	  `abc`
//	  `d.e.f/gh`
//	  `snake_case`
//	  `@true`
//	  `@nil`
//
// hex-string atom type:
//
//	optionally starts with a leading '^' followed by <integer> to specify the decoded data length
//	leading '#' and trailing '#'
//	encodes an octet-string with each octet described in hexadecimal with 2*n hex-digits (n >= 0)
//	only hex-digits may appear between leading '#' and trailing '#'
//	no white-space or other characters allowed to simplify encoding and decoding logic
//	octets are encoded as 2 hex digits in sequence, most significant digit first followed by least significant
//	if an odd number of hex-digits is encountered, the last digit is assumed to be the most-significant digit
//	of a 2-digit octet and the least-significant digit is assumed to be 0.
//
//	examples:
//	  `#616263#`
//	  `#001022#`
//	  `^3#616263#`
//	  `^$a#0102030405060708090a#`
//
// quoted-string atom type:
//
//	optionally starts with a leading '^' followed by <integer> to specify the decoded data length
//	leading '"' and trailing '"'
//	may contain any ASCII and non-ASCII character except '\' and '"'
//	a '\' is treated as the start of an escape sequence followed by one of:
//		'\' = '\\'
//		'"' = '\"'
//		'r' = '\r'
//		'n' = '\n'
//		't' = '\t'
//		'x' <hex-digit> <hex-digit> = escape of 8-bit character encoded in hexadecimal
//
//	examples:
//	  "abc\ndef\t\"123\"\x00\xff"
//	  ^5"12345"
//
// BNF:
//
//	<sexpr>           :: <nil> | <bool> | <integer> | <token> | <string> | <list> ;
//
//	<list>            :: "(" ( <sexpr> | <whitespace> )* ")" ;
//
//	<nil>             :: "n" "i" "l" ;
//
//	<bool>            :: <bool-true> | <bool-false> ;
//	<bool-true>       :: "t" "r" "u" "e" ;
//	<bool-false>      :: "f" "a" "l" "s" "e" ;
//
//	<integer>         :: <decimal> | <hexadecimal> ;
//
//	<decimal>         :: <decimal-digit>+ ;
//	<decimal-digit>   :: "0" | ... | "9" ;
//
//	<hexadecimal>     :: "$" <hex-digit>+ ;
//	<hex-digit>       :: <decimal-digit> | "a" | ... | "f" ;
//
//	<string>          :: ( "^" <integer> )? <data-string> ;
//	<data-string>     :: <hex-string> | <quoted-string> ;
//
//	<hex-string>      :: "#" <hex-digit>* "#" ;
//
//	<quoted-string>   :: "\"" ( <quoted-char> | <quoted-escape> )* "\"" ;
//	<quoted-char>     :: [any 8-bit char except "\"", "\\", "\r", "\n"] ;
//	<quoted-escape>   :: "\\" ( <escape-single> | <escape-hex> ) ;
//	<escape-single>   :: "\\" | "\"" | "n" | "r" | "t" ;
//	<escape-hex>      :: "x" <hex-digit> <hex-digit> ;
//
//	<whitespace>      :: <whitespace-char>* ;
//	<whitespace-char> :: " " | "\t" ;
//
//	<token>           :: ( "@" )? <token-start> <token-remainder>* ;
//	<token-start>     :: <alpha> | <simple-punc> | <non-ascii> ;
//	<token-remainder> :: <token-start> | <decimal-digit> ;
//	<alpha>           :: <upper-case> | <lower-case> ;
//	<lower-case>      :: "a" | ... | "z" ;
//	<upper-case>      :: "A" | ... | "Z" ;
//	<simple-punc>     :: "_" | "." | "/" | "?" | "!" ;
//	<non-ascii>       :: [128 <= char <= 255] ;
package sexp
