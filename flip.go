package imgaug

import (
	"image"

	"github.com/disintegration/imaging"
)

type FlipLR struct{}

func (f FlipLR) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	img = imaging.FlipH(img)
	out := []interface{}{}
	for _, l := range labels {
		switch v := l.(type) {
		case image.Rectangle:
			dx := img.Bounds().Dx()
			r := image.Rectangle{
				Min: image.Pt(dx-v.Max.X, v.Min.Y),
				Max: image.Pt(dx-v.Min.X, v.Max.Y),
			}
			out = append(out, r)
		}
	}
	return img, out
}

type FlipUD struct{}

func (f FlipUD) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	img = imaging.FlipV(img)
	out := []interface{}{}
	for _, l := range labels {
		switch v := l.(type) {
		case image.Rectangle:
			dy := img.Bounds().Dy()
			r := image.Rectangle{
				Min: image.Pt(v.Min.X, dy-v.Max.Y),
				Max: image.Pt(v.Max.X, dy-v.Min.Y),
			}
			out = append(out, r)
		}
	}
	return img, out
}
