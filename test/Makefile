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
%: %.in ../s-lang
	@echo "$<"
	@echo " compiling $< to $@"
	@../s-lang compile -output $@ $<
	@echo " executing $@ > $@.out"
	@./$@ > $@.out 2>&1 || true
	@echo " comparing $@.out to $@.expected"
	@cmp $@.out $@.expected || exit "program $< did not match"
	@echo " cleanup"
	@rm $@.out $@
