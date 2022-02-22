package imgaug

import (
	"image"

	"github.com/disintegration/imaging"
)

func FlipLR() Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		img = imaging.FlipH(img)
		newLabels := Labels{}
		for _, v := range labels.BBoxes {
			dx := img.Bounds().Dx()
			r := image.Rectangle{
				Min: image.Pt(dx-v.Max.X, v.Min.Y),
				Max: image.Pt(dx-v.Min.X, v.Max.Y),
			}
			newLabels.BBoxes = append(newLabels.BBoxes, r)
		}
		return img, newLabels
	})
}

func FlipUD() Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		img = imaging.FlipV(img)
		newLabels := Labels{}
		for _, v := range labels.BBoxes {
			dy := img.Bounds().Dy()
			r := image.Rectangle{
				Min: image.Pt(v.Min.X, dy-v.Max.Y),
				Max: image.Pt(v.Max.X, dy-v.Min.Y),
			}
			newLabels.BBoxes = append(newLabels.BBoxes, r)
		}
		return img, newLabels
	})
}
