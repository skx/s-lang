#!/usr/bin/perl -w
#
# Trivial compiler that understands variable assignment, and printing registers.
#
# We can also return a single value as exit-code.
#
#

use strict;
use warnings;


#
# Parse the program
#
my $prog =<<EOP;
  let a = 17
  let b = 4;
  let c = 3;

  print( b );
  newline();

  print(a);
  newline();

  return c;
EOP

#
# Write header to output file
#
writeHeader("out.s");

#
# Process line by line
#
foreach my $line (split(/\n/, $prog) ){

    # Remove newline
    chomp($line);

    # Remove leading/trailing whitespace
    $line =~ s/^\s+|\s+$//g;

    # Skip comments
    next if ( $line =~ /^#/ );

    compile($line);
}

writeFooter("out.s");

# now compile
system( "as -msyntax=intel -mnaked-reg out.s -o out.o");
system( "as -msyntax=intel -mnaked-reg stdlib.s -o stdlib.o");
system("ld -s -o out out.o stdlib.o");
system("./out");


sub compile {
    my($line) = (@_);

    # newline?
    if ( $line =~ /^newline\s*\(\s*\);*\s*$/ ) {
        open( OUT, ">>", "out.s")
          or die "cannot compile $!";
        print OUT<<EOP;
        # Print newline
        call newline

EOP
        close(OUT);
        return;
    }

    # print?
    if ( $line =~ /^print\s*\(\s*([a-z])\s*\)\s*;*\s*$/ ) {
        my $reg = lc($1);
        my $num = ord($reg) - ord('a');
        open( OUT, ">>", "out.s")
          or die "cannot compile $!";
        print OUT<<EOP;
        # Print register $reg [$num]
        lea rcx, vars
        mov rax, [rcx + $num*8]
        call print_rax

EOP
        close(OUT);
        return;
    }

    # assign
    if (  $line =~ /^let\s+([a-z])\s*=\s*([0-9]+)\s*;*\s*$/ ) {
        my $reg = lc($1);
        my $val = $2;
        my $num = ord($reg) - ord('a');
        open( OUT, ">>", "out.s")
          or die "cannot compile $!";
        print OUT<<EOS;
        # Set register $reg [$num] = $val
        lea rcx, vars
        mov rax, $val
        mov [rcx + $num*8], rax

EOS
        close(OUT);

        return;
    }

    # return
    if (  $line =~ /^return\s*([a-z])\s*;*\s*$/ ) {
        my $reg = lc($1);
        my $num = ord($reg) - ord('a');
        open( OUT, ">>", "out.s")
          or die "cannot compile $!";
        print OUT<<EOP;
        # Exit with status from $reg [$num]
        lea rcx, vars
        mov rax, [rcx + $num*8]
        call exit_with_status

EOP
        close(OUT);
        return;
    }
}

sub writeHeader {
    my($name) = (@_);
    open(FILE, ">", $name )
      or die "failed to write to $name $!";
    print FILE <<EOH;
        # Define our entry-point
        .global _start

        # Declare functions in our STDLIB.s file
        .extern print_rax
        .extern newline
        .extern exit_with_status

        # Writeable data storage
        .section .bss

        # storage for our variables
vars:
        .skip 26 * 8

        #
        # Code
        #
        .section .text
_start:

EOH
    close(FILE);
}




sub writeFooter {
    my($name) = (@_);
    open(FILE, ">>", $name )
      or die "failed to write to $name $!";
    print FILE <<EOH;

        #
        # Exit just in case the program was missing a
        # terminating RETURN statement.
        #
        mov rax, 60
        mov rdi, 0
        syscall

EOH
    close(FILE);
}
