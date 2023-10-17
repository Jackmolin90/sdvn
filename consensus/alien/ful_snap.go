package alien

import (
	"github.com/seaskycheng/sdvn/common"
	"github.com/seaskycheng/sdvn/ethdb"
	"math/big"
)

type FulState interface {
	Set(addr common.Address, amount *big.Int)
	Add(addr common.Address, amount *big.Int)
	Sub(addr common.Address, amount *big.Int) error
	Get(addr common.Address) *big.Int
	Del(addr common.Address)
	Copy() FulState
	Load(db ethdb.Database, hash common.Hash) error
	Save(db ethdb.Database) (common.Hash, error)
	Root() common.Hash
	GetAll() map[common.Address]*big.Int
}

func NewFUL(root common.Hash,db ethdb.Database) (FulState,error) {
	state ,err:= NewTrieFulState(root,db)
	return state,err
}

func NewTrieFulState(root common.Hash, db ethdb.Database) (FulState,error) {
	return NewFulTrie(root,db)
}