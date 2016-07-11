
# Adds build information from git repo
#
# as suggested by tatsushid in
# https://github.com/spf13/hugo/issues/540

COMMIT_HASH=`git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE=`date +%FT%T%z`
LDFLAGS=-ldflags "-X github.com/spf13/hugo/hugolib.CommitHash=${COMMIT_HASH} -X github.com/spf13/hugo/hugolib.BuildDate=${BUILD_DATE}"

build:
	go build -o markdownlint main.go

shell: docker-build
	docker run --rm -it -v $(CURDIR):/go/src/github.com/SvenDowideit/markdownlint markdownlint bash

docker-build:
	rm -f markdownlint.zip
	docker build -t markdownlint .

docker: docker-build
	docker rm markdownlint-build || true
	docker run --name markdownlint-build markdownlint ls
	docker cp markdownlint-build:/go/src/github.com/SvenDowideit/markdownlint/markdownlint.zip .
	rm -f markdownlint
	unzip -o markdownlint.zip

run:
	./markdownlint .

validate:
	docker run \
		-v $(CURDIR)/markdownlint:/usr/bin/markdownlint \
		--volumes-from docsdockercom_data_1 \
		--rm -it \
			debian /usr/bin/markdownlint /docs/content/

AWSTOKENSFILE ?= ../aws.env
-include $(AWSTOKENSFILE)
export GITHUB_USERNAME GITHUB_TOKEN

RELEASE_DATE=`date +%F`

release: docker
	# TODO: check that we have upstream master, bail if not
	docker run --rm -it -e GITHUB_TOKEN markdownlint \
		github-release release --user docker --repo markdownlint --tag $(RELEASE_DATE)
	docker run --rm -it -e GITHUB_TOKEN markdownlint \
		github-release upload --user docker --repo markdownlint --tag $(RELEASE_DATE) \
			--name markdownlint \
			--file markdownlint
	docker run --rm -it -e GITHUB_TOKEN markdownlint \
		github-release upload --user docker --repo markdownlint --tag $(RELEASE_DATE) \
			--name markdownlint-osx \
			--file markdownlint.app
	docker run --rm -it -e GITHUB_TOKEN markdownlint \
		github-release upload --user docker --repo markdownlint --tag $(RELEASE_DATE) \
			--name markdownlint.exe \
			--file markdownlint.exe
