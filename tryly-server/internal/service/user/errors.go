package user

import "errors"

var ErrReviewImagesInvalid = errors.New("image_urls must contain at most 5 unique urls")
