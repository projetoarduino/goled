package goled

import (
	"fmt"

	"github.com/d2r2/go-i2c"
)

const OLED_ADDR = 0x3C

const ssd1306Width = 127
const ssd1306Height = 63

const ssd1306PageSize = 8

const DISPLAY_OFF = 0xAE
const DISPLAY_ON = 0xAF
const SET_DISPLAY_CLOCK_DIV = 0xD5
const SET_MULTIPLEX = 0xA8
const SET_DISPLAY_OFFSET = 0xD3
const SET_START_LINE = 0x00
const CHARGE_PUMP = 0x8D
const EXTERNAL_VCC = false
const MEMORY_MODE = 0x20
const SEG_REMAP = 0xA1 // using 0xA0 will flip screen
const COM_SCAN_DEC = 0xC8
const COM_SCAN_INC = 0xC0
const SET_COM_PINS = 0xDA
const SET_CONTRAST = 0x81
const SET_PRECHARGE = 0xd9
const SET_VCOM_DETECT = 0xDB
const DISPLAY_ALL_ON_RESUME = 0xA4
const NORMAL_DISPLAY = 0xA6
const COLUMN_ADDR = 0x21
const PAGE_ADDR = 0x22
const INVERT_DISPLAY = 0xA7
const ACTIVATE_SCROLL = 0x2F
const DEACTIVATE_SCROLL = 0x2E
const SET_VERTICAL_SCROLL_AREA = 0xA3
const RIGHT_HORIZONTAL_SCROLL = 0x26
const LEFT_HORIZONTAL_SCROLL = 0x27
const VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
const VERTICAL_AND_LEFT_HORIZONTAL_SCROLL = 0x2A

var ssd1306InitSequence []byte = []byte{
	NORMAL_DISPLAY,
	DISPLAY_OFF,
	SET_DISPLAY_CLOCK_DIV, 0x80,
	SET_MULTIPLEX, 0x3F, // set the last value dynamically based on screen size requirement
	SET_DISPLAY_OFFSET, 0x00, // sets offset pro to 0
	SET_START_LINE,
	CHARGE_PUMP, 0x14, // charge pump val
	MEMORY_MODE, 0x00, // 0x0 act like ks0108
	SEG_REMAP,          // screen orientation
	COM_SCAN_DEC,       // screen orientation change to INC to flip
	SET_COM_PINS, 0x12, // com pins val sets dynamically to match each screen size requirement
	SET_CONTRAST, 0x8F, // contrast val
	SET_PRECHARGE, 0xF1, // precharge val
	SET_VCOM_DETECT, 0x40, // vcom detect
	DISPLAY_ALL_ON_RESUME,
	NORMAL_DISPLAY,
	DISPLAY_ON,
}

// DisplayBuffer represents the display buffer intermediate memory
type DisplayBuffer struct {
	Width, Height int32
	buffer        []byte
}

var conn, _ = i2c.NewI2C(OLED_ADDR, 0)

func (s *DisplayBuffer) Init() {
	conn.WriteBytes(ssd1306InitSequence) //Init

	conn.WriteBytes([]byte{COLUMN_ADDR, 0, // Start at 0,
		byte(s.Width) - 1, // End at last column (127?)
	})

	conn.WriteBytes([]byte{PAGE_ADDR, 0, // Start at 0,
		(byte(s.Height) / ssd1306PageSize) - 1, // End at page 7
	})
}

// NewDisplayBuffer creates a new DisplayBuffer
func NewDisplayBuffer(Width, Height int32) *DisplayBuffer {
	s := &DisplayBuffer{
		Width:  Width,
		Height: Height,
	}
	s.buffer = make([]byte, s.Size())
	return s
}

// On turns display on
func On() {
	conn.WriteBytes([]byte{DISPLAY_ON})
}

// Off turns display off
func (s *DisplayBuffer) Off() {
	s.buffer = make([]byte, s.Size())
	conn.WriteBytes(append([]byte{0x40}, s.buffer...)) //Init
}

// Size returns the memory size of the display buffer
func (s *DisplayBuffer) Size() int32 {
	return (s.Width * s.Height) / ssd1306PageSize
}

// Clear the contents of the display buffer
func (s *DisplayBuffer) Clear() {
	s.buffer = make([]byte, s.Size())
}

// Set sets the x, y pixel with c color
func (s *DisplayBuffer) Set(x, y int32) {
	idx := x + (y/ssd1306PageSize)*s.Width
	bit := uint(y) % ssd1306PageSize
	s.buffer[idx] |= (1 << bit)
	//fmt.Println(s.buffer)
}

// Display sends the memory buffer to the display
func (s *DisplayBuffer) Display() {
	// Write the buffer
	conn.WriteBytes(append([]byte{0x40}, s.buffer...)) //Init
}

// draw a filled rectangle on the oled
func (s *DisplayBuffer) fillrect(x, y, w, h int32) {
	// one iteration for each column of the rectangle
	for i := x; i < x+w; i++ {
		// draws a vert line
		s.Draw_line(i, y, i, y+h-1)
	}
}

//Desenha linhas no lcd
func (s *DisplayBuffer) Draw_line(start_x, start_y, end_x, end_y int32) {

	// Bresenham's
	var cx int32 = start_x
	var cy int32 = start_y

	var dx int32 = end_x - cx
	var dy int32 = end_y - cy
	if dx < 0 {
		dx = 0 - dx
	}
	if dy < 0 {
		dy = 0 - dy
	}

	var sx int32
	var sy int32
	if cx < end_x {
		sx = 1
	} else {
		sx = -1
	}
	if cy < end_y {
		sy = 1
	} else {
		sy = -1
	}
	var err int32 = dx - dy

	var n int
	for n = 0; n < 1024; n++ {
		s.Set(cx, cy)
		if (cx == end_x) && (cy == end_y) {
			return
		}
		var e2 int32 = 2 * err
		if e2 > (0 - dy) {
			err = err - dy
			cx = cx + sx
		}
		if e2 < dx {
			err = err + dx
			cy = cy + sy
		}
	}
}

//Gera icones a partir de um array de bytes
func (s *DisplayBuffer) GenIcon(byteArray []byte, x int32, y int32) { //Recebe um array de bytes
	var binary []string
	//Isso aqui é pra voltar o cursor quando mudar de linha, imagine que ele funciona como uma maquina de escrever sempre voltando ao inicio quando mudo de linha
	cursor := x
	count := 1

	for i := 0; i < len(byteArray); i++ { //Transforma um array de bytes em uma representação de string
		bitRepres := fmt.Sprintf("%08b", byteArray[i])
		binary = append(binary, bitRepres)
	}

	for i := 0; i < len(binary); i++ { // percorre o array de strings em busca de 1 ou melhor 49 em ascii
		count++
		for j := 0; j < len(binary[i]); j++ {

			if binary[i][j] == 49 { //se For "1" seta o pixel
				s.Set(x, y)
				x++ //Vai para a proxima posição da linha
			} else {
				x++ //mesmo que for zero segue para a proxima posição do eixo x
			}
		}
		if count > 3 { // a cada 3 array segue para a proxima linha
			count = 1  //Reseta o contador
			x = cursor //Volta o cursor para a posição inicial
			y++        //Proxima linha significa incrementar o eixo y
		}
	}
}

func (s *DisplayBuffer) WriteString(text string, x int32, y int32, size int32) {
	cursor := x

	for i := 0; i < len(text); i++ {
		//fmt.Println(string(text[i]))
		t := findCharBuf(string(text[i]))
		letter := readCharBytes(t)
		s.drawChar(letter, cursor, y, size)
		cursor = cursor + int32(Font6x8.Width)*size //Espaçamento entre as letras
	}
}

// draw an individual character to the screen
func (s *DisplayBuffer) drawChar(byteArray [][]byte, xpos int32, ypos int32, size int32) {

	for i := 0; i < len(byteArray); i++ {
		for j := 0; j < len(byteArray[i]); j++ {
			if byteArray[i][j] == 1 {

				if size == 1 {
					i2 := int32(i)
					j2 := int32(j)
					s.Set(i2+xpos, j2+ypos)
				} else {
					// MATH! Calculating pixel size multiplier to primitively scale the font
					x := xpos + (int32(i) * size)
					y := ypos + (int32(j) * size)
					s.fillrect(x, y, size, size)
				}
				//fmt.Println(i2+xpos, j2+ypos, i, j, byteArray[i])
			}
		}
	}
}

// find where the character exists within the font object
func findCharBuf(c string) []byte {
	pos := indexOf(c, Font6x8.Lookup)
	// use the lookup array as a ref to find where the current char bytes start
	cBufPos := pos * Font6x8.Width
	// slice just the current char's bytes out of the fontData array and return
	cBuf := Font6x8.Font[cBufPos : cBufPos+Font6x8.Width]
	//fmt.Println(cBuf)
	return cBuf
}

// get character bytes from the supplied font object in order to send to framebuffer [][]byte
func readCharBytes(byteArray []byte) [][]byte {
	binaryLetter := [][]byte{}
	// loop through each byte supplied for a char
	for i := 0; i < len(byteArray); i++ {
		// set current byte
		bitArr := buildArr(byteArray[i])
		// push to array containing flattened bit sequence
		binaryLetter = append(binaryLetter, bitArr)
		//fmt.Println(i)
	}
	return binaryLetter
}

//Ajuda o readCharBytes por algum motivo não consegui um bom resultado com for dentro de for
func buildArr(cbyte byte) []byte {
	bitArr := []byte{}
	// read each byte
	for j := 0; j <= Font6x8.Height; j++ {
		// shift bits right until all are read
		bit := cbyte >> byte(j) & 1
		bitArr = append(bitArr, bit)
		//fmt.Println(cbyte >> byte(j) & 1)

	}
	return bitArr
}

//Famoso indexof que não existe no go
func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
