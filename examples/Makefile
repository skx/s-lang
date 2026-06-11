PROGRAMS := $(basename $(wildcard *.in))
TESTS := $(addsuffix .test,$(PROGRAMS))
ASMFILES := $(addsuffix .s,$(PROGRAMS))

.PHONY: all clean

all: $(PROGRAMS)

test: $(TESTS)

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

# Generic rule for testing each binary
%.test: %
	./$< > $<.out 2>&1 < $<.in || true
	cmp $<.out $<.expected
	rm -f $<.out


# We can't run "life" so we have a more specific rule for that
# binary.
life.test: life.in ../s-lang
	$(GEN_ASM_CMD)
	../s-lang compile -output $@ $<
	strip --strip-all $@
	@echo "Skipping execution of $@"
