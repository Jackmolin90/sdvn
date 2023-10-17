package alien

import (
	"errors"
	"fmt"
	"github.com/seaskycheng/sdvn/common"
	"github.com/seaskycheng/sdvn/consensus"
	"github.com/seaskycheng/sdvn/core/state"
	"github.com/seaskycheng/sdvn/core/types"
	"github.com/seaskycheng/sdvn/crypto"
	"github.com/seaskycheng/sdvn/ethdb"
	"github.com/seaskycheng/sdvn/log"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sort"
	"strings"
)

var (
	calFlowToFULRatio= uint64(13671875000000)//0.014 FUL/GB
)

func (a *Alien) processFlowCustomTx(txDataInfo []string, headerExtra HeaderExtra, txSender common.Address, tx *types.Transaction, receipts []*types.Receipt, snapCache *Snapshot, number *big.Int, state *state.StateDB, chain consensus.ChainHeaderReader,fulBalances map[common.Address]*big.Int) HeaderExtra {
	if  txDataInfo[posCategory]==nfcEventFlowReportEn {
		headerExtra.FlowReport = a.processFlowReportEn (headerExtra.FlowReport, txDataInfo,number.Uint64(),snapCache,txSender, tx, receipts,fulBalances)
	}
	return headerExtra
}


func (a *Alien) processFlowReportEn(flowReport []MinerFlowReportRecord, txDataInfo []string, number uint64, snap *Snapshot, txSender common.Address, tx *types.Transaction, receipts []*types.Receipt,fulBalances map[common.Address]*big.Int) []MinerFlowReportRecord {
	if len(txDataInfo) <= 4 {
		log.Warn("En Flow report", "parameter number", len(txDataInfo))
		return flowReport
	}
	position := 3
	enAddr := txSender
	if _, ok := snap.FlowPledge[enAddr]; !ok {
		log.Warn("En Flow report", "enAddr is not in FlowPledge", enAddr)
		return flowReport
	}
	position++
	verifyArr :=strings.Split(txDataInfo[position],"|")
	if len(verifyArr)==0 {
		log.Warn("En Flow report", " verifyArr len = 0", txSender, len(verifyArr))
		return flowReport
	}
	census := MinerFlowReportRecord{
		ChainHash: common.Hash{},
		ReportTime: number,
		ReportContent: []MinerFlowReportItem{},
	}
	zeroAddr:=common.Address{}
	var verifyResult []int
	flowrecordlen:=4
	for index,verifydata :=range verifyArr {
		if verifydata == "" {
			continue
		}
		flowrecord :=strings.Split(verifydata,",")
		if len(flowrecord)!=flowrecordlen{
			log.Warn("En Flow report ", "flowrecordlen", len(flowrecord),"index", index)
			continue
		}
		i:=uint8(0)
		reportNumber, err := decimal.NewFromString(flowrecord[i])
		if err !=nil|| reportNumber.Cmp(decimal.Zero)<=0{
			log.Warn("En Flow report ", "reportNumber is wrong", flowrecord[i],"index", index)
			continue
		}
		if !reportNumber.BigInt().IsUint64()||reportNumber.BigInt().Uint64()==uint64(0){
			log.Warn("En Flow report ", "reportNumber is not uint64 or is zero",reportNumber,"checkReportNumber index", index)
			continue
		}
		if !snap.checkReportNumber(reportNumber,number) {
			log.Warn("En Flow report ", "checkReportNumber index", index)
			continue
		}
		i++
		deviceId, err := decimal.NewFromString(flowrecord[i])
		if err !=nil|| deviceId.Cmp(decimal.Zero)<=0{
			log.Warn("En Flow report ", "deviceId is wrong", flowrecord[i],"index", index)
			continue
		}
		if !deviceId.BigInt().IsUint64(){
			log.Warn("En Flow report ", "deviceId is not uint64", flowrecord[i],"index", index)
			continue
		}
		i++
		flowValue, err := decimal.NewFromString(flowrecord[i])
		if err !=nil||flowValue.Cmp(decimal.Zero)<=0{
			log.Warn("En Flow report ", "flowValue is wrong", flowrecord[i],"index", index)
			continue
		}
		if !flowValue.BigInt().IsUint64()||flowValue.BigInt().Uint64()==uint64(0){
			log.Warn("En Flow report ", "flowValue is not uint64 or is zero", flowrecord[i],"index", index)
			continue
		}
		i++
		rsv:=flowrecord[i]
		if len(rsv)==0{
			log.Warn("En Flow report ", "len(vrs) is zero", flowrecord[i],"index", index)
			continue
		}
		sig:=common.FromHex(rsv)
		if len(sig) != crypto.SignatureLength {
			log.Warn("En Flow report ", "wrong size for signature", flowrecord[i],"index", index)
			continue
		}
		var from common.Address
		if singer,ok:=isCheckFlowRecordSign(reportNumber,deviceId,enAddr,flowValue,sig);!ok{
			log.Warn("En Flow report ", "checkFlowRecordSign index", index)
			continue
		}else{
			from=singer
		}
		if from==zeroAddr{
			log.Warn("En Flow report ","index",index)
			continue
		}
		if _,ok:=fulBalances[from];!ok{
			fulBal:=snap.Ful.Get(from)
			fulBalances[from]=new(big.Int).Set(fulBal)
		}
		costFul:=common.Big0
		if ful,ok:=snap.checkFulEnoughItem(flowValue.BigInt().Uint64(),fulBalances[from]) ;!ok{
			log.Warn("En Flow report ", "CheckFulEnoughItem index", index)
			continue
		}else{
			costFul=new(big.Int).Set(ful)
		}

		flowValueM:=flowValue.BigInt().Uint64()
		flowReportItem:=MinerFlowReportItem {
			Target:enAddr,
			ReportNumber:0,
			FlowValue1:flowValueM,
			FlowValue2:0,
		}
		flowReportItem2:=MinerFlowReportItem {
			Target:from,
			ReportNumber:0,
			FlowValue1:0,
			FlowValue2:flowValueM,
		}
		census.ReportContent=append(census.ReportContent,flowReportItem)
		census.ReportContent=append(census.ReportContent,flowReportItem2)
		verifyResult=append(verifyResult, index)
		fulBalances[from]=new(big.Int).Sub(fulBalances[from],costFul)
	}
	if len(census.ReportContent)>0{
		flowReport = append(flowReport, census)
		topicdata := ""
		sort.Ints(verifyResult)
		for _, val := range verifyResult {
			if topicdata == "" {
				topicdata =fmt.Sprintf("%d", val)
			} else {
				topicdata += "," + fmt.Sprintf("%d", val)
			}
		}
		topics := make([]common.Hash, 1)
		topics[0].UnmarshalText([]byte("0xea40f050c9c577748d5ddcdb6a19aab17cacb2fa5f63f3747c516b06b597afd1"))//web3.sha3("Flwrpten(address,uint256)")
		a.addCustomerTxLog(tx, receipts, topics, []byte(topicdata))
	}
	return flowReport
}

func isCheckFlowRecordSign(reportNumber decimal.Decimal,deviceId decimal.Decimal, toAddress common.Address, flowValue decimal.Decimal,sig []byte) (common.Address,bool) {
	zeroAddr:=common.Address{}
	var hash common.Hash
	hasher := sha3.NewLegacyKeccak256()
    toAddressStr:=strings.ToLower(toAddress.String())
	msg := toAddressStr[2:]+reportNumber.String()+deviceId.String()+flowValue.String()
	hasher.Write([]byte(msg))
	hasher.Sum(hash[:0])
	var rBig=new(big.Int).SetBytes(sig[:32])
	var sBig=new(big.Int).SetBytes(sig[32:64])
	if !crypto.ValidateSignatureValues(sig[64], rBig, sBig, true) {
		log.Warn("isCheckFlowRecordSign", "crypto validateSignatureValues fail ", "sign wrong")
		return zeroAddr,false
	}
	pubkey, err := crypto.Ecrecover(hash.Bytes(), sig)
	if err != nil {
		log.Warn("isCheckFlowRecordSign", "crypto.Ecrecover", err)
		return zeroAddr,false
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
	return signer,true
}

func (s *Snapshot) calFulHashVer(roothash common.Hash, number uint64, db ethdb.Database) (*Snapshot,error) {
	if isGeFulTrieNumber(number) {
		if s.Ful.Root() != roothash {
			return s, errors.New("Ful root hash is not same,head:" + roothash.String() + "cal:" + s.Ful.Root().String())
		}
	}
	return s,nil
}

func (s *Snapshot) initFulBalance() {
	delFul:=make([]common.Address,0)
	for target, item := range s.FULBalance {
		s.Ful.Add(target,item.Balance)
		delFul=append(delFul,target)
	}
	for _,target:=range delFul{
		delete(s.FULBalance,target)
	}
}

func (snap *Snapshot) updateFULBalanceCost(flowReport []MinerFlowReportRecord, headerNumber *big.Int) {
	for _, items := range flowReport {
		for _, flow := range items.ReportContent {
			if flow.FlowValue2 > 0 {
				address :=flow.Target
				cost:=snap.calCostFul(flow.FlowValue2)
				balance:=snap.Ful.Get(address)
				if balance.Cmp(cost)>0 {
					snap.Ful.Sub(address,cost)
				}else{
					snap.Ful.Del(address)
				}
			}
		}
	}
}

func (s *Snapshot) calCostFul(value uint64) *big.Int {
	flowValue:=value
	cost := new(big.Int).Mul(new(big.Int).SetUint64(flowValue), new(big.Int).SetUint64(calFlowToFULRatio))
	return cost
}

func (snap *Snapshot) AddFlowReportCostFul(flowReport []MinerFlowReportRecord, costFul map[common.Address]*big.Int) map[common.Address]*big.Int {
	for _, items := range flowReport {
		for _, flow := range items.ReportContent {
			if flow.FlowValue2 > 0 {
				address :=flow.Target
				cost:=snap.calCostFul(flow.FlowValue2)
				if _, ok := costFul[address]; !ok {
					costFul[address]=new(big.Int).Set(cost)
				}else{
					costFul[address]=new(big.Int).Add(costFul[address],cost)
				}
			}
		}
	}
	return costFul
}

func (snap *Snapshot) CheckFulEnough(flowReport []MinerFlowReportRecord, report MinerFlowReportRecord) bool {
	costFul:=make(map[common.Address]*big.Int)
	costFul=snap.AddFlowReportCostFul(flowReport,costFul)
	costFul=snap.AddFlowReportCostFul([]MinerFlowReportRecord{report},costFul)
	fulEnough:=true
	for target,value:=range costFul {
		fulBal:=snap.Ful.Get(target)
		if fulBal.Cmp(value)<0 {
			fulEnough=false
		}
	}
	return fulEnough
}

func (snap *Snapshot) checkFulEnoughItem(flowValue uint64,fromFulBalance *big.Int) (*big.Int,bool) {
	costFul:=snap.calCostFul(flowValue)
	fulEnough:=true
	if fromFulBalance.Cmp(costFul)<0 {
		fulEnough=false
	}
	return costFul,fulEnough
}

func (snap *Snapshot) calAllCostFul(flowReport []MinerFlowReportRecord, report MinerFlowReportRecord) map[common.Address]*big.Int {
	costFul:=make(map[common.Address]*big.Int)
	costFul=snap.AddFlowReportCostFul(flowReport,costFul)
	costFul=snap.AddFlowReportCostFul([]MinerFlowReportRecord{report},costFul)
	return costFul
}

func (s *Snapshot) checkReportNumber(reportNumber decimal.Decimal, number uint64) bool {
	reportDay:=reportNumber.BigInt().Uint64()/s.getBlockPreDay()
	blockDay:=number/s.getBlockPreDay()
	return (reportDay==blockDay||reportDay==(blockDay-1))&&reportNumber.BigInt().Uint64()<=number
}