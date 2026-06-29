package embedded

import "embed"

//go:embed resources/*
var Resources embed.FS

//go:embed all:skills
var Skills embed.FS
