:construction: Toward building a chain configuration standard that can be implemented by
any Ethereum-ready client to describe and configure chain parameters.

#### Develop

```shell
./parity/chainspec-validation-test.sh
```

#### Why

The two most popular clients in the Ethereum ecosystem, [Parity](https://github.com/paritytech/parity-ethereum) and [Geth](https://github.com/ethereum/go-ethereum) use different patterns for external chain and network definitions. This is annoying. 

Without a standardized way to talk about chain configurations that can be understood by at least these two major clients, all cross-client and cross-network interfaces are limited to one-off solutions and/or canonical-only chain configurations.

This is, at least, a massive limitation for testing and development. For example, the canonical [ethererum/tests](http://github.com/ethereum/tests) suite which is run by both Geth and Parity, rely entirely on hardcoded, flat, and opaque chain definitions:

- https://github.com/ethereum/tests/blob/develop/JSONSchema/st-schema.json#L212-L236
- https://github.com/ethereum/tests/blob/develop/GeneralStateTests/stArgsZeroOneBalance/addNonConst.json#L18-L19

... which limits the validity and relevance of these tests enormously.

With a spec available for this configuration, a door opens to including x-chain configuration data in these tests, which would extend their relevance and applicability beyond their current ETH-only opinion. 


#### What

A standardized way to describe Ethereum-ecosystem chain configuration.

#### How

Develop and document a spec here, then propose via EIP and ECIP (et al?) channels.



#### Resources

- https://wiki.parity.io/Chain-specification
- https://github.com/keorn/parity-spec
- https://github.com/5chdn/crossclient-chainspec
-
