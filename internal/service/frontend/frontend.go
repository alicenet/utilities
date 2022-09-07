package frontend

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	alicev1 "github.com/alicenet/indexer/api/alice/v1"
	"github.com/alicenet/indexer/internal/alicenet"
)

const (
	defaultLimit = 100
	maxLimit     = 1024
)

// A validator will return an error for any misconfigured fields.
type validator interface {
	ValidateAll() error
}

// A fieldError for any invalid arguments found in validation.
type fieldError interface {
	Field() string
	Reason() string
}

// validate an input and return an error with details suitable to be returned by a GRPC method.
func validate[ME ~[]error, E fieldError](val validator) error {
	err := val.ValidateAll()
	if err == nil {
		return nil
	}

	st := status.New(codes.InvalidArgument, "invalid request")
	br := &errdetails.BadRequest{}

	//nolint: errorlint,forcetypeassert // Annoying casting of errors needed here as we always know the underlying type.
	errs := err.(ME)
	for _, e := range errs {
		//nolint: errorlint,forcetypeassert // Annoying casting of errors needed here as we always know the underlying type.
		err := e.(E)
		v := &errdetails.BadRequest_FieldViolation{
			Field:       err.Field(),
			Description: err.Reason(),
		}
		br.FieldViolations = append(br.FieldViolations, v)
	}

	st, err = st.WithDetails(br)
	if err != nil {
		panic(err)
	}

	return st.Err()
}

type Service struct {
	stores *alicenet.Stores
}

func NewService(stores *alicenet.Stores) *Service {
	return &Service{
		stores: stores,
	}
}

func (s *Service) ListStores(
	ctx context.Context, req *alicev1.ListStoresRequest) (
	*alicev1.ListStoresResponse, error,
) {
	if err := validate[alicev1.ListStoresRequestMultiError, alicev1.ListStoresRequestValidationError](req); err != nil {
		return nil, err
	}

	stores, err := s.stores.AccountStores.List(ctx, spanner.Key{req.Address}, maxLimit, 0)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.ListStoresResponse{}
	for _, v := range stores {
		resp.Indexes = append(resp.Indexes, v.Index)
	}

	return resp, nil
}

func (s *Service) GetStoreValue(
	ctx context.Context, req *alicev1.GetStoreValueRequest) (
	*alicev1.GetStoreValueResponse, error,
) {
	if err := validate[
		alicev1.GetStoreValueRequestMultiError, alicev1.GetStoreValueRequestValidationError,
	](req); err != nil {
		return nil, err
	}

	value, err := s.stores.AccountStores.Get(ctx, spanner.Key{req.Address, req.Index})
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.GetStoreValueResponse{
		Value:    value.Value,
		IssuedAt: uint32(value.IssuedAt),
	}

	return resp, nil
}

func (s *Service) ListTransactionsForAddress(
	ctx context.Context, req *alicev1.ListTransactionsForAddressRequest) (
	*alicev1.ListTransactionsForAddressResponse, error,
) {
	if err := validate[
		alicev1.ListTransactionsForAddressRequestMultiError,
		alicev1.ListTransactionsForAddressRequestValidationError,
	](req); err != nil {
		return nil, err
	}

	limit := int64(defaultLimit)
	if req.Limit > 0 {
		limit = req.Limit
	}

	transactions, err := s.stores.AccountTransactions.List(ctx, spanner.Key{req.Address}, limit, req.Offset)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp := &alicev1.ListTransactionsForAddressResponse{}

	for _, v := range transactions {
		resp.TransactionHashes = append(resp.TransactionHashes, v.TransactionHash)
	}

	return resp, nil
}

func (s *Service) GetBalance(
	ctx context.Context, req *alicev1.GetBalanceRequest) (
	*alicev1.GetBalanceResponse, error,
) {
	if err := validate[
		alicev1.GetBalanceRequestMultiError,
		alicev1.GetBalanceRequestValidationError,
	](req); err != nil {
		return nil, err
	}

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
	ctx context.Context, req *alicev1.GetTransactionRequest) (
	*alicev1.GetTransactionResponse, error,
) {
	if err := validate[
		alicev1.GetTransactionRequestMultiError,
		alicev1.GetTransactionRequestValidationError,
	](req); err != nil {
		return nil, err
	}

	resp := &alicev1.GetTransactionResponse{}

	txn, err := s.stores.Transactions.Get(ctx, spanner.Key{req.Transaction})
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	resp.Transaction = &alicev1.Transaction{
		Hash:        txn.TransactionHash,
		Height:      uint32(txn.Height),
		ObserveTime: timestamppb.New(txn.ObserveTime),
	}

	inputs, err := s.stores.TransactionInputs.List(ctx, spanner.Key{txn.TransactionHash}, 0, 0)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	for _, input := range inputs {
		newInput := &alicev1.Transaction_Input{
			TransactionHash: input.TransactionHash,
			// ?
			ChainId:                  uint32(input.ChainID),
			ConsumedTransactionHash:  input.ConsumedTransactionHash,
			ConsumedTransactionIndex: input.ConsumedTransactionIndex,
			Signature:                input.Signature,
		}
		resp.Transaction.Inputs = append(resp.Transaction.Inputs, newInput)
	}

	dataStores, err := s.stores.DataStores.List(ctx, spanner.Key{txn.TransactionHash}, 0, 0)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	for _, dataStore := range dataStores {
		newDataStore := &alicev1.Transaction_Output{
			UnspectTransactionOutput: &alicev1.Transaction_Output_DataStore_{
				DataStore: &alicev1.Transaction_Output_DataStore{
					Signature:           dataStore.Signature,
					TransactionHash:     dataStore.TransactionHash,
					ChainId:             uint32(dataStore.ChainID),
					Index:               dataStore.Index,
					IssuedAt:            uint32(dataStore.IssuedAt),
					Deposit:             dataStore.Deposit,
					RawData:             dataStore.RawData,
					TransactionOutIndex: uint32(dataStore.TransactionOutIndex),
					Owner:               dataStore.Owner,
					Fee:                 dataStore.Fee,
				},
			},
		}
		resp.Transaction.Outputs = append(resp.Transaction.Outputs, newDataStore)
	}

	valueStores, err := s.stores.ValueStores.List(ctx, spanner.Key{txn.TransactionHash}, 0, 0)
	if err != nil {
		fmt.Printf("err(%T): %v\n", err, err)

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	for _, valueStore := range valueStores {
		newValueStore := &alicev1.Transaction_Output{
			UnspectTransactionOutput: &alicev1.Transaction_Output_ValueStore_{
				ValueStore: &alicev1.Transaction_Output_ValueStore{
					TransactionHash:     valueStore.TransactionHash,
					ChainId:             uint32(valueStore.ChainID),
					Value:               valueStore.Value,
					TransactionOutIndex: uint32(valueStore.TransactionOutIndex),
					Owner:               valueStore.Owner,
					Fee:                 valueStore.Fee,
				},
			},
		}
		resp.Transaction.Outputs = append(resp.Transaction.Outputs, newValueStore)
	}

	return resp, nil
}

func (s *Service) ListTransactions(
	ctx context.Context, req *alicev1.ListTransactionsRequest) (
	*alicev1.ListTransactionsResponse, error,
) {
	if err := validate[
		alicev1.ListTransactionsRequestMultiError,
		alicev1.ListTransactionsRequestValidationError,
	](req); err != nil {
		return nil, err
	}

	limit := int64(defaultLimit)
	if req.Limit > 0 {
		limit = req.Limit
	}

	txns, err := s.stores.Transactions.List(ctx, nil, limit, req.Offset)
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
	if err := validate[alicev1.GetBlockRequestMultiError, alicev1.GetBlockRequestValidationError](req); err != nil {
		return nil, err
	}

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
			ObserveTime:         timestamppb.New(block.ObserveTime),
		},
	}

	return resp, nil
}

func (s *Service) ListBlocks(
	ctx context.Context, req *alicev1.ListBlocksRequest) (
	*alicev1.ListBlocksResponse, error,
) {
	if err := validate[
		alicev1.ListBlocksRequestMultiError,
		alicev1.ListBlocksRequestValidationError,
	](req); err != nil {
		return nil, err
	}

	limit := int64(defaultLimit)
	if req.Limit > 0 {
		limit = req.Limit
	}

	blocks, err := s.stores.Blocks.List(ctx, nil, limit, req.Offset)
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
