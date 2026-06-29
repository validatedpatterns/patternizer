package embedded

import "embed"

// Resources holds the embedded resources directory tree (pattern.sh, Makefile, etc.).
//
//go:embed resources/*
var Resources embed.FS

// Skills holds the embedded skills directory tree for IDE integrations.
//
//go:embed all:skills
var Skills embed.FS
