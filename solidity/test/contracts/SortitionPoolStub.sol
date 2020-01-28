pragma solidity ^0.5.10;

contract SortitionPoolStub {

    address payable[] operators;

    function selectGroup(uint256 groupSize, bytes32 seed) public view returns (address payable[] memory members)  {
        seed;

        members = new address payable[](groupSize);

        uint nextIndex = 0;
        for (uint i = 0; i < groupSize; i++) {
            members[i] = operators[nextIndex];
            nextIndex++;
            nextIndex = nextIndex % operators.length;
        }
    }

    function insertOperator(address payable operator, uint weight) public {
        weight;

        operators.push(operator);
    }

    function getOperators() public view returns (address payable[] memory) {
        return operators;
    }
}
