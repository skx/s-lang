PROGRAMS := $(basename $(wildcard *.in))
ASMFILES := $(addsuffix .s,$(PROGRAMS))

.PHONY: all clean

all: $(PROGRAMS)

clean:
	@rm -f $(PROGRAMS) $(ASMFILES)

# Ensure we have the compiler
../s-lang:
	cd .. && go build .

# Optional assembly generation
ifeq ($(ASM),1)
GEN_ASM_CMD = ../s-lang generate -output $@.s $<
else
GEN_ASM_CMD = @true
endif


# Generic rule - "foo.in" becomes "foo.s" and "foo"
%: %.in ../s-lang
	$(GEN_ASM_CMD)
	@# actually compile
	../s-lang compile -output $@ $<
	strip --strip-all $@
