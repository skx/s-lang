#
#  Define the functions we export from this file
#
        .global exit_with_status
        .global newline
        .global print_rax


#
# Storage area for misc static strings
# Read-only section
#
        .section .data
newline_msg: .ascii "\n"
newline_msg_end:


#
# Storage area for writeable values
#
        .section .bss
print_rax_buffer:
        .skip 32


#
# Storage area for code
#
        .section .text


        #
        # Write a newline to STDOUT.
        #
newline:
        mov rax, 1
        mov rdi, 1
        mov rsi, offset newline_msg
        mov rdx, newline_msg_end-newline_msg
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
print_rax:
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

        # write(fd=1, buf=rdi, len=rdx)
        mov rax, 1           # sys_write
        mov rsi, rdi         # buffer
        mov rdi, 1           # stdout
        syscall

        ret
