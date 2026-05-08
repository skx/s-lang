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
		case lexer.EQUALS:
			fmt.Fprintln(out, `
cmp rbx, rax
sete al
movzx rax, al`)

		case lexer.NOT_EQUALS:
			fmt.Fprintln(out, `
cmp rbx, rax
setne al
movzx rax, al`)

		case lexer.LT:
			fmt.Fprintln(out, `
cmp rbx, rax
setl al
movzx rax, al`)

		case lexer.LT_EQUALS:
			fmt.Fprintln(out, `
cmp rbx, rax
setle al
movzx rax, al`)

		case lexer.GT:
			fmt.Fprintln(out, `
cmp rbx, rax
setg al
movzx rax, al`)

		case lexer.GT_EQUALS:
			fmt.Fprintln(out, `
cmp rbx, rax
setge al
movzx rax, al`)
		}

	}
}

// generateStmt generates the assembly for a single statement, it is moved into
// its own function so that we can call it recursively for cases like "if" and "while".
func (g *generateCommand) generateStmt(out io.Writer, stmt parser.Statement) {

	switch stmt.(type) {

	case *parser.If:
		i := stmt.(*parser.If)

		// Generate a unique label
		n := g.whileCount
		g.whileCount++
		n++

		g.compileExpr(out, i.Expression)

		txt := fmt.Sprintf(`
	cmp rax, 0
	jz if_%d_end
`, n)
		fmt.Fprint(out, txt)

		// assemble the body
		for _, s := range i.Statements {
			g.generateStmt(out, s)
		}
		txt = fmt.Sprintf(`
if_%d_end:
`, n)
		fmt.Fprint(out, txt)

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

		n := g.whileCount
		g.whileCount++
		n++

		txt := fmt.Sprintf(`
while_%d_start:
`, n)
		fmt.Fprint(out, txt)

		g.compileExpr(out, whl.Expression)

		txt = fmt.Sprintf(`
	cmp rax, 0
	jnz while_%d_end
`, n)
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
