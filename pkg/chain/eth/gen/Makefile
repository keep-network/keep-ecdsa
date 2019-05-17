# In .../github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen

# Directory where the solidity contracts are.
SOL_DIR=../../../../solidity/contracts


# Relative directory where the Go code needs to go. This is relative to the
# ${SOL_DIR} directory.
OP_DIR=../../pkg/chain/eth/gen
GEN_DIR=../../pkg/chain/eth/gen/abi

generate_solidity_go:
	( cd ${SOL_DIR} ; make -f ${OP_DIR}/Makefile gen_go )

SOL=$(shell ls *.sol | sed -e 's/.sol/.go/' )

gen_go:
	mkdir -p ${GEN_DIR}
	make -f ${OP_DIR}/Makefile ${SOL}

%.go: %.sol
	solc --allow-paths . --overwrite --abi -o . $< 
	abigen --abi $*.abi --pkg abi --type $* --out ${GEN_DIR}/$@
