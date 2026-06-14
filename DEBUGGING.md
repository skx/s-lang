# Debugging Binaries

Our binaries are static, and have no debug information, so it might be a challenge to debug them



## GDB

Load the binary into GDB:

     $ gdb a.out
     Reading symbols from a.out...
     (No debugging symbols found in a.out)
     (gdb)

Start the binary, via "starti":

    (gdb) starti
    Starting program: /home/skx/s-lang/a.out

    Program stopped.
    0x0000000000401013 in ?? ()

You can now enter `stepi` to step forward, and keep pressing RETURN to repeat the command.  Or run `stepi 100` to step forward 100 instructions.

At any time you may show register contents via `info registers` or `info registers float`.

You might try the new GDB TUI options:

* `layout asm`
  * Show the disassembly at the point.
* `layout regs`
  * Show the register, and flag contents.



## Tracing

Save this file as `trace.gdb`:

```text
set pagination off

set logging file trace.log
set logging overwrite off
set logging redirect on
set logging enabled on

starti

display/i $pc

while 1
    stepi
end
```

Now launch GDB with a binary:

```
gdb  a.out -x trace.gdb
```
And look at `trace.log` to see a log of the instructions executed.



## Handy Tips

If you start tracing into a function you can enter `finish` to continue execution until the function returns.

Instead of running `stepi` you can enter `nexti` to skip over a call.



## Disassembly

I always prefer `set disassembly-flavor intel` because I'm old-school.

You can make that the default via:

    echo "set disassembly-flavor intel" >> ~/.gdbinit
