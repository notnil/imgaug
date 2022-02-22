package imgaug

import "image"

func Sequential(ts ...Transformer) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		for _, t := range ts {
			img, labels = t.Transform(cfg, img, labels)
		}
		return img, labels
	})
}

func Sometimes(p float64, t Transformer) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		if p > cfg.r.Float64() {
			return t.Transform(cfg, img, labels)
		}
		return Noop().Transform(cfg, img, labels)
	})
}

func SomeOf(r IntRange, ts ...Transformer) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		cp := append([]Transformer{}, ts...)
		cfg.r.Shuffle(len(ts), func(i, j int) {
			cp[i], cp[j] = cp[j], cp[i]
		})
		n := r.Int(cfg.r)
		if n < 0 {
			n = 0
		}
		if n >= len(cp) {
			n = len(cp) - 1
		}
		return Sequential(cp[:n]...).Transform(cfg, img, labels)
	})
}

func OneOf(ts ...Transformer) Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		return SomeOf(IntRange{Min: 1, Max: 1}, ts...).Transform(cfg, img, labels)
	})
}

func Noop() Transformer {
	return TransformerFunc(func(cfg *Config, img image.Image, labels Labels) (image.Image, Labels) {
		return img, labels
	})
}
