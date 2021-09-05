package main

import (
	"github.com/faiface/pixel"
	"github.com/racerxdl/riscv-emulator/devices/vga"
	"golang.org/x/image/colornames"
	"image/color"
)

// PixelVGA is a VGA Device to PixelGL bridge
type PixelVGA struct {
	VGA       *vga.VGA
	buffer    *pixel.PictureData
	colorbuff []color.RGBA
}

// Creates a new PixelVGA with a screen buffer of the specified size
func MakePixelVGA(width, height int) *PixelVGA {
	p := &PixelVGA{
		VGA:    vga.NewVGA(width, height),
		buffer: pixel.MakePictureData(pixel.R(0, 0, float64(width), float64(height))),
	}
	ClearPictureData(p.buffer, colornames.Black)
	return p
}

// GetPicture returns a pixel picturedata and increments the VGA frame count
func (vga *PixelVGA) GetPicture() *pixel.PictureData {
	vga.buffer.Pix = vga.VGA.GetBuffer(vga.buffer.Pix)
	vga.VGA.IncFrame()
	return vga.buffer
}
