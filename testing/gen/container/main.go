package main

import (
	"os"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func main() {
	alice := must(ed25519.Generate())
	space := must(ed25519.Generate())
	service := must(did.Parse("did:web:example.com"))

	command := must(command.Parse("/fruits/purchase"))

	dlg := must(
		delegation.Delegate(
			space,
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

	arguments := must(datamodel.NewMap(datamodel.WithValue("fruits", []string{"apple", "banana"})))
	meta := must(
		datamodel.NewMap(datamodel.WithValues(map[string]any{
			"id":   must(ed25519.Generate()).DID().String(),
			"root": must(cid.Parse("bafkreigh2akiscaildcqabsyg3dfr6chu3fgpregiymsck7e7aqa4s52zy")),
			"name": "test",
			"size": int64(1000),
			"blob": must(datamodel.NewMap(datamodel.WithValue("digest", []byte{1, 2, 3}))),
		})),
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
	os.Stdout.Write(must(container.Encode(container.Base64, ct)))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
