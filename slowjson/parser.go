package slowjson

import (
	"fmt"
	"strings"
	"unicode"
)

// NodeType represents the kind of JSON node.
type NodeType int

const (
	NodeUnknown NodeType = iota
	NodeObject
	NodeArray
	NodeString
	NodeNumber
	NodeBoolean
	NodeNull
)

// Node is a parsed JSON element, with start/end line/column info.
// For objects and arrays, Children holds contained items.
// For strings, numbers, booleans, and null, Value holds the literal.
// StartLine, StartCol, EndLine, EndCol indicate where the node begins/ends.
// The entire JSON source is stored in Source for easy context extraction.
type Node struct {
	Type     NodeType
	Value    string
	Children []*Node

	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int

	// Store entire input for debug context. In practice you may store it externally.
	Source string
}

// DebugContext returns lines around the node to help in debugging.
// linesBefore and linesAfter let you specify how many lines of context to include.
func (n *Node) DebugContext(linesBefore, linesAfter int) string {
	sourceLines := strings.Split(n.Source, "\n")

	// 1-based indexing in lines.
	start := n.StartLine - 1 - linesBefore
	if start < 0 {
		start = 0
	}
	end := n.EndLine - 1 + linesAfter
	if end >= len(sourceLines) {
		end = len(sourceLines) - 1
	}

	var builder strings.Builder
	for i := start; i <= end; i++ {
		lineNum := i + 1
		builder.WriteString(fmt.Sprintf("%d: %s\n", lineNum, sourceLines[i]))
		// We can annotate the lines where the node starts or ends.
		// If we are on the start line or end line, we can place a caret.
		if lineNum == n.StartLine {
			caretPos := n.StartCol
			if caretPos > 0 && caretPos <= len(sourceLines[i]) {
				builder.WriteString(fmt.Sprintf("%s^ start\n", strings.Repeat(" ", caretPos)))
			}
		}
		if lineNum == n.EndLine {
			caretPos := n.EndCol
			if caretPos > 0 && caretPos <= len(sourceLines[i]) {
				builder.WriteString(fmt.Sprintf("%s^ end\n", strings.Repeat(" ", caretPos)))
			}
		}
	}

	return builder.String()
}

// Parser implements a simple JSON parser that tracks line/column positions.
type Parser struct {
	runes  []rune
	pos    int
	line   int
	col    int
	length int
	// original input
	source string
}

// NewParser creates a Parser from the given JSON string.
func NewParser(input string) *Parser {
	r := []rune(input)
	return &Parser{
		runes:  r,
		pos:    0,
		line:   1, // 1-based indexing for line
		col:    1, // 1-based indexing for column
		length: len(r),
		source: input,
	}
}

// Parse parses the entire input and returns the root Node.
// If any error occurs, a partial node might still be returned.
func (p *Parser) Parse() (*Node, error) {
	n, err := p.parseValue()
	if err != nil {
		return n, err
	}
	p.skipWhitespace()
	// If we haven't consumed all runes, we can ignore or return an error.
	if !p.isEOF() {
		// We'll ignore trailing characters, or you can return an error.
	}
	return n, nil
}

func (p *Parser) parseValue() (*Node, error) {
	p.skipWhitespace()
	if p.isEOF() {
		// Return an error node
		return nil, fmt.Errorf("unexpected end of input")
	}

	switch p.peekChar() {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		return p.parseString()
	case 't', 'f':
		return p.parseBoolean()
	case 'n':
		return p.parseNull()
	default:
		// Might be a number?
		return p.parseNumber()
	}
}

func (p *Parser) parseObject() (*Node, error) {
	n := &Node{
		Type:      NodeObject,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	p.consumeChar() // consume '{'
	p.skipWhitespace()

	n.Children = []*Node{}

	// Check for empty object
	if p.peekChar() == '}' {
		p.consumeChar()
		n.EndLine = p.line
		n.EndCol = p.col
		return n, nil
	}

	for {
		p.skipWhitespace()
		if p.peekChar() != '"' {
			return n, fmt.Errorf("expected string key at line %d col %d", p.line, p.col)
		}
		keyNode, err := p.parseString()
		if err != nil {
			return n, err
		}

		p.skipWhitespace()
		if p.peekChar() != ':' {
			return n, fmt.Errorf("expected ':' after object key at line %d col %d", p.line, p.col)
		}
		p.consumeChar() // consume ':'

		valueNode, err := p.parseValue()
		if err != nil {
			return n, err
		}

		// We can store key as a child with a single child representing its value
		// or we can store them differently. We'll create a node of type NodeString for key
		// and attach the value as its child.
		keyNode.Children = []*Node{valueNode}
		n.Children = append(n.Children, keyNode)

		p.skipWhitespace()
		if p.peekChar() == '}' {
			p.consumeChar()
			n.EndLine = p.line
			n.EndCol = p.col
			return n, nil
		}
		if p.peekChar() != ',' {
			return n, fmt.Errorf("expected ',' or '}' in object at line %d col %d", p.line, p.col)
		}
		p.consumeChar() // consume ','
	}
}

func (p *Parser) parseArray() (*Node, error) {
	n := &Node{
		Type:      NodeArray,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	p.consumeChar() // consume '['
	p.skipWhitespace()

	n.Children = []*Node{}

	// Check for empty array
	if p.peekChar() == ']' {
		p.consumeChar()
		n.EndLine = p.line
		n.EndCol = p.col
		return n, nil
	}

	for {
		valueNode, err := p.parseValue()
		if err != nil {
			return n, err
		}
		n.Children = append(n.Children, valueNode)

		p.skipWhitespace()
		if p.peekChar() == ']' {
			p.consumeChar()
			n.EndLine = p.line
			n.EndCol = p.col
			return n, nil
		}
		if p.peekChar() != ',' {
			return n, fmt.Errorf("expected ',' or ']' in array at line %d col %d", p.line, p.col)
		}
		p.consumeChar() // consume ','
		p.skipWhitespace()
	}
}

func (p *Parser) parseString() (*Node, error) {
	n := &Node{
		Type:      NodeString,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	p.consumeChar() // consume '"'

	var sb strings.Builder
	for {
		if p.isEOF() {
			n.Value = sb.String()
			n.EndLine = p.line
			n.EndCol = p.col
			return n, fmt.Errorf("unexpected end of input in string")
		}
		ch := p.peekChar()
		if ch == '"' {
			p.consumeChar()
			break
		}
		if ch == '\\' {
			// handle escape
			p.consumeChar() // consume '\'
			if p.isEOF() {
				return n, fmt.Errorf("unexpected end of input in string escape")
			}
			escaped := p.peekChar()
			// simplistic approach: just append the character after '\' as-is
			sb.WriteRune(escaped)
			p.consumeChar()
		} else {
			sb.WriteRune(ch)
			p.consumeChar()
		}
	}

	n.Value = sb.String()
	n.EndLine = p.line
	n.EndCol = p.col
	return n, nil
}

func (p *Parser) parseNumber() (*Node, error) {
	n := &Node{
		Type:      NodeNumber,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	var sb strings.Builder

	// We'll parse until we hit a non-number-related character.
	// This is naive and does not strictly enforce JSON numeric format.

	for !p.isEOF() {
		ch := p.peekChar()
		if ch == '-' || ch == '+' || ch == '.' || unicode.IsDigit(ch) {
			sb.WriteRune(ch)
			p.consumeChar()
		} else {
			break
		}
	}

	n.Value = sb.String()
	n.EndLine = p.line
	n.EndCol = p.col

	// We could validate if it's a valid number.
	return n, nil
}

func (p *Parser) parseBoolean() (*Node, error) {
	n := &Node{
		Type:      NodeBoolean,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	if strings.HasPrefix(p.remaining(), "true") {
		n.Value = "true"
		// consume 4 chars
		for i := 0; i < 4; i++ {
			p.consumeChar()
		}
	} else if strings.HasPrefix(p.remaining(), "false") {
		n.Value = "false"
		// consume 5 chars
		for i := 0; i < 5; i++ {
			p.consumeChar()
		}
	} else {
		return n, fmt.Errorf("invalid boolean at line %d col %d", p.line, p.col)
	}

	n.EndLine = p.line
	n.EndCol = p.col
	return n, nil
}

func (p *Parser) parseNull() (*Node, error) {
	n := &Node{
		Type:      NodeNull,
		Source:    p.source,
		StartLine: p.line,
		StartCol:  p.col,
	}

	if strings.HasPrefix(p.remaining(), "null") {
		n.Value = "null"
		// consume 4 chars
		for i := 0; i < 4; i++ {
			p.consumeChar()
		}
	} else {
		return n, fmt.Errorf("invalid null at line %d col %d", p.line, p.col)
	}

	n.EndLine = p.line
	n.EndCol = p.col
	return n, nil
}

func (p *Parser) skipWhitespace() {
	for !p.isEOF() {
		ch := p.peekChar()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			p.consumeChar()
		} else {
			break
		}
	}
}

func (p *Parser) peekChar() rune {
	if p.pos >= p.length {
		return 0
	}
	return p.runes[p.pos]
}

func (p *Parser) consumeChar() {
	if p.pos >= p.length {
		return
	}

	ch := p.runes[p.pos]
	p.pos++
	// update line/col
	if ch == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
}

func (p *Parser) remaining() string {
	if p.pos >= p.length {
		return ""
	}
	return string(p.runes[p.pos:])
}

func (p *Parser) isEOF() bool {
	return p.pos >= p.length
}

// Example usage:
// func main() {
// 	input := `{"hello": [1, 2, 3], "world": true, "nested": {"foo": "bar"}}`
// 	p := NewParser(input)
// 	n, err := p.Parse()
// 	if err != nil {
// 	\tfmt.Println("Parse error:", err)
// 	}
// 	// Debug context for the root node.
// 	fmt.Println(n.DebugContext(1, 1))
// }
