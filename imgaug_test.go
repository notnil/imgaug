package imgaug_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"math/rand"
	"os"
	"testing"

	"github.com/notnil/imgaug"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/fogleman/gg.v1"
)

type testCase struct {
	Name   string
	Tm     imgaug.Transformer
	Labels imgaug.Labels
}

var (
	StandardLabels = imgaug.Labels{
		BBoxes: []image.Rectangle{image.Rect(26, 9, 110, 129)},
	}
)

var (
	testCases = []testCase{
		{
			Name:   "noop_001",
			Tm:     imgaug.Noop{},
			Labels: StandardLabels,
		},
		{
			Name:   "crop_001",
			Tm:     imgaug.Crop(image.Rect(25, 25, 100, 100)),
			Labels: StandardLabels,
		},
		{
			Name:   "pad_001",
			Tm:     imgaug.Pad{Side: imgaug.Left, Pixels: 20},
			Labels: StandardLabels,
		},
		{
			Name:   "flip_lr_001",
			Tm:     imgaug.FlipLR{},
			Labels: StandardLabels,
		},
		{
			Name:   "flip_ud_001",
			Tm:     imgaug.FlipUD{},
			Labels: StandardLabels,
		},
		{
			Name: "resize_001",
			Tm: imgaug.Resize{
				Sizer: imgaug.FixedResizer{Width: 50, Height: 50},
				Algs:  []imgaug.ResizeAlg{imgaug.NearestNeighbor},
			},
			Labels: imgaug.Labels{},
		},
		{
			Name: "seqential_001",
			Tm: imgaug.Sequential(
				[]imgaug.Transformer{
					imgaug.FlipLR{},
					imgaug.Pad{Side: imgaug.Right, Pixels: 10},
					imgaug.FlipUD{},
					imgaug.Crop(image.Rect(10, 10, 140, 120)),
				},
			),
			Labels: StandardLabels,
		},
		{
			Name: "seqential_sometimes_001",
			Tm: imgaug.Sequential(
				[]imgaug.Transformer{
					imgaug.Sometimes{P: 0.33, Transformer: imgaug.FlipLR{}},
					imgaug.Sometimes{P: 0.33, Transformer: imgaug.Pad{Side: imgaug.Right, Pixels: 10}},
					imgaug.Sometimes{P: 0.33, Transformer: imgaug.FlipUD{}},
					imgaug.Sometimes{P: 0.33, Transformer: imgaug.Crop(image.Rect(10, 10, 140, 120))},
				},
			),
			Labels: StandardLabels,
		},
	}
)

func TestTransforms(t *testing.T) {
	g := goldie.New(t)
	img := getImage()
	r := rand.New(rand.NewSource(42))
	cfg := imgaug.NewConfig(r, 20, 0.1)
	for _, tc := range testCases {
		t.Log(tc.Name)
		imgOut, labelsOut := tc.Tm.Transform(cfg, img, tc.Labels)
		drawImg := drawLabels(imgOut, labelsOut)
		g.Assert(t, tc.Name+".jpeg", imgToBytes(imgOut))
		g.Assert(t, tc.Name+".viz.jpeg", imgToBytes(drawImg))
		g.AssertJson(t, tc.Name+".json", labelsOut)
	}

}

func imgToBytes(img image.Image) []byte {
	buf := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func drawLabels(img image.Image, labels imgaug.Labels) image.Image {
	dc := gg.NewContextForImage(img)
	dc.SetRGB(1.0, 0.0, 0.0)
	dc.SetLineWidth(2)
	for _, v := range labels.BBoxes {
		dc.DrawLine(float64(v.Min.X), float64(v.Min.Y), float64(v.Max.X), float64(v.Min.Y))
		dc.Stroke()

		dc.DrawLine(float64(v.Max.X), float64(v.Min.Y), float64(v.Max.X), float64(v.Max.Y))
		dc.Stroke()

		dc.DrawLine(float64(v.Max.X), float64(v.Max.Y), float64(v.Min.X), float64(v.Max.Y))
		dc.Stroke()

		dc.DrawLine(float64(v.Min.X), float64(v.Max.Y), float64(v.Min.X), float64(v.Min.Y))
		dc.Stroke()
	}
	return dc.Image()
}

func getImage() image.Image {
	// 192x129
	f, err := os.Open("fixtures/image.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}
