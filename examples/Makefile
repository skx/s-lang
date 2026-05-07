ALL: factorial fizzbuzz

# Ensure we have the compiler
../s-lang:
	cd .. ; go build . ; cd examples

# factorial 1-20
factorial: factorial.in ../s-lang
	../s-lang compile -output factorial factorial.in

# fizzbuzz 1-100
fizzbuzz: fizzbuzz.in ../s-lang
	../s-lang compile -output fizzbuzz fizzbuzz.in
