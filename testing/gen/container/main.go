package main

import (
	"maps"
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/principal/signer"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/builder"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
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

	pol := must(builder.Build(
		builder.All(
			".fruits",
			builder.Or(
				builder.Equal(".", "apple"),
				builder.Equal(".", "orange"),
				builder.Equal(".", "banana"),
			),
		),
	))

	dlg := must(
		delegation.Delegate(
			market,
			alice,
			command,
			delegation.WithSubject(market),
			delegation.WithPolicy(pol),
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

	inv := must(
		invocation.Invoke(
			alice,
			market,
			command,
			arguments,
			invocation.WithProofs(dlg.Link()),
			invocation.WithMetadata(meta),
			invocation.WithExpiration(ucan.Now()+30),
		),
	)

	// Execute ///////////////////////////////////////////////////////////////////

	out := result.OK[int64, any](42)
	rcpt := must(receipt.Issue(market, inv.Task().Link(), out))

	// Transport /////////////////////////////////////////////////////////////////

	ct := must(
		container.New(
			container.WithDelegations(dlg),
			container.WithInvocations(inv),
			container.WithReceipts(rcpt),
		),
	)
	os.Stdout.Write(must(container.Encode(container.Base64, ct)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
