define README_PREFIX
# gotwitch
*Twitch client written in Golang*

It is a fairly simple approach to interacting with Twitch.tv through a command-line interface.
It is written in [Go](https://golang.org/) and is a [free software](https://www.fsf.org/about/what-is-free-software).

## Help

```
endef

define README_POSTFIX
```
endef

export README_PREFIX
export README_POSTFIX

gotwitch: gotwitch.go
	go build
	@echo "$$README_PREFIX" > README.md
	./gotwitch --help-long >> README.md
	@echo "$$README_POSTFIX" >> README.md
