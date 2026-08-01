package main

import (
	"os"

	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/multikey/ed25519"
	tdm "github.com/fil-forge/ucantone/testutil/datamodel"
	"github.com/fil-forge/ucantone/ucan/command"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
)

func main() {
	issuer := must(ed25519.GenerateIssuer())
	subject := must(did.Parse("did:key:z6MkrYxEAeY8bQGaxaY2S5QuN7skMSAyye3XacFxk2iMFw5G"))
	audience := must(did.Parse("did:web:example.com"))
	command := must(command.Parse("/test/invoke"))
	arguments := tdm.TestArgs{
		ID:    must(ed25519.GenerateIssuer()).DID(),
		Link:  must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
		Str:   "test",
		Num:   1000,
		Bytes: []byte{1, 2, 3},
		Obj: tdm.TestObject{
			Bytes: []byte{4, 5, 6},
		},
		List: []string{"one", "two", "three"},
	}

	inv := must(invocation.Invoke(issuer, subject, command, &arguments, invocation.WithAudience(audience)))
	os.Stdout.Write(must(invocation.Encode(inv)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
