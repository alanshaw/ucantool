package main

import (
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func main() {
	alice := must(ed25519.Generate())
	space := must(ed25519.Generate())
	service := must(did.Parse("did:web:example.com"))

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

	meta := datamodel.NewMap()
	meta.SetValue("foo", "bar")

	dlg := must(
		delegation.Delegate(
			space,
			alice,
			command,
		),
	)
	dlgBytes := must(delegation.Encode(dlg))
	prf := must(
		cid.Prefix{
			Version:  1,
			Codec:    dagcbor.Code,
			MhType:   multihash.SHA2_256,
			MhLength: -1,
		}.Sum(dlgBytes),
	)

	inv := must(
		invocation.Invoke(
			alice,
			space,
			command,
			arguments,
			invocation.WithAudience(service),
			invocation.WithProofs(prf),
			invocation.WithMetadata(meta),
			invocation.WithExpiration(ucan.Now()+30),
			invocation.WithCause(
				must(cid.Parse("bafkreibtsod63vtdyyq5iwfblycy6gk2te5n3lr6k6orymxp23x6cken3e")),
			),
		),
	)

	ct := must(
		container.New(
			container.WithDelegations(dlg),
			container.WithInvocations(inv),
		),
	)
	os.Stdout.Write(must(container.Encode(container.Base64Gzip, ct)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
