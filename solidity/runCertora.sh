certoraRun  specs/harness/FullyBackedBondingHarness.sol:FullyBackedBondingHarness  specs/harness/Beneficiary.sol:Beneficiary \
 --structLink FullyBackedBondingHarness:2=Beneficiary \
--solc solc5.17  --verify FullyBackedBondingHarness:specs/fullyBackedBonding.spec \
--settings -recursionErrorAsAssert=false,-deleteSMTFile=true,-graphDrawLimit=0 \
--staging --cache FullyBackedBonding --msg "$1"
