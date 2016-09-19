
rm libapi.a libparser.a libparser.so libapi.so cgotest ddd
rm *.o
g++ -fPIC -c api.c
g++ -fPIC -shared -Wl,-soname,libapi.so -o libapi.so  api.o 
ar -rv libapi.a api.o  

ar -rv libparser.a  ../build/*.o 
g++ -fPIC -shared -Wl,-soname,libparser.so -o libparser.so  ../build/*.o 

#g++ -fPIC -g -O1 -Wall -DPF_STATS -I. -o ddd api.o ../build/*.o

go build
