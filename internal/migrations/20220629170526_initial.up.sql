CREATE TABLE Blocks (
    ChainID             INT64 NOT NULL,
    Height              INT64 NOT NULL,
    TransactionCount    INT64 NOT NULL,
    PreviousBlockHash   STRING(MAX) NOT NULL,
    TransactionRootHash STRING(MAX) NOT NULL,
    StateRootHash       STRING(MAX) NOT NULL,
    HeaderRootHash      STRING(MAX) NOT NULL,
    GroupSignatureHash  STRING(MAX) NOT NULL,
    TransactionHashes   ARRAY<STRING(MAX)>,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (Height DESC);

CREATE TABLE Transactions (
    Height          INT64 NOT NULL,
    TransactionHash STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (TransactionHash);

CREATE TABLE TransactionInputs (
    TransactionHash          STRING(MAX) NOT NULL,
    TransactionIndex         INT64 NOT NULL,
    ChainID                  INT64 NOT NULL,
    ConsumedTransactionHash  STRING(MAX) NOT NULL,
    ConsumedTransactionIndex INT64 NOT NULL,
    Signature                STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (TransactionHash, TransactionIndex),
  INTERLEAVE IN PARENT Transactions ON DELETE CASCADE;

CREATE TABLE AtomicSwaps (
    TransactionHash     STRING(MAX) NOT NULL,
    ChainID             INT64 NOT NULL,
    Value               STRING(MAX) NOT NULL,
    TransactionOutIndex INT64 NOT NULL,
    IssuedAt            INT64 NOT NULL,
    Exp                 INT64 NOT NULL,
    Owner               STRING(MAX) NOT NULL,
    Fee                 STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (TransactionHash, TransactionOutIndex),
  INTERLEAVE IN PARENT Transactions ON DELETE CASCADE;

CREATE TABLE ValueStores (
    TransactionHash     STRING(MAX) NOT NULL,
    ChainID             INT64 NOT NULL,
    Value               STRING(MAX) NOT NULL,
    TransactionOutIndex INT64 NOT NULL,
    Owner               STRING(MAX) NOT NULL,
    Fee                 STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (TransactionHash, TransactionOutIndex),
  INTERLEAVE IN PARENT Transactions ON DELETE CASCADE;

CREATE TABLE DataStores (
    Signature           STRING(MAX) NOT NULL,
    TransactionHash     STRING(MAX) NOT NULL,
    ChainID             INT64 NOT NULL,
    Index               STRING(MAX) NOT NULL,
    IssuedAt            INT64 NOT NULL,
    Deposit             STRING(MAX) NOT NULL,
    RawData             STRING(MAX) NOT NULL,
    TransactionOutIndex INT64 NOT NULL,
    Owner               STRING(MAX) NOT NULL,
    Fee                 STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (TransactionHash, TransactionOutIndex),
  INTERLEAVE IN PARENT Transactions ON DELETE CASCADE;

CREATE TABLE Accounts (
    Address STRING(MAX) NOT NULL,
    Balance STRING(MAX) NOT NULL,
) PRIMARY KEY (Address);

CREATE TABLE AccountTransactions (
    Address         STRING(MAX) NOT NULL,
    TransactionHash STRING(MAX) NOT NULL,
    ObserveTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (Address, TransactionHash),
  INTERLEAVE IN PARENT Accounts ON DELETE CASCADE;
