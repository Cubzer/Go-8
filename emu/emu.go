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
	Display [32][64]byte // 64x32 Display
	keys    [16]bool     // Array of Keys 1-F

	pc          uint16     // Programm Counter
	mem         [4096]byte // Memory
	V           [16]byte   // Registers V0 -VF
	index_r     uint16     // Index Register I
	stack       [16]uint16 // Stack *only used for subroutines
	stack_p     byte       // Stack Pointer
	Flag_draw   bool
	delay_timer byte
	sound_timer byte
}

func Init() system {
	e := system{
		pc: 0x200,
	}
	return e
}

func (e *system) Get_display() [32][64]byte {
	return e.Display
}

func (e *system) Load_buildin_font() {
	for i, v := range fontset {
		e.mem[i] = v
	}
}

func (e *system) Load_rom(path string) {

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

func (e *system) Press_key(key uint8) {
	e.keys[key] = true
}

func (e *system) Release_key(key uint8) {
	e.keys[key] = false
}

func (e *system) Dec_timer() {
	if e.delay_timer > 0 {
		e.delay_timer -= 1
	}
}

func (e *system) Emulate_cycle() {
	// Fetch
	var opcode uint16
	opcode = uint16(e.mem[e.pc])<<8 | uint16(e.mem[e.pc+1])

	fmt.Printf("%x\n", opcode)

	// Decode and Execute
	switch opcode & 0xf000 {
	// 0nnn / Multiple Instructions
	case 0x0000:
		switch opcode & 0x00ff {
		// 00E0 CLR / Clear Display
		case 0xe0:
			for y := 0; y < len(e.Display); y++ {
				for x := 0; x < len(e.Display[y]); x++ {
					e.Display[y][x] = 0x0

				}
			}
		// 00EE RET / Return from Subroutine
		case 0xee:
			e.pc = e.stack[e.stack_p]
			e.stack_p--
			//e.pc -= 2
		}
	// 1nnn JP / Jump to Address nnn
	case 0x1000:
		e.pc = opcode & 0x0fff
		e.pc -= 2
	// 2nnn CALL / Call Subroutine at Address nnn
	case 0x2000:
		e.stack_p++
		e.stack[e.stack_p] = e.pc
		//fmt.Printf("JSR First in Stack now: %X\n", e.pc)
		e.pc = opcode & 0x0fff
		e.pc -= 2
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
			e.V[opcode&0x0f00>>8] = byte(result)
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
			if e.V[opcode&0x0f00>>8]&0x1 == 0x1 {
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
				real_y := (y + uint8(yl)) % 32
				real_x := (x + xl) % 64
				if pixel_byte&(0x80>>xl) != 0 {
					if e.Display[real_y][real_x] == 0x1 {
						e.V[0xf] = 1
					}
					e.Display[real_y][real_x] ^= 0x1
				}
			}
		}
		e.Flag_draw = true
	// E000 / Multiple Instructions
	case 0xe000:
		switch opcode & 0x00ff {
		// Ex9E SKP / Skip next Instruction if key with value Vx is pressed
		case 0x9e:
			if e.keys[e.V[opcode&0x0f00>>8]] == true {
				e.pc += 2
			}
		// ExA1 SKNP / Skip next Instruction if key with value Vx is not pressed
		case 0xa1:
			if e.keys[e.V[opcode&0x0f00>>8]] == false {
				e.pc += 2
			}
		}
	// F000 / Multiple instructions
	case 0xf000:
		switch opcode & 0x00ff {
		// Fx07 LD / Load Value of the Delay Timer into Vx
		case 0x07:
			e.V[opcode&0x0f00>>8] = e.delay_timer
		// Fx0A LD / Wait for a Key press and store the Value of the Key in Vx !!ALL EXECUTIONS STOPPED TILL KEY PRESSED!!
		case 0x0a:
			for i, val := range e.keys {
				if val == true {
					e.V[opcode&0x0f00>>8] = byte(i)
				} else {
					e.pc -= 2
				}
			}
		// Fx15 LD / Load Value of Vx into the Delay Timer
		case 0x15:
			e.delay_timer = e.V[opcode&0x0f00>>8]
		// Fx18 LD / Load Value of Vx into the Sound Timer
		case 0x18:
			e.sound_timer = e.V[opcode&0x0f00>>8]
		// Fx1E ADD / add the Values of I and Vx and Stores the Result in I
		case 0x1e:
			if uint16(e.V[opcode&0x0f00>>8])+e.index_r > 0xfff {
				e.V[0xf] = 1
			} else {
				e.V[0xf] = 0
			}
			e.index_r += uint16(e.V[opcode&0x0f00>>8])
		// Fx29 LD / Set I to the location on a hexadecimal Sprite corresponding to the Value of Vx
		case 0x29:
			e.index_r = uint16(e.V[opcode&0x0f00>>8]) * 5
		// Fx33 LD / Stores the Value of Vx encoded as BCD in the memory location I, I+1, I+3 (255 | i=2 | I+1=5 | I+2=5)
		case 0x33:
			val := e.V[opcode&0x0f00>>8]
			e.mem[e.index_r] = val / 100
			val = val % 100
			e.mem[e.index_r+1] = val / 10
			e.mem[e.index_r+2] = val % 10
		// Fx55 LD / Copies the Values of Registers V0 - Vx into memory locations starting from I
		case 0x55:
			for i := uint16(0); i <= (opcode & 0x0f00 >> 8); i++ {
				e.mem[e.index_r+i] = e.V[i]
			}
		// Fx65 LD / Copies the Values of the memory locations I - I+x into V0 - Vx
		case 0x65:
			for i := uint16(0); i <= (opcode & 0x0f00 >> 8); i++ {
				e.V[i] = e.mem[e.index_r+i]
			}
		}
	}

	e.pc += 2

	//fmt.Printf("%x\n", e.pc)
}

func (e *system) Print_mem() {
	fmt.Println(e.mem)
}
