EXE = Scarlet

all: $(EXE)

$(EXE): $(wildcard *.go)
	go build -o $@ $^

clean:
	rm -f $(EXE)
	find . -iname "*~" -exec rm -f {} \;

deps:
	go get github.com/simonz05/godis/redis

.PHONY: all clean deps
