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
var prev_row, prev_col int

var source_file string

var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var modified bool
var cursor_changed bool
var is_welcome_page bool

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
	is_welcome_page = false
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

func scrollTextBuffer() {
	if current_row < offset_row {
		offset_row = current_row
	}
	if current_col < offset_col {
		offset_col = current_col
	}
	if current_row >= offset_row+ROWS {
		offset_row = current_row - ROWS + 1
	}
	if current_col >= offset_col+COLS {
		offset_col = current_col - COLS + 1
	}
}

/*
Moves the cursor to the desired direction

direction can be = "up" | "down" | "left" | "right" | "home" | "end" | "pageUp" | "pageDown" | "top" | "bottom"
*/
func moveCursor(direction string) {
	if cursor_changed {
		prev_col = current_col
	}
	switch direction {
	case "up":
		cursor_changed = false
		current_col = prev_col
		if current_row != 0 {
			current_row--
		}
	case "down":
		cursor_changed = false
		current_col = prev_col
		if current_row < len(text_buffer)-1 {
			current_row++
		}
	case "left":
		cursor_changed = true
		if current_col != 0 {
			current_col--
		} else if current_row > 0 {
			current_row--
			current_col = len(text_buffer[current_row])
		}

	case "right":
		cursor_changed = true
		if current_col < len(text_buffer[current_row]) {
			current_col++
		} else if current_row < len(text_buffer)-1 {
			current_row++
			current_col = 0
		}
	case "home":
		cursor_changed = true
		current_col = 0
	case "end":
		cursor_changed = true
		current_col = len(text_buffer[current_row])
	case "pageUp":
		cursor_changed = false
		current_col = prev_col
		if current_row-int(ROWS/2) > 0 {
			current_row -= int(ROWS / 2)
		} else {
			current_row = 0
		}
	case "pageDown":
		cursor_changed = false
		current_col = prev_col
		if current_row+int(ROWS/2) < len(text_buffer)-1 {
			current_row += int(ROWS / 2)
		} else {
			current_row = len(text_buffer) - 1
		}
	case "top":
		cursor_changed = true
		current_col = 0
		current_row = 0
	case "bottom":
		cursor_changed = true
		current_row = len(text_buffer) - 1
		if len(text_buffer[current_row]) > 1 {
			current_col = len(text_buffer[current_row])
		} else {
			current_col = 0
		}
	default:
		termbox.Close()
		valid_directions := "\"up\", \"down\", \"left\", \"right\", \"left\", \"home\", \"end\", \"pageUp\", \"pageDown\", \"top\" or \"bottom\""
		panic("\tThe direction: \"" + direction + "\" is not a defined direction\n\tValid directions: " + valid_directions + "\n\tPlease check your code.")
	}
}

func displayTextBuffer() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + offset_row
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + offset_col
			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen)
				}
			} else if row+offset_row > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('~'), termbox.ColorBlue, termbox.ColorDefault)
			}
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
	printMessage(0, ROWS, termbox.ColorBlack, termbox.ColorBlue, message)
}

func getKey() termbox.Event {
	var key_event termbox.Event
	switch event := termbox.PollEvent(); event.Type {
	case termbox.EventKey:
		key_event = event
	case termbox.EventError:
		panic(event.Err)
	}
	return key_event
}

func processKeyPress() {
	key_event := getKey()
	key := key_event.Key
	ch := key_event.Ch
	if key == termbox.KeyEsc || key == termbox.KeyCtrlQ {
		mode = 0
	} else if ch != 0 {
		if mode == 1 {
			insertCharacter(key_event)
			modified = true
		} else {
			switch ch {
			case 'Q':
				termbox.Close()
				os.Exit(0)
			case 'i':
				if current_col != 0 {
					current_col--
				}
				mode = 1
			case 'I':
				moveCursor("home")
				mode = 1
			case 'a':
				if current_col < len(text_buffer[current_row]) {
					current_col++
				}
				mode = 1
			case 'A':
				moveCursor("end")
				mode = 1
			case 'j':
				moveCursor("down")
			case 'k':
				moveCursor("up")
			case 'l':
				moveCursor("right")
			case 'h':
				moveCursor("left")
			case 'g':
				key_event = getKey()
				if key_event.Ch == 'g' {
					moveCursor("top")
				}
			case 'G':
				moveCursor("bottom")
			}
		}
	} else {
		switch key {
		case termbox.KeyBackspace:
			deleteCharacter()
			modified = true
		case termbox.KeyBackspace2:
			deleteCharacter()
			modified = true
		case termbox.KeyTab:
			if mode == 1 {
				for i := 0; i < 2; i++ {
					insertCharacter(key_event)
				}
				modified = true
			}
		case termbox.KeySpace:
			if mode == 1 {
				insertCharacter(key_event)
				modified = true
			}
		case termbox.KeyHome:
			moveCursor("home")
		case termbox.KeyEnd:
			moveCursor("end")
		case termbox.KeyPgup:
			moveCursor("pageUp")
		case termbox.KeyPgdn:
			moveCursor("pageDown")
		case termbox.KeyArrowDown:
			moveCursor("down")
		case termbox.KeyArrowUp:
			moveCursor("up")
		case termbox.KeyArrowLeft:
			moveCursor("left")
		case termbox.KeyArrowRight:
			moveCursor("right")
		}
		if current_col > len(text_buffer[current_row]) {
			current_col = len(text_buffer[current_row])
		}
	}
}

func insertCharacter(event termbox.Event) {
	insert_rune := make([]rune, len(text_buffer[current_row])+1)
	copy(insert_rune[:current_col], text_buffer[current_row][:current_col])
	if event.Key == termbox.KeySpace {
		insert_rune[current_col] = rune(' ')
	} else if event.Key == termbox.KeyTab {
		// TODO: insert tab or spaces depending on preferences
		insert_rune[current_col] = rune('\t')
	} else {
		insert_rune[current_col] = rune(event.Ch)
	}
	copy(insert_rune[current_col+1:], text_buffer[current_row][current_col:])
	text_buffer[current_row] = insert_rune
	current_col++
}

func deleteCharacter() {
	if current_col > 0 {
		current_col--
		delete_line := make([]rune, len(text_buffer[current_row])-1)
		copy(delete_line[:current_col], text_buffer[current_row][:current_col])
		copy(delete_line[current_col:], text_buffer[current_row][current_col+1:])
		text_buffer[current_row] = delete_line
	} else if current_row > 0 {
		append_line := make([]rune, len(text_buffer[current_row]))
		copy(append_line, text_buffer[current_row][current_col:])
		new_text_buffer := make([][]rune, len(text_buffer)-1)
		copy(new_text_buffer[:current_row], text_buffer[:current_row])
		copy(new_text_buffer[current_row:], text_buffer[current_row+1:])
		text_buffer = new_text_buffer
		current_row--
		current_col = len(text_buffer[current_row])
		insert_line := make([]rune, len(text_buffer[current_row])+len(append_line))
		copy(insert_line[:len(text_buffer[current_row])], text_buffer[current_row])
		copy(insert_line[len(text_buffer[current_row]):], append_line)
		text_buffer[current_row] = insert_line
	}
}

func displayWelcomePage(display bool) {
	if display {
		message := "GoEdit - A minimalistic text editor written in GoLang"
		printMessage(int((COLS-len(message))/2), int(ROWS/2)-1, termbox.ColorBlue, termbox.ColorDefault, message)
		message = "What are you waiting for? GoEdit those files :)"
		printMessage(int((COLS-len(message))/2), int(ROWS/2)+1, termbox.ColorBlue, termbox.ColorDefault, message)
	}
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
		source_file = "welcome.txt"
		text_buffer = append(text_buffer, []rune{})
		is_welcome_page = true
	}
	for {
		COLS, ROWS = termbox.Size()
		ROWS--
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		scrollTextBuffer()
		displayTextBuffer()
		displayWelcomePage(is_welcome_page)
		displayStatusBar()
		termbox.SetCursor(current_col-offset_col, current_row-offset_row)
		termbox.Flush()
		processKeyPress()
	}
}

func main() {
	goEdit()
}
