
#
# This block makes GDB exit after making the trace
#
set $_exitcode = -999
set height 0
handle SIGTERM nostop print pass
handle SIGPIPE nostop
define hook-stop
    if $_exitcode != -999
        quit
    end
end

#
# Remove previous trace
#
shell rm trace.log || true

#
# Set logging and disassembly options
#
set pagination off
set disassembly-flavor intel
set logging file trace.log
set logging overwrite off
set logging redirect on
set logging enabled on


#
# Start the program
#
starti

display/i $pc

while 1
    stepi

#
# Dumping the registers is noisy and slow
#   info registers
#
end

