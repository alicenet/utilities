package alicenet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/alicenet/alicenet/proto"
	"github.com/alicenet/indexer/internal/store"
)

// Paths to access various alicenet functionality.
const (
	blockNumberPath      = "v1/get-block-number"
	blockHeaderPath      = "v1/get-block-header"
	minedTransactionPath = "v1/get-mined-transaction"
)

// An APIError returned from alicenet.
type APIError struct {
	Status  string
	Message []byte
}

// Error reported from the API that isn't a low-level socket error.
func (a APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Status, a.Message)
}

// The Interface to working with alicenet.
//
//go:generate go run github.com/golang/mock/mockgen -package mocks -destination ../mocks/alicenet.mockgen.go . Interface
type Interface interface {
	Height(context.Context) (uint32, error)
	BlockHeader(context.Context, uint32) (*proto.BlockHeader, error)
	Transaction(context.Context, string) (*MinedTransactionResponse, error)
}

// Client to interact with alicenet.
// Note: This is a temporary solution as the GRPC endpoint is currently unavailable.
type Client struct {
	baseURL string
}

// Ensure Client matches the package Interface.
var _ Interface = &Client{}

// Connect to AliceNet.
func Connect(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// do handles boilerplate of calling the alicenet local state API.
func do[In, Out any](ctx context.Context, base, path string, request In) (Out, error) {
	url := fmt.Sprintf("https://%s/%s", base, path)

	var out Out

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(request); err != nil {
		return out, fmt.Errorf("encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return out, fmt.Errorf("request: %w", err)
	}

	rawResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("response: %w", err)
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(rawResp.Body)
		err := APIError{Status: rawResp.Status, Message: msg}

		return out, fmt.Errorf("response: %w", err)
	}

	decoder := json.NewDecoder(rawResp.Body)

	if err := decoder.Decode(&out); err != nil {
		return out, fmt.Errorf("response: %w", err)
	}

	return out, nil
}

// Height of chain.
func (c *Client) Height(ctx context.Context) (uint32, error) {
	resp, err := do[
		proto.BlockNumberRequest,
		proto.BlockNumberResponse,
	](ctx, c.baseURL, blockNumberPath, proto.BlockNumberRequest{})
	if err != nil {
		return 0, fmt.Errorf("height: %w", err)
	}

	return resp.BlockHeight, nil
}

// BlockHeader at a given height.
func (c *Client) BlockHeader(ctx context.Context, height uint32) (*proto.BlockHeader, error) {
	resp, err := do[
		proto.BlockHeaderRequest,
		proto.BlockHeaderResponse,
	](ctx, c.baseURL, blockHeaderPath, proto.BlockHeaderRequest{Height: height})
	if err != nil {
		return nil, fmt.Errorf("header: %w", err)
	}

	return resp.BlockHeader, nil
}

// Transaction for a given hash.
func (c *Client) Transaction(ctx context.Context, hash string) (*MinedTransactionResponse, error) {
	resp, err := do[
		proto.MinedTransactionRequest,
		MinedTransactionResponse,
	](ctx, c.baseURL, minedTransactionPath, proto.MinedTransactionRequest{TxHash: hash})
	if err != nil {
		return nil, fmt.Errorf("transaction: %w", err)
	}

	return &resp, nil
}

// MinedTransactionResponse is a hack put in place to allow for json deserialization because the proto models will
// eat the contents of Vout. This will not be necessary after moving to using GRPC clients.
type MinedTransactionResponse struct {
	Tx struct {
		Fee string
		Vin []struct {
			TXInLinker struct {
				TXInPreImage struct {
					ChainID        uint32
					ConsumedTxIdx  uint32
					ConsumedTxHash string
				}
				TxHash string
			}
			Signature string
		}
		Vout []struct {
			DataStore *struct {
				DSLinker struct {
					DSPreImage struct {
						ChainID  uint32
						Index    string
						IssuedAt uint32
						Deposit  string
						RawData  string
						TXOutIdx uint32
						Owner    string
						Fee      string
					}
					TxHash string
				}
				Signature string
			}
			ValueStore *struct {
				VSPreImage struct {
					ChainID  uint32
					Value    string
					TXOutIdx uint32
					Owner    string
					Fee      string
				}
				TxHash string
			}
		}
	}
}

// A Block model for storage in Spanner.
type Block struct {
	ChainID             int64
	Height              int64
	TransactionCount    int64
	PreviousBlockHash   string
	TransactionRootHash string
	StateRootHash       string
	HeaderRootHash      string
	GroupSignatureHash  string
	TransactionHashes   []string
	ObserveTime         time.Time
}

// Key for the Block.
func (b Block) Key() spanner.Key {
	return spanner.Key{b.Height}
}

// Table to store Blocks.
func (Block) Table() string {
	return "Blocks"
}

// List statement for Blocks.
func (Block) List(_ spanner.Key, limit, offset int64) spanner.Statement {
	stmt := spanner.NewStatement("SELECT * FROM Blocks ORDER BY Height DESC LIMIT @limit OFFSET @offset")
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	return stmt
}

// A Transaction model for storage in Spanner.
type Transaction struct {
	Height          int64
	TransactionHash string
	ObserveTime     time.Time
}

// Key for the Transaction.
func (t Transaction) Key() spanner.Key {
	return spanner.Key{t.Height, t.TransactionHash}
}

// Table to store Transactions.
func (Transaction) Table() string {
	return "Transactions"
}

// List statement for Transactions.
func (Transaction) List(_ spanner.Key, limit, offset int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM Transactions ORDER BY Height DESC, TransactionHash DESC LIMIT @limit OFFSET @offset",
	)
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	return stmt
}

// A TransactionInput for storage in Spanner.
type TransactionInput struct {
	TransactionHash          string
	TransactionIndex         int64
	ChainID                  int64
	ConsumedTransactionHash  string
	ConsumedTransactionIndex int64
	Signature                string
	ObserveTime              time.Time
}

// Key for the Transaction.
func (t TransactionInput) Key() spanner.Key {
	return spanner.Key{t.TransactionHash, t.TransactionIndex}
}

// Table to store TransactionInputs.
func (TransactionInput) Table() string {
	return "TransactionInputs"
}

// List statement for TransactionInputs.
func (TransactionInput) List(prefix spanner.Key, _, _ int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM TransactionInputs WHERE TransactionHash = @transactionHash " +
			"ORDER BY TransactionIndex DESC",
	)
	stmt.Params["transactionHash"] = prefix[0]

	return stmt
}

// A ValueStore model to store in Spanner.
type ValueStore struct {
	TransactionHash     string
	ChainID             int64
	Value               string
	TransactionOutIndex int64
	Owner               string
	Fee                 string
	ObserveTime         time.Time
}

// Key for the ValueStores.
func (v ValueStore) Key() spanner.Key {
	return spanner.Key{v.TransactionHash, v.TransactionOutIndex}
}

// Table to store ValueStores.
func (ValueStore) Table() string {
	return "ValueStores"
}

// List statement for ValueStores.
func (ValueStore) List(prefix spanner.Key, _, _ int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM ValueStores WHERE TransactionHash = @transactionHash " +
			"ORDER BY TransactionOutIndex DESC",
	)
	stmt.Params["transactionHash"] = prefix[0]

	return stmt
}

// A DataStore model to store in Spanner.
type DataStore struct {
	Signature           string
	TransactionHash     string
	ChainID             int64
	Index               string
	IssuedAt            int64
	Deposit             string
	RawData             string
	TransactionOutIndex int64
	Owner               string
	Fee                 string
	ObserveTime         time.Time
}

// Key for the DataStores.
func (d DataStore) Key() spanner.Key {
	return spanner.Key{d.TransactionHash, d.TransactionOutIndex}
}

// Table to store DataStores.
func (DataStore) Table() string {
	return "DataStores"
}

// List statement for DataStores.
func (DataStore) List(prefix spanner.Key, _, _ int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM DataStores WHERE TransactionHash = @transactionHash " +
			"ORDER BY TransactionOutIndex DESC",
	)
	stmt.Params["transactionHash"] = prefix[0]

	return stmt
}

// An Account model to store in Spanner.
type Account struct {
	Address string
	Balance string
}

// Key for the Account.
func (a Account) Key() spanner.Key {
	return spanner.Key{a.Address}
}

// Table to store Accounts.
func (Account) Table() string {
	return "Accounts"
}

// List statement for Accounts.
func (Account) List(_ spanner.Key, limit, offset int64) spanner.Statement {
	stmt := spanner.NewStatement("SELECT * FROM Accounts ORDER BY Address LIMIT @limit OFFSET @offset")
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	return stmt
}

// An AccountTransaction model to store in Spanner.
type AccountTransaction struct {
	Address         string
	TransactionHash string
	ObserveTime     time.Time
}

// Key for the AccountTransaction.
func (a AccountTransaction) Key() spanner.Key {
	return spanner.Key{a.Address, a.TransactionHash}
}

// Table to store AccountTransactions.
func (AccountTransaction) Table() string {
	return "AccountTransactions"
}

// List statement for AccountTransactions.
func (AccountTransaction) List(prefix spanner.Key, limit, offset int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM AccountTransactions WHERE Address = @address ORDER BY TransactionHash LIMIT @limit OFFSET @offset")
	stmt.Params["address"] = prefix[0]
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	return stmt
}

// An AccountStore model to store in Spanner.
type AccountStore struct {
	Address     string
	Index       string
	IssuedAt    int64
	Value       string
	ObserveTime time.Time
}

// Key for the AccountStore.
func (a AccountStore) Key() spanner.Key {
	return spanner.Key{a.Address, a.Index}
}

// Table to store AccountStores.
func (AccountStore) Table() string {
	return "AccountStores"
}

// List statement for AccountStores.
func (AccountStore) List(prefix spanner.Key, limit, offset int64) spanner.Statement {
	stmt := spanner.NewStatement(
		"SELECT * FROM AccountStores WHERE Address = @address ORDER BY Index LIMIT @limit OFFSET @offset")
	stmt.Params["address"] = prefix[0]
	stmt.Params["limit"] = limit
	stmt.Params["offset"] = offset

	return stmt
}

// Stores is a collection of all alicenet Storable objects.
type Stores struct {
	Blocks              store.Store[Block]
	Transactions        store.Store[Transaction]
	TransactionInputs   store.Store[TransactionInput]
	DataStores          store.Store[DataStore]
	ValueStores         store.Store[ValueStore]
	Accounts            store.Store[Account]
	AccountTransactions store.Store[AccountTransaction]
	AccountStores       store.Store[AccountStore]
}

// InSpanner storage of all alicenet resources.
func InSpanner(client *spanner.Client) *Stores {
	return &Stores{
		Blocks:              store.InSpanner[Block](client),
		Transactions:        store.InSpanner[Transaction](client),
		TransactionInputs:   store.InSpanner[TransactionInput](client),
		DataStores:          store.InSpanner[DataStore](client),
		ValueStores:         store.InSpanner[ValueStore](client),
		Accounts:            store.InSpanner[Account](client),
		AccountTransactions: store.InSpanner[AccountTransaction](client),
		AccountStores:       store.InSpanner[AccountStore](client),
	}
}
