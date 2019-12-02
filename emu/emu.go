package emu

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

var fontset = [...]byte{
	0xf0, 0x90, 0x90, 0x90, 0xf0, // "0"
	0x20, 0x60, 0x20, 0x20, 0x70, // "1"
	0xf0, 0x10, 0xf0, 0x10, 0x10, // "2"
	0xf0, 0x10, 0xf0, 0x10, 0xf0, // "3"
	0x90, 0x90, 0xf0, 0x10, 0x10, // "4"
	0xf0, 0x80, 0xf0, 0x10, 0xf0, // "5"
	0xf0, 0x80, 0xf0, 0x90, 0xf0, // "6"
	0xf0, 0x10, 0x20, 0x40, 0x40, // "7"
	0xf0, 0x90, 0xf0, 0x90, 0xf0, // "8"
	0xf0, 0x90, 0xf0, 0x10, 0xf0, // "9"
	0xf0, 0x90, 0xf0, 0x90, 0x90, // "A"
	0xe0, 0x90, 0xe0, 0x90, 0xe0, // "B"
	0xf0, 0x80, 0x80, 0x80, 0xf0, // "C"
	0xe0, 0x90, 0x90, 0x90, 0xe0, // "D"
	0xf0, 0x80, 0xf0, 0x80, 0xf0, // "E"
	0xf0, 0x80, 0xf0, 0xe0, 0xe0, // "F"
}

type system struct {
	display [32][64]byte // 64x32 Display

	pc          uint16   // Programm Counter
	mem         []byte   // Memory
	V           []byte   // Registers V0 -VF
	index_r     uint16   // Index Register I
	stack       []uint16 // Stack *only used for subroutines
	stack_p     byte     // Stack Pointer
	flag_draw   bool
	delay_timer byte
	sound_time  byte
}

var e system

func init() {
	e.pc = 0x200
	e.mem = make([]byte, 4096)
	e.V = make([]byte, 16)
	e.index_r = 0x0
	e.stack = make([]uint16, 16)
	e.stack_p = 0x0
}

func load_buildin_font() {
	for i, v := range fontset {
		e.mem[i] = v
	}
}

func load_rom(path string) {

	file, err := os.Open(path)

	if err != nil {
		log.Fatal("Error while opening file", err)
	}

	data, err := ioutil.ReadAll(file)
	for i, v := range data {
		e.mem[0x200+i] = v
	}

	defer file.Close()
}

func emulate_cycle() {
	// Fetch
	fmt.Printf("%x\n", e.pc)
	var opcode uint16
	opcode = uint16(e.mem[e.pc])<<8 | uint16(e.mem[e.pc+1])

	// Decode
	switch opcode & 0xf000 {
	// 1nnn JP / Jump to Address nnn
	case 0x1000:
		e.pc = opcode & 0x0fff
	// 2nnn CALL / Call Subroutine at Address nnn
	case 0x2000:
		e.stack_p++
		e.stack[e.stack_p] = e.pc
		e.pc = opcode & 0x0fff
	// 3xkk SE / Skip next Instruction if Register Vx == kk
	case 0x3000:
		if e.V[(opcode&0x0f00)>>8] == byte(opcode&0x00ff) {
			e.pc += 2
		}
	// 4xkk SNE / Skip next Instruction if Register Vx != kk
	case 0x4000:
		if e.V[(opcode&0x0f00)>>8] != byte(opcode&0x00ff) {
			e.pc += 2
		}
	// 5xy0 SE / Skip next instruction if Register Vx == Vy
	case 0x5000:
		if e.V[(opcode&0x0f00)>>8] == e.V[(opcode&0x00f0)>>4] {
			e.pc += 2
		}
	// 6xkk LD / Load Value kk in Register Vx
	case 0x6000:
		e.V[(opcode&0x0f00)>>8] = byte(opcode & 0x00ff)
	// 7xkk ADD / Adds Value kk to the Value stored in Vx
	case 0x7000:
		e.V[(opcode&0x0f00)>>8] += byte(opcode & 0x00ff)
	// 8000 Multiple Instructions
	case 0x8000:
		switch opcode & 0x000f {
		// 8xy1 OR / bitwise OR on Vx and Vy, Result is stored in Vx
		case 0x1:
			e.V[opcode&0x0f00>>8] |= e.V[opcode&0x00f0>>4]
		// 8xy2 AND / bitwise AND on Vx and Vy, Result is stored in Vx
		case 0x2:
			e.V[opcode&0x0f00>>8] &= e.V[opcode&0x00f0>>4]
		// 8xy3 XOR / bitwise XOR on Vx and Vy, Result is stored in Vx
		case 0x3:
			e.V[opcode&0x0f00>>8] ^= e.V[opcode&0x00f0>>4]
		// 8xy4 ADD / Adds the Value from Vx to the Value of Vy, Result is stored in Vx, Carry to VF
		case 0x4:
			var result uint16 = uint16(e.V[opcode&0x0f00>>8] + e.V[opcode&0x00f0>>4])
			if result > 0xff {
				e.V[0xf] = 0x1
				result = result & 0x00ff
			} else {
				e.V[0xf] = 0x0
			}
			e.V[opcode&0x0f00] = byte(result)
		// 8xy5 SUB / If Value of Vx > Value of Vy then VF = 1 else 0, Result of Vx - Vy is stored in Vx
		case 0x5:
			if e.V[opcode&0x0f00>>8] > e.V[opcode&0x00f0>>4] {
				e.V[0xf] = 0x1
			} else {
				e.V[0xf] = 0x0
			}
			e.V[opcode&0x0f00>>8] -= e.V[opcode&0x00f0>>4]
		// 8xy6 SHR / If LSB of Vx is 1 then VF = 1 else 0, then shift Value of Vx 1bit to the right
		case 0x6:
			if e.V[opcode&0x0f00>>8]&0xfe == 0x1 {
				e.V[0xf] = 0x1
			} else {
				e.V[0xf] = 0x0
			}
			e.V[opcode&0x0f00>>8] = e.V[opcode&0x0f00>>8] >> 1
		// 8xy7 SUBN / If Value of Vy > Value of Vx then VF = 1 else 0, Result of Vy - Vx is stored in Vx
		case 0x7:
			if e.V[opcode&0x00f0>>4] > e.V[opcode&0x0f00>>8] {
				e.V[0xf] = 0x1
			} else {
				e.V[0xf] = 0x0
			}
			e.V[opcode&0x0f00>>8] = e.V[opcode&0x00f0>>4] - e.V[opcode&0x0f00>>8]
		// 8xyE SHL / If MSB of Vx is 1 then VF = 1 else 0, then shift Value of Vx 1bit to the left
		case 0xe:
			if e.V[opcode&0x0f00>>8]&0x7f == 0x1 {
				e.V[0xf] = 0x1
			} else {
				e.V[0xf] = 0x0
			}
			e.V[opcode&0x0f00>>8] = e.V[opcode&0x0f00>>8] << 1
		}
	// 9xy0 SNE / Skip next Instruction if Vx != Vy
	case 0x9000:
		if e.V[opcode&0x0f00>>8] != e.V[opcode&0x00f0>>4] {
			e.pc += 2
		}
	// Annn LD / Load nnn in I
	case 0xa000:
		e.index_r = opcode & 0x0fff
	// Bnnn JP / Jump to nnn + V0
	case 0xb000:
		e.pc = (opcode & 0x0fff) + uint16(e.V[0x0])
	// Cxkk RND / Random Value 0-255 which is ANDed with the Value kk, Result is stored in Vx
	case 0xc000:
		e.V[opcode&0x0f00>>8] = uint8((opcode & 0x00ff)) & uint8(rand.Intn(255))
	// Dxyn DRW / Draw a Sprite of height n Bytes Starting from I at position Vx Vy
	case 0xd000:
		height := opcode & 0x000f
		var x, y uint8 = e.V[opcode&0x0f00>>8], e.V[opcode&0x00f0>>4]

		e.V[0xf] = 0
		for yl := uint16(0); yl < height; yl++ {
			pixel_byte := uint8(e.mem[yl+e.index_r])
			for xl := uint8(0); xl < 8; xl++ {
				if pixel_byte&(0x80>>xl) != 0 {
					if e.display[y+uint8(yl)][x+xl] == 0x1 {
						e.V[0xf] = 1
					}
					e.display[y+uint8(yl)][x+xl] ^= 0x1
				}
			}
		}
		e.flag_draw = true
	// E000 / Multiple Instructions
	case 0xe000:
		switch opcode & 0x00ff {
		// Ex9E SKP / Skip next Instruction if key with value Vx is pressed
		case 0x90:
		}
	}

	fmt.Printf("%x\n", e.pc)
}

func print_mem() {
	fmt.Println(e.mem)
}
