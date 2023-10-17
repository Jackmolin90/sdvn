// Copyright 2021 The sdvn Authors
// This file is part of the sdvn library.
//
// The sdvn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The sdvn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the sdvn library. If not, see <http://www.gnu.org/licenses/>.

package alien

import (
	"fmt"
	"github.com/seaskycheng/sdvn/core/rawdb"
	"github.com/seaskycheng/sdvn/core/state"
	"github.com/seaskycheng/sdvn/core/types"
	"github.com/seaskycheng/sdvn/params"
	"github.com/shopspring/decimal"
	"math/big"
	"testing"

	"github.com/seaskycheng/sdvn/common"
)

func TestAlien_PenaltyTrantor(t *testing.T) {
	tests := []struct {
		last    string
		current string
		queue   []string
		lastQ   []string
		result  []string // the result of missing

	}{
		{
			/* 	Case 0:
			 *  simple loop order, miss nothing
			 *  A -> B -> C
			 */
			last:    "A",
			current: "B",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{},
			result:  []string{},
		},
		{
			/* 	Case 1:
			 *  same loop, missing B
			 *  A -> B -> C
			 */
			last:    "A",
			current: "C",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{},
			result:  []string{"B"},
		},
		{
			/* 	Case 2:
			 *  same loop, not start from the first one
			 *  C -> A -> B
			 */
			last:    "C",
			current: "B",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{},
			result:  []string{"A"},
		},
		{
			/* 	Case 3:
			 *  same loop, missing two
			 *  A -> B -> C
			 */
			last:    "C",
			current: "C",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{},
			result:  []string{"A", "B"},
		},
		{
			/* 	Case 4:
			 *  cross loop
			 *  B -> A -> B -> C -> A
			 */
			last:    "B",
			current: "B",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{"C", "A", "B"},
			result:  []string{"A"},
		},
		{
			/* 	Case 5:
			 *  cross loop, nothing missing
			 *  A -> C -> A -> B -> C
			 */
			last:    "A",
			current: "C",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{"C", "A", "B"},
			result:  []string{},
		},
		{
			/* 	Case 6:
			 *  cross loop, two signers missing in last loop
			 *  C -> B -> C -> A
			 */
			last:    "C",
			current: "A",
			queue:   []string{"A", "B", "C"},
			lastQ:   []string{"C", "A", "B"},
			result:  []string{"B", "C"},
		},
	}

	// Run through the test
	for i, tt := range tests {
		// Create the account pool and generate the initial set of all address in addrNames
		accounts := newTesterAccountPool()
		addrQueue := make([]common.Address, len(tt.queue))
		for j, signer := range tt.queue {
			addrQueue[j] = accounts.address(signer)
		}

		extra := HeaderExtra{SignerQueue: addrQueue}
		var lastExtra HeaderExtra
		if len(tt.lastQ) > 0 {
			lastAddrQueue := make([]common.Address, len(tt.lastQ))
			for j, signer := range tt.lastQ {
				lastAddrQueue[j] = accounts.address(signer)
			}
			lastExtra = HeaderExtra{SignerQueue: lastAddrQueue}
		}

		missing := getSignerMissingTrantor(accounts.address(tt.last), accounts.address(tt.current), &extra, &lastExtra)

		signersMissing := make(map[string]bool)
		for _, signer := range missing {
			signersMissing[accounts.name(signer)] = true
		}
		if len(missing) != len(tt.result) {
			t.Errorf("test %d: the length of missing not equal to the length of result, Result is %v not %v  ", i, signersMissing, tt.result)
		}

		for j := 0; j < len(missing); j++ {
			if _, ok := signersMissing[tt.result[j]]; !ok {
				t.Errorf("test %d: the signersMissing is not equal Result is %v not %v ", i, signersMissing, tt.result)
			}
		}
	}
}

func TestAlien_Penalty(t *testing.T) {
	tests := []struct {
		last    string
		current string
		queue   []string
		newLoop bool
		result  []string // the result of current snapshot
	}{
		{
			/* 	Case 0:
			 *  simple loop order
			 */
			last:    "A",
			current: "B",
			queue:   []string{"A", "B", "C"},
			newLoop: false,
			result:  []string{},
		},
		{
			/* 	Case 1:
			 * simple loop order, new loop, no matter which one is current signer
			 */
			last:    "C",
			current: "A",
			queue:   []string{"A", "B", "C"},
			newLoop: true,
			result:  []string{},
		},
		{
			/* 	Case 2:
			 * simple loop order, new loop, no matter which one is current signer
			 */
			last:    "C",
			current: "B",
			queue:   []string{"A", "B", "C"},
			newLoop: true,
			result:  []string{},
		},
		{
			/* 	Case 3:
			 * simple loop order, new loop, missing in last loop
			 */
			last:    "B",
			current: "C",
			queue:   []string{"A", "B", "C"},
			newLoop: true,
			result:  []string{"C"},
		},
		{
			/* 	Case 4:
			 * simple loop order, new loop, two signers missing in last loop
			 */
			last:    "A",
			current: "C",
			queue:   []string{"A", "B", "C"},
			newLoop: true,
			result:  []string{"B", "C"},
		},
	}

	// Run through the test
	for i, tt := range tests {
		// Create the account pool and generate the initial set of all address in addrNames
		accounts := newTesterAccountPool()
		addrQueue := make([]common.Address, len(tt.queue))
		for j, signer := range tt.queue {
			addrQueue[j] = accounts.address(signer)
		}

		extra := HeaderExtra{SignerQueue: addrQueue}
		//missing := getSignerMissing(accounts.address(tt.last), accounts.address(tt.current), extra, tt.newLoop)
		missing := getSignerMissing(0,0,0,accounts.address(tt.last), accounts.address(tt.current), extra, tt.newLoop)

		signersMissing := make(map[string]bool)
		for _, signer := range missing {
			signersMissing[accounts.name(signer)] = true
		}
		if len(missing) != len(tt.result) {
			t.Errorf("test %d: the length of missing not equal to the length of result, Result is %v not %v  ", i, signersMissing, tt.result)
		}

		for j := 0; j < len(missing); j++ {
			if _, ok := signersMissing[tt.result[j]]; !ok {
				t.Errorf("test %d: the signersMissing is not equal Result is %v not %v ", i, signersMissing, tt.result)
			}
		}

	}
}

func  TestAlien_BlockReward(t *testing.T) {
	var period =uint64(3)
	blockNumPerYear := secondsPerYear /period
	tests := []struct {
		maxSignerCount uint64
		number         uint64
		coinbase    common.Address
		Period      uint64
		expectValue  *big.Int
	}{
		{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			expectValue:big.NewInt(int64(1089767370173551512)),
		},{
			maxSignerCount :3,
			number     :blockNumPerYear*2-1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			expectValue:big.NewInt(int64(970872353158784960)),
		},{
			maxSignerCount :3,
			number     :blockNumPerYear*3-1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			expectValue:big.NewInt(int64(864948934907056792)),
		}}

	for _, tt := range tests {
		snap := &Snapshot{
			config:   &params.AlienConfig{
				    MaxSignerCount: tt.maxSignerCount,
				    Period: tt.Period,
			  },
			Number:   tt.number,
			LCRS:     1,
			Tally:    make(map[common.Address]*big.Int),
			Punished: make(map[common.Address]uint64),
		}
		header := &types.Header{
			Coinbase: tt.coinbase,
			Number: big.NewInt(int64(tt.number)),
		}
		//
		state := &state.StateDB{
		}
		refundGas:= RefundGas{}
		gasReward:= big.NewInt(0)
		//currentLockRd := &LockRewardRecord{
		//	Target: tt.coinbase,
		//	Amount:big.NewInt(0)  ,
		//	IsReward: 1,
		//	FlowValue1: 0,
		//	FlowValue2:0,
		//}
		alienConfig :=&params.AlienConfig{
			Period: tt.Period,
		}
		config := &params.ChainConfig{
			ChainID:big.NewInt(128),
			Alien:alienConfig,
		}

		currentHeaderExtra := HeaderExtra{}
		//RefundGas := make(map[common.Address]*big.Int)
		//accumulateRewards(currentLockReward []LockRewardRecord, config *params.ChainConfig, header *types.Header, snap *Snapshot)
		LockRewardRecord :=accumulateRewards(currentHeaderExtra.LockReward, config,state, header, snap,refundGas,gasReward)

		yearCount := header.Number.Uint64() / blockNumPerYear
		if yearCount*blockNumPerYear!=header.Number.Uint64() {
			yearCount++
		}
		for index := range LockRewardRecord {
			if LockRewardRecord[index].Amount.Int64()==tt.expectValue.Int64() {
				t.Logf("blocknumber : %d ,%d th years,coinbase %s,Block reward calculation pass,expect %d NFC,but act %d" ,header.Number,yearCount,(LockRewardRecord[index].Target).Hex(),tt.expectValue,LockRewardRecord[index].Amount)
			}else {
				t.Errorf("blocknumber : %d ,%d th years,coinbase %s,Block reward calculation error,expect %d NFC,but act %d" ,header.Number,yearCount,(LockRewardRecord[index].Target).Hex(),tt.expectValue,LockRewardRecord[index].Amount)
			}

		}

	}
}
func  TestAlien_bandwidthReward(t *testing.T) {
	var period =uint64(10)

	blockNumPerYear := secondsPerYear /period
	tests := []struct {
		maxSignerCount uint64
		number         uint64
		coinbase    common.Address
		Period      uint64
		expectValue  decimal.Decimal
		bandwidth uint64
	}{
		{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(100),
			expectValue: decimal.NewFromFloat(0.1* 32291.4913).Round(4),

		},{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(400),
			expectValue: decimal.NewFromFloat(0.2* 32291.4913).Round(4),

		},{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(1000),
			expectValue: decimal.NewFromFloat(0.3* 32291.4913).Round(4),

		},{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(2000),
			expectValue: decimal.NewFromFloat(0.4* 32291.4913).Round(4),

		},
		{
			maxSignerCount :3,
			number     :blockNumPerYear*2-1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(100),
			expectValue: decimal.NewFromFloat(0.1* 30479.1092).Round(4),
		},{
			maxSignerCount :3,
			number     :blockNumPerYear*2-1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			bandwidth:uint64(400),
			expectValue: decimal.NewFromFloat(0.2* 30479.1092).Round(4),
		},{
			maxSignerCount :3,
			number     :blockNumPerYear*2-1,
			coinbase : common.HexToAddress("NX7A4539Ed8A0b8B4583EAd1e5a3F604e83419f502"),
			Period :period,
			bandwidth:uint64(1000),
			expectValue: decimal.NewFromFloat(0.3* 30479.1092).Round(4),
		},{
			maxSignerCount :3,
			number     :blockNumPerYear*2-1,
			coinbase : common.HexToAddress("NX7A4539Ed8A0b8B4583EAd1e5a3F604e83419f502"),
			Period :period,
			bandwidth:uint64(2000),
			expectValue: decimal.NewFromFloat(0.4* 30479.1092).Round(4),
		},
	}

	for _, tt := range tests {
		snap := &Snapshot{
			config:   &params.AlienConfig{
				MaxSignerCount: tt.maxSignerCount,
				Period: tt.Period,
			},
			Number:   tt.number,
			LCRS:     1,
			Tally:    make(map[common.Address]*big.Int),
			Punished: make(map[common.Address]uint64),
			Bandwidth:make(map[common.Address]*ClaimedBandwidth),
			FlowMiner:&FlowMinerSnap{
				FlowMinerPrev: make(map[common.Address]map[common.Hash]*FlowMinerReport),
			},
		}
		header := &types.Header{
			Coinbase: tt.coinbase,
			Number: big.NewInt(int64(tt.number)),
		}

		alienConfig :=&params.AlienConfig{
			Period: tt.Period,
		}
		config := &params.ChainConfig{
			ChainID:big.NewInt(128),
			Alien:alienConfig,
		}
		oldBandwidth100 :=&ClaimedBandwidth{
			ISPQosID:4,
			BandwidthClaimed:uint32(tt.bandwidth),
		}
		snap.Bandwidth[tt.coinbase] = oldBandwidth100
		snap.FlowMiner.FlowMinerPrev[common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502")]=make(map[common.Hash]*FlowMinerReport)
		snap.FlowMiner.FlowMinerPrev[common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502")][common.Hash{}]=&FlowMinerReport{
			FlowValue1:uint64(10),
		}
		db:=rawdb.NewMemoryDatabase()
		currentHeaderExtra := HeaderExtra{}
		LockRewardRecord :=accumulateBandwidthRewards(currentHeaderExtra.LockReward, config, header, snap,db)
		yearCount := header.Number.Uint64() / blockNumPerYear
		if yearCount*blockNumPerYear!=header.Number.Uint64() {
			yearCount++
		}
		for index := range LockRewardRecord {
		actReward:=decimal.NewFromBigInt(LockRewardRecord[index].Amount,0).Div(decimal.NewFromInt(1E+18)).Round(4)
			if actReward.Cmp(tt.expectValue)==0{
				t.Logf("blocknumber : %d ,%d th years,coinbase %s,Bandwidth reward calculation pass,expect %d NFC,but act %d" ,header.Number,yearCount,(LockRewardRecord[index].Target).Hex(),tt.expectValue.BigFloat(),actReward.BigFloat())
			}else {
				t.Errorf("blocknumber : %d ,%d th years,coinbase %s,Bandwidth reward calculation error,expect %d NFC,but act %d" ,header.Number,yearCount,(LockRewardRecord[index].Target).Hex(),tt.expectValue.BigFloat(),actReward.BigFloat())
			}

		}
	}
}
func    TestAlien_FwReward(t *testing.T) {
	//
	var ebval =float64(1099511627776)
	tests := []struct {
		flowTotal decimal.Decimal
		expectValue decimal.Decimal
	}{
		{
			flowTotal:decimal.NewFromFloat(10995116),
			expectValue:decimal.NewFromFloat(60),
		},{
			flowTotal:decimal.NewFromFloat(ebval+1024),
			expectValue:decimal.NewFromFloat(60.945),
		},{
		flowTotal:decimal.NewFromFloat(ebval*2+1024),
			expectValue:decimal.NewFromFloat(61.904),
	    },{
		flowTotal:decimal.NewFromFloat(ebval*3+1024),
			expectValue:decimal.NewFromFloat(62.878),
	    },{
			flowTotal:decimal.NewFromFloat(ebval*4+1024),
			expectValue:decimal.NewFromFloat(63.868),
		},
	}
	for _, tt := range tests {
		fwreward := getFlowRewardScale(tt.flowTotal)
		rewardgb:= decimal.NewFromFloat(1).Div(fwreward.Mul(decimal.NewFromFloat(1024)).Div(decimal.NewFromFloat(1e+18))).Round(3)
		totalEb :=tt.flowTotal.Div(decimal.NewFromInt(1099511627776))
		var nebCount=totalEb.Round(0)
		if totalEb.Cmp(nebCount)>0 {
			nebCount= nebCount.Add(decimal.NewFromInt(1))
		}
		if rewardgb.Cmp(tt.expectValue)==0 {
			fmt.Println("Flow mining reward test pass ，",nebCount,"th EB，1 NFC=",rewardgb,"GB flow")
		}else{
			//t.Errorf("Flow mining reward test failed #{%d},",nebCount,"th EB，1 NFC=",rewardgb,"GB,But the actual need",tt.expectValue,"GB")
			t.Errorf("test: Flow mining reward test failed,theory 1 NFC need %d GB act need %d GB",tt.expectValue.BigFloat(),rewardgb.BigFloat())
		}

	}

}

func  TestAlien_accumulateFlowRewards(t *testing.T) {
	var period =uint64(10)
	dayReward1:=60.000*1024
	FlowValue1:=uint64(100000)
	FlowValue2:=uint64(400000)
	bandwidth1:=uint64(100)
	bandwidth2:=uint64(400)
	tests := []struct {
		maxSignerCount uint64
		number         uint64
		coinbase    common.Address
		Period      uint64
		expectValue  decimal.Decimal
		FlowValue1 uint64
		bandwidth uint64
	}{
		{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX7a4539ed8a0b8b4583ead1e5a3f604e83419f502"),
			Period :period,
			FlowValue1:FlowValue1,
			expectValue: decimal.NewFromFloat(float64(FlowValue1)/float64(dayReward1)).Round(10),
			bandwidth:bandwidth1,
		},{
			maxSignerCount :3,
			number     :1,
			coinbase : common.HexToAddress("NX2aD0559Afade09a22364F6380f52BF57E9057B8D"),
			Period :period,
			bandwidth:bandwidth2,
			FlowValue1:FlowValue2,
			expectValue: decimal.NewFromFloat(float64(FlowValue2)/float64(dayReward1)).Round(10),
		},
	}

	snap := &Snapshot{
		LCRS:      1,
		Tally:     make(map[common.Address]*big.Int),
		Punished:  make(map[common.Address]uint64),
		Bandwidth: make(map[common.Address]*ClaimedBandwidth),
		FlowMiner:&FlowMinerSnap{
			FlowMinerPrev:make(map[common.Address]map[common.Hash]*FlowMinerReport),
		},
		FlowTotal:big.NewInt(0),
	}
	for _, tt := range tests {
		oldBandwidth100 := &ClaimedBandwidth{
			ISPQosID:         4,
			BandwidthClaimed: uint32(tt.bandwidth),
		}
		snap.Bandwidth[tt.coinbase] = oldBandwidth100
	}
	var LockRewardRecord[] LockRewardRecord
	for _, tt := range tests {
		snap.config = &params.AlienConfig{
			MaxSignerCount: tt.maxSignerCount,
			Period:         tt.Period,
		}
		snap.Number = tt.number
		snap.FlowMiner.FlowMinerPrev[tt.coinbase]=make(map[common.Hash]*FlowMinerReport)
		snap.FlowMiner.FlowMinerPrev[tt.coinbase][common.Hash{}]=&FlowMinerReport{
			FlowValue1:tt.FlowValue1,
		}
	}
	db:=rawdb.NewMemoryDatabase()
	currentHeaderExtra := HeaderExtra{}
	LockRewardRecord, _ = accumulateFlowRewards(currentHeaderExtra.LockReward,  snap, db)
	for _, tt := range tests {
		for index := range LockRewardRecord {
			if tt.coinbase==LockRewardRecord[index].Target{
				actReward:=decimal.NewFromBigInt(LockRewardRecord[index].Amount,0).Div(decimal.NewFromInt(1E+18)).Round(10)
				cut:=actReward.Sub(tt.expectValue).Abs()
				if cut.Cmp(decimal.NewFromFloat(float64(0.2)))<=0{
					t.Logf("coinbase %s,Bandwidth reward calculation pass,expect %d utg,but act %d,cut is %d" ,(LockRewardRecord[index].Target).Hex(),tt.expectValue.BigFloat(),actReward.BigFloat(),cut.BigFloat())
				}else {
					t.Errorf("coinbase %s,Bandwidth reward calculation error,expect %d utg,but act %d,cut is %d" ,(LockRewardRecord[index].Target).Hex(),tt.expectValue.BigFloat(),actReward.BigFloat(),cut.BigFloat())
				}
			}
		}
	}
}