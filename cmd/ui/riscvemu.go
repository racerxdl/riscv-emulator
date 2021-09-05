package main

import (
	"context"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/racerxdl/riscv-emulator/core"
	"github.com/racerxdl/riscv-emulator/devices/ram"
	"github.com/racerxdl/riscv-emulator/devices/spi"
	"github.com/racerxdl/riscv-emulator/devices/uart"
	"github.com/racerxdl/riscv-emulator/disasm"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"io/ioutil"
	"strings"
	"time"
)

var log = logrus.New()
var riscv *core.RISCV
var disasmText *text.Text
var debugText *text.Text
var stackText *text.Text

// RefreshStack refreshes the picture box that shows the stack content pointed by SP (X2)
func RefreshStack() {
	stackText.Clear()
	stackText.Color = colornames.Black
	fmt.Fprintf(stackText, "Stack: \n\n")
	ctx := context.Background()

	sp := riscv.Registers.GetInteger(2)

	for i := 0; i < 32; i++ {
		off := sp + uint32(i*4)
		if int32(off) < 0 {
			continue
		}
		v, err := riscv.Bus.ReadWord(ctx, off)
		if err != nil {
			stackText.Color = colornames.Red
			fmt.Fprintf(stackText, "%08x: bus err\n", off)
			continue
		}
		fmt.Fprintf(stackText, "%08x: %08x\n", off, v)
	}
}

// RefreshDebug refreshes the picture box that shows the registers
func RefreshDebug() {
	ctx := context.Background()
	opc := riscv.GetPC()
	ins, _ := riscv.Bus.ReadWord(ctx, opc)
	asm := disasm.Disasm(opc, ins)[18:]

	debugText.Clear()
	debugText.Color = colornames.Black
	fmt.Fprintf(debugText, "Registers\n\n")
	for i := 0; i < 32; i++ {
		regName := core.GetIntRegisterName(i)
		debugText.Color = colornames.Black
		if strings.Contains(asm, regName) {
			debugText.Color = colornames.Blue
		}
		regStr := fmt.Sprintf("(%s)", core.GetIntRegisterName(i))
		fmt.Fprintf(debugText, "X%02d %6s = %08x\n", i, regStr, riscv.Registers.GetInteger(uint32(i)))
	}
}

// RefreshDisasm refreshes the pciture box that shows the disassemble
func RefreshDisasm() {
	opc := riscv.GetPC()
	offset := opc &^ 32
	offset -= 16

	ctx := context.Background()

	disasmText.Clear()
	disasmText.Color = colornames.Black
	fmt.Fprint(disasmText, "Disassembler: \n")

	for i := 0; i < 32; i++ {
		off := offset + uint32(i*4)
		if int32(off) < 0 {
			continue
		}
		v, err := riscv.Bus.ReadWord(ctx, off)
		if err != nil {
			disasmText.Color = colornames.Red
			fmt.Fprintf(disasmText, "%08x: bus err\n", off)
			continue
		}
		asm := disasm.Disasm(off, v) + "\n"
		disasmText.Color = colornames.Black
		if off == opc {
			disasmText.Color = colornames.Blue
		}
		fmt.Fprintf(disasmText, asm)
	}
}

func run() {
	log.SetLevel(logrus.DebugLevel)
	// Initialize pixelgl and invert the Y coordinate
	// PixelGL is weird, it uses Y coordinate starting at the bottom of the screen
	// we don't want that
	cfg := pixelgl.WindowConfig{
		Title:  "RISC-V Emulator",
		Bounds: pixel.R(0, 0, 1280, 720),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	screenOrigin = pixel.IM.
		ScaledXY(pixel.V(0, 0), pixel.V(1, -1)).
		Moved(pixel.V(0, win.Bounds().H()))
	// Create a new font atlas for using in strings
	atlas := text.NewAtlas(
		basicfont.Face7x13,
		text.ASCII,
	)

	// Get window width / height
	r := win.Bounds()
	w := r.Max.X
	h := r.Max.Y

	// Create Text Boxes
	memoryMapText := text.New(pixel.V(0, 0), atlas)
	memoryMapText.Color = colornames.Black

	uartText := text.New(pixel.V(0, 0), atlas)
	uartText.Color = colornames.Black
	uartText.WriteString("UART\n")

	disasmText = text.New(pixel.V(0, 0), atlas)
	debugText = text.New(pixel.V(0, 0), atlas)
	stackText = text.New(pixel.V(0, 0), atlas)

	// Instanciate the RISC-V Emulator
	riscv = core.CreateEmulator(log)

	vga := MakePixelVGA(320, 200)                   // Doom runs at 320x200 natively
	serial := uart.NewUART()                        // UART
	programRom := ram.NewROM("program", 1024*1024)  // Main Program, where the RISCV-DOOM is loaded
	mainRam := ram.NewRAM("main_ram", 16*1024*1024) // Main Ram, on ICE40 RISCV-DOOM it is the PSRAM, here is just a normal ram
	bram := ram.NewRAM("bram", 1024)                // FPGA BRAM, this is actually a program + data ram for the "bootloader"
	dspi := spi.NewDummySPI(log)                    // Dummy SPI since tnt's bootloader expects a SPI Flash, and doom as well. This does nothing but log the SPI configuration

	// Load the bootloader from ice40-playground/riscv_doom to BRAM
	prog, err := ioutil.ReadFile("/media/lucas/ELTNEXT/Works2/ice40-playground/projects/riscv_doom/fw_boot/boot.bin")
	if err != nil {
		panic(err)
	}
	copy(bram.Data, prog)

	// Load RISC-V Doom application
	prog, err = ioutil.ReadFile("/media/lucas/ELTNEXT/Works2/doom_riscv/src/riscv/doom-riscv.bin")
	if err != nil {
		panic(err)
	}
	copy(programRom.Data, prog)

	// Load DOOM WAD
	wad, err := ioutil.ReadFile("/media/ELTN/Games/DOOM/DOOM.WAD")
	if err != nil {
		panic(err)
	}

	// Create a WAD Memory and store the wad. In the ICE 40 this is stored in flash memory
	doomWad := ram.NewROM("doom_wad", len(wad))
	copy(doomWad.Data, wad)

	log.Infof("WAD Size: %d - Memory Size: %d\n", len(wad), len(doomWad.Data))

	// Map the BRAM at 0x00000000
	err = bram.Map(0, riscv.Bus)
	if err != nil {
		panic(err)
	}

	// Map the program at 0x40100000 (in ICE40 the Flash Base is 0x40000000
	err = programRom.Map(0x4010_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}
	// Map the WAD at 0x40200000
	err = doomWad.Map(0x4020_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}
	// Map the RAM at 0x40200000
	err = mainRam.Map(0x4100_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}
	// Map the DummySPI at 0x80000000
	err = dspi.Map(0x8000_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}
	// Map the VGA at 0x8100000
	err = vga.VGA.Map(0x8100_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}
	// Map the UART at 0x8200000
	err = serial.Map(0x8200_0000, riscv.Bus)
	if err != nil {
		panic(err)
	}

	// Fill the memory map text box with the bus mapping
	memoryMapText.Clear()
	_, _ = memoryMapText.WriteString("Bus Map\n")
	_, _ = memoryMapText.WriteString(riscv.Bus.String())

	log.Infof("Bus Map\n%s", riscv.Bus.String())

	// Fill the disassemble box
	RefreshDisasm()

	// Start the RISC-V Emulator gouroutine
	riscv.Start()
	//riscv.AddBreak(0x40118da4)
	//riscv.AddBreak(0x40126750)

	serialData := ""

	for !win.Closed() {
		vga.VGA.VBlank(false)
		win.Clear(colornames.Skyblue)
		vgaScreen := vga.GetPicture()

		pixel.NewSprite(vgaScreen, vgaScreen.Bounds()).
			Draw(win, MoveAndScaleTo(vgaScreen, 32, 32, 1.5))
		memoryMapText.Draw(win, pixel.IM.Moved(pixel.V(w-300, h-32)))
		uartText.Draw(win, pixel.IM.Moved(pixel.V(0, 300)))
		disasmText.Draw(win, pixel.IM.Moved(pixel.V(w-700, h-25)))
		debugText.Draw(win, pixel.IM.Moved(pixel.V(w-380, h-250)))
		stackText.Draw(win, pixel.IM.Moved(pixel.V(w-180, h-250)))

		b := serial.ReadOutputBuffer()
		if len(b) > 0 {
			serialData += string(b)
			lines := strings.Split(serialData, "\n")
			if len(lines) > 15 {
				diff := len(lines) - 15
				lines = lines[diff:]
			}
			uartText.Clear()
			for _, line := range lines {
				uartText.WriteString(line + "\n")
			}
		}
		if win.JustPressed(pixelgl.KeyR) {
			log.Debug("Reset")
			riscv.Reset()
		}

		if win.JustPressed(pixelgl.KeyC) {
			log.Debug("Continue")
			riscv.Continue()
		}

		if win.JustPressed(pixelgl.KeyS) {
			log.Debug("Step")
			riscv.Step()
		}

		if win.JustPressed(pixelgl.KeyP) {
			log.Debug("Pause")
			riscv.Pause()
		}

		if riscv.Paused() {
			RefreshDisasm()
			RefreshStack()
		}
		RefreshDebug()
		win.Update()
		vga.VGA.VBlank(true)
		time.Sleep(time.Second / 60)
	}
	riscv.Stop()
}

func main() {
	pixelgl.Run(run)
}
