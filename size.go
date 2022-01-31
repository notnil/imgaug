package imgaug

import (
	"image"
	"image/draw"
	"math"

	"github.com/disintegration/imaging"
)

type Crop image.Rectangle

func (c Crop) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	img = imaging.Crop(img, image.Rectangle(c))
	out := []interface{}{}
	for _, l := range labels {
		switch v := l.(type) {
		case image.Rectangle:
			r := v.Intersect(image.Rectangle(c))
			r = image.Rectangle{
				Min: r.Min.Sub(c.Min),
				Max: r.Max.Sub(c.Min),
			}
			if cfg.keepBbox(image.Rectangle(c), r) {
				out = append(out, r)
			}
		}
	}
	return img, out
}

type Side int8

const (
	Left Side = iota
	Top
	Right
	Bottom
)

type Pad struct {
	Side   Side
	Pixels int
}

func (p Pad) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	bnds := img.Bounds()
	r := image.Rectangle{}
	switch p.Side {
	case Left:
		offset := image.Pt(p.Pixels, 0)
		bnds = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Add(offset)}
		r = image.Rectangle{Min: offset, Max: bnds.Max}
	case Right:
		offset := image.Pt(p.Pixels, 0)
		bnds = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Add(offset)}
		r = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Sub(offset)}
	case Top:
		offset := image.Pt(0, p.Pixels)
		bnds = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Add(offset)}
		r = image.Rectangle{Min: offset, Max: bnds.Max}
	case Bottom:
		offset := image.Pt(0, p.Pixels)
		bnds = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Add(offset)}
		r = image.Rectangle{Min: image.Pt(0, 0), Max: bnds.Max.Sub(offset)}
	}
	dImg := image.NewRGBA(bnds)
	draw.Draw(dImg, r, img, image.ZP, draw.Over)
	out := []interface{}{}
	for _, l := range labels {
		switch v := l.(type) {
		case image.Rectangle:
			r := image.Rectangle{
				Min: v.Min.Add(r.Min),
				Max: v.Max.Add(r.Min),
			}
			out = append(out, r)
		}
	}
	return dImg, out
}

type Sizer interface {
	Size(cfg *Config, bnds image.Point) image.Point
}

type FixedResizer struct {
	Width  int
	Height int
}

func (fr FixedResizer) Size(cfg *Config, bnds image.Point) image.Point {
	return image.Pt(fr.Width, fr.Height)
}

type MultiplyResizer struct {
	Width  FloatRange
	Height FloatRange
}

func (mr MultiplyResizer) Size(cfg *Config, bnds image.Point) image.Point {
	wx := mr.Width.Float(cfg.r)
	hx := mr.Height.Float(cfg.r)
	return resizePoint(bnds, wx, hx)
}

type Resize struct {
	Width int
	Heigh int
	Alg   imaging.ResampleFilter
}

func (r Resize) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	ogBnds := img.Bounds()
	img = imaging.Resize(img, r.Width, r.Heigh, r.Alg)
	xRatio := float64(img.Bounds().Dx()) / float64(ogBnds.Dx())
	yRatio := float64(img.Bounds().Dy()) / float64(ogBnds.Dy())
	out := []interface{}{}
	for _, l := range labels {
		switch v := l.(type) {
		case image.Rectangle:
			r := image.Rectangle{
				Min: resizePoint(v.Min, xRatio, yRatio),
				Max: resizePoint(v.Max, xRatio, yRatio),
			}
			out = append(out, r)
		}
	}
	return img, out
}

func resizePoint(pt image.Point, xRatio, yRatio float64) image.Point {
	x := int(math.Round(float64(pt.X) * xRatio))
	y := int(math.Round(float64(pt.Y) * yRatio))
	return image.Pt(x, y)
}

type RandResize struct {
}
