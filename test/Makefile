# Some of our test-code uses environmental variables, so
# we need to ensure we have something setup.
export FOO=BAR
export NAME=Steve

PROGRAMS := $(basename $(wildcard *.in))

.PHONY: all clean

all: $(PROGRAMS)

clean:
	@rm -f $(PROGRAMS) *.out

# Ensure we have the compiler
../s-lang:
	cd .. && go build .

# Generic rule
#   "foo.in" becomes "foo"
#   "foo" is executed to produce "foo.out"
#   "foo.out" is compared to "foo.expected" which is the expected output
#
# NOTE: We disable compile-time type-checking, otherwise the run-time checking
# cannot be used - Because the compilation will be terminate.
#
%: %.in ../s-lang
	@echo "$<"
	@echo " compiling $< to $@"
	@../s-lang compile -check-stdlib=false -output $@ $<
	@strip --strip-all $@
	@echo " executing $@ > $@.out"
	@./$@ > $@.out 2>&1 || true
	@echo " comparing $@.out to $@.expected"
	@cmp $@.out $@.expected || exit "program $< did not match"
	@echo " cleanup"
	@rm $@.out $@
