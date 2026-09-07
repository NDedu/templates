CC       ?= cc
CFLAGS   ?= -std=c11 -D_POSIX_C_SOURCE=200809L -Wall -Wextra -Wpedantic -g -O0 -pthread
LDFLAGS  ?= -pthread
LDLIBS   ?= -lm

BIN := algorithms
SRC := $(wildcard *.c)
OBJ := $(SRC:.c=.o)

all: $(BIN)

$(BIN): $(OBJ)
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LDLIBS)

%.o: %.c algorithms.h
	$(CC) $(CPPFLAGS) $(CFLAGS) -c $< -o $@

run: $(BIN)
	./$(BIN)

# Run a single section, e.g. make only SECTION=ARRAY
only:
	$(MAKE) clean
	$(MAKE) CPPFLAGS="$(CPPFLAGS) -DTEST_ALL=0 -DTEST_$(SECTION)=1" run

# The same checks under AddressSanitizer and UBSan, which catch the leaks,
# the out of bounds writes and the use after frees that C otherwise lets slide
asan:
	$(MAKE) clean
	$(MAKE) CFLAGS="$(CFLAGS) -fsanitize=address,undefined -fno-omit-frame-pointer" \
	        LDFLAGS="$(LDFLAGS) -fsanitize=address,undefined" run

# Type check main.c on its own, before any algorithm is implemented
check:
	$(CC) $(CPPFLAGS) $(CFLAGS) -fsyntax-only main.c

clean:
	rm -f $(OBJ) $(BIN)

.PHONY: all run only asan check clean
