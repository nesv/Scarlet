all: scarletd

scarletd: $(wildcard scarlet/*.go)
	go build -o $@ $^

clean:
	rm scarletd
	find . -iname "*~" -exec rm -f {} \;

.PHONY: all clean
