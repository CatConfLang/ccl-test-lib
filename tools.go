//go:build tools

package main

//go:generate go install github.com/atombender/go-jsonschema
//go:generate go install gotest.tools/gotestsum

import (
	_ "github.com/atombender/go-jsonschema"
	_ "gotest.tools/gotestsum"
)
