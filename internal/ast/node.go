package ast

import "github.com/out-lang/out/internal/lexer"

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

type LetStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	return ls.Name.String() + " = " + ls.Value.String()
}

type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	return "return " + rs.ReturnValue.String()
}

type ImportStatement struct {
	Token lexer.Token
	Path  string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string       { return "import " + is.Path }

type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type Boolean struct {
	Token lexer.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return "null" }
func (nl *NullLiteral) String() string       { return "null" }

type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

type IfExpression struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out string
	out += "if" + ie.Condition.String() + " " + ie.Consequence.String()
	if ie.Alternative != nil {
		out += " else " + ie.Alternative.String()
	}
	return out
}

type WhileExpression struct {
	Token     lexer.Token
	Condition Expression
	Body      *BlockStatement
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	return "while " + we.Condition.String() + " " + we.Body.String()
}

type ForExpression struct {
	Token   lexer.Token
	Name    *Identifier
	Start   Expression
	End     Expression
	Body    *BlockStatement
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	return "for " + fe.Name.String() + " in " + fe.Start.String() + ".." + fe.End.String() + " " + fe.Body.String()
}

type FunctionLiteral struct {
	Token      lexer.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out string
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out += "def"
	out += "("
	out += joinStrings(params, ", ")
	out += ") "
	out += fl.Body.String()
	return out
}

type CallExpression struct {
	Token     lexer.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out string
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out += ce.Function.String()
	out += "("
	out += joinStrings(args, ", ")
	out += ")"
	return out
}

type IndexExpression struct {
	Token lexer.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

type AssignExpression struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) String() string {
	return ae.Name.String() + " = " + ae.Value.String()
}

type ArrayLiteral struct {
	Token    lexer.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out string
	elems := []string{}
	for _, e := range al.Elements {
		elems = append(elems, e.String())
	}
	out += "["
	out += joinStrings(elems, ", ")
	out += "]"
	return out
}

type HashLiteral struct {
	Token lexer.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out string
	pairs := []string{}
	for key, val := range hl.Pairs {
		pairs = append(pairs, key.String()+": "+val.String())
	}
	out += "{"
	out += joinStrings(pairs, ", ")
	out += "}"
	return out
}

type MemberAccess struct {
	Token  lexer.Token
	Object Expression
	Member string
}

func (ma *MemberAccess) expressionNode()      {}
func (ma *MemberAccess) TokenLiteral() string { return ma.Token.Literal }
func (ma *MemberAccess) String() string {
	return ma.Object.String() + "::" + ma.Member
}

type TryExpression struct {
	Token       lexer.Token
	TryBlock    *BlockStatement
	CatchVar    *Identifier
	CatchBlock  *BlockStatement
}

func (te *TryExpression) expressionNode()      {}
func (te *TryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TryExpression) String() string {
	var out string
	out += "try " + te.TryBlock.String()
	if te.CatchVar != nil {
		out += " catch(" + te.CatchVar.String() + ") " + te.CatchBlock.String()
	}
	return out
}

type ThrowExpression struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (te *ThrowExpression) expressionNode()      {}
func (te *ThrowExpression) TokenLiteral() string { return te.Token.Literal }
func (te *ThrowExpression) String() string {
	return "throw " + te.ReturnValue.String()
}

type SafeAccess struct {
	Token  lexer.Token
	Object Expression
	Member string
}

func (sa *SafeAccess) expressionNode()      {}
func (sa *SafeAccess) TokenLiteral() string { return sa.Token.Literal }
func (sa *SafeAccess) String() string {
	return sa.Object.String() + "?. " + sa.Member
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
