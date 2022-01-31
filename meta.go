package imgaug

import "image"

type Sequential []Transformer

func (s Sequential) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	for _, t := range s {
		img, labels = t.Transform(cfg, img, labels)
	}
	return img, labels
}

type Sometimes struct {
	P           float64
	Transformer Transformer
}

func (s Sometimes) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	if s.P > cfg.r.Float64() {
		return s.Transformer.Transform(cfg, img, labels)
	}
	return Noop{}.Transform(cfg, img, labels)
}

type SomeOf struct {
	N IntRange
	T []Transformer
}

func (s SomeOf) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	cp := append([]Transformer{}, s.T...)
	cfg.r.Shuffle(len(s.T), func(i, j int) {
		cp[i], cp[j] = cp[j], cp[i]
	})
	n := s.N.Int(cfg.r)
	return Sequential(cp[:n]).Transform(cfg, img, labels)
}

type OneOf struct {
	T []Transformer
}

func (o OneOf) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	return SomeOf{N: IntRange{Min: 1, Max: 1}, T: o.T}.Transform(cfg, img, labels)
}

type Noop struct{}

func (n Noop) Transform(cfg *Config, img image.Image, labels []interface{}) (image.Image, []interface{}) {
	return img, labels
}
