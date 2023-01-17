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
//   (test-exp abc d.e.f/gh snake_case #616263# 1023 $3ff "abc\ndef\t\"123\"")
//
// BNF:
//  <sexpr>           :: <token> | <string> | <integer> | <list> ;
//
//  <string>          :: <hex-string> | <quoted-string> ;
//
//  <hex-string>      :: "#" ( <hex-digit> | <whitespace> )* "#" ;
//
//  <quoted-string>   :: "\"" ( <quoted-char> | <quoted-escape> )* "\""
//  <quoted-char>     :: <any 8-bit char except "\"", "\\", "\r", "\n"> ;
//  <quoted-escape>   :: "\\" ( <escape-single> | <escape-hex> ) ;
//  <escape-single>   :: "\\" | "\"" | "n" | "r" | "t" | "v" | "f" ;
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
//  <list>            :: "(" ( <sexpr> | <whitespace> )* ")" ;
//
//  <whitespace>      :: <whitespace-char>* ;
//  <whitespace-char> :: " " | "\t" | "\v" | "\f" ;
//
//  <token>           :: <token-char>+ ;
//  <token-char>      :: <alpha> | <decimal-digit> | <simple-punc> ;
//  <alpha>           :: <upper-case> | <lower-case> ;
//  <lower-case>      :: "a" | ... | "z" ;
//  <upper-case>      :: "A" | ... | "Z" ;
//  <simple-punc>     :: "-" | "." | "/" | "_" | ":" | "*" | "+" | "=" ;

package sexp
