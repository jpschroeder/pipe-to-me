TARGET = client

OBJS= client.o

CC = gcc

CFLAGS = -c -g

# sudo apt install libcurl4-openssl-dev libssl-dev
LIBS = -lcurl -lssl -lcrypto

$(TARGET) : $(OBJS)
	$(CC)  -o $(TARGET) $(OBJS) $(LDFLAGS) $(LIBS)

client.o : client.c
	$(CC) $(CFLAGS) $<