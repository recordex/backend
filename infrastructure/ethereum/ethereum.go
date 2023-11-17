package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/recordex/smartcontract/pkg/record"
	"golang.org/x/xerrors"

	"github.com/recordex/backend/config"
)

var ethClient *ethclient.Client

func init() {
	infuraAPIKei := os.Getenv("INFURA_API_KEY")
	if infuraAPIKei == "" {
		log.Fatalf("環境変数 INFURA_API_KEY が設定されていません。INFURA_API_KEY -> %s", infuraAPIKei)
	}

	var err error
	ethClient, err = ethclient.Dial(fmt.Sprintf("%s/%s", config.Get().EthereumNodeURL, infuraAPIKei))
	if err != nil {
		log.Fatalf("イーサリアムノードへの接続に失敗しました。: %v", err)
	}
}

// GetEthClient はイーサリアムクライアントを返します。
func GetEthClient() *ethclient.Client {
	return ethClient
}

// IsRecordTransactionHashValid はトランザクション ID が正しいかどうかをチェックします。
// 引数で指定されたトランザクション ID のデータに記録されている FileHash と引数の fileHash が一致するかどうかをチェックし、一致していた場合は true を返します。
func IsRecordTransactionHashValid(transactionHash string, fileHash string) (bool, error) {
	// transactionHash がイーサリアムネットワークに送られているかを確認
	commonTxHash := common.HexToHash(transactionHash)
	_, isPending, err := GetEthClient().TransactionByHash(context.Background(), commonTxHash)
	if err != nil {
		return false, xerrors.Errorf("トランザクションの取得に失敗しました。transactionHash -> %s: %w", transactionHash, err)
	}
	if isPending {
		return false, xerrors.Errorf("トランザクションが未確定です。transactionHash -> %s", transactionHash)
	}

	// トランザクションレシートから引数で指定されているトランザクションが成功したかを確認
	receipt, err := GetEthClient().TransactionReceipt(context.Background(), commonTxHash)
	if err != nil {
		return false, xerrors.Errorf("トランザクションレシートの取得に失敗しました。transactionHash -> %s: %w", transactionHash, err)
	}
	if receipt.Status != 1 {
		return false, xerrors.Errorf("トランザクションが eth ネットワーク上で失敗しています。transactionHash -> %s", transactionHash)
	}

	contractAddress := common.HexToAddress(config.Get().RecordContractAddress)
	recordContract, err := record.NewRecordFilterer(contractAddress, GetEthClient())
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: receipt.BlockNumber,
		ToBlock:   receipt.BlockNumber,
	}
	logs, err := GetEthClient().FilterLogs(context.Background(), query)
	if err != nil {
		return false, xerrors.Errorf("トランザクションログの取得に失敗しました。transactionHash -> %s: %w", transactionHash, err)
	}
	if len(logs) == 0 {
		return false, xerrors.Errorf("トランザクションログが存在しません。transactionHash -> %s", transactionHash)
	}

	for _, vLog := range logs {
		event, err := recordContract.ParseFileAdded(vLog)
		if err != nil {
			log.Println(err)
			continue
		}

		hexHash := hex.EncodeToString(event.FileHash[:])

		if hexHash == fileHash {
			return true, nil
		}
	}

	return false, nil
}
