// +build tools

package tools

// This package records the versions of the tools that are used during development of this module.
// To install the tools:
//	go install golang.org/x/tools/cmd/stringer
//
// Also see https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md

import (
	_ "golang.org/x/tools/cmd/stringer"
)
