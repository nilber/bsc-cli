package ethcli

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

//sudo apt-get install libhidapi-dev

type Native struct {
	client *ethclient.Client
}

func NewNative(client *ethclient.Client) *Native {
	return &Native{
		client: client,
	}
}

func (t *Native) GetAddress(mnemonic string, pathDerivation string) (*common.Address, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath(pathDerivation)
	account, err := wallet.Derive(path, true)
	if err != nil {
		return nil, fmt.Errorf("account %s", err.Error())
	}

	return &account.Address, nil

}

func (t *Native) BalanceOf(address string) (*big.Float, error) {
	account := common.HexToAddress(address)
	balance, err := t.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	return weiToEther(balance, params.Ether), nil
}

func (t *Native) Transfer(req *TransferOpts) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(req.Mnemonic)
	if err != nil {
		return "", fmt.Errorf("mnemonic %s", err.Error())
	}

	path := hdwallet.MustParseDerivationPath(req.Path)
	account, err := wallet.Derive(path, true)
	if err != nil {
		return "", fmt.Errorf("account %s", err.Error())
	}

	fromAddress := account.Address
	nonce, err := t.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("nonce %s", err.Error())
	}

	total := big.NewFloat(req.Amount)
	value := etherToWei(total, 18)
	var data []byte
	gasLimit := uint64(21000) // in units
	gasPrice, err := t.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("gasPrice %s", err.Error())
	}

	toAddress := common.HexToAddress(req.Address)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	sign, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		return "", fmt.Errorf("sign %s", err.Error())
	}
	err = t.client.SendTransaction(context.Background(), sign)
	if err != nil {
		return "", fmt.Errorf("tx %s", err.Error())
	}
	return sign.Hash().Hex(), nil

}
