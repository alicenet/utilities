//go:build tools

package main

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/golang/mock/mockgen"
	_ "golang.org/x/tools/cmd/stringer"
)
