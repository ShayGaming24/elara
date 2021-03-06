package parserlegacy

import (
	"github.com/ElaraLang/elara/lexer"
	"strconv"
	"strings"
)

type Expr interface{ exprNode() }

type BinaryExpr struct {
	Lhs Expr
	Op  TokenType
	Rhs Expr
}

type UnaryExpr struct {
	Op  TokenType
	Rhs Expr
}

type GroupExpr struct {
	Group Expr
}

type VariableExpr struct {
	Identifier string
}

type AssignmentExpr struct {
	Context    Expr
	Identifier string
	Value      Expr
}

type InvocationExpr struct {
	Invoker Expr
	Args    []Expr
}

type ContextExpr struct {
	Context  Expr
	Variable VariableExpr
}

type TypeCastExpr struct {
	Expr Expr
	Type Type
}

type TypeCheckExpr struct {
	Expr Expr
	Type Type
}

type IfElseExpr struct {
	Condition  Expr
	IfBranch   []Stmt
	IfResult   Expr
	ElseBranch []Stmt
	ElseResult Expr
}

type FuncDefExpr struct {
	Arguments  []FunctionArgument
	ReturnType Type
	Statement  Stmt
}

type AccessExpr struct {
	Expr  Expr
	Index Expr
}

type CollectionExpr struct {
	Elements []Expr
}

type MapExpr struct {
	Entries []MapEntry
}

type MapEntry struct {
	Key   Expr
	Value Expr
}

type StringLiteralExpr struct {
	Value string
}
type CharLiteralExpr struct {
	Value rune
}

type IntegerLiteralExpr struct {
	Value int64
}

type FloatLiteralExpr struct {
	Value float64
}

type BooleanLiteralExpr struct {
	Value bool
}

func (FuncDefExpr) exprNode()        {}
func (AccessExpr) exprNode()         {}
func (CollectionExpr) exprNode()     {}
func (MapExpr) exprNode()            {}
func (StringLiteralExpr) exprNode()  {}
func (CharLiteralExpr) exprNode()    {}
func (IntegerLiteralExpr) exprNode() {}
func (FloatLiteralExpr) exprNode()   {}
func (BooleanLiteralExpr) exprNode() {}
func (UnaryExpr) exprNode()          {}
func (BinaryExpr) exprNode()         {}
func (GroupExpr) exprNode()          {}
func (ContextExpr) exprNode()        {}
func (IfElseExpr) exprNode()         {}
func (InvocationExpr) exprNode()     {}
func (AssignmentExpr) exprNode()     {}
func (VariableExpr) exprNode()       {}
func (TypeCastExpr) exprNode()       {}
func (TypeCheckExpr) exprNode()      {}

func (p *Parser) expression() Expr {
	if p.peek().TokenType == lexer.If {
		return p.ifElseExpression()
	}
	return p.assignment()
}

func (p *Parser) assignment() (expr Expr) {
	expr = p.typeCast()

	if p.check(lexer.Equal) {
		eqlTok := p.advance()
		rhs := p.typeCast()

		switch v := expr.(type) {
		case VariableExpr:
			expr = AssignmentExpr{
				Identifier: v.Identifier,
				Value:      rhs,
			}
			break
		case ContextExpr:
			expr = AssignmentExpr{
				Context:    v.Context,
				Identifier: v.Variable.Identifier,
				Value:      rhs,
			}
			break
		default:
			panic(ParseError{
				token:   eqlTok,
				message: "Invalid type found behind assignment",
			})
		}
	}
	return
}

func (p *Parser) typeCast() Expr {
	expr := p.typeCheck()
	for p.match(lexer.As) {
		expr = TypeCastExpr{
			Expr: expr,
			Type: p.typeContractDefinable(),
		}
	}
	return expr
}

func (p *Parser) typeCheck() Expr {
	expr := p.logicalOr()
	if p.match(lexer.Is) {
		expr = TypeCheckExpr{
			Expr: expr,
			Type: p.typeContractDefinable(),
		}
	}
	return expr
}

func (p *Parser) logicalOr() (expr Expr) {
	expr = p.logicalAnd()

	for p.match(lexer.Or) {
		op := p.previous()
		rhs := p.logicalAnd()
		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return
}

func (p *Parser) logicalAnd() Expr {
	expr := p.referenceEquality()

	for p.match(lexer.And) {
		op := p.previous()
		rhs := p.referenceEquality()

		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return expr
}

func (p *Parser) referenceEquality() (expr Expr) {
	expr = p.comparison()

	for p.match(lexer.Equals, lexer.NotEquals) {
		op := p.previous()
		rhs := p.comparison()

		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return
}

func (p *Parser) comparison() (expr Expr) {
	expr = p.addition()

	for p.match(lexer.GreaterEqual, lexer.RAngle, lexer.LesserEqual, lexer.LAngle) {
		op := p.previous()
		rhs := p.addition()

		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return
}

func (p *Parser) addition() (expr Expr) {
	expr = p.multiplication()

	for p.match(lexer.Add, lexer.Subtract) {
		op := p.previous()
		rhs := p.multiplication()
		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return
}

func (p *Parser) multiplication() (expr Expr) {
	expr = p.unary()

	for p.match(lexer.Multiply, lexer.Slash, lexer.Mod) {
		op := p.previous()
		rhs := p.unary()
		expr = BinaryExpr{
			Lhs: expr,
			Op:  op.TokenType,
			Rhs: rhs,
		}
	}
	return
}

func (p *Parser) unary() (expr Expr) {
	if p.match(lexer.Subtract, lexer.Not, lexer.Add) {
		op := p.previous()
		rhs := p.unary()
		expr = UnaryExpr{
			Op:  op.TokenType,
			Rhs: rhs,
		}
		return
	}
	expr = p.invoke()
	return
}

func (p *Parser) invoke() (expr Expr) {
	expr = p.funDef()

	for p.match(lexer.LParen, lexer.Dot, lexer.LSquare) {
		switch p.previous().TokenType {
		case lexer.LParen:
			separator := lexer.Comma
			args := p.invocationParameters(&separator)

			expr = InvocationExpr{
				Invoker: expr,
				Args:    args,
			}
		case lexer.Dot:
			id := p.consumeValidIdentifier("Expected identifier inside context getter/setter")

			expr = ContextExpr{
				Context:  expr,
				Variable: VariableExpr{Identifier: string(id.Text)},
			}
		case lexer.LSquare:
			expr = AccessExpr{
				Expr:  expr,
				Index: p.expression(),
			}
			p.consume(lexer.RSquare, "Expected ']' after access index")
		}
	}
	return
}

func (p *Parser) funDef() Expr {
	tok := p.peek()
	switch tok.TokenType {
	case lexer.LParen:
		args := p.functionArguments()
		var typ Type
		p.consume(lexer.Arrow, "Expected arrow at function definition")

		if p.check(lexer.Identifier) && p.isBlockPresent() {
			typ = p.typeContract()
		}
		return FuncDefExpr{
			Arguments:  args,
			ReturnType: typ,
			Statement:  p.statement(),
		}
	case lexer.LBrace:
		mapExpr := p.tryParseMapLiteral()
		if mapExpr != nil {
			return mapExpr
		}
		if p.previous().TokenType == lexer.Arrow {
			panic(ParseError{
				token:   tok,
				message: "Single line function expected, found block function",
			})
		}
		return FuncDefExpr{
			Arguments:  make([]FunctionArgument, 0),
			ReturnType: nil,
			Statement:  p.blockStatement(),
		}
	case lexer.Arrow:
		p.advance()
		return FuncDefExpr{
			Arguments:  make([]FunctionArgument, 0),
			ReturnType: nil,
			Statement:  p.exprStatement(),
		}
	default:
		return p.collection()
	}
}

func (p *Parser) tryParseMapLiteral() Expr {
	p.advance()
	//Peek until reaching a closing brace
	count := 0
	seenColon := false
	for {
		count++
		next := p.advance().TokenType
		if next == lexer.Colon {
			seenColon = true
			break
		}
		if next == lexer.RBrace {
			break
		}
	}
	for i := 0; i < count; i++ {
		p.reverse()
	}
	p.reverse()     // Undo the lbrace read
	if !seenColon { //it's not a map literal
		return nil
	}
	return p.mapLiteral()
}

func (p *Parser) mapLiteral() Expr {
	p.consume(lexer.LBrace, "Expected { in map literal")
	p.cleanNewLines()
	consumeEntry := func() (Expr, Expr) {
		key := p.expression()
		p.consume(lexer.Colon, "Expected colon between map literal key and value")
		val := p.expression()
		return key, val
	}
	entries := make([]MapEntry, 0)

	for {
		if p.peek().TokenType == lexer.RBrace {
			break
		}
		key, val := consumeEntry()
		entries = append(entries, MapEntry{key, val})
		if p.peek().TokenType == lexer.Comma {
			p.advance()
		}
		//p.consume(lexer.Comma, "Expected comma after map literal entry")
		p.cleanNewLines()
	}
	p.consume(lexer.RBrace, "Expected } to close map literal")
	return MapExpr{Entries: entries}
}

func (p *Parser) collection() (expr Expr) {
	if p.match(lexer.LSquare) {
		col := make([]Expr, 0)
		for {
			col = append(col, p.expression())
			p.cleanNewLines()
			if !p.match(lexer.Comma) {
				break
			}
		}
		p.consume(lexer.RSquare, "Expected ']' at end of collection literal")
		return CollectionExpr{
			Elements: col,
		}
	}
	expr = p.primary()
	return
}

func (p *Parser) primary() (expr Expr) {
	var err error
	switch p.peek().TokenType {
	case lexer.String:
		str := p.consume(lexer.String, "Expected string")
		text := string(str.Text)
		text = strings.ReplaceAll(text, "\\n", "\n")
		//TODO other special characters

		expr = StringLiteralExpr{Value: text}
		break
	case lexer.Char:
		charTok := p.consume(lexer.Char, "Expected char")
		char := charTok.Text[0]
		expr = CharLiteralExpr{Value: char}
	case lexer.BooleanTrue:
		p.consume(lexer.BooleanTrue, "Expected BooleanTrue")
		expr = BooleanLiteralExpr{Value: true}
		break
	case lexer.BooleanFalse:
		p.consume(lexer.BooleanFalse, "Expected BooleanFalse")
		expr = BooleanLiteralExpr{Value: false}
		break
	case lexer.Int:
		str := p.consume(lexer.Int, "Expected integer")
		var integer int64
		integer, err = strconv.ParseInt(string(str.Text), 10, 64)
		expr = IntegerLiteralExpr{Value: integer}
		break
	case lexer.Float:
		str := p.consume(lexer.Float, "Expected float")
		var float float64
		float, err = strconv.ParseFloat(string(str.Text), 64)
		expr = FloatLiteralExpr{Value: float}
		break
	case lexer.Identifier:
		str := p.consume(lexer.Identifier, "Expected identifier")
		expr = VariableExpr{Identifier: string(str.Text)}
		break

	case lexer.If:
		return p.ifElseExpression()
	case lexer.LParen:
		p.advance()
		expr = GroupExpr{Group: p.expression()}
		p.consume(lexer.RParen, "Expected ')' after grouped expression")
	}

	if err != nil {
		panic(ParseError{
			token:   p.previous(),
			message: "Expected literal",
		})
	}

	if expr == nil {
		panic(ParseError{
			token:   p.peek(),
			message: "Invalid expression",
		})
	}
	return
}

func (p *Parser) ifElseExpression() Expr {
	p.consume(lexer.If, "Expected if at beginning of if expression")
	condition := p.logicalOr()
	if p.peek().TokenType == lexer.Arrow {
		p.consume(lexer.Arrow, "")
		mainResult := p.expression()

		elseBranch, elseResult := p.elseExpression()

		return IfElseExpr{
			Condition:  condition,
			IfBranch:   nil,
			IfResult:   mainResult,
			ElseBranch: elseBranch,
			ElseResult: elseResult,
		}
	}

	mainBranch := p.blockStatement()
	mainResult := mainBranch.Stmts[len(mainBranch.Stmts)-1]
	_, isExpr := mainResult.(ExpressionStmt)
	if !isExpr {
		panic(ParseError{message: "Last line in an `if` block must be an expression"})
	}

	elseBranch, elseResult := p.elseExpression()

	return IfElseExpr{
		Condition:  condition,
		IfBranch:   mainBranch.Stmts[:len(mainBranch.Stmts)-1], //drop the last result
		IfResult:   mainResult.(ExpressionStmt).Expr,
		ElseBranch: elseBranch,
		ElseResult: elseResult,
	}
}

func (p *Parser) elseExpression() ([]Stmt, Expr) {
	p.cleanNewLines()
	p.consume(lexer.Else, "if expression must follow with else expression")
	if p.peek().TokenType == lexer.Arrow {
		p.advance()
		return nil, p.expression()
	} else if p.peek().TokenType == lexer.If {
		return nil, p.ifElseExpression()
	} else {
		elseBranch := p.blockStatement()
		elseResult := elseBranch.Stmts[len(elseBranch.Stmts)-1]

		_, isExpr := elseResult.(ExpressionStmt)
		if !isExpr {
			panic(ParseError{message: "Last line in an `else` expression block must be an expression"})
		}

		return elseBranch.Stmts[:len(elseBranch.Stmts)-1], elseResult.(ExpressionStmt).Expr
	}
}
