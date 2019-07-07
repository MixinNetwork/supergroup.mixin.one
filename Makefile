plugins: plugins/hello.so

.SECONDEXPANSION:
plugins/%.so: $$(wildcard plugins/%/*.go)
	go build -buildmode=plugin -o $@ $^

clean-plugins:
	rm plugins/*.so
