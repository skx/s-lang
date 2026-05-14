// Package compiler implements our compiler.
//
// The compiler makes use of our parser and lexer packages
// and largely walks trees generating snippets of assembly
// language as it goes.
package compiler

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"s-lang/lexer"
	"s-lang/parser"
)

// templateFS holds the templates for our prelude/prologue and standard library
//
//go:embed templates/*.tmpl
//go:embed templates/stdlib/*.tmpl
var templateFS embed.FS

// Option defines a config-setting option for our constructor.
//
// We use the decorator-pattern to allow flexible updates for the
// configuration values we allow.
type Option func(*Compiler) error

// WithConstantFolding allows specifying whether constant folding
// is applied
func WithConstantFolding(enable bool) Option {
	return func(c *Compiler) error {
		c.constantFolding = enable
		return nil
	}
}

// WithSource allows specifying the source code to compile
func WithSource(source string) Option {
	return func(c *Compiler) error {
		c.Source = source
		return nil
	}
}

// Compiler holds our internal compiler state.
type Compiler struct {

	// Source holds the source program we'll work on.
	Source string

	// buff is the writer object we use to send our
	// output to, as it is generated.
	buff bytes.Buffer

	// labelCount is used for generating unique labels,
	// these are used when compiling "if" and "while"
	// statements.
	labelCount int

	// inWhile is incremented every time we enter a new
	// while-scope.  This is required because a BREAK
	// or CONTINUE statement is only valid inside such
	// a loop.
	whiles []int

	// stringTable holds (interned) strings
	//
	// To ensure that each string has a stable and safe
	// name we actually hash the string-contents and
	// refer to them by that; this has the side-effect
	// of providing interning - the same string might be
	// defined/used multiple times, but only appear
	// within the source code we generate a single time.
	stringTable *StringTable

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

	// rawData stores raw data from `data { .. }` blocks
	// inside the programs.
	// We save them so that we can generate them, in order,
	// at the end of our file.
	rawData []string

	// globalCount stores the count of global variables.
	globalCount int

	// constantFolding determines whether we try to
	// optimize our AST, folding constants, before we
	// generate code
	constantFolding bool
}

// New creates a new compiler instance, to compile the
// given program.
func New(options ...Option) (*Compiler, error) {
	tmp := &Compiler{
		stringTable:     NewStringTable(),
		scope:           NewScope(nil),
		constantFolding: true,
	}

	// Allow options to override our defaults
	for _, option := range options {
		err := option(tmp)
		if err != nil {
			return tmp, err
		}
	}

	return tmp, nil
}

// Compile produces, and returns, an assembly language
// implementation of the program which was passed to
// our constructor.
func (c *Compiler) Compile() (string, error) {

	//
	// Create a buffer which will be used to include
	// all of our standard library files - in order
	//
	// Which we can then include in our prologue.
	//
	var buf strings.Builder
	buf.WriteString(`{{define "stdlib"}}`)
	entries, err := fs.Glob(templateFS, "templates/stdlib/*.tmpl")
	if err != nil {
		return "", err
	}
	for _, f := range entries {
		buf.WriteString(fmt.Sprintf(`{{template "%s" .}}`, filepath.Base(f)))
	}
	buf.WriteString(`{{end}}`)

	//
	// Load all our templates
	//
	tmpl := template.Must(
		template.New("stdlib").
			ParseFS(templateFS, "templates/*.tmpl", "templates/stdlib/*.tmpl"),
	)

	//
	// Ensure we have the generated template too,
	// which defines the "stdlib" inclusion text
	//
	tmpl = template.Must(tmpl.Parse(buf.String()))

	// Create a lexer and parser
	lex := lexer.NewLexer(string(c.Source))
	parse := parser.New(lex)

	// now parse the program
	program, err := parse.ParseProgram()
	if err != nil {
		return "", err
	}

	// Render the header into our output
	err = tmpl.ExecuteTemplate(&c.buff, "header.tmpl", nil)
	if err != nil {
		return "", err
	}

	// compile each statement
	for _, stmt := range program.Statements {
		err = c.generateStmt(stmt)
		if err != nil {
			return "", err
		}
	}

	// Helper struct
	type FooterData struct {

		// String data for the template rendering
		StringTable []StringEntry

		// GlobalVars has global variable storage
		Globals []*GlobalVariable

		// Data holds raw data for the file footer
		Data []string
	}

	vars := &FooterData{
		StringTable: c.stringTable.GetAll(),
		Globals:     c.globalVariables,
		Data:        c.rawData,
	}

	// Render the footer, which will also include
	// our standard library
	err = tmpl.ExecuteTemplate(&c.buff, "footer.tmpl", vars)
	if err != nil {
		return "", err
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

// newGlobalLabel returns a suitable label for the global
// variable named "name".
//
// TODO: The numeric suffix is probably not necessary.
func (c *Compiler) newGlobalLabel(name string) string {
	lbl := fmt.Sprintf("global_%s_%d", name, c.globalCount)
	c.globalCount++
	return lbl
}

// optimizeExpr optimizes constant expressions.
func (c *Compiler) optimizeExpr(expr parser.Expr) parser.Expr {
	switch v := expr.(type) {

	case *parser.BinaryExpr:
		// First recursively fold children
		v.Left = c.optimizeExpr(v.Left)
		v.Right = c.optimizeExpr(v.Right)

		// Check if both sides are now integers
		l, ok1 := v.Left.(*parser.IntegerExpr)
		r, ok2 := v.Right.(*parser.IntegerExpr)

		if ok1 && ok2 {
			switch v.Op {

			case lexer.PLUS:
				return &parser.IntegerExpr{
					Value: l.Value + r.Value,
				}

			case lexer.MINUS:
				return &parser.IntegerExpr{
					Value: l.Value - r.Value,
				}

			case lexer.MULTIPLY:
				return &parser.IntegerExpr{
					Value: l.Value * r.Value,
				}

			case lexer.DIVIDE:
				if r.Value != 0 {
					return &parser.IntegerExpr{
						Value: l.Value / r.Value,
					}
				}
			}
		}

		return v

	default:
		return expr
	}
}

// compileExpr handles compiling an expression.
func (c *Compiler) compileExpr(e parser.Expr) error {

	// Convert simple expressions to their results
	if c.constantFolding {
		e = c.optimizeExpr(e)
	}

	switch v := e.(type) {

	case *parser.IntegerExpr:
		// This is an integer
		fmt.Fprintf(&c.buff, `
	mov rax, %d  # mov rax, %d + typing`, v.Value<<2, v.Value)

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
	mov rax, %d   # ABI: RAX contains argument count
	call %s
	add rsp, %d
`, len(v.Arguments), v.Name, 8*len(v.Arguments))

	case *parser.VariableExpr:
		return c.emitLoadVariable(v.Name)

	case *parser.StringExpr:
		str := v.Value
		id := c.stringTable.Add(str)

		txt := fmt.Sprintf(`
	mov rax, offset %s
	or rax, 1   # tagged as a string
`, id)
		fmt.Fprint(&c.buff, txt)

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
	# +
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	add rax, rbx
	shl rax, 2  # add typing`)

		case lexer.MINUS:
			fmt.Fprintln(&c.buff, `
	# -
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	sub rbx, rax
	mov rax, rbx
	shl rax, 2  # add typing`)

		case lexer.MULTIPLY:
			fmt.Fprintln(&c.buff, `
	# *
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	imul rax, rbx
	shl rax, 2  # add typing`)

		case lexer.DIVIDE:
			fmt.Fprintln(&c.buff, `
	# /
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	mov rdx, 0
	mov rcx, rax
	mov rax, rbx
	idiv rcx
	shl rax, 2  # add typing`)
		case lexer.EQUALS:
			fmt.Fprintln(&c.buff, `
	# ==
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	sete al
	movzx rax, al
	shl rax, 2  # add typing`)

		case lexer.NOT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# !=
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	setne al
	movzx rax, al
	shl rax, 2  # add typing`)

		case lexer.LT:
			fmt.Fprintln(&c.buff, `
	# <
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	setl al
	movzx rax, al
	shl rax, 2  # add typing`)

		case lexer.LT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# <=
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	setle al
	movzx rax, al
	shl rax, 2  # add typing`)

		case lexer.GT:
			fmt.Fprintln(&c.buff, `
	# >
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	setg al
	movzx rax, al
	shl rax, 2  # add typing`)

		case lexer.GT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# >=
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	cmp rbx, rax
	setge al
	movzx rax, al
	shl rax, 2  # add typing`)
		case lexer.AND:
			fmt.Fprintln(&c.buff, `
	# &&
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	imul rax, rbx
	shl rax, 2  # add typing`)
		case lexer.OR:
			fmt.Fprintln(&c.buff, `
	# ||
	sar rax, 2  # undo the type-storage
	sar rbx, 2  # undo the type-storage
	or rax, rbx
	cmp rax, 0
	setne al
	movzx rax, al
	shl rax, 2  # add typing`)
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

	case *parser.Break:
		if len(c.whiles) == 0 {
			return fmt.Errorf("BREAK outside WHILE")
		}
		label := c.whiles[len(c.whiles)-1]

		// Jump to end of the while-loop
		txt := fmt.Sprintf(`
	# BREAK
	jmp while_%d_end
`, label)
		fmt.Fprint(&c.buff, txt)

	case *parser.Continue:
		if len(c.whiles) == 0 {
			return fmt.Errorf("CONTINUE outside WHILE")
		}

		label := c.whiles[len(c.whiles)-1]
		txt := fmt.Sprintf(`
	# CONTINUE
	jmp while_%d_start
`, label)
		fmt.Fprint(&c.buff, txt)

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
	mov rax, %d   # ABI: RAX contains argument count
	call %s
	add rsp, %d
`, len(s.Arguments), s.Name, 8*len(s.Arguments))

	case *parser.Function:

		// Add the name of this function to the end
		// of the list.
		c.functions = append(c.functions, s.Name)

		fmt.Fprintf(&c.buff, `
	# Skip inline function implementation
	jmp over_function_%s
%s:`, s.Name, s.Name)

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
		fmt.Fprint(&c.buff, `
	push rbp
	mov rbp, rsp
	sub rsp, 8 * 64  # Space for local variables
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

over_function_%s:
`, s.Name, s.Name)

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
	sar rax, 2  # undo the type-storage
	cmp rax, 0
	jz if_%d_false
`, n)
		fmt.Fprint(&c.buff, txt)

		c.pushScope()

		// assemble the body
		for _, st := range s.True {
			err := c.generateStmt(st)
			if err != nil {
				return err
			}
		}
		txt = fmt.Sprintf(`
	jmp if_%d_end
`, n)
		fmt.Fprint(&c.buff, txt)

		c.popScope()

		// else - might be empty
		txt = fmt.Sprintf(`
if_%d_false:
`, n)
		fmt.Fprint(&c.buff, txt)
		c.pushScope()

		if len(s.False) > 0 {
			// assemble the body
			for _, st := range s.False {
				err := c.generateStmt(st)
				if err != nil {
					return err
				}
			}
		}
		c.popScope()

		txt = fmt.Sprintf(`
if_%d_end:
`, n)
		fmt.Fprint(&c.buff, txt)

	case *parser.Inline:
		fmt.Fprint(&c.buff, "\n"+s.Text+"\n")

	case *parser.Data:
		c.rawData = append(c.rawData, s.Text)

	case *parser.Let:

		// Compile the expression, leaving
		// the result in RAX
		err := c.compileExpr(s.Expression)
		if err != nil {
			return err
		}

		_, exists := c.scope.Lookup(s.Name)

		// Create a lable for the value, if necessary
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

		// Actually store RAX into the value
		err = c.emitStoreVariable(s.Name)
		if err != nil {
			return err
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
	# RETURN
	jmp %s_cleanup
`
			fmt.Fprintf(&c.buff, txt, c.functions[len(c.functions)-1])
		} else {

			txt := `
	call exit
`
			fmt.Fprint(&c.buff, txt)
		}
	case *parser.While:

		n := c.labelCount
		c.labelCount++

		// record the number of this while statement
		c.whiles = append(c.whiles, n)

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

		// remove the number from the most recent while-list
		c.whiles = c.whiles[:len(c.whiles)-1]

	default:
		return fmt.Errorf("unhandled token in generateStmt %v", stmt)
	}
	return nil
}
