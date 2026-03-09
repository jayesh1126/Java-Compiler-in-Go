package lexer

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar gives us the next character and advance our position in the input string
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for 'NUL', signifies end of file/input
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = newToken(ASSIGN, l.ch)
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '*':
		tok = newToken(ASTERISK, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case ',':
		tok = newToken(COMMA, l.ch)
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '.':
		tok = newToken(DOT, l.ch)
	case '[':
		tok = newToken(LBRACKET, l.ch)
	case ']':
		tok = newToken(RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			tok.Type = LookupIdent(literal)
			tok.Literal = literal
			return tok
		} else if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// Helper to create a new token from a single character
func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	// For now we expect only ASCII characters.
	return ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'))
}

func isDigit(ch byte) bool {
	// Again only ASCII here
	return (ch >= '0' && ch <= '9')
}

func (l *Lexer) readIdentifier() string {
	var identifier []byte
	for l.ch == '_' || isLetter(l.ch) {
		identifier = append(identifier, l.ch)
		l.readChar()
	}

	return string(identifier)
}

func (l *Lexer) readNumber() string {
	var number []byte
	for isDigit(l.ch) {
		number = append(number, l.ch)
		l.readChar()
	}

	return string(number)
}
