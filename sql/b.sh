#rm ddd
bison -dy parse.y; mv y.tab.c parse.c
flex scan.l; mv lex.yy.c scan.c
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c scan.c -o build/scan.o
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c parse.c -o build/parse.o
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c nodes.c -o build/nodes.o
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c interp.cc -o build/interp.o
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c ql_error.cc -o build/ql_error.o
g++ -c -fPIC -g -O1 -Wall -DPF_STATS -I. -c ddd.c -o build/ddd.o


g++ -fPIC -shared -Wl,-soname,libparser.so -o ../lib/libparser.so  build/*.o
#g++ -fPIC -g -O1 -Wall -DPF_STATS -I. -o ddd ./build/*.o