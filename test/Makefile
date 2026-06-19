# Some of our test-code uses environmental variables, so
# we need to ensure we have something setup.
export FOO=BAR
export NAME=Steve

PROGRAMS := $(basename $(wildcard *.in))
TESTS := $(addsuffix .test,$(PROGRAMS))
ASMFILES := $(addsuffix .s,$(PROGRAMS))

.PHONY: all clean

all: $(PROGRAMS)

clean:
	@rm -f $(PROGRAMS) *.out

test: $(TESTS)

# Ensure we have the compiler
../s-lang:
	cd .. && go build .

# Optional assembly generation
ifeq ($(ASM),1)
GEN_ASM_CMD = @../s-lang generate -check-stdlib=false -output $@.s $<
else
GEN_ASM_CMD = @true
endif


# Generic rule
#   "foo.in" becomes "foo"
#   "foo" is executed to produce "foo.out"
#   "foo.out" is compared to "foo.expected" which is the expected output
#
# NOTE: We disable compile-time type-checking, otherwise the run-time checking
# cannot be used - Because the compilation will be terminate.
#
%: %.in ../s-lang
	@echo "compiling $@"
	$(GEN_ASM_CMD)
	@../s-lang compile -check-stdlib=false -output $@ $<
	@strip --strip-all $@

# Generic rule for testing each binary
%.test: %
	@echo "executing $@ > $@.out"
	@./$< > $<.out 2>&1 < $<.in || true
	@echo "  comparing $<.out to $<.expected"
	@cmp $<.out $<.expected || (echo "program $< did not match" ; cat $<.out ; exit 1)
	@rm $<.out
