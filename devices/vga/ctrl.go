package vga

// IncFrame increments frame counter in VGA
func (vga *VGA) IncFrame() {
	vga.frameCount++
}

// VBlank sets the vblank status flag
func (vga *VGA) VBlank(on bool) {
	vga.vblank = 0
	if on {
		vga.vblank = 1
	}
}
