# AliceNet Indexer

[![GitHub Release](https://img.shields.io/github/v/release/alicenet/indexer)](https://github.com/alicenet/indexer/releases)
[![License](https://img.shields.io/github/license/alicenet/indexer)](./LICENSE)
[![codecov](https://codecov.io/gh/alicenet/indexer/branch/main/graph/badge.svg?token=GSJJFZB9WV)](https://codecov.io/gh/alicenet/indexer)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/bkhW2KUWDu)

AliceNet is a Proof-Of-Stake, UTXO based blockchain written in golang that enables high-speed bridging between Layer 1 and Layer 2 protocols while emphasizing strong security and identity standards.

The Indexer is designed to process all events on every layer for AliceNet and make it available for easy access, similar to [Etherscan](https://etherscan.io) does for Ethereum.

To learn more, check out our official [website](https://www.alice.net/) and join our official [Discord community](https://discord.gg/bkhW2KUWDu).

## Indexer Components

### Worker

The worker is designed to run continuously and poll the state of AliceNet on all layers.
As events are detected, they are processed and stored in long-term storage (currently
limited to just [Google Cloud Spanner](https://cloud.google.com/spanner)).

### Frontend

The frontend runs a combination GRPC/REST endpoint that can be called to return the
information stored by the worker.

## How to contribute

AliceNet is still under development and contributions are always welcome! Please make sure to check [Contributing](./CONTRIBUTING.md) if you want to help.

## License

AliceNet is licensed under [MIT license](./LICENSE).
