pragma solidity ^0.5.4;

import "../../contracts/external/ITokenStaking.sol";

/// @title Token Staking Stub
/// @dev This contract is for testing purposes only.
contract TokenStakingStub is ITokenStaking {
    mapping(bytes32 => uint256) balance;

    /// @dev Sets balance variable value.
    function setBalance(
        address _operator,
        address _operatorContract,
        uint256 _balance
    ) public {
        bytes32 mapKey = keccak256(
            abi.encodePacked(_operator, _operatorContract)
        );

        balance[mapKey] = _balance;
    }

    /// @dev Returns balance variable value.
    function eligibleStake(address _operator, address _operatorContract)
        public
        view
        returns (uint256)
    {
        bytes32 mapKey = keccak256(
            abi.encodePacked(_operator, _operatorContract)
        );

        return balance[mapKey];
    }
}
