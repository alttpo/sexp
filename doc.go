// s-expressions encoder and decoder (custom format)
// author: jsd1982
// date:   2023-01-07
//
// this s-expression encoding adheres to the following restrictions:
//   1. newline-related whitespace characters ('\r', '\n') MUST NOT appear in the
//      serialized form of an S-expression
//   2. 8-bit clean characters. only the standard 7-bit ASCII characters have special meaning.
//      all other characters are passed through verbatim.
//
// examples:
//
//   (test-exp abc d.e.f/gh snake_case #616263# ^3#616263# ^$a#00 01 02 03 04 05 06 07 08 09 0a# 1023 $3ff "abc\ndef\t\"123\"" () (1 3))
//
// token:
//   alpha-numeric identifier must start with alpha char
//   may contain digits and special punctuation chars
//
//   examples:
//     `test-exp`
//     `abc`
//     `d.e.f/gh`
//     `snake_case`
//
// integer:
//   base-10 or base-16 integer value
//   base-10 is the default base
//   base-16 indicated by leading '$'
//
//   examples:
//     `1023`
//     `$3ff`
//
// hex-string:
//   leading '#' and trailing '#' with hex-digits and white-space in between
//   octets are encoded as pairs of 2 hex digits regardless of white-space in between
//   white-space is ignored and is not interpreted as an octet separator
//   if an odd number of hex-digits is encountered, the right-most digit is assumed to be a '0'
//
//   examples:
//     `#616263#`        = #61 62 63#
//     `#0 0102 2#`      = #00 10 22#
//
// length-prefixed hex-string:
//   leading '^' indicates length follows as an <integer> followed by the <hex-string>
//   length can be decimal or hexadecimal
//
//   examples:
//     `^3#616263#`
//     `^$a#00 01 02 03 04 05 06 07 08 09 0a#`
//
// quoted-string:
//   leading '"' and trailing '"'
//
// BNF:
//  <sexpr>           :: <token> | <string> | <integer> | <list> ;
//
//  <list>            :: "(" ( <sexpr> | <whitespace> )* ")" ;
//
//  <string>          :: ( "^" <integer> )? <data-string> ;
//  <data-string>     :: <hex-string> | <quoted-string> ;
//
//  <hex-string>      :: "#" ( <hex-digit> | <whitespace> )* "#" ;
//
//  <quoted-string>   :: "\"" ( <quoted-char> | <quoted-escape> )* "\"" ;
//  <quoted-char>     :: [32 <= char <= 255, except "\"", "\\"] ;
//  <quoted-escape>   :: "\\" ( <escape-single> | <escape-hex> ) ;
//  <escape-single>   :: "\\" | "\"" | "n" | "r" | "t" ;
//  <escape-hex>      :: "x" <hex-digit> <hex-digit> ;
//
//  <integer>         :: <decimal> | <hexadecimal> ;
//
//  <decimal>         :: <decimal-digit>+ ;
//  <decimal-digit>   :: "0" | ... | "9" ;
//
//  <hexadecimal>     :: "$" <hex-digit>+ ;
//  <hex-digit>       :: <decimal-digit> | "A" | ... | "F" | "a" | ... | "f" ;
//
//  <whitespace>      :: <whitespace-char>* ;
//  <whitespace-char> :: " " | "\t" ;
//
//  <token>           :: <token-start> <token-char>* ;
//  <token-start>     :: <alpha> | <simple-punc> ;
//  <token-char>      :: <token-start> | <decimal-digit> ;
//  <alpha>           :: <upper-case> | <lower-case> ;
//  <lower-case>      :: "a" | ... | "z" ;
//  <upper-case>      :: "A" | ... | "Z" ;
//  <simple-punc>     :: "-" | "." | "/" | "_" ;

package sexp
