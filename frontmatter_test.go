package main

import (
	"io"
	"strings"
	"testing"
)

const OK_TOPIC = `<!--[metadata]>
+++
title = "Dockerfile reference"
description = "Dockerfiles use a simple DSL which allows you to automate the steps you would normally manually take to create an image."
keywords = ["builder, docker, Dockerfile, automation, image creation"]
[menu.main]
parent = "mn_reference"
+++
<![end-metadata]-->
# Dockerfile reference
`

func TestFrontmatterFound(t *testing.T) {
	reader := strings.NewReader(OK_TOPIC)
	sectionReader := io.NewSectionReader(reader, 0, 2048)
	length, err := checkHugoFrontmatter(sectionReader)

	if err != nil {
		t.Errorf("ERROR parsing: %v", err)
	}
	if length != 12 {
		t.Errorf("ERROR wrong length for frontmatter: %d", length)
	}
}
