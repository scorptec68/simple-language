package simple_language

import (
	"errors"
	"go/parser"
	"golang.org/x/tools/go/gcimporter15/testdata"
	"strconv"
)

type ValueType int
type StatementType int
type LoopType int
type ExpressionType int
type IntExpressionType int
type IntOperatorType int
type BoolExpressionType int
type BoolOperatorType int
type StringExpressionType int
type StringOperatorType int

const (
	ValueInteger ValueType = iota
    ValueString
    ValueBoolean
)

const (
    StmtLoop StatementType = iota
    StmtIf
    StmtAssignment
    StmtPrint
)

const (
	LoopForever LoopType = iota
	LoopWhile
	LoopTimes
)

const (
    ExprnIdentifier ExpressionType = iota
    ExprnInteger
    ExprnBoolean
    ExprnString
)

const (
	IntExprnValue IntExpressionType = iota
	IntExprnVariable
	IntExprnBinary
	IntExprnUnary
)

const (
	IntUnaryOpNegative IntOperatorType = iota
	IntBinaryOp
	IntBinaryOpAdd
	IntBinaryOpMinus
	IntBinaryOpTimes
	IntBinaryOpDivide
)

const (
	BoolExprnValue BoolExpressionType = iota
	BoolExprnVariable
	BoolExprnIntBinary
	BoolExprnBoolBinary
	BoolExprnUnary
)

const (
	BoolUnaryOpNot IntOperatorType = iota
	BoolBinaryOp
	BoolBinaryOpOr
	BoolBinaryOpAnd
)

const (
	StringExprnValue StringExpressionType = iota
	StringExprnVariable
	StringExprnBinary
)

const (
	StringBinaryOpAdd StringOperatorType = iota
	StringBinaryOpMinus
)

type Program struct {
	variables *Variables
	stmtList []Statement
}

type Variables struct {
	values map[string]*Value
}

type Parser struct {
	lex *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

// nextItem returns the nextItem token.
func (p *Parser) nextItem() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *Parser) backup() {
	p.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (p *Parser) backup2(t1 item) {
	p.token[1] = t1
	p.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (p *Parser) backup3(t2, t1 item) { // Reverse order: we're pushing back.
	p.token[1] = t1
	p.token[2] = t2
	p.peekCount = 3
}

// peek returns but does not consume the nextItem token.
func (p *Parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func (parser *Parser) ParseProgram() (*Program, error) {
	var err error
	prog := new(Program)
	prog.variables, err = parser.parseVariables()
	if err != nil {
		return nil, err
	}
	item := parser.nextItem()
	if item.typ != itemRun {
	    err := errors.New("Missing Run keyword")
	    return nil, err
	}

	prog.stmtList, err = parser.parseStementList()
	if err != nil {
		return nil, err
	}

	item = parser.nextItem()
	if item.typ != itemEndRun {
		err := errors.New("Missing EndRun keyword")
		return nil, err
	}
	return prog, nil
}

func (parser *Parser) parseVariables() (*Variables, error) {
    vars := new(Variables)
    vars.values = make(map[string]*Value)
    item := parser.nextItem()
    if item.typ != itemVar {
    	// no variables to process
		parser.backup()
    	return vars, nil
	}
	// we have potentially some variables (could be empty)
	for {
		item = parser.nextItem()
		switch item.typ {
		case itemEndVar:
			// end of variable declaration
			return vars, nil
		case itemEOF:
			// end of any input which is an error
			err := errors.New("Cannot find EndVar")
			return nil, err
		case itemIdentifier:
            idStr := item.val
			item = parser.nextItem()
			if item.typ != itemColon {
				err := errors.New("Expecting colon")
				return nil, err
			}
			value, err := parser.parseValue()
			if err != nil {
				return nil, err
			}
	        vars.values[idStr] = value
		}
	}
}

func (parser *Parser) parseValue() (*Value, error) {
	var value Value
	item := parser.nextItem()
	if !(item.typ == itemString || item.typ == itemBoolean || item.typ == itemInteger) {
		err := errors.New("Expecting a variable type")
		return nil, err
	}
	switch item.typ {
	case itemString:
		value = Value{valueType: ValueString, stringVal: item.val}
	case itemInteger:
		i, _ := strconv.Atoi(item.val)
		value = Value{valueType: ValueInteger, intVal: i}
	case itemBoolean:
		b, _ := strconv.ParseBool(item.val)
		value = Value{valueType: ValueBoolean, boolVal: b}
	}
	return &value, nil
}

func isEndKeyword(i item) bool {
	return i.typ == itemEndRun || i.typ == itemEndLoop || i.typ == itemEndIf;

}

func (parser *Parser) parseStatementList() ([]*Statement, error) {
	var stmtList []*Statement
	for {
		stmt, err := parser.parseStatement()
		if err != nil {
			return nil, err
		}
		stmtList = append(stmtList, stmt)
		item := parser.peek()
		if isEndKeyword(item) {
			return stmtList, nil
		}
	}
}

func (parser *Parser) parseStatement() (stmt *Statement, err error) {
	var stmtType StatementType

	var assignStmt *AssignmentStatement
	var ifStmt *IfStatement
	var loopStmt *LoopStatement
	var printStmt *PrintStatement

    item := parser.nextItem()
    switch item.typ {
	case itemIdentifier:
		stmtType = StmtAssignment
		parser.backup()
		assignStmt, err = parser.parseAssignment()
	case itemIf:
		stmtType = StmtIf
		ifStmt, err = parser.parseIfStatement()
	case itemLoop:
		stmtType = StmtLoop
		loopStmt, err = parser.parseLoopStatement()
	case itemPrint:
		stmtType = StmtPrint
		printStmt, err = parser.parsePrintStatement()
	}

	return &Statement{stmtType, assignStmt, ifStmt, loopStmt, printStmt}, err
}

func (parser *Parser) parseAssignment() (assign *AssignmentStatement, err error) {
	idItem := parser.nextItem()
	op := parser.nextItem()
	if op.typ != itemEquals {
	    err	= errors.New("Missing expected equals assign")
	}
	exprn, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
    return &AssignmentStatement{idItem.val, exprn}, nil
}

func (parser *Parser) parseBoolExpression() (boolExprn *BoolExpression, err error) {
	item := parser.nextItem()
	switch item.typ {
	case itemTrue:
		boolExprn.boolExprnType = BoolExprnValue
		boolExprn.boolValue = true
	case itemFalse:
		boolExprn.boolExprnType = BoolExprnValue
		boolExprn.boolValue = false
	case itemIdentifier:
		boolExprn.boolExprnType = BoolExprnVariable
		boolExprn.identifier = item.val
	case itemNot:
		lhsBoolExprn, err := parser.parseBoolExpression()
		if err == nil {
			boolExprn.boolExprnType = BoolExprnUnary
			boolExprn.lhsBoolExprn = lhsBoolExprn
			return boolExprn, nil
		} else {
			parser.backup()
			err = errors.New("Missing boolean expression for unary not")
			return nil, err
		}
	default:
		lhsBoolExprn, err := parser.parseBoolExpression()
		if err == nil {
			// check for boolean operators
		    item = parser.nextItem()
		    if item.typ == itemAnd || item.typ == itemOr {
		    	rhsBoolExprn, err := parser.parseBoolExpression()
		    	if err == nil {
		    	    // we have a boolean expression, yay!
					boolExprn.boolExprnType = BoolExprnBoolBinary
					boolExprn.lhsBoolExprn = lhsBoolExprn
					boolExprn.rhsBoolExprn = rhsBoolExprn
					return boolExprn, nil
				} else {
					parser.backup()
					err = errors.New("right hand side is not a boolean expression")
					return nil, err
				}
			} else {
				parser.backup()
				err = errors.New("missing operator for right hand boolean expression")
				return nil, err
			}
		} else {
			// try the integer expression
			lhsIntExprn, err := parser.parseIntExpression()
			if err == nil {
				item = parser.nextItem()
				if item.typ == itemLessThan || item.typ == itemGreaterThan || item.typ == itemLessEquals ||
					item.typ == itemGreaterEquals {
					rhsIntExprn, err := parser.parseIntExpression()
					if err == nil {
						// we have an integer expression, yay!
						boolExprn.boolExprnType = BoolExprnIntBinary
						boolExprn.lhsIntExprn = lhsIntExprn
						boolExprn.rhsIntExprn = rhsIntExprn
						return boolExprn, nil
					} else {
						parser.backup()
						err = errors.New("right hand side is not an integer expression")
						return nil, err
					}
				} else {
					parser.backup()
					err = errors.New("missing operator for right hand integer expression")
					return nil, err
				}
			} else {
				parser.backup()
				err = errors.New("Couldn't parse boolean expression")
				return nil, err
			}
		}
	}
}

func (parser *Parser) parseIfStatement() (ifStmt *IfStatement, err error) {
	var elseifList []ElseIf
	var elseStmtList []Statement

    exprn, err := parser.parseBoolExpression()
    if err != nil {
    	return nil, err
	}

    stmtList, err := parser.parseStatementList()
	if err != nil {
		return nil, err
	}

	// test for next optional parts or the end
    item := parser.nextItem()

    // optional elseif
    if item.typ == itemElseIf {
		elseifList, err = parser.parseElseIfList()
		if err != nil {
			return nil, err
		}
		item = parser.nextItem()
	}

	// optional else
	if item.typ == itemElse {
		elseStmtList, err = parser.parseElse()
		if err != nil {
			return nil, err
		}
		item = parser.nextItem()
	}

	// mandatory endif
	if item.typ != itemEndIf {
	    err = errors.New("Missing endif")
	    return nil, err
	}

	return &IfStatement{exprn, stmtList, elseifList, elseStmtList}, nil
}

type Value struct {
	valueType ValueType

	intVal int
	stringVal string
	boolVal bool
}

type Statement struct {
	stmtType StatementType

	assignmentStmt *AssignmentStatement
	ifStmt   *IfStatement
	loopStmt *LoopStatement
	printStmt *PrintStatement
}

type LoopStatement struct {
	loopType LoopType

	intExpression *IntegerExpression
	boolExpression *BoolExpression
	stmtList []Statement
}

type IfStatement struct {
	boolExpression *BoolExpression
	stmtList []Statement
    elsifList []ElseIf
    elseStmtList []Statement
}

type ElseIf struct {
	boolExpression BoolExpression
	stmtList []Statement
}

type AssignmentStatement struct {
	identifier string
	exprn *Expression
}

type PrintStatement struct {
	exprn *StringExpression
}

type Expression struct {
	exprnType ExpressionType

	intExpression *IntegerExpression
	boolExpression *BoolExpression
	stringExpression *StringExpression
}

type IntegerExpression struct {
	intExprnType IntExpressionType

    intVal int
	identifier string
    lhsExprn *IntegerExpression
    rhsExprn *IntegerExpression
    operator IntOperatorType
}

type BoolExpression struct {
	boolExprnType BoolExpressionType

	boolValue bool
	identifier string
	lhsBoolExprn *BoolExpression
	rhsBoolExprn *BoolExpression
	lhsIntExprn *IntExpression
	rhsIntExprn *IntExpression
	operator BoolOperatorType
}

type StringExpression struct {
	strExprnType StringExpressionType

	strVal string
	identifier string
	lhsExprn StringExpression
	rhsExprn StringExpression
	operator StringOperatorType
}
