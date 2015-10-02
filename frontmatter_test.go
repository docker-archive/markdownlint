package main

import (
	"bufio"
	"strings"
	"testing"
)

// NOTE: this has some spaces and tabs as well as newlines at the start. this is intentional
const OK_TOPIC = `

  
	
<!--[metadata]>
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
	err := checkHugoFrontmatter(bufio.NewReader(reader))

	if err != nil {
		t.Errorf("ERROR parsing: %v", err)
	}
}
