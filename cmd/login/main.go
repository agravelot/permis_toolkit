package main

import (
	"github.com/agravelot/permis_toolkit/ornikar"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	// TODO pass env
	token, _ := ornikar.Login("", "")

	spew.Dump(token)
}
