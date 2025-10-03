package main

import (
	"maps"
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/principal/signer"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func main() {
	alice := must(ed25519.Generate())
	market := must(
		signer.Wrap(
			must(ed25519.Generate()),
			must(did.Parse("did:web:fruit.market")),
		),
	)

	command := must(command.Parse("/fruits/purchase"))

	// Delegate //////////////////////////////////////////////////////////////////

	dlg := must(
		delegation.Delegate(
			market,
			alice,
			command,
			delegation.WithPolicy(
				policy.All(
					must(selector.Parse(".fruits")),
					policy.Or(
						policy.Equal(must(selector.Parse(".")), "apple"),
						policy.Equal(must(selector.Parse(".")), "orange"),
						policy.Equal(must(selector.Parse(".")), "banana"),
					),
				),
			),
		),
	)

	// Invoke ////////////////////////////////////////////////////////////////////

	arguments := datamodel.NewMap(datamodel.WithEntry("fruits", []string{"apple", "banana"}))
	meta := datamodel.NewMap(
		datamodel.WithEntries(
			maps.All(map[string]any{
				"id":   must(ed25519.Generate()).DID().String(),
				"root": must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
				"name": "test",
				"size": int64(1000),
				"blob": datamodel.NewMap(datamodel.WithEntry("digest", []byte{1, 2, 3})),
			}),
		),
	)

	prf := must(
		cid.Prefix{
			Version:  1,
			Codec:    dagcbor.Code,
			MhType:   multihash.SHA2_256,
			MhLength: -1,
		}.Sum(must(delegation.Encode(dlg))),
	)

	inv := must(
		invocation.Invoke(
			alice,
			market,
			command,
			arguments,
			invocation.WithProofs(prf),
			invocation.WithMetadata(meta),
			invocation.WithExpiration(ucan.Now()+30),
		),
	)

	// Execute ///////////////////////////////////////////////////////////////////

	out := result.Ok[int64, any](42)
	rcpt := must(receipt.Issue(market, inv.Task(), out))

	// Transport /////////////////////////////////////////////////////////////////

	ct := must(
		container.New(
			container.WithDelegations(dlg),
			container.WithInvocations(inv),
			container.WithReceipts(rcpt),
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
