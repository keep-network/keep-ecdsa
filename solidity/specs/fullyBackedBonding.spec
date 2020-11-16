/*  Declaration of methods that are used in the rules.
    envfree indicate that the method is not dependent on the environment (msg.value, msg.sender)
    methods that are not declared here are assumed to be dependent on env
*/
methods {
    withdraw(uint256,address)
    deposit(address)
    beneficiaryOf(address) returns address envfree
    ownerOf(address) returns address envfree
    authorizerOf(address) returns address envfree
    unbondedValue(address) returns uint256 envfree
    bondAmount(address, address, uint256) returns uint256 envfree
    balanceOf(address) returns uint256 envfree
    getDelegatedAuthority(address) returns address envfree
    everDeposited(address) returns uint256 envfree
    init_state() envfree
    createBond(address, address, uint256, uint256, address)
    otherBeneficiary() returns address envfree
}

// a few macros
definition MAXINT() returns uint256 =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;

definition safeAdd(uint256 x, uint256 y) returns bool =
        x + y >= y && x + y >=x && x+y <= MAXINT();

/* A ghost can be thought as additional variable that is in the background */
// A mapping from bondID to operator. e.g. mapping(uint256 => address)
ghost bondIDToOperator(uint256) returns address;
// A mapping from operator to the sum of all bonds e.g. mapping(address => uint256)
ghost totalLockedBonds(uint256) returns uint256;
// total amount of lockedBonds of all operators
ghost allLockedBond() returns uint256;
// total amount of unbounded of all operators
ghost allUnbounded() returns uint256;

/* update to bondIdToOperator is on updates to operatorHolderRefToBondID
    e.g. after,
    operatorHolderRefToBondID[operatorY][holder][referenceID] = bondID;
    the following is simulated:
    bondIDToOperator[bondID] = operatorY;
*/
hook Sload uint bondID operatorHolderRefToBondID[KEY uint operatorY][KEY uint holder][KEY uint referenceID] STORAGE {
    havoc bondIDToOperator assuming bondIDToOperator@new(bondID) == operatorY &&
            (forall uint otherBondID. otherBondID != bondID => bondIDToOperator@new(otherBondID) == bondIDToOperator@old(otherBondID));
}

/* update to lockedBonds is on updates to lockedBonds based on the old value of both lockedBonds and totalLockedBonds
    e.g. after,
    lockedBonds[bondID] = x ;
    the following is simulated:
    totalLockedBonds[bondIDToOperator[bondID]] = new_value;
    ....
*/
hook Sstore lockedBonds[KEY uint bondID] uint value (uint old_value) STORAGE {
    uint256 operatorX = bondIDToOperator(bondID);
    havoc totalLockedBonds assuming totalLockedBonds@new(operatorX) == totalLockedBonds@old(operatorX) + value - old_value &&
            (forall uint256 otherOperator. otherOperator != operatorX => totalLockedBonds@new(otherOperator) == totalLockedBonds@old(otherOperator));
    havoc allLockedBond assuming allLockedBond@new() == allLockedBond@old() + value - old_value;
}

hook Sstore unbondedValue[KEY uint operatorY] uint value (uint old_value) STORAGE {
    havoc allUnbounded assuming allUnbounded@new() == allUnbounded@old() + value - old_value;
}

/**
@title Valid Operator ✔️
@notice Zero cannot be a operator
       beneficiaryOf(o) ≠ 0  ⇔
          ( o ≠ 0 ⋀ ownerOf(o) ≠ 0  ⋀ authorizerOf(o) ≠ 0 )


*/
invariant validOperator(address operator)
        beneficiaryOf(operator) != 0  <=>  ( operator!=0 && ownerOf(operator) != 0  && authorizerOf(operator) !=0 )

/**
@title Valid state of a operator ❌
@notice Operator with assets must have an owner, beneficiary, and an authorizer

    (unbondedValue(o) + lockedBonds(o)) > 0 ⟹
	    ( ownerOf(o) ≠ 0 ⋀ beneficiaryOf(o) ≠ 0 ⋀ authorizerOf(o) ≠ 0 )
*/
/*
invariant validStateOfOperator(address operator)
        (unbondedValue(operator)  +  totalLockedBonds(operator) > 0) =>  beneficiaryOf(operator) != 0
        {
        preserved freeBond(env e, address _operator, uint256 _referenceID)  {
             require _operator == operator &&
             totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, _referenceID)  &&
             (unbondedValue(operator)  +  totalLockedBonds(operator) > 0) ;
        }
        preserved init_state() {
            require totalLockedBonds(operator) == 0;
            require unbondedValue(operator) == 0;
        }
}
*/
//todo - move to the invariant when preserved in working
rule validState(address operator, method f) {
     env e;
     calldataarg args;
     require unbondedValue(operator) + totalLockedBonds(operator) > 0 => (ownerOf(operator) != 0 && beneficiaryOf(operator) != 0 );
     require e.msg.sender != 0 ;
     assert unbondedValue(operator) + totalLockedBonds(operator) > 0 => (ownerOf(operator) != 0 && beneficiaryOf(operator) != 0 );
}

/**
@title: No bankruptcy of the system (no money lost)  ✔*
@notice The balance of the system is more than the obligations
(the total assets deposited in the system and is in either unbounded or in a locked state )

     ethSystemBalance(a) ≥  sum for all operator o. ( unbondedValue(o) + lockedBonds(o) )
*/
//todo - move to invariant when preserved in working
rule noBankruptcyAsRule(method f) {
    env e;
    calldataarg args;
    require e.msg.sender != currentContract;
    require balanceOf(currentContract) >= allUnbounded() + allLockedBond();
    sinvoke f(e,args);
    assert balanceOf(currentContract) >= allUnbounded() + allLockedBond();
}


/**
@title User can not gain  assets ✔*
@notice The total assets of operator can not be more than ever deposited to operator
        everDeposited(o) ≥ (unbondedValue(o) + lockedBonds(o))

*/
/*
invariant assetsLessThanEverDeposited(address operator, address holder, uint256 referenceID )
        everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator) &&
        safeAdd(unbondedValue(operator), totalLockedBonds(operator))   {

        preserved createBond(env e, address _operator, address _holder, uint256 _referenceID,
                        uint256 amount, address authorizedSortitionPool) {
            require _operator == operator && _holder == holder && _referenceID == referenceID;
        }

        preserved reassignBond(env e, address _operator, uint256 _referenceID,
                        address newHolder, uint256 newReferenceID) {
            require _operator == operator &&
            totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, _referenceID) ;

        }
        preserved seizeBond(env e, address _operator, uint256 _referenceID, uint256 amount,
                        address destination) {
            require _operator == operator &&
            totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, _referenceID) ;
        }
        preserved freeBond(env e, address _operator, uint256 _referenceID)  {
            require _operator == operator &&
            totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, _referenceID) ;
        }
}
*/
//todo - move to invariant when preserved in working
rule assetsLessThanEverDepositedAsRule(address operator, method f, mathint t,
                    address holder, uint256 referenceID ) {
    env e;
    calldataarg args;
    require  t == totalLockedBonds(operator);
    require totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, referenceID);
    require safeAdd(unbondedValue(operator), totalLockedBonds(operator));
    require everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
    require f.selector==init_state().selector => totalLockedBonds(operator)==0;
    uint256 before =  bondAmount(operator, e.msg.sender, referenceID);
    sinvoke f(e, args);
    require (f.selector==freeBond(address,uint256).selector ||
             f.selector==seizeBond(address,uint256,uint256,address).selector  ||
             f.selector==createBond(address,address,uint256,uint256,address).selector ||
             f.selector==reassignBond(address, uint256, address, uint256).selector )
             => before !=  bondAmount(operator, e.msg.sender, referenceID);
    assert safeAdd(unbondedValue(operator), totalLockedBonds(operator)) &&
            everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
   }

/**
@title  No cyclic authorization
@notice There is  no cycle in delegating authority, as it can cause denial of service

        ㄱ (delegatedAuthority(a) = b ⋀ delegatedAuthority(b) = c ⋀ delegatedAuthority(c) = a )

    This rule is violated and is an excepted behavior
*/
/*
invariant noCyclicDelegatedAuthority(address a, address b, address c)
     (a!=0 && b!=0 && c!=0 => !(getDelegatedAuthority(a)==b && getDelegatedAuthority(b)==c && getDelegatedAuthority(c)==a)) &&
     getDelegatedAuthority(0) == 0
*/


/**
@title Total assets of a user is the sum of assets within the system (either locked or unbounded) and outside the system

@notice The total assets of an operator within the system is preserved, except on deposits and on withdrawal and seizing.
    {  b = unbondedValue(o) + lockedBonds(o) }
        op(u, x)
    { b =  unbondedValue(o) + lockedBonds(o) + x}

    Note that this property holds for every successful operation performed by any user or holder, u,
    possibly sending in x amount of wai (msg.value), except for sizeBond and withdraw which are defined differently:
*/

rule totalAssetsPreserved(address operator, address owner, address beneficiary,  method f) {
    env e;
    calldataarg args;
    mathint totalAssetsBefore = totalLockedBonds(operator) + unbondedValue(operator) ;
    require f.selector != seizeBond(address,uint256,uint256,address).selector &&
             f.selector != withdraw(uint256,address).selector;
    sinvoke f(e, args);
    mathint totalAssetsAfter = totalLockedBonds(operator) + unbondedValue(operator);
    assert  totalAssetsAfter == totalAssetsBefore ||  totalAssetsAfter == totalAssetsBefore + e.msg.value,
    "$f can change the total assets of operator in an unexpected way";
}

/**
On withdraw, the total assets within the system of operator  and the balance of the beneficiary is perserved

{  b = unbondedValue(o) + lockedBonds(o) + beneficiaryOf(o).balance}
	withdraw(x ,o)
{ b = unbondedValue(o) + lockedBonds(o) + beneficiaryOf(o).balance }

*/

rule totalAssetsPreservedOnWithdraw(address operator, address owner, address beneficiary,  uint256 amount) {
     env e;
     require beneficiaryOf(operator) == beneficiary;
     require beneficiary != currentContract;
     mathint totalAssets = totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(beneficiary);
     sinvoke withdraw(e, amount, operator);
     assert  totalAssets == totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(beneficiary),
     "withdraw can change the total assets of operator in an unexpected way";
}
/**
On seizing of a bond, the total assets within the system of operator and the destination address receiving the bond value is preserved

{  b = unbondedValue(o) + lockedBonds(o) + destination.balance}
	seizeBond(o, ref, x, destination)
{ b = unbondedValue(o) + lockedBonds(o) + destination.balance }
*/
rule totalAssetsPreservedOnseizeBond(address operator, address owner, uint256 referenceID,
            uint256 amount, address destination,  method f) {
    env e;
    //require destination == otherBeneficiary();
    mathint totalAssetsBefore = totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(destination);
    sinvoke seizeBond(e, operator, referenceID, amount, destination);
    assert  totalAssetsBefore == totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(destination),
    "seizeBond can change the total assets of operator in an unexpected way";
}

/**
@title Integrity of withdraw
@notice successful withdraw of a value x of an operator o decreases o's unbonded value by x and transfers x to the beneficiary of the operator
             { 	beneficiary = beneficiaryOf(o)  ⋀
                u = unbondedValue(o)  ⋀
                b = beneficiary.balanceOf()
            }
            withdraw(x, o)
            {
                  u - x = unbondedValue(o)  ⋀
                  b + x = beneficiary.balanceOf()
            }
*/
rule integrityOfWithdraw(address operator,address beneficiary, uint256 x) {
    env e;
    require beneficiaryOf(operator) == beneficiary;
    require beneficiary != currentContract;
    //todo change to requireinvariant
    require safeAdd(unbondedValue(operator),totalLockedBonds(operator));
    require everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(beneficiary) ;
    uint s = balanceOf(currentContract);
    sinvoke withdraw(e, x, operator);
    assert u - x == unbondedValue(operator) &&  b + x == balanceOf(beneficiary) &&
           s - x == balanceOf(currentContract) && x <= u &&
           x <= everDeposited(operator),
           "withdraw integrity does not hold";
}
/**
@title Maximum withdraw
@notice When withdrawing the total of ever deposited the assets of operator is zeroed out
        { }
        r = withdraw(everDeposited(o), o)
        {
             r ⟹ ( unbondedValue(o) = 0 ⋀ lockedBonds(o) = 0 )
        }
*/
rule maximumWithdraw(address operator ) {
    env e;
    //todo change to requireinvariant
    require safeAdd(unbondedValue(operator),totalLockedBonds(operator));
    require everDeposited(operator) >= (unbondedValue(operator) + totalLockedBonds(operator));
    uint256 x = everDeposited(operator);
    invoke withdraw(e, x, operator);
    assert !lastReverted => (unbondedValue(operator) == 0 && totalLockedBonds(operator)==0),
    "maximum withdraw does not hold";
}

/**
@title Additivity of withdraw
@notice Withdrawing is additive, i.e., it can be performed either all at once or in two steps

( withdraw(x, o) ; withdraw(y, o) ) ~ withdraw(x+y, o)

Here we expect the effect of withdrawing and x and then withdrawing y to be the same as withdrawing them simultaneously.
The correctness of this rule on all inputs increases the confidence that the protocol is less fragile, e.g., to rounding errors.
*/
rule additiveWithdraw(address operator, address owner, address beneficiary, uint256 x, uint y) {
    env e;
    require ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary;
    require beneficiary != currentContract;
    require x + y < MAXINT();
    require balanceOf(beneficiary) + x + y <= MAXINT();

    storage init_storage = lastStorage; //store the current state
    invoke withdraw(e, x, operator);
    bool successX = !lastReverted;
    invoke withdraw(e, y, operator);
    bool successY = !lastReverted;
    uint256 uScenario1 = unbondedValue(operator);
    uint256 bScenario1 = balanceOf(beneficiary);

    invoke withdraw(e, x + y, operator) at init_storage; //back to the initial state
    bool successXY = !lastReverted;
    uint256 uScenario2 = unbondedValue(operator);
    uint256 bScenario2 = balanceOf(beneficiary);
    assert (successX && successY) <=> successXY;
    assert successXY => (uScenario1 == uScenario2 && bScenario1 == bScenario2),
    "withdraw is not additive";
}

/**
@title No front running on withdraw
@notice If one can withdraw x amount from the unbounded amount of operator o1, then same one should be able to withdraw after
another user has performed an operation as operator o2.

    r1 = withdraw(x, o1)   ∼r1 = r2   (  f ; r1 =withdraw(x, o1) )

Here we compare two arbitrary executions of the program: one with a single withdrawal,
and another in which another operation is performed by another operator.  We require that the withdrawal will succeed.

*/
rule noFrontRunningOnWithdraw(address operator, address owner, address beneficiary, address otherOperator, uint256 x, method f) {
    env eF;
    calldataarg args;
    uint256 referenceID;

    require ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary;
    require otherOperator != operator;

    require safeAdd(unbondedValue(operator), unbondedValue(otherOperator)) &&
            allUnbounded()  >= unbondedValue(operator) + unbondedValue(otherOperator);

    require safeAdd(totalLockedBonds(operator), totalLockedBonds(otherOperator)) &&
            allLockedBond()  >= totalLockedBonds(operator) + totalLockedBonds(otherOperator);

    require safeAdd(allUnbounded(), allLockedBond()) &&
            balanceOf(currentContract) >= allUnbounded() + allLockedBond();

    require totalLockedBonds(otherOperator) >=  bondAmount(otherOperator, eF.msg.sender, referenceID);

    uint256 allUserAssets = allUnbounded() + allLockedBond();
    require safeAdd(balanceOf(currentContract), allUserAssets);
    uint256 allAssets = balanceOf(currentContract) +  allUserAssets;
   require safeAdd(balanceOf(beneficiary), allAssets);
    require beneficiary != currentContract;

    storage init_storage = lastStorage;
    env eWithdraw;
    sinvoke withdraw(eWithdraw, x, operator) ;
    bool succSceanrio1 = !lastReverted;

    //f should only change other operator and if changed totalLocked than the specific bondID used in the require
    uint256 u = unbondedValue(otherOperator) at init_storage;
    uint256 l = totalLockedBonds(otherOperator);
    uint256 bondBefore =  bondAmount(otherOperator, eF.msg.sender, referenceID);
    sinvoke f(eF,args) ;
    require u != unbondedValue(otherOperator) || l != totalLockedBonds(otherOperator);
    require f.selector==seizeBond(address,uint256,uint256,address).selector  =>
                bondBefore !=  bondAmount(otherOperator, eF.msg.sender, referenceID);
    invoke withdraw(eWithdraw, x, operator);
    bool succSceanrio2 = !lastReverted;
    assert  succSceanrio2 ;
}

/**
@title Deposit and withdraw are inverse functions
@notice Withdraw is the inverse function of deposit with respect to the system balance and the unbounded value of operator
    {  u = unbondedValue(o)  ⋀
       b = ethSystemBalance()
    }
    ( deposit(x, o) ; r = withdraw(x, o) )
    {  r ⋀ u = unbondedValue(o)  ⋀
       b = ethSystemBalance()
    }
    in addition withdraw is possible under the limitations:
                (i) an allowed user with no msg.value,
                (ii) after the delegation lock period passed,
                (iii) beneficiary can accept the transfer
*/

rule inverseOfDepositAndWithdraw(address operator, address owner, address beneficiary, uint256 x ) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(currentContract) ;
    //sucessfully depoist
    env eDeposit;
    require eDeposit.msg.value == x;
    require (eDeposit.msg.sender != currentContract);
    sinvoke deposit(eDeposit, operator);
    env eWithdraw;
    require eWithdraw.msg.sender == operator ||  eWithdraw.msg.sender == owner;
    require eWithdraw.msg.value == 0;
    require _hasDelegationLockPassed(eWithdraw, operator);
    require safeAdd(balanceOf(beneficiary), x);
    invoke withdraw(eWithdraw, x, operator);
    bool succ = !lastReverted;
    //verify that succeeded and the the value is as expected
    assert succ &&  u  == unbondedValue(operator) && b == balanceOf(currentContract);
}


/**
@title Valid change to BalanceOf ✔️
@notice On operator op by user s  sending x amount, the Eth balance of a user (operator or not)  changes as follows:

             {  b = o.balance }
            op(s,x)
            {  b = o.balance ∨
               ( b + x = o.balance ∧ isDepositOp(op) ) ∨
               ( b ≤ o.balance   ∧  (isSeizeBond(op) || isWithdraw(op) ) )
            }

where isDepositOp is true for any of the deposit operations (deposit, delegate or topup)

*/
rule validChangeToBalanceOf(address o,  method f) {
    env e;
    calldataarg args;
    require (o != currentContract);
    uint256 before = balanceOf(o) ;
    sinvoke f(e, args);
    uint256 after = balanceOf(o) ;
    assert before == after ||
        // depositing case
        ( (f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
           f.selector==delegate(address, address, address).selector  ) &&
           after == before - e.msg.value ) ||
        // withdrawing case
           ( ( f.selector==seizeBond(address,uint256,uint256,address).selector || f.selector==withdraw(uint256,address).selector) &&
             after >= before  ) ;
}

/**
@title Valid change to unboundedValue ✔️
@notice On operator op by user u sending x amount, the unbounded value of operator o changes as follows:


        {  u = unbondedValue(o) }
        op(s,x)
        {  u = unbondedValue(o) ∨
          ( u + x = unbondedValue(o) ∧ isDepositOp(op) ) ∨
          ( l ≥ unbondedValue(o) ∧ isWithdraw(op) ∧ (s=o ∨ u=owner(s) )  ∨
                       (  l ≥ unbondedValue(o) ∧ isCreateBond(op)
}
*/
rule validChangeToTotalBondAmount(address o, method f ) {
    uint256 lockedBondsBefore = totalLockedBonds(o) ;
    uint256 unbondedBefore = unbondedValue(o);
    env e;
    calldataarg args;
    sinvoke f(e, args);
    uint256 lockedBondsAfter = totalLockedBonds(o);
    assert lockedBondsBefore == lockedBondsAfter ||
           // case of unlocking
           ( (f.selector==freeBond(address,uint256).selector ||
                 f.selector==seizeBond(address,uint256,uint256,address).selector ) &&
              lockedBondsAfter <= lockedBondsBefore )
           ||
           // case of locking
           ( f.selector==createBond(address,address,uint256,uint256,address).selector  &&
             lockedBondsAfter >= lockedBondsBefore && lockedBondsAfter <= lockedBondsBefore + unbondedBefore )
           ;
}

/**
@title Valid change to unboundedValue ✔️
@notice On operator op by user u sending x amount, the unbounded value of operator o changes as follows:

        {  u = unbondedValue(o) }
        op(s,x)
        {  u = unbondedValue(o) ∨
          ( u + x = unbondedValue(o) ∧ isDepositOp(op) ) ∨
          ( u ≥ unbondedValue(o) ∧ isWithdraw(op) ∧ (s=o ∨ u=owner(s) )  ∨
                       (  u ≥ unbondedValue(o) ∧ isCreateBond(op)
        }
*/
rule validChangeToUnbondedValue(address o, address owner,  method f) {
    uint256 unbondedBefore = unbondedValue(o);
    env e;
    calldataarg args;
    sinvoke f(e, args);
    uint256 unbondedAfter = unbondedValue(o);
    assert  unbondedAfter  == unbondedBefore ||
        //cases for deposit
        (( f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
           f.selector==delegate(address, address, address).selector  ) &&
            unbondedAfter == unbondedBefore + e.msg.value ) ||
        //cases for withdraw
        ( f.selector==withdraw(uint256,address).selector && (e.msg.sender==o || e.msg.sender == ownerOf(o)) &&
         unbondedAfter <= unbondedBefore) ||
        // cases for holder free
        ( f.selector==freeBond(address,uint256).selector &&  unbondedAfter >= unbondedBefore ) ||
        // case for holder locking
        ( f.selector==createBond(address,address,uint256,uint256,address).selector   &&
           unbondedAfter <= unbondedBefore  );
}

/**
@title Valid change to everDeposited ✔️
On operator op by user u sending x amount, the unbounded value of operator o changes as follows:
        {  e = everDeposited(o) }
        op(s,x)
        {  e = everDeposited(o) ∨
          ( e + x = everDeposited(o)  ∧ isDepositOp(op) )
        }

*/
rule validChangeToEverDeposited(address o,  method f) {
    env e;
    calldataarg args;
    uint256 before = everDeposited(o);
    sinvoke f(e, args);
    uint256 after = everDeposited(o);
    assert after == before ||
            ( ( f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
                f.selector==delegate(address, address, address).selector  ) &&
                after == before + e.msg.value );
}
