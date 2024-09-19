package assets

import _ "embed"

//go:embed cacert/cacert.pem
var CaCerts []byte // Embeds the PEM file into the Go binary
