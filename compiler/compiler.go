// Package compiler implements our compiler.
//
// The compiler makes use of our parser and lexer packages
// and largely walks trees generating snippets of assembly
// language as it goes.
package compiler

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"

	"s-lang/lexer"
	"s-lang/parser"
)

//go:embed header.s.txt
var header string

//go:embed footer.s.txt
var footer string

// Compiler holds our internal compiler state.
type Compiler struct {

	// Source holds the source program we'll work on.
	Source string

	// buff is the writer object we use to send our
	// output to, as it is generated.
	buff bytes.Buffer

	// labelCount is used for generating labels
	labelCount int

	// stringTable holds (interned) strings
	stringTable map[string]string

	// functions stores the name of functions we're compiling
	// We should only be compiling one function at a time,
	// but this way we could allow nested functions if we
	// wanted to.
	//
	// The most recent function we've entered is the LAST
	// item on the array.
	functions []string

	// scope stores stack-frames which are used to hold
	// symbols.
	scope *Scope

	// globalVariables _should_ use the same stack frame,
	// however for quickness they are here.
	globalVariables []*GlobalVariable

	// globalCount stores the count of global variables
	globalCount int
}

// New creates a new compiler instance, to compile the
// given program.
func New(source string) *Compiler {
	global := NewScope(nil)
	tmp := &Compiler{
		Source:      source,
		stringTable: make(map[string]string),
		scope:       global,
	}

	return tmp
}

// Compile produces, and returns, an assembly language
// implementation of the program which was passed to
// our constructor.
func (c *Compiler) Compile() (string, error) {

	// Create a lexer and parser
	lex := lexer.NewLexer(string(c.Source))
	parse := parser.New(lex)

	// now parse the program
	program, err := parse.ParseProgram()
	if err != nil {
		return "", err
	}

	// Write the header
	fmt.Fprintf(&c.buff, "%s", header)

	// compile each statement
	for _, stmt := range program.Statements {
		err := c.generateStmt(stmt)
		if err != nil {
			return "", err
		}
	}

	// Write the footer
	fmt.Fprintf(&c.buff, "%s", footer)

	// If there are any string-table entries then we write them
	// to the foot of the file.
	if len(c.stringTable) > 0 {

		// The header
		fmt.Fprintf(&c.buff, "\n# generated string table\n.section .data\n")

		// The actual strings.
		for k, v := range c.stringTable {
			v = strings.ReplaceAll(v, "\n", "\\n")
			fmt.Fprintf(&c.buff, "  %s: .ascii \"%s\"\n", k, v)
			fmt.Fprintf(&c.buff, ".byte 00  # null byte at end of string\n")
			fmt.Fprintf(&c.buff, "  %s_end:\n", k)
		}
	}

	if len(c.globalVariables) > 0 {

		fmt.Fprintf(&c.buff, `
# globals
.section .data
`)

		for _, g := range c.globalVariables {

			fmt.Fprintf(&c.buff, `
%s:
    .quad 0
`, g.Label)
		}
	}

	return c.buff.String(), nil
}

// pushScope enters a new scope for compiling a function body.
func (c *Compiler) pushScope() {
	c.scope = NewScope(c.scope)
}

// popScope returns to the parent scope, when compiling a function body is complete.
func (c *Compiler) popScope() {
	c.scope = c.scope.Parent
}

func (c *Compiler) emitLoadVariable(name string) error {

	sym, ok := c.scope.Lookup(name)
	if !ok {
		return fmt.Errorf("undefined variable: %s", name)
	}

	switch v := sym.(type) {

	case *FunctionVariable:

		if v.Offset < 0 {
			fmt.Fprintf(&c.buff, `
	mov rax, [rbp-%d]
`, -v.Offset)
		} else {
			fmt.Fprintf(&c.buff, `
	mov rax, [rbp+%d]
`, v.Offset)
		}

	case *GlobalVariable:

		fmt.Fprintf(&c.buff, `
	mov rax, [%s]
`, v.Label)

	default:
		return fmt.Errorf("unknown symbol type")
	}

	return nil
}

func (c *Compiler) emitStoreVariable(name string) error {

	sym, ok := c.scope.Lookup(name)
	if !ok {
		return fmt.Errorf("undefined variable: %s", name)
	}

	switch v := sym.(type) {

	case *FunctionVariable:

		if v.Offset < 0 {
			fmt.Fprintf(&c.buff, `
mov [rbp-%d], rax
`, -v.Offset)
		} else {
			fmt.Fprintf(&c.buff, `
mov [rbp+%d], rax
`, v.Offset)
		}

	case *GlobalVariable:

		fmt.Fprintf(&c.buff, `
mov [%s], rax
`, v.Label)

	default:
		return fmt.Errorf("unknown symbol type")
	}

	return nil
}

func (c *Compiler) newGlobalLabel(name string) string {
	lbl := fmt.Sprintf("global_%s_%d", name, c.globalCount)
	c.globalCount++
	return lbl
}

// hashString is used to generate a stable identifier for any string literals within
// the program we generate - this is used primarily to allow string interning.
func (c *Compiler) hashString(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("msg_%s", sha)
}

// compileExpr handles compiling an expression.
func (c *Compiler) compileExpr(e parser.Expr) error {
	switch v := e.(type) {

	case *parser.IntegerExpr:
		fmt.Fprintf(&c.buff, `
mov rax, %d`, v.Value)

	case *parser.FunctionCallExpr:

		// We have to loop over the arguments in reverse
		for i := len(v.Arguments) - 1; i >= 0; i-- {
			// push each argument to the stack
			err := c.compileExpr(v.Arguments[i])
			if err != nil {
				return err
			}
			fmt.Fprintf(&c.buff, `
	push rax
`)

		}
		fmt.Fprintf(&c.buff, `
call %s
add rsp, %d
`, v.Name, 8*len(v.Arguments))

	case *parser.VariableExpr:
		return c.emitLoadVariable(v.Name)

	case *parser.StringExpr:
		return fmt.Errorf("compileExpr cannot handle a string-expression")

	case *parser.BinaryExpr:

		err := c.compileExpr(v.Left)
		if err != nil {
			return err
		}

		fmt.Fprintln(&c.buff, `
	push rax`)

		err = c.compileExpr(v.Right)
		if err != nil {
			return err
		}

		fmt.Fprintln(&c.buff, `
	pop rbx`)

		switch v.Op {

		case lexer.PLUS:
			fmt.Fprintln(&c.buff, `
	add rax, rbx`)

		case lexer.MINUS:
			fmt.Fprintln(&c.buff, `
	sub rbx, rax
	mov rax, rbx`)

		case lexer.MULTIPLY:
			fmt.Fprintln(&c.buff, `
	imul rax, rbx`)

		case lexer.DIVIDE:
			fmt.Fprintln(&c.buff, `
	mov rdx, 0
	mov rcx, rax
	mov rax, rbx
	idiv rcx`)
		case lexer.EQUALS:
			fmt.Fprintln(&c.buff, `
	cmp rbx, rax
	sete al
	movzx rax, al`)

		case lexer.NOT_EQUALS:
			fmt.Fprintln(&c.buff, `
	cmp rbx, rax
	setne al
	movzx rax, al`)

		case lexer.LT:
			fmt.Fprintln(&c.buff, `
	cmp rbx, rax
	setl al
	movzx rax, al`)

		case lexer.LT_EQUALS:
			fmt.Fprintln(&c.buff, `
	cmp rbx, rax
	setle al
	movzx rax, al`)

		case lexer.GT:
			fmt.Fprintln(&c.buff, `
	cmp rbx, rax
	setg al
	movzx rax, al`)

		case lexer.GT_EQUALS:
			fmt.Fprintln(&c.buff, `
cmp rbx, rax
setge al
movzx rax, al`)
		case lexer.AND:
			fmt.Fprintln(&c.buff, `
imul rax, rbx`)
		case lexer.OR:
			fmt.Fprintln(&c.buff, `
or rax, rbx
cmp rax, 0
setne al
movzx rax, al`)
		default:
			return fmt.Errorf("unhandled token in compileExpr: %v", v)
		}

	}
	return nil
}

// generateStmt generates the assembly for a single statement, it is moved into
// its own function so that we can call it recursively for cases like "if" and "while".
func (c *Compiler) generateStmt(stmt parser.Statement) error {

	switch s := stmt.(type) {

	case *parser.FunctionCallExpr:

		// We have to loop over the arguments in reverse
		for i := len(s.Arguments) - 1; i >= 0; i-- {
			// push each argument to the stack
			err := c.compileExpr(s.Arguments[i])
			if err != nil {
				return err
			}
			fmt.Fprintf(&c.buff, `
push rax
`)

		}
		fmt.Fprintf(&c.buff, `
call %s
add rsp, %d
`, s.Name, 8*len(s.Arguments))

	case *parser.Function:

		n := c.labelCount
		c.labelCount++

		// Add the name of this function to the end
		// of the list.
		c.functions = append(c.functions, s.Name)

		fmt.Fprintf(&c.buff, `
	jmp function_%d

%s:
`, n, s.Name)

		// new function scope
		c.pushScope()

		// define parameters
		for i, p := range s.Parameters {
			_, err := c.scope.DefineArgument(p.Value.(string), i)
			if err != nil {
				return err
			}
		}

		// crude stack reservation for now
		fmt.Fprintf(&c.buff, `
	push rbp
	mov rbp, rsp
	sub rsp, 8 * 64
`)

		for _, stm := range s.Statements {
			err := c.generateStmt(stm)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&c.buff, `
%s_cleanup:
	mov rsp, rbp
	pop rbp
	ret

function_%d:
`, s.Name, n)

		c.popScope()

		// remove the last function from the list
		c.functions = c.functions[:len(c.functions)-1]

	case *parser.If:

		// Generate a unique label
		n := c.labelCount
		c.labelCount++

		// Compile the expression, masking off strings.
		switch s.Expression.(type) {

		case *parser.StringExpr:
			return fmt.Errorf("'if' only permits a numerical expression")
		default:
			err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}
		txt := fmt.Sprintf(`
	cmp rax, 0
	jz if_%d_end
`, n)
		fmt.Fprint(&c.buff, txt)

		c.pushScope()

		// assemble the body
		for _, st := range s.Statements {
			err := c.generateStmt(st)
			if err != nil {
				return err
			}
		}

		c.popScope()

		txt = fmt.Sprintf(`
if_%d_end:
`, n)
		fmt.Fprint(&c.buff, txt)

	case *parser.Inline:
		fmt.Fprint(&c.buff, "\n"+s.Text+"\n")

	case *parser.Let:

		// Compile the expression, masking off strings.
		// which require special handling.
		switch v := s.Expression.(type) {

		case *parser.StringExpr:
			str := v.Value
			hsh := c.hashString(str)
			c.stringTable[hsh] = str

			txt := fmt.Sprintf(`
       mov rax, offset %s
`, hsh)
			fmt.Fprint(&c.buff, txt)
		default:
			err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}

		_, exists := c.scope.Lookup(s.Name)

		// If we're compiling a function..
		if len(c.functions) > 0 {

			// define local only if it doesn't exist already
			if !exists {
				_, err := c.scope.DefineLocal(s.Name)
				if err != nil {
					return err
				}
			}

		} else {

			if !exists {

				label := c.newGlobalLabel(s.Name)

				g := &GlobalVariable{
					Name:  s.Name,
					Label: label,
				}

				err := c.scope.Define(g)
				if err != nil {
					return err
				}

				c.globalVariables = append(c.globalVariables, g)

			}
		}

		return c.emitStoreVariable(s.Name)

	case *parser.Print:
		for _, item := range s.Values {
			switch v := item.(type) {

			case *parser.StringExpr:
				str := v.Value
				hsh := c.hashString(str)
				c.stringTable[hsh] = str

				txt := fmt.Sprintf(`
	mov rsi, offset %s
	mov rdx, %s_end-%s
	call print_string_with_length

`, hsh, hsh, hsh)
				fmt.Fprint(&c.buff, txt)
			default:
				err := c.compileExpr(v)
				if err != nil {
					return err
				}
				txt := `
	call print_number
`
				fmt.Fprint(&c.buff, txt)

			}
		}

		if s.NewLine {
			fmt.Fprint(&c.buff, `
	call newline
`)
		}
	case *parser.Return:

		// Compile the expression, masking off strings.
		switch s.Expression.(type) {

		case *parser.StringExpr:
			return fmt.Errorf("'return' only permits a numerical expression")
		default:
			err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}

		// If we're compiling a function we don't
		// terminate the program with a return.
		// Instead we just jump to the stack-cleanup
		// code.
		if len(c.functions) > 0 {

			txt := `
	jmp %s_cleanup
`
			fmt.Fprintf(&c.buff, txt, c.functions[len(c.functions)-1])
		} else {

			txt := `
	call exit_with_status
`
			fmt.Fprint(&c.buff, txt)
		}
	case *parser.While:

		n := c.labelCount
		c.labelCount++

		txt := fmt.Sprintf(`
while_%d_start:
`, n)
		fmt.Fprint(&c.buff, txt)

		// Compile the expression, masking off strings.
		switch s.Expression.(type) {

		case *parser.StringExpr:
			return fmt.Errorf("'while' only permits a numerical expression")
		default:
			err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}

		txt = fmt.Sprintf(`
	cmp rax, 0
	jz while_%d_end
`, n)
		fmt.Fprint(&c.buff, txt)

		c.pushScope()

		// assemble the body
		for _, s := range s.Statements {
			err := c.generateStmt(s)
			if err != nil {
				return err
			}
		}

		c.popScope()

		txt = fmt.Sprintf(`
	jmp while_%d_start
while_%d_end:
`, n, n)
		fmt.Fprint(&c.buff, txt)

	default:
		return fmt.Errorf("unhandled token in generateStmt %v", stmt)
	}
	return nil
}
