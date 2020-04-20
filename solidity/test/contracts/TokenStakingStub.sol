pragma solidity 0.5.17;

import "@keep-network/sortition-pools/contracts/api/IStaking.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";


/// @title Token Staking Stub
/// @dev This contract is for testing purposes only.
contract TokenStakingStub is IStaking {
    using SafeMath for uint256;

    uint256 public minimumStake = 200000 * 1e18;

    mapping(address => address payable) operatorToMagpie;

    mapping(address => uint256) stakes;

    mapping(address => int256) public operatorLocks;

    // Authorized operator contracts.
    mapping(address => mapping(address => bool)) internal authorizations;

    // Map of operator -> owner.
    mapping(address => address) owners;

    address public delegatedAuthority;

    bool slashingShouldFail;

    function setSlashingShouldFail(bool _shouldFail) public {
        slashingShouldFail = _shouldFail;
    }

    /// @dev Sets balance variable value.
    function setBalance(address _operator, uint256 _balance) public {
        stakes[_operator] = _balance;
    }

    /// @dev Returns balance variable value.
    function eligibleStake(address _operator, address)
        public
        view
        returns (uint256)
    {
        return stakes[_operator];
    }

    function setMagpie(address _operator, address payable _magpie) public {
        operatorToMagpie[_operator] = _magpie;
    }

    function magpieOf(address _operator) public view returns (address payable) {
        address payable magpie = operatorToMagpie[_operator];
        if (magpie == address(0)) {
            return address(uint160(_operator));
        }
        return magpie;
    }

    function slash(uint256 _amount, address[] memory _misbehavedOperators)
        public
    {
        if (slashingShouldFail) {
            // THIS SHOULD NEVER HAPPEN WITH REAL TOKEN STAKING
            revert("slashing failed");
        }
        for (uint256 i = 0; i < _misbehavedOperators.length; i++) {
            address operator = _misbehavedOperators[i];
            stakes[operator] = stakes[operator].sub(_amount);
        }
    }

    function lockStake(address operator, uint256 duration) public {
        operatorLocks[operator] = int256(duration);
    }

    function unlockStake(address operator) public {
        // We set it to negative value to be sure in tests that the function is
        // actually called and not just default `0` value is returned.
        operatorLocks[operator] = -1;
    }

    function authorizeOperatorContract(
        address _operator,
        address _operatorContract
    ) public {
        authorizations[_operatorContract][_operator] = true;
    }

    function isAuthorizedForOperator(
        address _operator,
        address _operatorContract
    ) public view returns (bool) {
        return authorizations[_operatorContract][_operator];
    }

    function authorizerOf(address _operator) public view returns (address) {
        return _operator;
    }

    function setOwner(address _operator, address _owner) public {
        owners[_operator] = _owner;
    }

    function ownerOf(address _operator) public view returns (address) {
        return owners[_operator];
    }

    function claimDelegatedAuthority(address delegatedAuthoritySource) public {
        delegatedAuthority = delegatedAuthoritySource;
    }
}
