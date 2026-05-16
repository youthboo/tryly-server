package service

import (
	"errors"

	"github.com/yourusername/wemake/internal/helper"
)

var ErrReviewImagesInvalid = errors.New("image_urls must contain at most 5 unique urls")

var (
	thailandLocation = helper.ThailandLocation
	roundCurrency    = helper.RoundCurrency
	percentOf        = helper.PercentOf
	derefString      = helper.DerefString
)
