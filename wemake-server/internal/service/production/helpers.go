package production

import (
	"time"

	"github.com/yourusername/wemake/internal/domainutil"
)

var thailandLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

func roundCurrency(v float64) float64 {
	return domainutil.RoundMoney(v)
}

func percentOf(amount, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return roundCurrency((amount / total) * 100)
}
