pasm -b spislave.p
gcc -g -fPIC -c -o lib/spislave.o lib/spislave.c
gcc -g -fPIC -shared -o libspislave.so lib/spislave.o
