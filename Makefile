build:
	go build -buildmode=c-shared -o dist/pygohcl_go.so pygohcl.go
	python3 build.py > dist/pygohcl_go.c
	cd dist && gcc  -shared -g -lc $(shell python3-config --cflags) $(shell python3-config --ldflags) -fPIC pygohcl_go.c pygohcl_go.so -o pygohcl.so

docker-build:
	docker build --tag pygohcl .

docker-gen: docker-build
	rm -rf dist
	mkdir dist
	docker run -it -v ${PWD}:/app pygohcl

test:
	docker run -it -v ${PWD}:/app pygohcl /bin/bash -c 'LD_LIBRARY_PATH=/app/dist PYTHONPATH=/app/dist python3 ./examples/pygohcl_example.py'
