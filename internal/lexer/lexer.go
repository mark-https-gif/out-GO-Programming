package lexer

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	linePosition int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, linePosition: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.linePosition = 0
	} else {
		l.linePosition++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for {
		if l.ch == '#' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			if l.ch == '\n' {
				l.readChar()
			}
			l.skipWhitespace()
			continue
		}
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			if l.ch == '\n' {
				l.readChar()
			}
			l.skipWhitespace()
			continue
		}
		if l.ch == '/' && l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			for {
				if l.ch == 0 {
					break
				}
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar()
					l.readChar()
					break
				}
				l.readChar()
			}
			l.skipWhitespace()
			continue
		}
		break
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComment()

	line := l.line
	pos := l.linePosition

	var tok Token

	if l.ch == 0 {
		tok = NewToken(EOF, "", line, pos)
		return tok
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(EQ, "==", line, pos)
		} else {
			tok = NewToken(ASSIGN, "=", line, pos)
		}
	case '+':
		tok = NewToken(PLUS, "+", line, pos)
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = NewToken(ARROW, "->", line, pos)
		} else {
			tok = NewToken(MINUS, "-", line, pos)
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(NEQ, "!=", line, pos)
		} else {
			tok = NewToken(BANG, "!", line, pos)
		}
	case '*':
		tok = NewToken(ASTERISK, "*", line, pos)
	case '/':
		tok = NewToken(SLASH, "/", line, pos)
	case '%':
		tok = NewToken(PERCENT, "%", line, pos)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(LTE, "<=", line, pos)
		} else {
			tok = NewToken(LT, "<", line, pos)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(GTE, ">=", line, pos)
		} else {
			tok = NewToken(GT, ">", line, pos)
		}
	case ',':
		tok = NewToken(COMMA, ",", line, pos)
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok = NewToken(DOUBLECOLON, "::", line, pos)
		} else {
			tok = NewToken(COLON, ":", line, pos)
		}
	case ';':
		tok = NewToken(SEMICOLON, ";", line, pos)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok = NewToken(DOTDOT, "..", line, pos)
		} else {
			tok = NewToken(DOT, ".", line, pos)
		}
	case '?':
		tok = NewToken(QUESTION, "?", line, pos)
	case '(':
		tok = NewToken(LPAREN, "(", line, pos)
	case ')':
		tok = NewToken(RPAREN, ")", line, pos)
	case '{':
		tok = NewToken(LBRACE, "{", line, pos)
	case '}':
		tok = NewToken(RBRACE, "}", line, pos)
	case '[':
		tok = NewToken(LBRACKET, "[", line, pos)
	case ']':
		tok = NewToken(RBRACKET, "]", line, pos)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString('"')
		tok.Line = line
		tok.Pos = pos
		return tok
	case '\'':
		tok.Type = STRING
		tok.Literal = l.readString('\'')
		tok.Line = line
		tok.Pos = pos
		return tok
	default:
		if isDigit(l.ch) {
			return l.readNumber()
		}
		if isLetter(l.ch) {
			return l.readIdent()
		}
		tok = NewToken(ILLEGAL, string(l.ch), line, pos)
	}

	l.readChar()
	return tok
}

func (l *Lexer) readNumber() Token {
	line := l.line
	pos := l.linePosition
	start := l.position
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[start:l.position]
	tok := Token{Line: line, Pos: pos}
	if isFloat {
		tok.Type = FLOAT
	} else {
		tok.Type = INT
	}
	tok.Literal = literal
	return tok
}

func (l *Lexer) readIdent() Token {
	line := l.line
	pos := l.linePosition
	start := l.position

	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	literal := l.input[start:l.position]
	return Token{Type: LookupIdent(literal), Literal: literal, Line: line, Pos: pos}
}

func (l *Lexer) readString(quote byte) string {
	l.readChar()
	var result []byte
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case quote:
				result = append(result, quote)
			case 'u':
				l.readChar()
				if code, ok := l.readHex4(); ok {
					result = append(result, []byte(string(rune(code)))...)
				}
				continue
			default:
				result = append(result, '\\', l.ch)
			}
			l.readChar()
			continue
		}
		result = append(result, l.ch)
		l.readChar()
	}
	if l.ch == quote {
		l.readChar()
	}
	return string(result)
}

func (l *Lexer) readHex4() (int, bool) {
	code := 0
	for i := 0; i < 4; i++ {
		d := hexDigit(l.ch)
		if d < 0 {
			return 0, false
		}
		code = code<<4 | d
		l.readChar()
	}
	return code, true
}

func hexDigit(ch byte) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch-'a') + 10
	case 'A' <= ch && ch <= 'F':
		return int(ch-'A') + 10
	default:
		return -1
	}
}

func NewToken(tokenType TokenType, literal string, line, pos int) Token {
	return Token{Type: tokenType, Literal: literal, Line: line, Pos: pos}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
