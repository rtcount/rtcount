rm ddd
g++ -m32 -g -O1 -Wall -DPF_STATS -I. -c scan.c -o build/scan.o
g++ -m32 -g -O1 -Wall -DPF_STATS -I. -c parse.c -o build/parse.o
g++ -m32 -g -O1 -Wall -DPF_STATS -I. -c nodes.c -o build/nodes.o
g++ -m32 -g -O1 -Wall -DPF_STATS -I. -c interp.cc -o build/interp.o
g++ -m32 -g -O1 -Wall -DPF_STATS -I. -c ql_error.cc -o build/ql_error.o


g++ -m32 -g -O1 -Wall -DPF_STATS -I. -o ddd ./build/*.o
