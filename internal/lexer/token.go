package lexer

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Pos     int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"
	FLOAT = "FLOAT"
	STRING = "STRING"

	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	PERCENT  = "%"

	LT  = "<"
	GT  = ">"
	EQ  = "=="
	NEQ = "!="
	LTE = "<="
	GTE = ">="

	COMMA       = ","
	COLON       = ":"
	DOUBLECOLON = "::"
	SEMICOLON   = ";"
	DOT       = "."
	DOTDOT    = ".."
	QUESTION  = "?"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
	ARROW    = "->"

	IF     = "if"
	ELSE   = "else"
	WHILE  = "while"
	FOR    = "for"
	IN     = "in"
	DEF    = "def"
	RETURN = "return"
	TRUE   = "true"
	FALSE  = "false"
	NULL   = "null"
	AND    = "and"
	OR     = "or"
	IMPORT = "import"
	TRY    = "try"
	CATCH  = "catch"
	THROW  = "throw"
)

var keywords = map[string]TokenType{
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"for":    FOR,
	"in":     IN,
	"def":    DEF,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"and":    AND,
	"or":     OR,
	"import": IMPORT,
	"try":    TRY,
	"catch":  CATCH,
	"throw":  THROW,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
