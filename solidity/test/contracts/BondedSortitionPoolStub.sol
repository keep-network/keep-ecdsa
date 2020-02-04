pragma solidity ^0.5.10;

contract BondedSortitionPoolStub {
    address payable[] operators;
    mapping(address => uint256) public operatorWeights;

    function isOperatorRegistered(address operator) public view returns (bool) {
        for (uint256 i = 0; i < operators.length; i++) {
            if (operators[i] == operator) {
                return true;
            }
        }
        return false;
    }

    function insertOperator(address payable operator, uint256 weight) public {
        operators.push(operator);
        operatorWeights[operator] = weight;
    }

    function getOperators() public view returns (address payable[] memory) {
        return operators;
    }
}
