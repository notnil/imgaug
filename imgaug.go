package imgaug

import (
	"image"
	"math/rand"
)

type Transformer interface {
	Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{})
}

type Config struct {
	r                 *rand.Rand
	bboxMinArea       int
	bboxMinVisibility float64
}

func NewConfig(r *rand.Rand, bboxMinArea int, bboxMinVisibility float64) *Config {
	return &Config{
		r:                 r,
		bboxMinArea:       bboxMinArea,
		bboxMinVisibility: bboxMinVisibility,
	}
}

func (cfg Config) keepBbox(src, dst image.Rectangle) bool {
	srcArea := src.Max.X * src.Max.Y
	dstArea := dst.Max.X * dst.Max.Y
	// check bboxMinArea
	if dstArea < cfg.bboxMinArea || dstArea == 0 || srcArea == 0 {
		return false
	}
	// check bboxMinVisibility
	ratio := float64(dstArea) / float64(srcArea)
	return ratio >= cfg.bboxMinVisibility
}

type IntRange struct {
	Min int
	Max int
}

func (ir IntRange) Int(r *rand.Rand) int {
	i := r.Intn(ir.Max - ir.Min)
	return i + ir.Min
}

type FloatRange struct {
	Min float64
	Max float64
}

func (fr FloatRange) Float(r *rand.Rand) float64 {
	f := r.Float64() * (fr.Max - fr.Min)
	return fr.Min + f
}
