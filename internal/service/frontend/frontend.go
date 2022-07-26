package frontend

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	alicev1 "github.com/alicenet/indexer/api/alice/v1"
	"github.com/alicenet/indexer/internal/alicenet"
)

const defaultLimit = 100

type Service struct {
	stores *alicenet.Stores
}

func NewService(stores *alicenet.Stores) *Service {
	return &Service{
		stores: stores,
	}
}

func (s *Service) ListStores(
	context.Context, *alicev1.ListStoresRequest) (
	*alicev1.ListStoresResponse, error,
) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *Service) GetStoreValue(
	context.Context, *alicev1.GetStoreValueRequest) (
	*alicev1.GetStoreValueResponse, error,
) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *Service) ListTransactionsForAddress(
	context.Context, *alicev1.ListTransactionsForAddressRequest) (
	*alicev1.ListTransactionsForAddressResponse, error,
) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *Service) GetBalance(
	ctx context.Context, req *alicev1.GetBalanceRequest) (
	*alicev1.GetBalanceResponse, error,
) {
	account, err := s.stores.Accounts.Get(ctx, spanner.Key{req.Address})
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		account = alicenet.Account{
			Balance: "0",
		}
	}

	resp := &alicev1.GetBalanceResponse{
		Balance: account.Balance,
	}

	return resp, nil
}

func (s *Service) GetTransaction(
	context.Context, *alicev1.GetTransactionRequest) (
	*alicev1.GetTransactionResponse, error,
) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *Service) ListTransactions(
	ctx context.Context, req *alicev1.ListTransactionsRequest) (
	*alicev1.ListTransactionsResponse, error,
) {
	limit := int64(defaultLimit)
	if req.Limit > 0 {
		limit = req.Limit
	}

	txns, err := s.stores.Transactions.List(ctx, limit, req.Offset)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.ListTransactionsResponse{}

	for _, v := range txns {
		resp.TransactionHashes = append(resp.TransactionHashes, v.TransactionHash)
	}

	return resp, nil
}

func (s *Service) GetBlock(
	ctx context.Context, req *alicev1.GetBlockRequest) (
	*alicev1.GetBlockResponse, error,
) {
	block, err := s.stores.Blocks.Get(ctx, spanner.Key{int64(req.Height)})
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.GetBlockResponse{
		Block: &alicev1.Block{
			ChainId:             uint32(block.ChainID),
			Height:              uint32(block.Height),
			TransactionCount:    uint32(block.TransactionCount),
			PreviousBlockHash:   block.PreviousBlockHash,
			TransactionRootHash: block.TransactionRootHash,
			StateRootHash:       block.StateRootHash,
			HeaderRootHash:      block.HeaderRootHash,
			GroupSignatureHash:  block.GroupSignatureHash,
			TransactionHashes:   block.TransactionHashes,
		},
	}

	return resp, nil
}

func (s *Service) ListBlocks(
	ctx context.Context, req *alicev1.ListBlocksRequest) (
	*alicev1.ListBlocksResponse, error,
) {
	limit := int64(defaultLimit)
	if req.Limit > 0 {
		limit = req.Limit
	}

	blocks, err := s.stores.Blocks.List(ctx, limit, req.Offset)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.ListBlocksResponse{}
	for _, v := range blocks {
		resp.Heights = append(resp.Heights, uint32(v.Height))
	}

	return resp, nil
}
