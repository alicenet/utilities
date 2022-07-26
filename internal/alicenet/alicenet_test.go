package alicenet

import (
	"context"
	"flag"
	"testing"
)

const addr = "api.alicenet.duckdns.org"

//nolint:gochecknoglobals // Flags are ok
var integration = flag.Bool("integration", false, "run integration tests")

func TestBlockHeight(t *testing.T) {
	if !*integration {
		t.Skip("not running integration tests. use -integration to enable")
	}

	t.Parallel()

	anet := Connect(addr)
	ctx := context.Background()

	height, err := anet.Height(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if height == 0 {
		t.Error("height unexpectedly zero")
	}

	t.Logf("got height: %v", height)
}

func TestBlockHeader(t *testing.T) {
	if !*integration {
		t.Skip("not running integration tests. use -integration to enable")
	}

	t.Parallel()

	anet := Connect(addr)
	ctx := context.Background()

	header, err := anet.BlockHeader(ctx, 200000)
	if err != nil {
		t.Fatal(err)
	}

	expectedSigGroup := "143dcf31d714ba56d370b74506065b631b2153a156ceeb5a7d7da58899700d5d" +
		"0a55782984f051a170bd14114abdcd57a5a5990cddffa95c389a7e76af704a9e" +
		"22d68a16da67f4e9ef5319c0ec20f6221c7cb40cc5ccc16d4eb4d1a5353ee2ae" +
		"17509781f7b80cab78d0d4743cb7d19b14564f3937aa810899b51cd518d8bb3a" +
		"031c614e21a6c8441f1bf822c59c14369f62377f4e816c96a9be308ef55c7b27" +
		"03f84fc11336741744c57092d73b82666afa7ccf85f41745b07de184272258b3"

	if header.SigGroup != expectedSigGroup {
		t.Errorf("header is not expected.\nwanted SigGroup %s\n   got SigGroup %s", expectedSigGroup, header.SigGroup)
	}

	t.Logf("header: %+v", header)
}

func TestTransaction(t *testing.T) {
	if !*integration {
		t.Skip("not running integration tests. use -integration to enable")
	}

	t.Parallel()

	anet := Connect(addr)
	ctx := context.Background()

	txn, err := anet.Transaction(ctx, "7633e92ab233af14e63517b0df66551303d49de2b06a67833ee67f01ae9ffa00")
	if err != nil {
		t.Fatal(err)
	}

	if len(txn.Tx.Vin) != 1 {
		t.Errorf("input transactions, want %v, got %v", 1, len(txn.Tx.Vin))
	}

	if len(txn.Tx.Vout) != 2 {
		t.Errorf("output transactions, want %v, got %v", 2, len(txn.Tx.Vout))
	}

	t.Logf("transaction: %+v", txn)
}
