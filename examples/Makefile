ALL: factorial fizzbuzz string while

# Ensure we have the compiler
../s-lang:
	cd .. ; go build . ; cd examples

# factorial 1-20
factorial: factorial.in ../s-lang
	../s-lang compile -output factorial factorial.in

# fizzbuzz 1-100
fizzbuzz: fizzbuzz.in ../s-lang
	../s-lang compile -output fizzbuzz fizzbuzz.in

# store a string in a variable, and print it
string: string.in ../s-lang
	../s-lang compile -output string string.in

# nested while loops
while: while.in ../s-lang
	../s-lang compile -output while while.in
