package compiler

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"s-lang/lexer"
	"s-lang/parser"
)

// Compiler holds our compiler state
type Compiler struct {

	// Source holds the source program we'll work on
	Source string

	// buff is where we write out data to
	buff bytes.Buffer

	// labelCount is used for generating labels
	labelCount int

	// stringTable holds (interned) strings
	stringTable map[string]string
}

// New creates a new compiler instance, to compile the
// given program.
func New(source string) *Compiler {
	tmp := &Compiler{
		Source:      source,
		stringTable: make(map[string]string),
	}

	return tmp
}

// Compiler produces the compiled string
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
	c.writeHeader(&c.buff)

	// compile each statement
	for _, stmt := range program.Statements {
		err := c.generateStmt(stmt)
		if err != nil {
			return "", err
		}
	}

	// write the footer
	c.writeFooter(&c.buff)

	// write String table
	if len(c.stringTable) > 0 {
		fmt.Fprintf(&c.buff, "\n# generated string table\n.section .data\n")
	}
	for k, v := range c.stringTable {
		v = strings.Replace(v, "\n", "\\n", -1)
		fmt.Fprintf(&c.buff, "  %s: .ascii \"%s\"\n", k, v)
		fmt.Fprintf(&c.buff, ".byte 00  # null byte at end of string\n")
		fmt.Fprintf(&c.buff, "  %s_end:\n", k)
	}

	return c.buff.String(), nil
}

// hashString is used to generate a stable identifier for any string literals within
// the program we generate - this is used primarily to allow string interning.
func (c *Compiler) hashString(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("msg_%s", sha)
}

// compileExpr handles compiling an expression
func (c *Compiler) compileExpr(e parser.Expr) error {
	switch v := e.(type) {

	case *parser.IntegerExpr:
		fmt.Fprintf(&c.buff, `
mov rax, %d`, v.Value)

	case *parser.VariableExpr:
		idx := rune(v.Name[0]) - 'a'

		fmt.Fprintf(&c.buff, `
lea rcx, vars
mov rax, [rcx + %d*8]
`, idx)

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

		// assemble the body
		for _, st := range s.Statements {
			err := c.generateStmt(st)
			if err != nil {
				return err
			}
		}
		txt = fmt.Sprintf(`
if_%d_end:
`, n)
		fmt.Fprint(&c.buff, txt)

	case *parser.Inline:
		fmt.Fprint(&c.buff, "\n"+s.Text+"\n")

	case *parser.Let:

		// Compile the expression, masking off strings.
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

		// The value will be in RAX, save it into the appropriate
		// variable
		idx := rune(s.Name[0]) - 'a'
		fmt.Fprintf(&c.buff, `
lea rcx, vars
mov [rcx + %d*8], rax
`, idx)

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

		txt := `
	call exit_with_status
`
		fmt.Fprint(&c.buff, txt)

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

		// assemble the body
		for _, s := range s.Statements {
			err := c.generateStmt(s)
			if err != nil {
				return err
			}
		}
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

// writeFooter generates the footer for our generated assembly language, which is
// a combination of our "standard library" routines, and a string-table of any string
// literals within the source file we're compiling.
func (c *Compiler) writeFooter(f io.Writer) {
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

	#
	# RSI should point to the start of the string
	#
	# RDX should have the length.
	#
print_string_with_length:
	mov rax, 1 # write
	mov rdi, 1 # STDOUT
	syscall
	ret


	#
	# Print the NULL-terminated string pointed to by RAX.
	#
	# The length will be dynamically discovered
print_string:
	mov rbx, rax

.loop:
	mov al, [rbx]         # load one byte
	test al, al           # is it zero?
	jz .done

	mov rax, 1            # sys_write
	mov rdi, 1            # stdout
	mov rsi, rbx          # pointer to char
	mov rdx, 1            # length = 1
	syscall

	inc rbx               # next character
	jmp .loop
.done:
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
# Clobbers: rax, rbx, rcx, rdx, rdi, rsi
#
print_number:
	mov rbx, 10
	lea rdi, [print_rax_buffer+31]
	mov byte ptr [rdi], 0
	dec rdi

	# Track sign in rcx
	xor rcx, rcx

	# If negative:
	test rax, rax
	jns .convert_loop_start

	neg rax
	mov rcx, 1              # remember number was negative

.convert_loop_start:

	# Special case for 0
	test rax, rax
	jnz .convert_loop

	mov byte ptr [rdi], '0'
	dec rdi
	jmp .after_convert

.convert_loop:
	xor rdx, rdx
	div rbx
	add dl, '0'
	mov [rdi], dl
	dec rdi
	test rax, rax
	jnz .convert_loop

.after_convert:

	# Add minus sign if needed
	test rcx, rcx
	jz .done_sign

	mov byte ptr [rdi], '-'
	dec rdi

.done_sign:

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
func (c *Compiler) writeHeader(f io.Writer) {
	text := `
	# Define our entry-point
	.global _start

	# Declare our library-functions
	.global print_number
	.global print_string
	.global print_string_with_length
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
