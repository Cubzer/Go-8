package main

import (
	"fmt"
	"time"

	"github.com/Cubzer/Go-8/emu"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	w_height = 320
	w_width  = 640
)

func do_every(d time.Duration, f func()) {
	for x := range time.Tick(d) {
		f()
		fmt.Printf("%t\n", x)
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w_width, w_height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	//rect := sdl.Rect{0, 0, 200, 200}

	cpu := emu.Init()
	cpu.Load_buildin_font()
	cpu.Load_rom("TETRIS.ch8")
	cpu.Print_mem()

	var display_arr [32][64]sdl.Rect // Array of Rect`s which represent the 32x64 Pixel

	// Fills the Display_arr with 10x10 rects as Pixels
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			display_arr[y][x] = sdl.Rect{int32(x * 10), int32(y * 10), 10, 10}
		}
	}

	running := true

	go do_every(2*time.Millisecond, cpu.Emulate_cycle)
	go do_every(17*time.Millisecond, cpu.Dec_timer)

	for running {
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			switch ev_t := ev.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				if ev_t.Type == sdl.KEYDOWN {
					switch ev_t.Keysym.Sym {
					case sdl.K_KP_1, sdl.K_1:
						cpu.Press_key(0x1)
					case sdl.K_KP_2, sdl.K_2:
						cpu.Press_key(0x2)
					case sdl.K_KP_3, sdl.K_3:
						cpu.Press_key(0x3)
					case sdl.K_KP_4, sdl.K_4:
						cpu.Press_key(0x4)
					case sdl.K_KP_5, sdl.K_5:
						cpu.Press_key(0x5)
					case sdl.K_KP_6, sdl.K_6:
						cpu.Press_key(0x6)
					case sdl.K_KP_7, sdl.K_7:
						cpu.Press_key(0x7)
					case sdl.K_KP_8, sdl.K_8:
						cpu.Press_key(0x8)
					case sdl.K_KP_9, sdl.K_9:
						cpu.Press_key(0x9)
					case sdl.K_a:
						cpu.Press_key(0xa)
					case sdl.K_b:
						cpu.Press_key(0xb)
					case sdl.K_c:
						cpu.Press_key(0xc)
					case sdl.K_d:
						cpu.Press_key(0xd)
					case sdl.K_e:
						cpu.Press_key(0xe)
					case sdl.K_SPACE:
						cpu.Emulate_cycle()
					}
				}
				if ev_t.Type == sdl.KEYUP {
					switch ev_t.Keysym.Sym {
					case sdl.K_KP_1, sdl.K_1:
						cpu.Release_key(0x1)
					case sdl.K_KP_2, sdl.K_2:
						cpu.Release_key(0x2)
					case sdl.K_KP_3, sdl.K_3:
						cpu.Release_key(0x3)
					case sdl.K_KP_4, sdl.K_4:
						cpu.Release_key(0x4)
					case sdl.K_KP_5, sdl.K_5:
						cpu.Release_key(0x5)
					case sdl.K_KP_6, sdl.K_6:
						cpu.Release_key(0x6)
					case sdl.K_KP_7, sdl.K_7:
						cpu.Release_key(0x7)
					case sdl.K_KP_8, sdl.K_8:
						cpu.Release_key(0x8)
					case sdl.K_KP_9, sdl.K_9:
						cpu.Release_key(0x9)
					case sdl.K_a:
						cpu.Release_key(0xa)
					case sdl.K_b:
						cpu.Release_key(0xb)
					case sdl.K_c:
						cpu.Release_key(0xc)
					case sdl.K_d:
						cpu.Release_key(0xd)
					case sdl.K_e:
						cpu.Release_key(0xe)
					}
				}
			}
		}

		if cpu.Flag_draw {
			cpu.Flag_draw = false
			for x := 0; x < 32; x++ {
				for y := 0; y < 64; y++ {
					if cpu.Display[x][y] != 0x0 {
						surface.FillRect(&display_arr[x][y], 0xFFFF)
					} else {
						surface.FillRect(&display_arr[x][y], 0x0)
					}
				}
			}
		}

		window.UpdateSurface()

		sdl.Delay(1000 / 60)

	}
}
