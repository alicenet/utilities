# AliceNet Utilities

[![GitHub Release](https://img.shields.io/github/v/release/alicenet/utilities)](https://github.com/alicenet/utilities/releases)
[![License](https://img.shields.io/github/license/alicenet/utilities)](./LICENSE)
[![codecov](https://codecov.io/gh/alicenet/utilities/branch/main/graph/badge.svg?token=GSJJFZB9WV)](https://codecov.io/gh/alicenet/utilities)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/bkhW2KUWDu)

AliceNet is a Proof-Of-Stake, UTXO based blockchain written in golang that enables high-speed bridging between Layer 1 and Layer 2 protocols while emphasizing strong security and identity standards.

This repository provides a number of utilities helpful for working with/hosting AliceNet.

To learn more, check out our official [website](https://www.alice.net/) and join our official [Discord community](https://discord.gg/bkhW2KUWDu).

## Indexer

The Indexer is designed to process all events on every layer for AliceNet and make it available for easy access, similar to [Etherscan](https://etherscan.io) does for Ethereum.

### Worker

The worker is designed to run continuously and poll the state of AliceNet on all layers.
As events are detected, they are processed and stored in long-term storage (currently
limited to just [Google Cloud Spanner](https://cloud.google.com/spanner)).

### Frontend

The frontend runs a combination GRPC/REST endpoint that can be called to return the
information stored by the worker.

## JSON RPC Proxy

A container image that will proxy JSON RPC requests to a remote path (not just host). This allows for hosting a proxy that will include an account key for a service such as [infura](https://infura.io).

## How to contribute

AliceNet is still under development and contributions are always welcome! Please make sure to check [Contributing](./CONTRIBUTING.md) if you want to help.

## License

AliceNet is licensed under [MIT license](./LICENSE).
