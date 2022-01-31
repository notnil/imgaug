package imgaug

import "image"

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}
