all: scarletd

scarletd: $(wildcard scarlet/*.go)
	go build -o $@ $^

clean:
	rm scarletd
