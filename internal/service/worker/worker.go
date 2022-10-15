package worker

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"cloud.google.com/go/spanner"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"

	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/utilities/internal/alicenet"
	"github.com/alicenet/utilities/internal/logz"
)

// Wait time between checks against alicenet.
const (
	loopWait = 5 * time.Second
	baseHex  = 16
)

var (
	//nolint:gochecknoglobals // Stats exempt
	highestBlock = stats.Int64("highest_block", "The highest seen block", "1")
	//nolint:gochecknoglobals // Stats exempt
	currentBlock = stats.Int64("current_block", "The current processed block", "1")
	//nolint:gochecknoglobals // Stats exempt
	views = []*view.View{
		{
			Name:        "highest_block_last",
			Measure:     highestBlock,
			Description: "The highest block seen",
			Aggregation: view.LastValue(),
		},
		{
			Name:        "current_block_last",
			Measure:     currentBlock,
			Description: "The current block seen",
			Aggregation: view.LastValue(),
		},
		{
			Name:        "blocks_count",
			Measure:     currentBlock,
			Description: "The number of blocks processed",
			Aggregation: view.Count(),
		},
	}
	//nolint:gochecknoglobals // Stats exempt
	setupStats sync.Once
)

// ParseError indicates a big.Int could not be parsed.
type ParseError string

// Error detailing what couldn't be parsed.
func (p ParseError) Error() string {
	return "could not parse: " + string(p)
}

// A Service that will periodically check alicenet for latest blocks and add them to the index.
type Service struct {
	stores  *alicenet.Stores
	client  alicenet.Interface
	highest int
}

// New Service from an alicenet client and stores.
func New(client alicenet.Interface, stores *alicenet.Stores) *Service {
	setupStats.Do(func() {
		for i := range views {
			if err := view.Register(views[i]); err != nil {
				panic(err)
			}
		}
	})

	return &Service{
		client:  client,
		stores:  stores,
		highest: 1,
	}
}

// Run the service.
func (s *Service) Run(ctx context.Context) {
	for {
		if err := s.process(ctx); err != nil {
			logz.WithDetail("err", err).Errorf("run error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(loopWait):
		}
	}
}

// process any new blocks found in alicenet.
func (s *Service) process(ctx context.Context) error {
	current, err := s.client.Height(ctx)
	if err != nil {
		return fmt.Errorf("processing: %w", err)
	}

	logz.WithDetails(logz.Details{"current": current, "highest": s.highest}).Info()
	stats.Record(ctx, highestBlock.M(int64(current)))

	for height := s.highest; height <= int(current); height++ {
		stats.Record(ctx, currentBlock.M(int64(height)))

		blockHeader, err := s.client.BlockHeader(ctx, uint32(height))
		if err != nil {
			return fmt.Errorf("processing: %w", err)
		}

		if err := s.pushBlock(ctx, blockHeader); err != nil {
			return err
		}

		for _, hash := range blockHeader.TxHshLst {
			txn, err := s.client.Transaction(ctx, hash)
			if err != nil {
				// Transaction has likely been purged from the chain. Mark it as missing and continue.
				logz.WithDetail("hash", hash).Warning("transaction missing, continuing")

				if err := s.pushMissingTransaction(ctx, height, hash); err != nil {
					return err
				}

				continue
			}

			if err := s.pushTransaction(ctx, height, hash, txn); err != nil {
				return err
			}
		}

		s.highest = height
	}

	return nil
}

// pushBlock to the permanent stores.
func (s *Service) pushBlock(ctx context.Context, blockHeader *proto.BlockHeader) error {
	logz.WithDetail("header", blockHeader).Info("got header")

	block := createBlock(blockHeader)

	if err := s.stores.Blocks.Insert(ctx, block); err != nil {
		return fmt.Errorf("pushing block: %w", err)
	}

	return nil
}

// createBlock from a Blockheader.
func createBlock(blockHeader *proto.BlockHeader) alicenet.Block {
	block := alicenet.Block{
		ChainID:             int64(blockHeader.BClaims.ChainID),
		Height:              int64(blockHeader.BClaims.Height),
		TransactionCount:    int64(blockHeader.BClaims.TxCount),
		PreviousBlockHash:   blockHeader.BClaims.PrevBlock,
		TransactionRootHash: blockHeader.BClaims.TxRoot,
		StateRootHash:       blockHeader.BClaims.StateRoot,
		HeaderRootHash:      blockHeader.BClaims.HeaderRoot,
		GroupSignatureHash:  blockHeader.SigGroup,
		TransactionHashes:   blockHeader.TxHshLst,
		ObserveTime:         spanner.CommitTimestamp,
	}

	return block
}

// pushTransaction to the permanent stores.
func (s *Service) pushTransaction(
	ctx context.Context,
	height int,
	hash string,
	txn *alicenet.MinedTransactionResponse,
) error {
	logz.WithDetail("transaction", txn).Info("got transaction")

	newTx := alicenet.Transaction{
		Height:          int64(height),
		TransactionHash: hash,
		ObserveTime:     spanner.CommitTimestamp,
	}

	if err := s.stores.Transactions.Insert(ctx, newTx); err != nil {
		return fmt.Errorf("pushing transaction: %w", err)
	}

	if err := s.pushTransactionInput(ctx, txn); err != nil {
		return fmt.Errorf("pushing transaction: %w", err)
	}

	if err := s.pushTransactionOutput(ctx, txn); err != nil {
		return fmt.Errorf("pushing transaction: %w", err)
	}

	return nil
}

// pushTransaction to the permanent stores.
func (s *Service) pushMissingTransaction(
	ctx context.Context,
	height int,
	hash string,
) error {
	logz.WithDetail("hash", hash).Info("writing missing transaction")

	missing := true

	newTx := alicenet.Transaction{
		Height:          int64(height),
		TransactionHash: hash,
		ObserveTime:     spanner.CommitTimestamp,
		Missing:         &missing,
	}

	if err := s.stores.Transactions.Insert(ctx, newTx); err != nil {
		return fmt.Errorf("pushing transaction: %w", err)
	}

	return nil
}

// pushTransactionOutput to permanent stores.
func (s *Service) pushTransactionOutput(ctx context.Context, txn *alicenet.MinedTransactionResponse) error {
	for _, vout := range txn.Tx.Vout {
		switch {
		case vout.DataStore != nil:
			output := alicenet.DataStore{
				Signature:           vout.DataStore.Signature,
				TransactionHash:     vout.DataStore.DSLinker.TxHash,
				ChainID:             int64(vout.DataStore.DSLinker.DSPreImage.ChainID),
				Index:               vout.DataStore.DSLinker.DSPreImage.Index,
				IssuedAt:            int64(vout.DataStore.DSLinker.DSPreImage.IssuedAt),
				Deposit:             vout.DataStore.DSLinker.DSPreImage.Deposit,
				RawData:             vout.DataStore.DSLinker.DSPreImage.RawData,
				TransactionOutIndex: int64(vout.DataStore.DSLinker.DSPreImage.TXOutIdx),
				Owner:               vout.DataStore.DSLinker.DSPreImage.Owner,
				Fee:                 vout.DataStore.DSLinker.DSPreImage.Fee,
				ObserveTime:         spanner.CommitTimestamp,
			}
			if err := s.stores.DataStores.Insert(ctx, output); err != nil {
				return fmt.Errorf("output: %w", err)
			}

			if err := s.pushAccount(
				ctx, vout.DataStore.DSLinker.DSPreImage.Owner,
				vout.DataStore.DSLinker.TxHash,
				"0",
			); err != nil {
				return fmt.Errorf("output: %w", err)
			}

			if err := s.pushStoredData(
				ctx,
				vout.DataStore.DSLinker.DSPreImage.Owner,
				vout.DataStore.DSLinker.DSPreImage.Index,
				int64(vout.DataStore.DSLinker.DSPreImage.IssuedAt),
				vout.DataStore.DSLinker.DSPreImage.RawData); err != nil {
				return fmt.Errorf("output: %w", err)
			}
		case vout.ValueStore != nil:
			output := alicenet.ValueStore{
				TransactionHash:     vout.ValueStore.TxHash,
				ChainID:             int64(vout.ValueStore.VSPreImage.ChainID),
				Value:               vout.ValueStore.VSPreImage.Value,
				TransactionOutIndex: int64(vout.ValueStore.VSPreImage.TXOutIdx),
				Owner:               vout.ValueStore.VSPreImage.Owner,
				Fee:                 vout.ValueStore.VSPreImage.Fee,
				ObserveTime:         spanner.CommitTimestamp,
			}
			if err := s.stores.ValueStores.Insert(ctx, output); err != nil {
				return fmt.Errorf("output: %w", err)
			}

			if err := s.pushAccount(
				ctx,
				vout.ValueStore.VSPreImage.Owner,
				vout.ValueStore.TxHash,
				vout.ValueStore.VSPreImage.Value,
			); err != nil {
				return fmt.Errorf("output: %w", err)
			}
		}
	}

	return nil
}

// pushAccount to permanent stores. Will associate transaction and amount stored.
func (s *Service) pushAccount(ctx context.Context, owner, hash, amount string) error {
	account, err := s.stores.Accounts.Get(ctx, spanner.Key{owner})
	if err != nil {
		account = alicenet.Account{
			Address: owner,
			Balance: "0",
		}
	}

	current, success := new(big.Int).SetString(account.Balance, baseHex)
	if !success {
		return ParseError(account.Balance)
	}

	added, success := new(big.Int).SetString(amount, baseHex)
	if !success {
		return ParseError(amount)
	}

	total := new(big.Int).Add(current, added)
	account.Balance = total.Text(baseHex)

	if err := s.stores.Accounts.Insert(ctx, account); err != nil {
		return fmt.Errorf("account: %w", err)
	}

	txn := alicenet.AccountTransaction{
		Address:         owner,
		TransactionHash: hash,
		ObserveTime:     spanner.CommitTimestamp,
	}

	if err := s.stores.AccountTransactions.Insert(ctx, txn); err != nil {
		return fmt.Errorf("account: %w", err)
	}

	return nil
}

// pushStoredData to permanent stores.
func (s *Service) pushStoredData(ctx context.Context, owner, index string, issuedAt int64, value string) error {
	accountStore := alicenet.AccountStore{
		Address:     owner,
		Index:       index,
		IssuedAt:    issuedAt,
		Value:       value,
		ObserveTime: spanner.CommitTimestamp,
	}

	if err := s.stores.AccountStores.Insert(ctx, accountStore); err != nil {
		return fmt.Errorf("account store: %w", err)
	}

	return nil
}

// pushTransactionInput to the permanent stores.
func (s *Service) pushTransactionInput(ctx context.Context, txn *alicenet.MinedTransactionResponse) error {
	for index, input := range txn.Tx.Vin {
		input := alicenet.TransactionInput{
			TransactionHash:          input.TXInLinker.TxHash,
			TransactionIndex:         int64(index),
			ChainID:                  int64(input.TXInLinker.TXInPreImage.ChainID),
			ConsumedTransactionHash:  input.TXInLinker.TXInPreImage.ConsumedTxHash,
			ConsumedTransactionIndex: int64(input.TXInLinker.TXInPreImage.ConsumedTxIdx),
			Signature:                input.Signature,
			ObserveTime:              spanner.CommitTimestamp,
		}

		if err := s.stores.TransactionInputs.Insert(ctx, input); err != nil {
			return fmt.Errorf("input: %w", err)
		}
	}

	return nil
}
