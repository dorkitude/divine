package deckdata

import "embed"

// FS contains all bundled deck files included in the binary.
//
//go:embed */*
var FS embed.FS

