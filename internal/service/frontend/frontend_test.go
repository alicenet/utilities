package frontend

import (
	"context"
	"testing"

	alicev1 "github.com/alicenet/utilities/api/alice/v1"
)

func TestListStores(t *testing.T) {
	t.Parallel()
	t.Skip("skipping until we add memory stores")

	s := Service{}
	ctx := context.Background()
	req := &alicev1.ListStoresRequest{
		Address: "123",
	}
	resp, err := s.ListStores(ctx, req)

	if err == nil {
		t.Error("expected error")
	}

	t.Logf("resp: %v\nerr: %v\n", resp, err)
}
