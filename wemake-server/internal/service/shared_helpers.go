package service

import (
	"errors"
	"time"

	"github.com/yourusername/wemake/internal/domainutil"
)

var ErrReviewImagesInvalid = errors.New("image_urls must contain at most 5 unique urls")

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
