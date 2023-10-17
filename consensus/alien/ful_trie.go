package alien

import (
	"errors"
	"github.com/seaskycheng/sdvn/common"
	"github.com/seaskycheng/sdvn/ethdb"
	"github.com/seaskycheng/sdvn/log"
	"github.com/seaskycheng/sdvn/rlp"
	"github.com/seaskycheng/sdvn/trie"
	"math/big"
)

var (
	errNotEnoughFul = errors.New("not enough ful")
)

type FulAccount struct {
	Address  common.Address
	Balance  *big.Int
}

func (s *FulAccount) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(s)
}

func decodeFulAccount(buf []byte) *FulAccount {
	s := &FulAccount{}
		err := rlp.DecodeBytes(buf, s)
		if err != nil {
			return nil
		}else {
			return s
		}
}

func (s *FulAccount) GetBalance() *big.Int {
	return s.Balance
}

func (s *FulAccount) SetBalance(amount *big.Int) {
	s.Balance = amount
}

func (s *FulAccount) AddBalance(amount *big.Int) {
	s.Balance = new(big.Int).Add(s.Balance, amount)
}

func (s *FulAccount) SubBalance(amount *big.Int) error {
	if s.Balance.Cmp(amount) < 0 {
		return errNotEnoughFul
	}
	s.Balance = new(big.Int).Sub(s.Balance, amount)
	return nil
}

//=========================================================================
type FulTrie struct {
	trie 	*trie.SecureTrie
	db      ethdb.Database
	triedb 	*trie.Database
}

func (s *FulTrie) GetOrNewAccount(addr common.Address) *FulAccount {
	var obj *FulAccount
	objData := s.trie.Get(addr.Bytes())
	obj = decodeFulAccount(objData)
	if obj != nil {
		return obj
	}
	obj = &FulAccount{
		Address: addr,
		Balance: common.Big0,
	}
	return obj
}

func (s *FulTrie) TireDB() *trie.Database {
	return s.triedb
}

func (s *FulTrie) getBalance(addr common.Address) *big.Int {
	obj := s.GetOrNewAccount(addr)
	if obj == nil {
		return common.Big0
	}
	return obj.GetBalance()
}

func (s *FulTrie) setBalance(addr common.Address, amount *big.Int) {
	obj := s.GetOrNewAccount(addr)
	if obj == nil {
		log.Warn("fultrie setbalance", "result", "failed")
		return
	}
	obj.SetBalance(amount)
	value, _ := obj.Encode()
	s.trie.Update(addr.Bytes(), value)
}

func (s *FulTrie) addBalance(addr common.Address, amount *big.Int) {
	obj := s.GetOrNewAccount(addr)
	if obj == nil {
		log.Warn("fultrie addbalance", "result", "failed")
		return
	}
	obj.AddBalance(amount)
	value, _ := obj.Encode()
	s.trie.Update(addr.Bytes(), value)
}

func (s *FulTrie) subBalance(addr common.Address, amount *big.Int) error{
	obj := s.GetOrNewAccount(addr)
	if obj == nil {
		log.Warn("fultrie subbalance", "result", "failed")
		return errNotEnoughFul
	}
	obj.SubBalance(amount)
	value, _ := obj.Encode()
	s.trie.Update(addr.Bytes(), value)
	return nil
}

func (s *FulTrie) cmpBalance(addr common.Address, amount *big.Int) int{
	obj := s.GetOrNewAccount(addr)
	if obj == nil {
		log.Warn("fultrie cmpbalance", "error", "load account failed")
		return -1
	}

	return obj.Balance.Cmp(amount)
}

func (s *FulTrie) Hash() common.Hash {
	return s.trie.Hash()
}

func (s *FulTrie) commit() (root common.Hash, err error){
	hash, err := s.trie.Commit(nil)
	if err != nil {
		return common.Hash{}, err
	}
	s.triedb.Commit(hash, true, nil)
	return hash, nil
}

//====================================================================================
func NewFulTrie(root common.Hash, db ethdb.Database) (*FulTrie, error) {
	triedb := trie.NewDatabase(db)
	tr, err := trie.NewSecure(root, triedb)
	if err != nil {
		log.Warn("fultrie open ful trie failed", "root", root)
		return nil, err
	}

	return &FulTrie{
		trie: tr,
		db: db,
		triedb: triedb,
	}, nil
}

func (s *FulTrie) Get (addr common.Address) *big.Int{
	return s.getBalance(addr)
}

func (s *FulTrie) Set (addr common.Address, amount *big.Int) {
	s.setBalance(addr, amount)
}

func (s *FulTrie) Add (addr common.Address, amount *big.Int) {
	s.addBalance(addr, amount)
}

func (s *FulTrie) Sub (addr common.Address, amount *big.Int) error{
	return s.subBalance(addr, amount)
}

func (s *FulTrie) Del(addr common.Address) {
	s.setBalance(addr, common.Big0)
	s.trie.Delete(addr.Bytes())
}

func (s *FulTrie) Copy() FulState {
	root, _ := s.Save(nil)
	trie, _ := NewTrieFulState(root, s.db)
	return trie
}

func (s *FulTrie) Load(db ethdb.Database, hash common.Hash) error{
	return nil
}

func (s *FulTrie) Save(db ethdb.Database) (common.Hash, error){
	return s.commit()
}

func (s *FulTrie) Root() common.Hash {
	return s.Hash()
}

func (s *FulTrie) GetAll() map[common.Address]*big.Int {
	found := make(map[common.Address]*big.Int)
	it := trie.NewIterator(s.trie.NodeIterator(nil))
	for it.Next() {
		acc := decodeFulAccount(it.Value)
		if nil != acc {
			found[acc.Address] = acc.Balance
		}
	}
	return found
}
