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

	"github.com/skx/s-lang/check"
	"github.com/skx/s-lang/lexer"
	"github.com/skx/s-lang/parser"
)

// templateFS holds the templates for our prelude/prologue and standard library
//
//go:embed templates/*.tmpl
//go:embed templates/stdlib/*.tmpl
var templateFS embed.FS

// Option defines a config-setting option for use with the compiler-constructor function, New.
//
// We use the decorator-pattern to allow flexible updates for the
// configuration values we allow.
type Option func(*Compiler) error

// WithConstantFolding allows specifying whether constant folding
// is applied after the compiler parses the program, and before it
// generates the assembly language for it.
func WithConstantFolding(enable bool) Option {
	return func(c *Compiler) error {
		c.constantFolding = enable
		return nil
	}
}

// WithSource allows specifying the source code to compile.
func WithSource(source string) Option {
	return func(c *Compiler) error {
		c.Source = source
		return nil
	}
}

// WithCompileChecking allows compile-time type-checking to
// be disabled.
func WithCompileChecking(enable bool) Option {
	return func(c *Compiler) error {
		c.checkTypes = enable
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

	// pragmas stores any pragma values we've received.
	//
	// Pragmas are nothing more than arbitrary "key = value"
	// pairs, which can be used by this compiler to alter
	// its behaviour.
	//
	// Currently we allow defining pragmas to specify the
	// size of values which are retrieved/stored within
	// arrays - we do this because we have no notion of
	// typing so we cannot define "u8", "u16", etc.
	//
	// It might be we'll extend their use further in the
	// future, but that is an open question.
	pragmas map[string]string

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

	// functionName stores the name of functions we're compiling
	//
	// We should only be compiling one function at a time,
	// and we've defined nested functions as being illegal
	// so we can catch them by testing against the value
	// here when we start a new one.
	//
	// We also need to know if we're compiling code within the
	// body of a function so we can handle the correct
	// generation of a "RETURN" statement.
	functionName string

	// knownFunctions keeps track of user-defined functions we know
	// about.  We need this so that we can handle any default arguments
	// which they might have.
	knownFunctions map[string][]*parser.FunctionParameter

	// scope stores stack-frames which are used to hold symbols.
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

	// checkTypes determines if we do compile-time type
	// checking for our standard library
	checkTypes bool

	// typeCheck holds our type checker
	typeCheck *check.Types
}

// New creates a new compiler instance.
//
// Using one of our Option methods source can be passed into the compiler,
// and options set.
func New(options ...Option) (*Compiler, error) {
	tmp := &Compiler{
		constantFolding: true,
		floatTable:      NewFloatTable(),
		knownFunctions:  make(map[string][]*parser.FunctionParameter),
		pragmas:         make(map[string]string),
		scope:           NewScope(nil),
		stringTable:     NewStringTable(),
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
	// checking of standard library types has
	// to be optional, so we can test the run-time
	// type-checking.
	//
	if c.checkTypes {
		c.typeCheck.RegisterStdLib()
	}

	//
	// Create a buffer which will be used to include
	// all of our standard library files - in order
	//
	// Which we can then include in our prologue.
	//
	var buf strings.Builder
	fmt.Fprint(&buf, `{{define "stdlib"}}`)
	entries, err := fs.Glob(templateFS, "templates/stdlib/*.tmpl")
	if err != nil {
		return "", err
	}
	for _, f := range entries {
		fmt.Fprintf(&buf, `{{template "%s" .}}`, filepath.Base(f))
	}
	fmt.Fprint(&buf, `{{end}}`)

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

	// Create a new parser
	parse := parser.New(string(c.Source))

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

	// Helper struct to populate the footer-template.
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

	// Create a concrete instance of the structure
	// above, with things we've created/updated
	// as we've parsed and compiled the user-program.
	vars := &FooterData{
		StringTable: c.stringTable.GetAll(),
		FloatTable:  c.floatTable.GetAll(),
		Globals:     c.globalVariables,
		Data:        c.rawData,
	}

	// Render the footer, which will also include
	// our standard library.
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

// emitLoadVariable emits the code for getting the value from a variable.
//
// The complexity here comes from determining if a variable is local or global.
func (c *Compiler) emitLoadVariable(name string) error {

	_, ok1 := c.knownFunctions[name]
	if ok1 {
		fmt.Fprintf(&c.buff, `
	mov rax, offset %s
`, name)
		return nil
	}

	sym, ok2 := c.scope.Lookup(name)
	if !ok2 {
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

// emitLoadIndex emits the code for "xx[N]".
func (c *Compiler) emitLoadIndex(expr *parser.IndexExpr) error {

	size := "byte"
	width := 1

	// See if we got a size pragma
	val, ok := c.pragmas[expr.Left.String()]
	if ok {
		switch val {
		case "size8":
			size = "byte"
			width = 1
		case "size16":
			size = "word"
			width = 2
		case "size32":
			size = "dword"
			width = 4
		default:
			return fmt.Errorf("unknown value in 'pragm %s %s'",
				expr.Left.String(), val)
		}
	}

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

	extra := ""
	if size == "word" {
		// 2 x index
		extra = "add rax, rax"
	}
	if size == "quad" {
		// 4 x index
		extra = "add rax, rax\n  add rax, rax\n"
	}

	fmt.Fprintf(&c.buff, `
	# index object -> integer
	sar rax, 2

	%s
	# restore base
	pop rbx
`, extra)

	check := ""

	if width == 1 {
		check = `
    cmp rax, rcx
    jae out_of_bounds
`
	} else {
		check = fmt.Sprintf(`
    lea rdx, [rax+%d]
    cmp rdx, rcx
    ja out_of_bounds
`, width)
	}

	//
	// rbx = tagged string
	// rax = integer index
	//

	fmt.Fprintf(&c.buff, `
	# untag string pointer
	and rbx, -4

	# load allocation size
	mov rcx, [rbx-8]

	# bounds check
	%s

	# compute address to read from
	add rbx, rax

	# load value
	xor rax, rax
	movzx rax, %s ptr [rbx]

	# mark as integer
	sal rax, 2
`, check, size)

	return nil
}

// emitStoreIndex generates the code for "x[N] = y"
func (c *Compiler) emitStoreIndex(expr *parser.IndexAssign) error {

	size := "byte"
	width := 1

	// See if we got a size pragma
	val, ok := c.pragmas[expr.Left.String()]
	if ok {
		switch val {
		case "size8":
			size = "byte"
			width = 1
		case "size16":
			size = "word"
			width = 2

		case "size32":
			size = "quad"
			width = 4

		default:
			return fmt.Errorf("unknown value in 'pragm %s %s'",
				expr.Left.String(), val)
		}
	}

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

	check := ""

	if width == 1 {
		check = `
	cmp rbx, rdx
	jae out_of_bounds
`
	} else {
		check = fmt.Sprintf(`
	lea r8, [rbx+%d]
	cmp r8, rdx
	ja out_of_bounds
`, width)
	}

	extra := ""
	register := "al"

	if size == "word" {
		// 2 x index
		extra = "add rax, rax"
		register = "ax"
	}
	if size == "quad" {
		// 4 x index
		extra = "add rax, rax\n  add rax, rax\n"
		register = "rax"
	}

	fmt.Fprintf(&c.buff, `
	# index object -> integer
	sar rax, 2
	%s
	push rax
`, extra)

	// compile value to set
	_, err = c.compileExpr(expr.Expression)
	if err != nil {
		return err
	}

	// rax == value
	// rbx == offset
	// rcx == base

	fmt.Fprintf(&c.buff, `
	pop rbx     # offset (already untagged)
	pop rcx     # base
	and rcx, -4 # untag base
	sar rax, 2  # untag value

	# load allocation size
	mov rdx, [rcx-8]

	# bounds check
	%s

	# compute address to update
	add rbx, rcx

	# save value
	mov %s ptr [rbx], %s

	# return the value
	sal rax, 2
`, check, size, register)

	return nil
}

// emitStoreVariable emits the code for "let name = ..:"
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

// emitFunctionCall expression handles generating a call to a function.
func (c *Compiler) emitFunctionCall(v *parser.FunctionCallExpr) error {

	// Store the types of the functions here
	callTypes := []check.Type{}

	// Confirm this is a known function, so that we
	// can get access to any default parameters it
	// might have defined.
	expected, ok := c.knownFunctions[v.Name]

	// If this is a known (user) function
	//
	// And the number of arguments we've received
	// does not match what we should be given then
	// we're in a situation where default arguments
	// are likely.
	//
	// So we take the default values and pretend they
	// were received.
	//
	if ok && len(expected) != len(v.Arguments) {

		// How many arguments we were given
		found := len(v.Arguments)

		// the maximum number of arguments to supply.
		max := len(expected)

		// Set the default value for each missing argument.
		for found < max {
			v.Arguments = append(v.Arguments, expected[found].Default)
			found++
		}
	}

	// We have to loop over the arguments in reverse
	for i := len(v.Arguments) - 1; i >= 0; i-- {

		// push each argument to the stack
		retType, err := c.compileExpr(v.Arguments[i])
		if err != nil {
			return err
		}
		callTypes = append(callTypes, retType)
		fmt.Fprintf(&c.buff, `
	push rax`)

	}

	// Type checking
	err := c.typeCheck.Check(v.Name, callTypes)
	if err != nil {
		return err
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
	return nil
}

// newGlobalLabel returns a suitable label for the global
// variable named "name".
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
				// Divisions always return a floating-point number
				return &parser.FloatLiteral{
					Value: float64(lI.Value) / float64(rI.Value),
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
				return &parser.FloatLiteral{
					Value: lF.Value / rF.Value,
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

	// left OP right
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
	call plus`)

		case lexer.MINUS:
			fmt.Fprintln(&c.buff, `
	call minus`)

		case lexer.MULTIPLY:
			fmt.Fprintln(&c.buff, `
	call multiply`)

		case lexer.DIVIDE:
			fmt.Fprintln(&c.buff, `
	call divide`)

		case lexer.EQUALS:
			fmt.Fprintln(&c.buff, `
	call equals`)

		case lexer.NOTEQUALS:
			fmt.Fprintln(&c.buff, `
	call not_equals`)

		case lexer.LT:
			fmt.Fprintln(&c.buff, `
	call less_than`)

		case lexer.LTEQUALS:
			fmt.Fprintln(&c.buff, `
	call less_equals`)

		case lexer.GT:
			fmt.Fprintln(&c.buff, `
	call greater_than`)

		case lexer.GTEQUALS:
			fmt.Fprintln(&c.buff, `
	call greater_equals`)

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
		case lexer.MODULUS:
			fmt.Fprintln(&c.buff, `
	call modulus`)
		case lexer.POWER:
			fmt.Fprintln(&c.buff, `
	call power`)

		default:
			return check.UNKNOWN, fmt.Errorf("unhandled BinaryExpr in compileExpr: %v", v)
		}

		// Binary operations will either return an integer, or a float.
		//
		// Since we don't know which it is effectively "unknown".
		return check.UNKNOWN, err

	// F
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

	// foo(..)
	case *parser.FunctionCallExpr:
		err := c.emitFunctionCall(v)
		return check.UNKNOWN, err

	// ptr[index] = value
	case *parser.IndexAssign:
		err := c.emitStoreIndex(v)
		return check.UNKNOWN, err

	// ptr[index]
	case *parser.IndexExpr:
		err := c.emitLoadIndex(v)
		return check.UNKNOWN, err

	// N
	case *parser.IntegerLiteral:

		fmt.Fprintf(&c.buff, `
	mov rax, %d  # mov rax, %d + typing`, v.Value<<2, v.Value)

		return check.INTEGER, nil

	// "str"
	case *parser.StringLiteral:
		str := v.Value
		id := c.stringTable.Add(str)

		txt := fmt.Sprintf(`
	mov rax, offset %s
	or rax, 1   # tagged as a string
`, id)
		fmt.Fprint(&c.buff, txt)

		return check.STRING, nil

	// id
	case *parser.VariableExpr:
		err := c.emitLoadVariable(v.Name)
		return check.UNKNOWN, err

	default:
		return check.UNKNOWN, fmt.Errorf("unhandled token in compileExpr: %v", v)
	}
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

	case *parser.Data:
		c.rawData = append(c.rawData, s.Text)

	case *parser.Function:

		// save this function away, so that it
		// is known and in the future we can
		// determine any default parameters
		// it might have.
		c.knownFunctions[s.Name] = s.Parameters

		// Are we already defining a function?
		//
		// That's illegal.
		if c.functionName != "" {
			return fmt.Errorf("nested functions are illegal, we're inside %s and trying to define %s",
				c.functionName, s.Name)
		}

		//
		// Record the name of the function we're defining
		//
		c.functionName = s.Name

		fmt.Fprintf(&c.buff, `
	# Skip inline function implementation
	jmp over_function_%s
.align 8
%s:`, s.Name, s.Name)

		// new function scope
		c.pushScope()

		// define parameters
		for i, p := range s.Parameters {
			_, err := c.scope.DefineArgument(p.Name, i)
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

		// we're no longer defining a function
		c.functionName = ""

	case *parser.FunctionCallExpr:

		err := c.emitFunctionCall(s)
		return err

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
	# IF condition - value in RAX
	call true
	jz  if_%d_false  # non-zero jump to the else
	jmp if_%d_true   # now we've tested we jump to the true block

if_%d_true:
`, n, n, n)

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

	case *parser.IndexAssign:
		err := c.emitStoreIndex(s)
		if err != nil {
			return err
		}

	case *parser.Inline:
		fmt.Fprint(&c.buff, "\n"+s.Text+"\n")

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

		// Create a label for the value, if necessary
		if c.functionName != "" {

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

	case *parser.Pragma:
		c.pragmas[s.Key] = s.Value

	case *parser.PostfixExpr:
		// get
		err := c.emitLoadVariable(s.Expr.String())
		if err != nil {
			return err
		}
		// mutate
		switch s.Op {
		case "++":
			fmt.Fprint(&c.buff, `
	# ++
	sar rax, 2
	inc rax
	sal rax, 2`)

		case "--":
			fmt.Fprint(&c.buff, `
	# --
	sar rax, 2
	dec rax
	sal rax, 2`)
		default:
			return fmt.Errorf("unknown postfix operation %s", s.Op)
		}
		// set
		err = c.emitStoreVariable(s.Expr.String())
		return err

	case *parser.Return:

		if c.functionName == "" {
			return fmt.Errorf("return can only be used within a function")
		}

		// Compile the expression if present
		//
		// Our return statement is special as it has two forms:
		//
		//    return;
		//    return( EXPR );
		//
		if s.Expression != nil {
			_, err := c.compileExpr(s.Expression)
			if err != nil {
				return err
			}
		}

		// Within a function a return just becomes a jump
		// to the cleanup/function exit.
		txt := `
	# RETURN
	jmp %s_cleanup
`
		fmt.Fprintf(&c.buff, txt, c.functionName)

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
	# WHILE condition - value in RAX
	call true
	jz while_%d_end      # zero?  Then we skip the body

while_%d_body:
`, n, n)
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

	case *parser.Switch:
		//
		// So we allow:
		//   switch a {
		//    case 2 { .. }
		//    case 3 { .. }
		//    default { }
		//
		// We compile the expression, a, and that will give the result in RAX.
		//
		// Then for each one we need to have a jump to the right block,
		// if equal.  And a fall-through to the default
		//
		//

		// create a label
		n := c.labelCount
		c.labelCount++

		fmt.Fprint(&c.buff, `
	# SWITCH `)

		// Generate the expression
		_, err := c.compileExpr(s.Value)
		if err != nil {
			return err
		}

		// Untag the value
		fmt.Fprint(&c.buff, `
	sar rax, 2 `)

		// Now we handle each of the case statements
		// skipping the default
		for i, cas := range s.Choices {

			// skip the default
			if cas.Default {
				continue
			}

			// At the moment we only handle integers
			val, ok := cas.Expression.(*parser.IntegerLiteral)
			if !ok {
				return fmt.Errorf("only integer literals for CASE statements")
			}

			fmt.Fprintf(&c.buff, `
	# CASE %d
	cmp rax, %d
	jz switch_%d_case_%d
`, val.Value, val.Value, n, i)

		}

		// Now we have the jump-tables for each case,
		// if none match we'll fall-through to the default
		// case
		for _, cas := range s.Choices {

			// skip the default
			if !cas.Default {
				continue
			}

			fmt.Fprint(&c.buff, `
	# FALL-THROUGH DEFAULT
`)
			// assemble the body
			for _, s := range cas.Statements {
				err := c.generateStmt(s)
				if err != nil {
					return err
				}
			}
		}

		//
		// Here we've finished the case-jumps
		// and we either have generated a default-case
		// or done nothing if there was none present.
		//
		// Either way we now need to skip over the implementations
		// of the specific, non-default, handlers.
		//
		fmt.Fprintf(&c.buff, `
	jmp switch_%d_end
`, n)

		// OK now we include each body
		// Now we handle each of the case statements
		// skipping the default
		for i, cas := range s.Choices {

			// skip the default
			if cas.Default {
				continue
			}

			fmt.Fprintf(&c.buff, `
switch_%d_case_%d:
`, n, i)
			for _, s := range cas.Statements {
				err := c.generateStmt(s)
				if err != nil {
					return err
				}
			}

			fmt.Fprintf(&c.buff, `
	jmp switch_%d_end
`, n)

		}

		// All over now
		fmt.Fprintf(&c.buff, `
switch_%d_end:
`, n)

	default:
		return fmt.Errorf("unhandled token in generateStmt %v", stmt)

	}

	return nil
}
