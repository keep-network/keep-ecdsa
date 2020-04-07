pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/KeepToken.sol";


// This file contains a workaround for a truffle bug described in [truffle#1250].
// We are not able to directly require artifacts from other packages, e.g.:
// `artifacts.require("@keep-network/keep-core/KeepTokenIntegration")`.
// We first need to import the contracts so they are compiled and placed in
// `build/contracts/` directory.
//
// [truffle#1250]:https://github.com/trufflesuite/truffle/issues/1250
contract KeepTokenIntegration is KeepToken {

}
