package lexer

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	// Special Tokens
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	CLASS    = "CLASS"
	PUBLIC   = "PUBLIC"
	STATIC   = "STATIC"
	VOID     = "VOID"
	INT_KW   = "INT_KW"
	RETURN   = "RETURN"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	DOT       = "."
	LBRACKET  = "["
	RBRACKET  = "]"
)

var keywords = map[string]TokenType{
	"let":    LET,
	"class":  CLASS,
	"public": PUBLIC,
	"static": STATIC,
	"void":   VOID,
	"int":    INT_KW,
	"return": RETURN,
}

func LookupIdent(Literal string) TokenType {
	tokenType, ok := keywords[Literal]
	if !ok {
		return IDENT
	}
	return tokenType
}
