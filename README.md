# Rounded census algorithm

![header](https://s6.imgcdn.dev/fNl6e.png)

## Context
At Vocdoni we are building a solution that makes it easy to create elections and cast votes for a variety of use cases. We support several types of elections, including anonymous voting using cryptographic techniques such as zero knowledge proofs (thanks to zk-snarks).

Thanks to the [Census3](https://github.com/vocdoni/census3) service, we provide a mechanism to create censuses for these elections based on the holders of a certain token or tokens (creating strategies to combine them), bringing support for use cases of the web3 world. Our technology ensures that the identity of the holders (their addresses) remains private during the voting process.

## The problem
In many cases, token holders can be identified not only by their address, but also by their account balances. This information must therefore be protected or obscured in order to prevent a voter's identity from being revealed.

## Our solution
To achieve this, we developed this Go package which implements an algorithm to create groups of holders with similar balances with at least 3 holders. This algorithm tries to hide the balances that uniquely identify the holders by including them in a group with the same balance (losing a percentage of their voting power).

The algorithm also calculates the accuracy of the resulting census by comparing the sum of its balances with the original one.

The algorithm is not perfect and cannot hide all token holders, it depends on the distribution of their balances. To avoid a huge loss of accuracy, it detects outliers before calculating the groups. The outliers keep their original balance, so they remain identifiable.

## Initial results
Checkout our initial results in [./tests](./tests) folder.