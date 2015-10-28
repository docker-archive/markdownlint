package checkers

import (
	"testing"

	"github.com/SvenDowideit/markdownlint/data"
	"github.com/SvenDowideit/markdownlint/linereader"
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
const MISSING_COMMENT_START_TOPIC = `

  
	
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
const MISSING_COMMENT_END_TOPIC = `

  
	
<!--[metadata]>
+++
title = "Dockerfile reference end comment missing"
description = "Dockerfiles use a simple DSL which allows you to automate the steps you would normally manually take to create an image."
keywords = ["builder, docker, Dockerfile, automation, image creation"]
[menu.main]
parent = "mn_reference"
+++
# Dockerfile reference
`

func TestFrontmatterFound(t *testing.T) {
	file := "ok"
	data.AllFiles = make(map[string]*data.FileDetails)
	data.AddFile(file, file)

	err := CheckHugoFrontmatter(linereader.ByteReader(OK_TOPIC), file)

	if err != nil {
		t.Errorf("ERROR parsing: %v", err)
	}
}

func TestFrontmatterError(t *testing.T) {
	file := "missing"
	data.AllFiles = make(map[string]*data.FileDetails)
	data.AddFile(file, file)

	err := CheckHugoFrontmatter(linereader.ByteReader(MISSING_COMMENT_END_TOPIC), file)

	if err == nil {
		t.Errorf("Expected error")
	} else {
		if err.Error() != "Did not find expected close metadata comment" {
			t.Errorf("unexpected ERROR parsing: %v", err)
		}
	}
}
