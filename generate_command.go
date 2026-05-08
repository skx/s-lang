package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"s-lang/lexer"
	"s-lang/parser"
	"strings"
)

// Structure for our options and state.
type generateCommand struct {

	// output will be the file to generate
	output string

	// hash table for strings to be generated
	stringTable map[string]string

	// whileCount is the count of while-expressions we've received
	whileCount int
}

// Arguments adds per-command args to the object.
func (g *generateCommand) Arguments(f *flag.FlagSet) {
	f.StringVar(&g.output, "output", "", "File to write our generated assembly to, STDOUT if not specified.")
}

// Info returns the name of this subcommand.
func (g *generateCommand) Info() (string, string) {
	return "generate", `Generate an assembly version of the given program.

Details:

This command allows you to generate an assembly version of the given program.

The assembly file is standalone and may then be compiled via 'as' and linked
with 'ld' as usual.  The built-in 'compile' sub-command will run the assembler
and linker for you.

Example:

    $ sysbox generate example.in
    $ sysbox generate -output example.s example.in
`
}

// hashString is used to generate a stable identifier for any string literals within
// the program we generate - this is used primarily to allow string interning.
func (g *generateCommand) hashString(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("msg_%s", sha)
}

// processFile takes the given source file and generates the appropriate assembly
// language from its contents, after parsing.
func (g *generateCommand) processFile(path string) error {

	// create the stringTable we'll need
	g.stringTable = make(map[string]string)

	// Read the file-contents
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Default output destination is the console
	var out io.Writer = os.Stdout

	// If a file was provided, use it instead
	if g.output != "" {
		file, err := os.Create(g.output)
		if err != nil {
			return err
		}
		defer file.Close()
		out = file
	}

	// Create a lexer and parser
	lex := lexer.NewLexer(string(data))
	parse := parser.New(lex)

	// now parse the program
	program, err2 := parse.ParseProgram()
	if err2 != nil {
		return err2
	}

	// Write the header
	g.writeHeader(out)

	// compile each statement
	for _, stmt := range program.Statements {
		g.generateStmt(out, stmt)
	}

	// write the footer
	g.writeFooter(out)

	// write String table
	if len(g.stringTable) > 0 {
		fmt.Fprintf(out, "\n# generated string table\n.section .data\n")
	}
	for k, v := range g.stringTable {
		v = strings.Replace(v, "\n", "\\n", -1)
		fmt.Fprintf(out, "  %s: .ascii \"%s\"\n", k, v)
		fmt.Fprintf(out, "  %s_end:\n", k)
	}
	return nil
}

// compileExpr handles compiling an expression
func (g *generateCommand) compileExpr(out io.Writer, e parser.Expr) {
	switch v := e.(type) {

	case *parser.NumberExpr:
		fmt.Fprintf(out, "mov rax, %d\n", v.Value)

	case *parser.VariableExpr:
		idx := rune(v.Name[0]) - 'a'

		fmt.Fprintf(out, `
lea rcx, vars
mov rax, [rcx + %d*8]
`, idx)

	case *parser.BinaryExpr:

		g.compileExpr(out, v.Left)

		fmt.Fprintln(out, "push rax")

		g.compileExpr(out, v.Right)

		fmt.Fprintln(out, "pop rbx")

		switch v.Op {

		case lexer.PLUS:
			fmt.Fprintln(out, "add rax, rbx")

		case lexer.MINUS:
			fmt.Fprintln(out, "sub rbx, rax")
			fmt.Fprintln(out, "mov rax, rbx")

		case lexer.MULTIPLY:
			fmt.Fprintln(out, "imul rax, rbx")

		case lexer.DIVIDE:
			fmt.Fprintln(out, `
mov rdx, 0
mov rcx, rax
mov rax, rbx
idiv rcx`)
		}

	}
}

// generateStmt generates the assembly for a single statement, it is moved into
// its own function so that we can call it recursively for cases like "if" and "while".
func (g *generateCommand) generateStmt(out io.Writer, stmt parser.Statement) {

	switch stmt.(type) {

	case *parser.If:
		// There are two ways we can run an "IF" test;
		//   if ( a ) { .. }
		//   if ( a OP b ) { .. }
		//
		i := stmt.(*parser.If)
		n := g.whileCount
		g.whileCount++
		n++

		if len(i.Condition) == 1 {

			cnd := i.Condition[0].Value.(string)
			num := rune(cnd[0]) - 'a'

			txt := fmt.Sprintf(`
	lea rcx, vars
	mov rax, [rcx + %d*8]
	cmp rax, 0
	jz if_%d_end
`, num, n)
			fmt.Fprint(out, txt)

			// assemble the body
			for _, s := range i.Statements {
				g.generateStmt(out, s)
			}
			txt = fmt.Sprintf(`
if_%d_end:
`, n)
			fmt.Fprint(out, txt)
		}

		if len(i.Condition) == 3 {

			// We want to handle: A OP B
			// A or B might be numbers or registers
			//
			// To simplify we'll store the values in RAX and RBX
			// then perform the operation before saving

			// RAX = A
			if i.Condition[0].Type == lexer.NUMBER {

				val := int64(i.Condition[0].Value.(float64))
				txt := fmt.Sprintf(`
	mov rax, %d
`, val)
				fmt.Fprint(out, txt)
			} else {

				// Value here is a register read
				src := strings.ToLower(i.Condition[0].Value.(string))
				srcN := rune(src[0]) - 'a'
				txt := fmt.Sprintf(`
	lea rcx, vars
	mov rax, [rcx + %d*8]
`, srcN)
				fmt.Fprint(out, txt)
			}

			// RBX = B
			if i.Condition[2].Type == lexer.NUMBER {

				val := int64(i.Condition[2].Value.(float64))
				txt := fmt.Sprintf(`
	mov rbx, %d
`, val)
				fmt.Fprint(out, txt)
			} else {

				// Value here is a register read
				src := strings.ToLower(i.Condition[2].Value.(string))
				srcN := rune(src[0]) - 'a'
				txt := fmt.Sprintf(`
	lea rcx, vars
	mov rbx, [rcx + %d*8]
`, srcN)
				fmt.Fprint(out, txt)
			}

			// Now we have A in RAX, and B in RCX
			// We need to handle the comparison operation.
			switch i.Condition[1].Value.(string) {
			case "==":
				txt := fmt.Sprintf(`
	cmp rax, rbx
	jnz if_%d_end`, n)
				fmt.Fprint(out, txt)

			case "!=":
				txt := fmt.Sprintf(`
	cmp rax, rbx
	jz if_%d_end`, n)
				fmt.Fprint(out, txt)

			case "<":
				txt := fmt.Sprintf(`
cmp rax, rbx
jge if_%d_end`, n)
				fmt.Fprint(out, txt)
			case "<=":
				txt := fmt.Sprintf(`
cmp rax, rbx
jg if_%d_end`, n)
				fmt.Fprint(out, txt)

			case ">":

				txt := fmt.Sprintf(`
cmp rax, rbx
jle if_%d_end`, n)
				fmt.Fprint(out, txt)

			case ">=":

				txt := fmt.Sprintf(`
cmp rax, rbx
jl if_%d_end`, n)
				fmt.Fprint(out, txt)
			default:
				panic(fmt.Sprintf("unknown condition for if-statement: %s", i.Condition[1].Value.(string)))
			}
			// assemble the body
			for _, s := range i.Statements {
				g.generateStmt(out, s)
			}

			txt := fmt.Sprintf(`
if_%d_end:
`, n)
			fmt.Fprint(out, txt)

		}

	case *parser.LetStatement:

		// We have two kinds of "LET f = XXX" values we handle
		// either XXX is a single thing, be it a register or an integer literal,
		// or it is a simple expression.
		//
		// We'll handle them both here
		l := stmt.(*parser.LetStatement)
		g.compileExpr(out, l.Expression)

		// The value will be in RAX, save it into the appropriate
		// variable
		idx := rune(l.Name[0]) - 'a'
		fmt.Fprintf(out, `
lea rcx, vars
mov [rcx + %d*8], rax
`, idx)

	case *parser.Print:
		prn := stmt.(*parser.Print)
		for _, item := range prn.Values {
			switch item.Type {

			case lexer.STRING:
				str := item.Value.(string)
				hsh := g.hashString(str)
				g.stringTable[hsh] = str

				txt := fmt.Sprintf(`
	mov rsi, offset %s
	mov rdx, %s_end-%s
	call print_string

`, hsh, hsh, hsh)
				fmt.Fprint(out, txt)

			case lexer.IDENT:
				reg := strings.ToLower(item.Value.(string))
				num := rune(reg[0]) - 'a'
				txt := fmt.Sprintf(`
	lea rcx, vars
	mov rax, [rcx + %d*8]
	call print_number`, num)
				fmt.Fprint(out, txt)

			case lexer.NUMBER:
				num := int64(item.Value.(float64))

				txt := fmt.Sprintf(`
	mov rax, %d
	call print_number`, num)
				fmt.Fprint(out, txt)
			default:
				fmt.Printf("Uknown token type %V\n", item.Value)
			}

			if prn.NewLine {
				fmt.Fprint(out, `
	call newline
`)
			}
		}
	case *parser.Return:
		rtn := stmt.(*parser.Return)
		switch rtn.Value.Type {

		case lexer.STRING:
			// TODO
			panic("invalid return value")

		case lexer.IDENT:
			reg := strings.ToLower(rtn.Value.Value.(string))
			num := rune(reg[0]) - 'a'
			txt := fmt.Sprintf(`
	lea rcx, vars
	mov rax, [rcx + %d*8]
	call exit_with_status
`, num)
			fmt.Fprint(out, txt)

		case lexer.NUMBER:
			num := int64(rtn.Value.Value.(float64))

			txt := fmt.Sprintf(`
	mov rax, %d
	call exit_with_status`, num)
			fmt.Fprint(out, txt)
		}

	case *parser.While:
		whl := stmt.(*parser.While)
		cnd := whl.Value.Value.(string)
		num := rune(cnd[0]) - 'a'

		n := g.whileCount
		g.whileCount++
		n++
		txt := fmt.Sprintf(`
while_%d_start:
	lea rcx, vars
	mov rax, [rcx + %d*8]
	cmp rax, 0
	jz while_%d_end
`, n, num, n)
		fmt.Fprint(out, txt)

		// assemble the body
		for _, s := range whl.Statements {
			g.generateStmt(out, s)
		}
		txt = fmt.Sprintf(`
	jmp while_%d_start
while_%d_end:
`, n, n)
		fmt.Fprint(out, txt)

		//

	default:
		fmt.Printf("Uknown %T %v\n", stmt, stmt)
	}
}

// writeFooter generates the footer for our generated assembly language, which is
// a combination of our "standard library" routines, and a string-table of any string
// literals within the source file we're compiling.
func (g *generateCommand) writeFooter(f io.Writer) {
	text := `

	#
	# Exit explicitly, just in case the program was missing a
	# terminating RETURN statement.
	#
	xor rax, rax
	jp exit_with_status



	#
	# Write a newline to STDOUT.
	#
	# Uses the "print_rax_buffer" as temporary
	# storage, and will trash it.
	#
newline:
	mov rdx, 1 # length
	lea rsi, print_rax_buffer
	mov byte ptr [rsi], '\n'
print_string:
	mov rax, 1 # write
	mov rdi, 1 # STDOUT
	syscall
	ret



	#
	# Exit with the given status.
	#
	# Status code to use is stored in RAX
	#
exit_with_status:
	mov rdi, rax
	mov rax, 60     # sys_exit
	syscall
	ret



	#
	# Convert the integer in RAX into
	# ASCII and print it to the console.
	#
	# Uses the "print_rax_buffer" as temporary
	# storage, and will trash it.
	#
print_number:
	mov rbx, 10
	lea rdi, [print_rax_buffer+31]
	mov byte ptr [rdi], 0
	dec rdi

	.convert_loop:
	xor rdx, rdx
	div rbx
	add dl, '0'
	mov [rdi], dl
	dec rdi
	test rax, rax
	jnz .convert_loop

	inc rdi              # rdi = pointer to string start

	# compute length in rdx
	lea rsi, [print_rax_buffer+31]
	mov rdx, rsi
	sub rdx, rdi

	mov rax, 1           # sys_write
	mov rsi, rdi         # buffer
	mov rdi, 1           # stdout
	syscall

	ret

`
	fmt.Fprintln(f, text)
}

// writeHeader generates the prolog which is prepended to the generated
// assembly language code.
func (g *generateCommand) writeHeader(f io.Writer) {
	text := `
	# Define our entry-point
	.global _start

	# Declare our library-functions
	.global print_number
	.global print_string
	.global newline
	.global exit_with_status


	# Writeable data storage
	.section .bss

	# Buffer for printing numbers, and newlines
print_rax_buffer:
	.skip 32

	# storage for our variables
vars:
	.skip 26 * 8

	#
	# Code
	#
	.section .text
_start:
`
	fmt.Fprintln(f, text)
}

// Execute is invoked if the user specifies `generate` as the subcommand.
func (g *generateCommand) Execute(args []string) int {

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: s-lang generate [options] source.in\n")
		return 1
	}

	err := g.processFile(args[0])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return 1
	}
	return 0
}
