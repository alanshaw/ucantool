package main

import (
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/principal/secp256k1"
	"github.com/alanshaw/ucantone/principal/signer"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
)

func main() {
	alice := must(secp256k1.Generate())
	market := must(
		signer.Wrap(
			must(ed25519.Generate()),
			must(did.Parse("did:web:fruit.market")),
		),
	)

	// Delegate //////////////////////////////////////////////////////////////////

	dlg := must(
		delegation.Delegate(
			market,
			alice,
			market,
			"/fruits/purchase",
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

	arguments := ipld.Map{
		"fruits": []string{"apple", "banana"},
	}
	meta := ipld.Map{
		"id":   must(ed25519.Generate()).DID().String(),
		"root": must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
		"name": "test",
		"size": int64(1000),
		"blob": ipld.Map{"digest": []byte{1, 2, 3}},
	}

	inv := must(
		invocation.Invoke(
			alice,
			market,
			"/fruits/purchase",
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

	ct := container.New(
		container.WithDelegations(dlg),
		container.WithInvocations(inv),
		container.WithReceipts(rcpt),
	)
	os.Stdout.Write(must(container.Encode(container.Base64, ct)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
