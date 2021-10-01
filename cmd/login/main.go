package main

import (
	"github.com/agravelot/permis_toolkit/ornikar"
)

func main() {
	// TODO pass env
	token, _ := ornikar.Login("", "")

	println(token)
}
