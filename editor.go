package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var mode int
var ROWS, COLS int
var offset_row, offset_col int
var current_row, current_col int

var source_file string

var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var modified bool

func readFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		source_file = filename
		text_buffer = append(text_buffer, []rune{})
		return
	}
	defer file.Close()
	sc := bufio.NewScanner(file)
	line_number := 0

	for sc.Scan() {
		line := sc.Text()
		text_buffer = append(text_buffer, []rune{})
		for i := 0; i < len(line); i++ {
			text_buffer[line_number] = append(text_buffer[line_number], rune(line[i]))
		}
		line_number++
	}
	if line_number == 0 {
		text_buffer = append(text_buffer, []rune{})
	}
}

func printMessage(col int, row int, fg termbox.Attribute, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}

func displayLineNumber(line_number int) {
	termbox.SetCell(0, line_number, rune(strconv.Itoa(line_number + 1)[0]), termbox.ColorBlack, termbox.ColorDefault)
}

func displayTextBuffer() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + offset_col
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + offset_row
			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col+2, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col+2, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen)
				}
			}
		}
		if row+offset_row > len(text_buffer) {
			termbox.SetCell(0, row, rune('~'), termbox.ColorBlue, termbox.ColorDefault)
		}
		termbox.SetChar(col, row, rune('\n'))
	}
}

func displayStatusBar() {
	var mode_status string
	var file_status string
	var copy_status string
	var undo_status string
	var cursor_status string
	var modified_status string
	const MAX_FILE_LENGTH int = 20
	if mode > 0 {
		mode_status = " INSERT: "
	} else {
		mode_status = " NORMAL: "
	}
	filename_length := len(source_file)
	if filename_length > MAX_FILE_LENGTH {
		filename_length = MAX_FILE_LENGTH
	}
	if modified {
		modified_status = " [*]"
	} else {
		modified_status = ""
	}
	file_status = source_file[:filename_length] + modified_status + " - " + strconv.Itoa(len(text_buffer)) + " lines"
	cursor_status = " " + strconv.Itoa(current_row+1) + ":" + strconv.Itoa(current_col+1) + " "
	if len(copy_buffer) > 0 {
		copy_status = "[Copy]"
	}
	if len(undo_buffer) > 0 {
		undo_status = "[Undo]"
	}
	used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len(undo_status)
	spaces := strings.Repeat(" ", COLS-used_space)
	message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status
	printMessage(0, ROWS, termbox.ColorBlack, termbox.ColorCyan, message)
}

func goEdit() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		source_file = os.Args[1]
		// remove autocomplete chars
		if runtime.GOOS == "windows" {
			source_file = strings.Replace(os.Args[1], ".\\", "", 1)
		} else {
			source_file = strings.Replace(os.Args[1], "./", "", 1)
		}
		readFile(source_file)
	} else {
		source_file = "out.txt"
		text_buffer = append(text_buffer, []rune{})
	}

	for {
		COLS, ROWS = termbox.Size()
		ROWS--
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		displayTextBuffer()
		displayStatusBar()
		termbox.Flush()
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey && (event.Key == termbox.KeyEsc || event.Key == termbox.KeyCtrlQ) {
			termbox.Close()
			break
		}
	}
}

func main() {
	goEdit()
}
