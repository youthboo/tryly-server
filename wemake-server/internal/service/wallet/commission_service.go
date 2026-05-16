package wallet

import (
	"errors"

	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
	platformrepo "github.com/yourusername/wemake/internal/repository/platform_config"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

var ErrCommissionConfigMissing = errors.New("COMMISSION_CONFIG_MISSING")

type Breakdown struct {
	Subtotal                 float64 `json:"subtotal"`
	DiscountAmount           float64 `json:"discount_amount"`
	ShippingCost             float64 `json:"shipping_cost"`
	PackagingCost            float64 `json:"packaging_cost"`
	ToolingMoldCost          float64 `json:"tooling_mold_cost"`
	PreVatBase               float64 `json:"pre_vat_base"`
	VatRate                  float64 `json:"vat_rate"`
	VatAmount                float64 `json:"vat_amount"`
	GrandTotal               float64 `json:"grand_total"`
	PlatformCommissionRate   float64 `json:"platform_commission_rate"`
	PlatformCommissionAmount float64 `json:"platform_commission_amount"`
	FactoryNetReceivable     float64 `json:"factory_net_receivable"`
	PlatformConfigID         int64   `json:"platform_config_id"`
}

type CommissionInput struct {
	Items          []domain.QuotationItem
	DiscountAmount float64
	ShippingCost   float64
	PackagingCost  float64
	ToolingCost    float64
	FactoryID      *int64
}

type CommissionService struct {
	configs   *platformrepo.PlatformConfigRepository
	overrides *walletrepo.CommissionRepository
}

func NewCommissionService(configs *platformrepo.PlatformConfigRepository, overrides *walletrepo.CommissionRepository) *CommissionService {
	return &CommissionService{configs: configs, overrides: overrides}
}

func (s *CommissionService) Calculate(in CommissionInput) (*Breakdown, error) {
	var cfg *domain.PlatformConfig
	var err error
	if in.FactoryID != nil {
		cfg, err = s.configs.GetByFactoryID(*in.FactoryID)
	} else {
		cfg, err = s.configs.GetDefault()
	}
	if err != nil {
		return nil, ErrCommissionConfigMissing
	}
	var lineSum float64
	for i := range in.Items {
		lineTotal := domainutil.RoundMoney(in.Items[i].Qty * in.Items[i].UnitPrice * (1 - in.Items[i].DiscountPct/100))
		in.Items[i].LineTotal = lineTotal
		lineSum += lineTotal
	}
	subtotal := domainutil.RoundMoney(lineSum - in.DiscountAmount)
	preVatBase := domainutil.RoundMoney(subtotal + in.ShippingCost + in.PackagingCost + in.ToolingCost)
	vatAmount := domainutil.RoundMoney(preVatBase * cfg.VatRate / 100)
	grandTotal := domainutil.RoundMoney(preVatBase + vatAmount)
	commissionRate := cfg.DefaultCommissionRate
	commissionAmount := domainutil.RoundMoney(grandTotal * commissionRate / 100)
	factoryNet := domainutil.RoundMoney(grandTotal - commissionAmount)

	return &Breakdown{
		Subtotal:                 subtotal,
		DiscountAmount:           domainutil.RoundMoney(in.DiscountAmount),
		ShippingCost:             domainutil.RoundMoney(in.ShippingCost),
		PackagingCost:            domainutil.RoundMoney(in.PackagingCost),
		ToolingMoldCost:          domainutil.RoundMoney(in.ToolingCost),
		PreVatBase:               preVatBase,
		VatRate:                  cfg.VatRate,
		VatAmount:                vatAmount,
		GrandTotal:               grandTotal,
		PlatformCommissionRate:   commissionRate,
		PlatformCommissionAmount: commissionAmount,
		FactoryNetReceivable:     factoryNet,
		PlatformConfigID:         cfg.ConfigID,
	}, nil
}
