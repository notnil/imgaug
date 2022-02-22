package imgaug

import (
	"image"
	"image/draw"
	"math"

	"github.com/disintegration/imaging"
)

type Side uint8

const (
	L Side = 1 << iota
	T
	R
	B
	LR  = L | R
	TB  = T | B
	All = LR | TB
)

type CropFunc func(cfg *Config, bnds image.Point) image.Rectangle

func FixedCrop(r image.Rectangle) CropFunc {
	return func(cfg *Config, bnds image.Point) image.Rectangle {
		return r
	}
}

func PercentCrop(sides map[Side]FloatRange) CropFunc {
	return func(cfg *Config, bnds image.Point) image.Rectangle {
		pxSides := percentSides(cfg, bnds, sides)
		return image.Rect(pxSides[L], pxSides[T], bnds.X-pxSides[R], bnds.Y-pxSides[B])
	}
}

func PixelCrop(sides map[Side]IntRange) CropFunc {
	return func(cfg *Config, bnds image.Point) image.Rectangle {
		pxSides := pixelSides(cfg, bnds, sides)
		return image.Rect(pxSides[L], pxSides[T], bnds.X-pxSides[R], bnds.Y-pxSides[B])
	}
}

func Crop(fn CropFunc) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		c := fn(cfg, img.Bounds().Max)
		img = imaging.Crop(img, c)
		newLabels := Labels{}
		for _, v := range labels.BBoxes {
			r := v.Intersect(image.Rectangle(c))
			r = image.Rectangle{
				Min: r.Min.Sub(c.Min),
				Max: r.Max.Sub(c.Min),
			}
			if cfg.keepBbox(image.Rectangle(c), r) {
				newLabels.BBoxes = append(newLabels.BBoxes, r)
			}
		}
		return img, newLabels
	})
}

type PadFunc func(cfg *Config, bnds image.Point) map[Side]int

func PercentPad(sides map[Side]FloatRange) PadFunc {
	return func(cfg *Config, bnds image.Point) map[Side]int {
		return percentSides(cfg, bnds, sides)
	}
}

func PixelPad(sides map[Side]IntRange) PadFunc {
	return func(cfg *Config, bnds image.Point) map[Side]int {
		return pixelSides(cfg, bnds, sides)
	}
}

func Pad(fn PadFunc) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		sides := fn(cfg, img.Bounds().Max)
		var bOffset image.Point
		var rOffset image.Point
		for s, px := range sides {
			switch s {
			case L:
				bOffset = bOffset.Add(image.Pt(px, 0))
				rOffset = rOffset.Add(image.Pt(px, 0))
			case R:
				bOffset = bOffset.Add(image.Pt(px, 0))
			case T:
				bOffset = bOffset.Add(image.Pt(0, px))
				rOffset = rOffset.Add(image.Pt(0, px))
			case B:
				bOffset = bOffset.Add(image.Pt(0, px))
			}
		}
		r := image.Rectangle{Min: rOffset, Max: rOffset.Add(img.Bounds().Max)}
		max := img.Bounds().Max.Add(bOffset)
		dImg := image.NewRGBA(image.Rect(0, 0, max.X, max.Y))
		draw.Draw(dImg, r, img, image.ZP, draw.Over)
		newLabels := Labels{}
		for _, v := range labels.BBoxes {
			r := image.Rectangle{
				Min: v.Min.Add(r.Min),
				Max: v.Max.Add(r.Min),
			}
			newLabels.BBoxes = append(newLabels.BBoxes, r)
		}
		return dImg, newLabels
	})
}

type ResizeFunc func(cfg *Config, bnds image.Point) image.Point

func FixedResize(w, h int) ResizeFunc {
	return func(cfg *Config, bnds image.Point) image.Point {
		return image.Pt(w, h)
	}
}

func PercentResize(w FloatRange, h FloatRange) ResizeFunc {
	return func(cfg *Config, bnds image.Point) image.Point {
		wx := w.Float(cfg.r)
		hx := h.Float(cfg.r)
		return resizePoint(bnds, wx, hx)
	}
}

func PixelResize(w IntRange, h IntRange) ResizeFunc {
	return func(cfg *Config, bnds image.Point) image.Point {
		wx := w.Int(cfg.r)
		hx := h.Int(cfg.r)
		return image.Pt(bnds.X+wx, bnds.Y+hx)
	}
}

func Resize(fn ResizeFunc, algs ...ResizeAlg) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		ogBnds := img.Bounds()
		pt := fn(cfg, img.Bounds().Max)
		alg := NearestNeighbor
		if len(algs) > 0 {
			alg = algs[cfg.r.Intn(len(algs))]
		}
		img = imaging.Resize(img, pt.X, pt.Y, alg.resampleFilter())
		xRatio := float64(img.Bounds().Dx()) / float64(ogBnds.Dx())
		yRatio := float64(img.Bounds().Dy()) / float64(ogBnds.Dy())
		newLabels := Labels{}
		for _, v := range labels.BBoxes {
			r := image.Rectangle{
				Min: resizePoint(v.Min, xRatio, yRatio),
				Max: resizePoint(v.Max, xRatio, yRatio),
			}
			newLabels.BBoxes = append(newLabels.BBoxes, r)
		}
		return img, newLabels
	})
}

func resizePoint(pt image.Point, xRatio, yRatio float64) image.Point {
	x := int(math.Round(float64(pt.X) * xRatio))
	y := int(math.Round(float64(pt.Y) * yRatio))
	return image.Pt(x, y)
}

type ResizeAlg int

const (
	NearestNeighbor ResizeAlg = iota
	Box
	Linear
	Hermite
	MitchellNetravali
	CatmullRom
	BSpline
	Gaussian
	Bartlett
	Lanczos
	Hann
	Hamming
	Blackman
	Welch
	Cosine
)

func (ra ResizeAlg) resampleFilter() imaging.ResampleFilter {
	switch ra {
	case NearestNeighbor:
		return imaging.NearestNeighbor
	case Box:
		return imaging.Box
	case Linear:
		return imaging.Linear
	case Hermite:
		return imaging.Hermite
	case MitchellNetravali:
		return imaging.MitchellNetravali
	case CatmullRom:
		return imaging.CatmullRom
	case BSpline:
		return imaging.BSpline
	case Gaussian:
		return imaging.Gaussian
	case Bartlett:
		return imaging.Bartlett
	case Lanczos:
		return imaging.Lanczos
	case Hann:
		return imaging.Hann
	case Hamming:
		return imaging.Hamming
	case Blackman:
		return imaging.Blackman
	case Welch:
		return imaging.Welch
	case Cosine:
		return imaging.Cosine
	}
	return imaging.NearestNeighbor
}

func percentSides(cfg *Config, bnds image.Point, sides map[Side]FloatRange) map[Side]int {
	m := map[Side]int{}
	cp := sides
	for s, ir := range sides {
		switch s {
		case L, R, T, B:
			cp[s] = ir
		case LR:
			cp[L] = ir
			cp[R] = ir
		case TB:
			cp[T] = ir
			cp[B] = ir
		case All:
			cp[L] = ir
			cp[R] = ir
			cp[T] = ir
			cp[B] = ir
		}
	}
	for s, fr := range cp {
		switch s {
		case L, R:
			f := fr.Float(cfg.r)
			f = math.Abs(f)
			m[s] = resizePoint(bnds, f, f).X
		case T, B:
			f := fr.Float(cfg.r)
			f = math.Abs(f)
			m[s] = resizePoint(bnds, f, f).Y
		}
	}
	return m
}

func pixelSides(cfg *Config, bnds image.Point, sides map[Side]IntRange) map[Side]int {
	m := map[Side]int{}
	cp := sides
	for s, ir := range sides {
		switch s {
		case L, R, T, B:
			cp[s] = ir
		case LR:
			cp[L] = ir
			cp[R] = ir
		case TB:
			cp[T] = ir
			cp[B] = ir
		case All:
			cp[L] = ir
			cp[R] = ir
			cp[T] = ir
			cp[B] = ir
		}
	}
	for s, ir := range cp {
		m[s] = ir.Int(cfg.r)
	}
	return m
}
