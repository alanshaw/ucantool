package main

import (
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
)

func main() {
	issuer := must(ed25519.Generate())
	subject := must(did.Parse("did:key:z6MkrYxEAeY8bQGaxaY2S5QuN7skMSAyye3XacFxk2iMFw5G"))
	audience := must(did.Parse("did:web:example.com"))
	command := must(command.Parse("/test/invoke"))
	arguments := must(datamodel.NewMapFromCBORMarshaler(
		&helpers.TestArgs{
			ID:    must(ed25519.Generate()).DID(),
			Link:  must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
			Str:   "test",
			Num:   1000,
			Bytes: []byte{1, 2, 3},
			Obj: helpers.TestObject{
				Bytes: []byte{4, 5, 6},
			},
			List: []string{"one", "two", "three"},
		},
	))

	inv := must(invocation.Invoke(issuer, subject, command, arguments, invocation.WithAudience(audience)))
	os.Stdout.Write(must(invocation.Encode(inv)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
