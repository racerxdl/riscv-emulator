package main

import (
	"github.com/faiface/pixel"
	"image/color"
)

var screenOrigin pixel.Matrix

// ClearPictureData sets all pixel colors for the specified color
func ClearPictureData(p *pixel.PictureData, c color.Color) {
	t := p.Bounds()
	t.Size()
	ts := t.Min
	te := t.Max
	nc := ToRGBA(c)
	for x := ts.X; x < te.X; x++ {
		for y := ts.Y; y < te.Y; y++ {
			p.Pix[int(x)+int(y)*p.Stride] = nc
		}
	}
}

// ToRGBA converts a generic color to RGBA color
func ToRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

// MoveAndScaleTo creates a matrix that moves the specified picture to specified position and scales the image.
// The scale affects the x/y coordinates
// The pixel coordinate is based in the object center
func MoveAndScaleTo(p pixel.Picture, x, y, s float64) pixel.Matrix {
	return pixel.IM.
		Moved(pixel.V(p.Bounds().W()/2+x/s, p.Bounds().H()/2+y/s)).
		Scaled(pixel.V(0, 0), s).
		Chained(screenOrigin)
}
