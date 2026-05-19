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

	"s-lang/check"
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

	// buff is the writer object we send all generated
	// assembly code to - as well as the static header,
	// footer, and standard library code.
	//
	// We use a handle so we may easily have the output
	// sent to an actual file, or STDOUT.
	buff bytes.Buffer

	// labelCount is used for generating unique labels,
	// these are used when compiling "if" and "while"
	// statements.
	labelCount int

	// whiles is updated every time we enter a new
	// while-scope.
	//
	// We need this because BREAK and CONTINUE statements
	// are only valid inside such a loop, and so we want
	// to do two things: 1.  Check we're inside one,
	// 2. Know _which_ one we're dealing with.
	//
	whiles []int

	// stringTable holds a table of all the (interned)
	// static strings we've encountered while parsing.
	//
	// To ensure that each string has a stable and safe
	// name we actually hash the string-contents and
	// refer to them by that; this has the side-effect
	// of providing interning - the same string might be
	// defined/used multiple times, but only appear
	// within the source code we generate a single time.
	stringTable *StringTable

	// floatTable serves basically the same purpose as
	// our string table.
	//
	// We intern literal floats, and generate a label
	// to hold them in our generated source.  This is
	// necessary because you cannot write "mov rax, 3.1",
	// instead you store the value in a memory location
	// and then run "mov rax, [float_name]".
	floatTable *FloatTable

	// functions stores the name of functions we're compiling
	// We should only be compiling one function at a time,
	// but this way we could allow nested functions if we
	// wanted to.
	//
	// The most recent function we've entered is the LAST
	// item on the array.
	//
	// We need to know if we're compiling code within the
	// body of a function so we can handle the correct
	// generation of a "RETURN" statement.
	functions []string

	// scope stores stack-frames which are used to hold
	// symbols.
	scope *Scope

	// globalVariables _should_ use the same stack frame,
	// however for quickness they are here.
	globalVariables []*GlobalVariable

	// rawData stores raw data from `data { .. }` blocks
	// inside the programs.
	//
	// We save them here as we encounter them, and then
	// when all our compilation is complete we can insert
	// them at the very end of the program.
	rawData []string

	// globalCount stores the count of global variables.
	globalCount int

	// constantFolding determines whether we try to
	// optimize our AST, folding constants, before we
	// generate code
	constantFolding bool

	// typeCheck holds our type checker
	typeCheck *check.Types
}

// New creates a new compiler instance, to compile the
// given program.
func New(options ...Option) (*Compiler, error) {
	tmp := &Compiler{
		stringTable:     NewStringTable(),
		floatTable:      NewFloatTable(),
		scope:           NewScope(nil),
		constantFolding: true,
		typeCheck:       check.New(),
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

		// Float data for the template rendering
		FloatTable []FloatEntry

		// GlobalVars has global variable storage
		Globals []*GlobalVariable

		// Data holds raw data for the file footer
		Data []string
	}

	vars := &FooterData{
		StringTable: c.stringTable.GetAll(),
		FloatTable:  c.floatTable.GetAll(),
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

func (c *Compiler) emitLoadIndex(expr *parser.IndexExpr) error {

	// Compile base expression
	_, err := c.compileExpr(expr.Left)
	if err != nil {
		return err
	}

	fmt.Fprint(&c.buff, `
	# save base object
	push rax
`)

	// Compile index expression
	_, err = c.compileExpr(expr.Index)
	if err != nil {
		return err
	}

	fmt.Fprint(&c.buff, `
	# index object -> integer
	sar rax, 2

	# restore base
	pop rbx
`)

	//
	// rbx = tagged string
	// rax = integer index
	//

	fmt.Fprint(&c.buff, `
	# untag string pointer
	and rbx, -4

	# compute address of character
	add rbx, rax

	# load byte
	movzx rax, byte ptr [rbx]

	# allocate boxed integer result
	and rax, 255
	sal rax, 2
`)

	return nil
}

func (c *Compiler) emitStoreIndex(expr *parser.IndexAssign) error {

	// Compile base expression
	_, err := c.compileExpr(expr.Left)
	if err != nil {
		return err
	}

	fmt.Fprint(&c.buff, `
	# save base object
	push rax
`)

	// Compile index expression
	_, err = c.compileExpr(expr.Index)
	if err != nil {
		return err
	}

	fmt.Fprint(&c.buff, `
	# index object -> integer
	sar rax, 2
	push rax
`)

	// compile value to set
	_, err = c.compileExpr(expr.Expression)
	if err != nil {
		return err
	}

	// rax == value
	// rbx == offset
	// rcx == base

	fmt.Fprint(&c.buff, `
	pop rbx     # offset (already untagged)
	pop rcx     # base
	and rcx, -4 # untag base
	sar rax, 2  # untag value

	# compute address of character
	add rbx, rcx

	# load byte
	mov byte ptr [rbx], al

	# return the value
	sal rax, 2
`)

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
		lI, okI1 := v.Left.(*parser.IntegerLiteral)
		rI, okI2 := v.Right.(*parser.IntegerLiteral)

		if okI1 && okI2 {
			switch v.Op {

			case lexer.PLUS:
				return &parser.IntegerLiteral{
					Value: lI.Value + rI.Value,
				}

			case lexer.MINUS:
				return &parser.IntegerLiteral{
					Value: lI.Value - rI.Value,
				}

			case lexer.MULTIPLY:
				return &parser.IntegerLiteral{
					Value: lI.Value * rI.Value,
				}

			case lexer.DIVIDE:
				if rI.Value != 0 {
					return &parser.IntegerLiteral{
						Value: lI.Value / rI.Value,
					}
				}
			}

			return v
		}

		// Check if both sides are now floats
		lF, okF1 := v.Left.(*parser.FloatLiteral)
		rF, okF2 := v.Right.(*parser.FloatLiteral)

		if okF1 && okF2 {
			switch v.Op {

			case lexer.PLUS:
				return &parser.FloatLiteral{
					Value: lF.Value + rF.Value,
				}

			case lexer.MINUS:
				return &parser.FloatLiteral{
					Value: lF.Value - rF.Value,
				}

			case lexer.MULTIPLY:
				return &parser.FloatLiteral{
					Value: lF.Value * rF.Value,
				}

			case lexer.DIVIDE:
				if rF.Value != 0 {
					return &parser.FloatLiteral{
						Value: lF.Value / rF.Value,
					}
				}
			}
			return v
		}
	}
	return expr
}

// compileExpr handles compiling an expression.
//
// Where the type of the expression can be known, statically, it is returned.
// For example compiling "3" will always return an INTEGER, and compiling
// a string literal will return STRING.   These return types are used for
// performing as much type and argument checking as is possible at compile-time.
func (c *Compiler) compileExpr(e parser.Expr) (check.Type, error) {

	// Convert simple expressions to their results
	if c.constantFolding {
		e = c.optimizeExpr(e)
	}

	switch v := e.(type) {

	// get str[index]
	case *parser.IndexExpr:
		err := c.emitLoadIndex(v)
		return check.UNKNOWN, err

	// str[index] = value
	case *parser.IndexAssign:
		err := c.emitStoreIndex(v)
		return check.UNKNOWN, err

	case *parser.IntegerLiteral:

		fmt.Fprintf(&c.buff, `
	mov rax, %d  # mov rax, %d + typing`, v.Value<<2, v.Value)

		return check.INTEGER, nil

	case *parser.FloatLiteral:

		id := c.floatTable.Add(v.Value)

		fmt.Fprintf(&c.buff, `
	# Float literal %f

	# allocate 8-byte boxed float
	call alloc8

	# load float constant
	movsd xmm0, [%s]

	# store payload
	movsd [rax], xmm0

	# tag pointer as float (10)
	or rax, 2
`, v.Value, id)
		return check.FLOAT, nil

	case *parser.FunctionCallExpr:
		// Store the types of the functions here
		callTypes := []check.Type{}

		// We have to loop over the arguments in reverse
		for i := len(v.Arguments) - 1; i >= 0; i-- {

			// push each argument to the stack
			retType, err := c.compileExpr(v.Arguments[i])
			if err != nil {
				return check.UNKNOWN, err
			}
			callTypes = append(callTypes, retType)
			fmt.Fprintf(&c.buff, `
	push rax
`)

		}

		// Type checking
		err := c.typeCheck.Check(v.Name, callTypes)
		if err != nil {
			return check.UNKNOWN, err
		}

		fmt.Fprintf(&c.buff, `
	mov rax, %d   # ABI: RAX contains argument count
	call %s
`, len(v.Arguments), v.Name)

		if len(v.Arguments) > 0 {
			fmt.Fprintf(&c.buff, `
	add rsp, %d
`,
				8*len(v.Arguments))
		}

		return check.UNKNOWN, nil

	case *parser.VariableExpr:
		err := c.emitLoadVariable(v.Name)
		return check.UNKNOWN, err

	case *parser.StringLiteral:
		str := v.Value
		id := c.stringTable.Add(str)

		txt := fmt.Sprintf(`
	mov rax, offset %s
	or rax, 1   # tagged as a string
`, id)
		fmt.Fprint(&c.buff, txt)

		return check.STRING, nil

	case *parser.BinaryExpr:

		_, err := c.compileExpr(v.Left)
		if err != nil {
			return check.UNKNOWN, err
		}

		fmt.Fprintln(&c.buff, `
	push rax`)

		_, err = c.compileExpr(v.Right)
		if err != nil {
			return check.UNKNOWN, err
		}

		fmt.Fprintln(&c.buff, `
	pop rbx`)

		switch v.Op {

		case lexer.PLUS:
			fmt.Fprintln(&c.buff, `
	# +
	sar rax, 2   # untag type
	sar rbx, 2   # untag type

	add rax, rbx # compute

	sal rax, 2   # add typing`)

		case lexer.MINUS:
			fmt.Fprintln(&c.buff, `
	# -
	sar rax, 2   # untag type
	sar rbx, 2   # untag type

	sub rbx, rax # compute
	mov rax, rbx
	sal rax, 2   # add typing`)

		case lexer.MULTIPLY:
			fmt.Fprintln(&c.buff, `
	# *
	sar rax, 2     # untag type
	sar rbx, 2     # untag type
	imul rax, rbx  # compute
	sal rax, 2     # add typing`)

		case lexer.DIVIDE:
			fmt.Fprintln(&c.buff, `
	# /
	sar rax, 2    # untag type
	sar rbx, 2    # untag type

	mov rcx, rax  # compute
	mov rax, rbx
	xor rdx, rdx
	idiv rcx

	sal rax, 2    # add typing`)
		case lexer.EQUALS:
			fmt.Fprintln(&c.buff, `
	# ==
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax  # compute
	sete al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.NOT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# !=
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax  # compute
	setne al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.LT:
			fmt.Fprintln(&c.buff, `
	# <
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax
	setl al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.LT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# <=
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax
	setle al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.GT:
			fmt.Fprintln(&c.buff, `
	# >
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax
	setg al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.GT_EQUALS:
			fmt.Fprintln(&c.buff, `
	# >=
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	cmp rbx, rax
	setge al
	movzx rax, al
	sal rax, 2    # add type`)

		case lexer.AND:
			fmt.Fprintln(&c.buff, `
	# &&
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	and rax, rbx
	cmp rax, 0
	setne al
	movzx rax, al
	sal rax, 2    # add type`)
		case lexer.OR:
			fmt.Fprintln(&c.buff, `
	# ||
	sar rax, 2    # untag type
	sar rbx, 2    # untag type
	or rax, rbx
	cmp rax, 0
	setne al
	movzx rax, al
	sal rax, 2    # add type`)
		default:
			return check.UNKNOWN, fmt.Errorf("unhandled token in compileExpr: %v", v)
		}
		return check.INTEGER, err

	}
	return check.UNKNOWN, nil

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

		// Store the types of the functions here
		callTypes := []check.Type{}

		// We have to loop over the arguments in reverse
		for i := len(s.Arguments) - 1; i >= 0; i-- {

			// push each argument to the stack
			retType, err := c.compileExpr(s.Arguments[i])
			if err != nil {
				return err
			}
			callTypes = append(callTypes, retType)
			fmt.Fprintf(&c.buff, `
	push rax
`)
		}

		// Type checking
		err := c.typeCheck.Check(s.Name, callTypes)
		if err != nil {
			return err
		}

		fmt.Fprintf(&c.buff, `
	mov rax, %d   # ABI: RAX contains argument count
	call %s
`, len(s.Arguments), s.Name)

		if len(s.Arguments) > 0 {
			fmt.Fprintf(&c.buff, `
	add rsp, %d
`,
				8*len(s.Arguments))
		}

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

		// We can record the count of function arguments
		// here, to allow later checking.
		c.typeCheck.AddUserFunction(s.Name, len(s.Parameters))

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

		case *parser.StringLiteral:
			return fmt.Errorf("'if' only permits a numerical expression")
		default:
			_, err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}
		txt := fmt.Sprintf(`
	# IF condition
	sar rax, 2      # untag
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

	case *parser.IndexAssign:
		c.emitStoreIndex(s)

	case *parser.Data:
		c.rawData = append(c.rawData, s.Text)

	case *parser.Let:

		// Compile the expression, leaving
		// the result in RAX
		_, err := c.compileExpr(s.Expression)
		if err != nil {
			return err
		}
		nm := ""
		switch s.Left.(type) {
		case *parser.VariableExpr:
			nm = s.Left.(*parser.VariableExpr).Name
		default:
			return fmt.Errorf("cannot assign to %T", s.Left)
		}

		_, exists := c.scope.Lookup(nm)

		// Create a lable for the value, if necessary
		if len(c.functions) > 0 {

			// define local only if it doesn't exist already
			if !exists {
				_, err = c.scope.DefineLocal(nm)
				if err != nil {
					return err
				}
			}

		} else {

			if !exists {

				label := c.newGlobalLabel(nm)

				g := &GlobalVariable{
					Name:  nm,
					Label: label,
				}

				err = c.scope.Define(g)
				if err != nil {
					return err
				}

				c.globalVariables = append(c.globalVariables, g)

			}
		}

		// Actually store RAX into the value
		err = c.emitStoreVariable(nm)
		if err != nil {
			return err
		}
	case *parser.Return:

		// Compile the expression, masking off strings.
		switch s.Expression.(type) {

		case *parser.StringLiteral:
			return fmt.Errorf("'return' only permits a numerical expression")
		default:
			_, err := c.compileExpr(s.Expression)
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

		case *parser.StringLiteral:
			return fmt.Errorf("'while' only permits a numerical expression")
		default:
			_, err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}

		txt = fmt.Sprintf(`
	# WHILE condition
	sar rax, 2            # remove type
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
