package ethsvc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/clog"
)

type (

	// Ethereum is a ethereum-compatible network service
	Ethereum interface {
		// Balance returns balance of the network coin (ETH for Ethereum, BNB for BNB Smart Chain)
		Balance(ctx context.Context, address string) (decimal.Decimal, error)
	}

	// Ethereum-compatible network service
	// BNB Smart chain, Polygon etc networks is also supported with this services
	ethereumSvc struct {
		url      string // you can consult for this parameter at https://chainlist.org/
		decimals int32
	}
)

// NewEthereum creates a service
func NewEthereum(url string) Ethereum {
	return &ethereumSvc{
		url:      url,
		decimals: 18, // Ethereum standard decimals
	}
}

// Balance returns balance of the network coin (ETH for Ethereum, BNB for BNB Smart Chain)
func (s *ethereumSvc) Balance(ctx context.Context, address string) (decimal.Decimal, error) {
	client, err := ethclient.DialContext(ctx, s.url)
	if err != nil {
		return decimal.Zero, err
	}
	defer client.Close() // it should be good in the future to preserve client object between requests

	l := clog.Ctx(ctx).With().Str("url", s.url).Logger()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		l.Warn().Err(err).Msg("Unable to acquire chainID")
	} else {
		l.Debug().Str("url", s.url).Msgf("ChainID is %d", chainID)
		l = l.With().Stringer("chainID", chainID).Logger()
	}

	addr := common.HexToAddress(address)

	balance, err := client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(balance, -1*s.decimals), nil
}
