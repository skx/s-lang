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

You can use the included [trace.gdb](trace.gdb) file to automate single-stepping and logging execution.

Launch GDB with a binary and the script like so:

```
gdb  a.out -x trace.gdb
```

Then your output will be saved into the file `trace.log`.  You can update the trace-script to log registers,
but that gets quite noisy.



## Handy Tips

If you start tracing into a function you can enter `finish` to continue execution until the function returns.

Instead of running `stepi` you can enter `nexti` to skip over a call.



## Disassembly

I always prefer `set disassembly-flavor intel` because I'm old-school.

You can make that the default via:

    echo "set disassembly-flavor intel" >> ~/.gdbinit
