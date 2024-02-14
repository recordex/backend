package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

// GetNewestFileMetadata は引数で指定されたファイル名の最新のファイルメタデータを取得します。
func GetNewestFileMetadata(ctx context.Context, fileName string) (*record.RecordFileMetadata, error) {
	contractAddress := common.HexToAddress(config.Get().RecordContractAddress)
	recordContract, err := record.NewRecordCaller(contractAddress, GetEthClient())
	if err != nil {
		return nil, xerrors.Errorf("recordContract の初期化に失敗しました。contractAddress -> %s: %+v", contractAddress, err)
	}

	fileMedata, err := recordContract.GetFileMetadataHistory(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	}, fileName)
	if err != nil {
		return nil, xerrors.Errorf("fileMedata の取得に失敗しました。fileName -> %s: %+v", fileName, err)
	}

	return &fileMedata[len(fileMedata)-1], nil
}

// IsTransactionPending は引数で指定されたトランザクションがイーサリアムネットワークに送られ、そのトランザクションが現在進行中かどうかをチェックします。
func IsTransactionPending(ctx context.Context, transactionHash string) (bool, error) {
	commonTxHash := common.HexToHash(transactionHash)
	_, isPending, err := GetEthClient().TransactionByHash(ctx, commonTxHash)
	if err != nil {
		return false, xerrors.Errorf("トランザクションの取得に失敗しました。transactionHash -> %s: %+v", transactionHash, err)
	}
	return isPending, nil
}

// IsRecordTransactionHashValid はトランザクション ID が正しいかどうかをチェックします。
// 引数で指定されたトランザクション ID のデータに記録されている FileHash と引数の fileHash が一致するかどうかをチェックし、一致していた場合は true を返します。
func IsRecordTransactionHashValid(ctx context.Context, transactionHash string, fileHash string) (bool, error) {
	// transactionHash がイーサリアムネットワークに送られているかを確認
	commonTxHash := common.HexToHash(transactionHash)
	isPending, err := IsTransactionPending(ctx, transactionHash)
	if err != nil {
		return false, xerrors.Errorf("トランザクションが進行中かの確認に失敗しました。transactionHash -> %s: %+v", transactionHash, err)
	}
	if isPending {
		return false, xerrors.Errorf("トランザクションが未確定です。transactionHash -> %s", transactionHash)
	}

	// トランザクションレシートから引数で指定されているトランザクションが成功したかを確認
	receipt, err := GetEthClient().TransactionReceipt(ctx, commonTxHash)
	if err != nil {
		return false, xerrors.Errorf("トランザクションレシートの取得に失敗しました。transactionHash -> %s: %+v", transactionHash, err)
	}
	if receipt.Status != 1 {
		return false, xerrors.Errorf("トランザクションが eth ネットワーク上で失敗しています。transactionHash -> %s", transactionHash)
	}

	contractAddress := common.HexToAddress(config.Get().RecordContractAddress)
	recordFilterer, err := record.NewRecordFilterer(contractAddress, GetEthClient())
	if err != nil {
		return false, xerrors.Errorf("recordFilterer の初期化に失敗しました。contractAddress -> %s: %+v", contractAddress, err)
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: receipt.BlockNumber,
		ToBlock:   receipt.BlockNumber,
	}
	logs, err := GetEthClient().FilterLogs(ctx, query)
	if err != nil {
		return false, xerrors.Errorf("トランザクションログの取得に失敗しました。transactionHash -> %s: %+v", transactionHash, err)
	}
	if len(logs) == 0 {
		return false, xerrors.Errorf("トランザクションログが存在しません。transactionHash -> %s", transactionHash)
	}

	for _, vLog := range logs {
		event, err := recordFilterer.ParseFileAdded(vLog)
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
