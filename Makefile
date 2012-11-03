EXE = Scarlet

all: $(EXE)

$(EXE): $(wildcard *.go)
	go build -o $@ $^

clean:
	rm -f $(EXE)
	find . -iname "*~" -exec rm -f {} \;

deps:
	go get github.com/garyburd/redigo/redis

.PHONY: all clean deps
