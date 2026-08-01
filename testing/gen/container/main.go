package main

import (
	"os"

	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/ipld/datamodel"
	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/multikey/ed25519"
	"github.com/fil-forge/ucantone/multikey/secp256k1"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/command"
	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/delegation/policy"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
)

func main() {
	alice := must(secp256k1.GenerateIssuer())
	market := multikey.NewIssuer(
		must(did.Parse("did:web:fruit.market")),
		must(ed25519.Generate()),
	)

	// Delegate //////////////////////////////////////////////////////////////////

	dlg := must(
		delegation.Delegate(
			market,
			alice.DID(),
			market.DID(),
			command.MustParse("/fruits/purchase"),
			delegation.WithPolicyBuilder(
				policy.All(
					".fruits",
					policy.Or(
						policy.Equal(".", "apple"),
						policy.Equal(".", "orange"),
						policy.Equal(".", "banana"),
					),
				),
			),
		),
	)

	// Invoke ////////////////////////////////////////////////////////////////////

	arguments := datamodel.Map{
		"fruits": []string{"apple", "banana"},
	}
	meta := datamodel.Map{
		"id":   must(ed25519.GenerateIssuer()).DID().String(),
		"root": must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
		"name": "test",
		"size": int64(1000),
		"blob": datamodel.Map{"digest": []byte{1, 2, 3}},
	}

	inv := must(
		invocation.Invoke(
			alice,
			market.DID(),
			command.MustParse("/fruits/purchase"),
			arguments,
			invocation.WithProofs(dlg.Link()),
			invocation.WithMetadata(meta),
			invocation.WithExpiration(ucan.Now()+30),
		),
	)

	// Execute ///////////////////////////////////////////////////////////////////

	rcpt := must(receipt.IssueOK(market, inv.Task().Link(), datamodel.NewAny(42)))

	// Transport /////////////////////////////////////////////////////////////////

	ct := container.New(
		container.WithDelegations(dlg),
		container.WithInvocations(inv),
		container.WithReceipts(rcpt),
	)
	os.Stdout.Write(must(container.Encode(container.Base64urlGzip, ct)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
