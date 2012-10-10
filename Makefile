all: scarletd

scarletd: $(wildcard scarlet/*.go)
	go build -o $@ $^

clean:
	rm -f scarletd
	find . -iname "*~" -exec rm -f {} \;

deps:
	for dep in `cat deps.list`; do echo Installing $$dep; go get $$dep; done

.PHONY: all clean deps
