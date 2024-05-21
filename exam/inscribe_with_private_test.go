package exam

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/simbahebinbo/go-ord-tx/ord"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestInscribeWithPrivate(t *testing.T) {
	// 使用 regtest 参数配置
	netParams := &chaincfg.RegressionNetParams

	connCfg := &rpcclient.ConnConfig{
		Host:         "52.221.9.230:28332/wallet/newwallet",
		User:         "testuser",
		Pass:         "123456",
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Shutdown()

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	// you can get from `client.ListUnspent()`
	utxoAddress := "bcrt1q2ak2wdyqaysxplyp0dt7l9hjgnxqs2va2f5uyp"
	address, err := btcutil.DecodeAddress(utxoAddress, netParams)
	if err != nil {
		log.Fatalf("decode address err %v", err)
	}
	log.Printf("address: %v", address)
	unspentList, err := client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{address})

	if err != nil {
		log.Fatalf("list unspentList err %v", err)
	}

	if len(unspentList) == 0 {
		log.Fatalf("unspentList is empty")
	}

	for i := range unspentList {
		inTxid, err := chainhash.NewHashFromStr(unspentList[i].TxID)
		if err != nil {
			log.Fatalf("decode in hash err %v", err)
		}
		commitTxOutPointList = append(commitTxOutPointList, wire.NewOutPoint(inTxid, unspentList[i].Vout))
	}
	log.Printf("commitTxOutPointList: %v", commitTxOutPointList)

	if len(commitTxOutPointList) == 0 {
		log.Fatalf("commitTxOutPointList is empty")
	}

	dataList := make([]ord.InscriptionData, 0)

	dataList = append(dataList, ord.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte("Create for Alice"),
		Destination: "tb1p3m6qfu0mzkxsmaue0hwekrxm2nxfjjrmv4dvy94gxs8c3s7zns6qcgf8ef",
	})

	dataList = append(dataList, ord.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte("Create for Bob"),
		Destination: "tb1pkz6c8cpsszcdq8n2qf8msk45qxmgpl8prwrs544305ew6vrrwc8spraf2z",
	})

	dataList = append(dataList, ord.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte("Create for Charlie"),
		Destination: "tb1pvxylf6kejgfa0jnp0e98xhajwwuqw55m0v37p0d8ywr6ang03hhqxmmfh2",
	})

	dataList = append(dataList, ord.InscriptionData{
		ContentType: "image/jpeg",
		Body:        readFile(),
		Destination: "tb1p3m6qfu0mzkxsmaue0hwekrxm2nxfjjrmv4dvy94gxs8c3s7zns6qcgf8ef",
	})

	request := ord.InscriptionRequest{
		CommitTxOutPointList: commitTxOutPointList,
		CommitFeeRate:        2,
		FeeRate:              1,
		DataList:             dataList,
		SingleRevealTxOnly:   false,
	}

	tool, err := ord.NewInscriptionTool(netParams, client, &request)
	if err != nil {
		log.Fatalf("Failed to create inscription tool: %v", err)
	}

	commitTxHash, revealTxHashList, inscriptions, fees, err := tool.Inscribe()
	if err != nil {
		log.Fatalf("send tx err, %v", err)
	}
	log.Println("commitTxHash, " + commitTxHash.String())
	for i := range revealTxHashList {
		log.Println("revealTxHash, " + revealTxHashList[i].String())
	}
	for i := range inscriptions {
		log.Println("inscription, " + inscriptions[i])
	}
	log.Println("fees: ", fees)
}

func readFile() []byte {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory, %v", err)
	}
	filePath := fmt.Sprintf("%s/1.jpeg", workingDir)
	// if file size too max will return sendrawtransaction RPC error: {"code":-26,"message":"tx-size"}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file %v", err)
	}

	contentType := http.DetectContentType(fileContent)
	log.Printf("file contentType %s", contentType)

	return fileContent
}
