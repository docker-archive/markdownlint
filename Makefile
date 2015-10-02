
# Adds build information from git repo
#
# as suggested by tatsushid in
# https://github.com/spf13/hugo/issues/540

COMMIT_HASH=`git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE=`date +%FT%T%z`
LDFLAGS=-ldflags "-X github.com/spf13/hugo/hugolib.CommitHash=${COMMIT_HASH} -X github.com/spf13/hugo/hugolib.BuildDate=${BUILD_DATE}"

shell: docker-build
	docker run --rm -it -v $(CURDIR):/go/src/github.com/SvenDowideit/doccheck doccheck bash

docker-build:
	rm -f doccheck.gz
	docker build -t doccheck .

docker: docker-build
	docker run --name doccheck-build doccheck gzip doccheck
	docker cp doccheck-build:/go/src/github.com/SvenDowideit/doccheck/doccheck.gz .
	docker rm doccheck-build
	gunzip doccheck.gz

run:
	./doccheck .
