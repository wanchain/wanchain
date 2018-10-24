package step

import (
	"github.com/wanchain/go-wanchain/log"
	"github.com/wanchain/go-wanchain/p2p/discover"
	mpcprotocol "github.com/wanchain/go-wanchain/storeman/storemanmpc/protocol"
	mpcsyslog "github.com/wanchain/go-wanchain/storeman/syslog"
	"math/big"
	"math/rand"
	"time"
)

type RequestMpcStep struct {
	BaseStep
	messageType int64
	accType     []byte
	txHash      big.Int
	address     big.Int
	chainID     big.Int
	chainType   []byte
	signType	[]byte
	txCode      []byte
	message     map[discover.NodeID]bool
}

func (req *RequestMpcStep) InitStep(result mpcprotocol.MpcResultInterface) error {
	log.Info("RequestMpcStep.InitStep begin")
	mpcsyslog.Info("RequestMpcStep.InitStep begin")

	if req.messageType == mpcprotocol.MpcCreateLockAccountLeader {
		findMap := make(map[uint64]bool)
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < len(*req.peers); i++ {
			for {
				(*req.peers)[i].Seed = (uint64)(rand.Intn(0x0FFFFFE) + 1)
				_, exist := findMap[(*req.peers)[i].Seed]
				if exist {
					continue
				}

				findMap[(*req.peers)[i].Seed] = true
				break
			}
		}

		accType, err := result.GetByteValue(mpcprotocol.MpcStmAccType)
		if err != nil {
			return err
		}

		req.accType = accType
		log.Info("RequestMpcStep.InitStep", "accType", string(accType[:]))

	} else if req.messageType == mpcprotocol.MpcTXSignLeader {
		addr, err := result.GetValue(mpcprotocol.MpcAddress)
		if err != nil {
			return err
		}

		req.address = addr[0]
		req.chainType, err = result.GetByteValue(mpcprotocol.MpcChainType)
		if err != nil {
			return err
		}

		req.signType, err = result.GetByteValue(mpcprotocol.MpcSignType)
		if err != nil {
			return err
		}

		req.txCode, err = result.GetByteValue(mpcprotocol.MpcTransaction)
		if err != nil {
			return err
		}

		if string(req.chainType) != "BTC" {
			hash, err := result.GetValue(mpcprotocol.MpcTxHash + "_0")
			if err != nil {
				return err
			}

			req.txHash = hash[0]
			chainID, err := result.GetValue(mpcprotocol.MpcChainID)
			if err != nil {
				return err
			}

			req.chainID = chainID[0]
		}
	}

	return nil
}

func CreateRequestMpcStep(peers *[]mpcprotocol.PeerInfo, messageType int64) *RequestMpcStep {
	return &RequestMpcStep{BaseStep: *CreateBaseStep(peers, len(*peers)-1), messageType: messageType, message: make(map[discover.NodeID]bool)}
}

func (req *RequestMpcStep) CreateMessage() []mpcprotocol.StepMessage {
	msg := mpcprotocol.StepMessage{
		Msgcode:mpcprotocol.RequestMPC,
		PeerID:nil,
		Peers:req.peers,
		Data:nil,
		BytesData:nil}
	msg.Data = make([]big.Int, 1)
	msg.Data[0].SetInt64(req.messageType)
	if req.messageType == mpcprotocol.MpcTXSignLeader {
		msg.Data = append(msg.Data, req.txHash)
		msg.Data = append(msg.Data, req.address)
		msg.Data = append(msg.Data, req.chainID)
		msg.BytesData = make([][]byte, 3)
		msg.BytesData[0] = req.chainType
		msg.BytesData[1] = req.txCode
		msg.BytesData[2] = req.signType
	} else if req.messageType == mpcprotocol.MpcCreateLockAccountLeader {
		msg.BytesData = make([][]byte, 1)
		msg.BytesData[0] = req.accType
	}

	return []mpcprotocol.StepMessage{msg}
}

func (req *RequestMpcStep) FinishStep(result mpcprotocol.MpcResultInterface, mpc mpcprotocol.StoremanManager) error {
	err := req.BaseStep.FinishStep()
	if err != nil {
		return err
	}

	data := make([]big.Int, 1)
	data[0].SetInt64(req.messageType)
	result.SetValue(mpcprotocol.MPCActoin, data)
	return nil
}

func (req *RequestMpcStep) HandleMessage(msg *mpcprotocol.StepMessage) bool {
	log.Info("RequestMpcStep.HandleMessage begin", "peerID", msg.PeerID)
	_, exist := req.message[*msg.PeerID]
	if exist {
		log.Error("RequestMpcStep.HandleMessage, get message from peerID fail", "peer", msg.PeerID.String())
		mpcsyslog.Err("RequestMpcStep.HandleMessage, get message from peerID fail. peer:%s", msg.PeerID.String())
		return false
	}

	req.message[*msg.PeerID] = true
	return true
}

