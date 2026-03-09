package parser

import (
	"go-java-compiler/src/ast"
	"go-java-compiler/src/lexer"
	"strconv"
	"fmt"
)

type Parser struct {
	l      *lexer.Lexer
	curTok lexer.Token
	peekTok lexer.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	for p.curTok.Type != lexer.EOF {
		if p.curTok.Type == lexer.CLASS {
			classDel := p.parseClass()
			if classDel != nil {
				program.Classes = append(program.Classes, classDel)
			}
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseClass() *ast.ClassDecl {
	// expect "class"
	if p.curTok.Type != lexer.CLASS {
		fmt.Printf("ERROR: Expected CLASS, got %s (%q)\n", p.curTok.Type, p.curTok.Literal)
		return nil
	}
	p.nextToken()

	// expect class name
	if p.curTok.Type != lexer.IDENT {
		fmt.Printf("ERROR: Expected IDENT for class name, got %s (%q)\n", p.curTok.Type, p.curTok.Literal)
		return nil
	}
	className := p.curTok.Literal
	p.nextToken()

	// expect '{'
	if p.curTok.Type != lexer.LBRACE {
		fmt.Printf("ERROR: Expected LBRACE, got %s (%q)\n", p.curTok.Type, p.curTok.Literal)
		return nil
	}
	p.nextToken()

	classDec := &ast.ClassDecl{Name: className}

	// parse methods/fields
	for p.curTok.Type != lexer.RBRACE && p.curTok.Type != lexer.EOF {
		if p.curTok.Type == lexer.PUBLIC {
			method := p.parseMethod()
			if method != nil {
				classDec.Methods = append(classDec.Methods, method)
			}
		} else {
			p.nextToken()
		}
	}

	return classDec
}

func (p *Parser) parseMethod() *ast.MethodDecl {
	method := &ast.MethodDecl{IsPublic: true}

	// expect "public"
	if p.curTok.Type != lexer.PUBLIC {
		return nil
	}
	p.nextToken()

	// expect "static"
	if p.curTok.Type == lexer.STATIC {
		method.IsStatic = true
		p.nextToken()
	}

	// expect return type
	if p.curTok.Type == lexer.VOID {
		method.ReturnType = "void"
		p.nextToken()
	} else if p.curTok.Type == lexer.INT_KW {
		method.ReturnType = "int"
		p.nextToken()
	} else if p.curTok.Type == lexer.IDENT {
		method.ReturnType = p.curTok.Literal
		p.nextToken()
	} else {
		return nil
	}

	// expect method name
	if p.curTok.Type != lexer.IDENT {
		return nil
	}
	method.Name = p.curTok.Literal
	p.nextToken()

	// expect '('
	if p.curTok.Type != lexer.LPAREN {
		return nil
	}
	p.nextToken()

	// parse parameters
	for p.curTok.Type != lexer.RPAREN && p.curTok.Type != lexer.EOF {
		paramType := ""
		if p.curTok.Type == lexer.IDENT {
			paramType = p.curTok.Literal
			p.nextToken()
			// Handle array notation
			if p.curTok.Type == lexer.LBRACKET {
				p.nextToken()
				if p.curTok.Type == lexer.RBRACKET {
					paramType += "[]"
					p.nextToken()
				}
			}
		} else if p.curTok.Type == lexer.INT_KW {
			paramType = "int"
			p.nextToken()
		}

		if p.curTok.Type == lexer.IDENT {
			paramName := p.curTok.Literal
			method.Params = append(method.Params, &ast.ParamDecl{Name: paramName, Type: paramType})
			p.nextToken()

			if p.curTok.Type == lexer.COMMA {
				p.nextToken()
			}
		}
	}

	// expect ')'
	if p.curTok.Type != lexer.RPAREN {
		return nil
	}
	p.nextToken()

	// expect '{'
	if p.curTok.Type != lexer.LBRACE {
		return nil
	}
	p.nextToken()

	// parse method body
	for p.curTok.Type != lexer.RBRACE && p.curTok.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			method.Body = append(method.Body, stmt)
		}
	}

	// expect '}'
	if p.curTok.Type != lexer.RBRACE {
		return nil
	}
	p.nextToken()

	return method
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curTok.Type {
	case lexer.INT_KW:
		return p.parseVarDecl()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVarDecl() *ast.VarDecl {
	varDecl := &ast.VarDecl{}

	if p.curTok.Type == lexer.INT_KW {
		varDecl.Type = "int"
		p.nextToken()
	} else {
		return nil
	}

	if p.curTok.Type != lexer.IDENT {
		return nil
	}
	varDecl.Name = p.curTok.Literal
	p.nextToken()

	if p.curTok.Type == lexer.ASSIGN {
		p.nextToken()
		varDecl.Value = p.parseExpression()
	}

	if p.curTok.Type == lexer.SEMICOLON {
		p.nextToken()
	}

	return varDecl
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	expr := p.parseExpression()

	if p.curTok.Type == lexer.SEMICOLON {
		p.nextToken()
	}

	return &ast.ExpressionStatement{Expr: expr}
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(precedence int) ast.Expression {
	leftExpr := p.parsePrimary()

	for p.curTok.Type == lexer.PLUS || p.curTok.Type == lexer.MINUS ||
		p.curTok.Type == lexer.ASTERISK || p.curTok.Type == lexer.SLASH {

		opPrecedence := p.getPrecedence(p.curTok.Type)
		if opPrecedence < precedence {
			break
		}

		op := p.curTok.Type
		p.nextToken()

		rightExpr := p.parseBinaryExpr(opPrecedence + 1)
		leftExpr = &ast.BinaryExpr{Left: leftExpr, Op: op, Right: rightExpr}
	}

	return leftExpr
}

func (p *Parser) parsePrimary() ast.Expression {
	switch p.curTok.Type {
	case lexer.INT:
		val, _ := strconv.ParseInt(p.curTok.Literal, 10, 64)
		p.nextToken()
		return &ast.IntegerLiteral{Value: val}
	case lexer.STRING:
		val := p.curTok.Literal
		p.nextToken()
		return &ast.StringLiteral{Value: val}
	case lexer.IDENT:
		name := p.curTok.Literal
		p.nextToken()

		// Check if it's a method call
		if p.curTok.Type == lexer.DOT {
			p.nextToken()
			if p.curTok.Type == lexer.IDENT {
				obj := &ast.Identifier{Name: name}
				methodName := p.curTok.Literal
				p.nextToken()

				// Check if it's a method call with arguments
				if p.curTok.Type == lexer.LPAREN {
					p.nextToken()
					args := []ast.Expression{}
					for p.curTok.Type != lexer.RPAREN && p.curTok.Type != lexer.EOF {
						args = append(args, p.parseExpression())
						if p.curTok.Type == lexer.COMMA {
							p.nextToken()
						}
					}
					if p.curTok.Type == lexer.RPAREN {
						p.nextToken()
					}
					return &ast.MethodCall{Object: obj, Method: methodName, Arguments: args}
				}
			}
		}

		return &ast.Identifier{Name: name}
	case lexer.LPAREN:
		p.nextToken()
		expr := p.parseExpression()
		if p.curTok.Type == lexer.RPAREN {
			p.nextToken()
		}
		return expr
	}

	return &ast.IntegerLiteral{Value: 0} // default
}

func (p *Parser) getPrecedence(tokenType lexer.TokenType) int {
	switch tokenType {
	case lexer.PLUS, lexer.MINUS:
		return 1
	case lexer.ASTERISK, lexer.SLASH:
		return 2
	default:
		return 0
	}
}
