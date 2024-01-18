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
var command_positions = 1

var source_file string

var error_buffer []string
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

func displayLineNumber() {
	for line := 0; line < len(text_buffer); line++ {
		line_str := strconv.Itoa(line + 1 + offset_row)
		for col, ch := range line_str {
			termbox.SetCell(col, line, ch, termbox.ColorBlack, termbox.ColorDefault)
		}
	}
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
func moveCursor(direction string, positions int) {
	if cursor_changed {
		prev_col = current_col
	}
	switch direction {
	case "up":
		cursor_changed = false
		current_col = prev_col
		if current_row-positions >= 0 {
			current_row -= positions
		} else if current_row-positions < 0 {
			current_row = 0
			setErrorMessage("Careful the top line is not that far :)", 2)
		}
	case "down":
		cursor_changed = false
		current_col = prev_col
		if current_row+positions <= len(text_buffer)-1 {
			current_row += positions
		} else if current_row+positions > len(text_buffer)-1 {
			current_row = len(text_buffer) - 1
			setErrorMessage("Careful the bottom line is not that far :)", 2)
		}
	case "left":
		cursor_changed = true
		if current_col != 0 {
			current_col -= positions
		} else if current_row-positions > 0 {
			current_row--
			current_col = len(text_buffer[current_row])
		}

	case "right":
		cursor_changed = true
		if current_col < len(text_buffer[current_row]) {
			current_col += positions
		} else if current_row+positions < len(text_buffer)-1 {
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
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
				}
			} else if row+offset_row > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('~'), termbox.ColorBlue, termbox.ColorDefault)
			}
		}
		termbox.SetChar(col, row, rune('\n'))
	}
}

type ErrorLevels int

const (
	LOG = iota
	ERROR
	WARN
)

func setErrorMessage(error_message string, error_level ErrorLevels) {
	switch error_level {
	case 0:
		error_message = " [LOG: " + error_message + "] "
	case 1:
		error_message = " [ERROR: " + error_message + "] "
	case 2:
		error_message = " [WARN: " + error_message + "] "
	default:
		break
	}
	error_buffer = append(error_buffer, error_message)
}

func getErrorMessage(index int) string {
	return error_buffer[index]
}

func getErrorColor(error_message string) (termbox.Attribute, termbox.Attribute) {
	var fg termbox.Attribute
	var bg termbox.Attribute
	if error_buffer != nil {
		switch string(error_message[2]) {
		case "L":
			fg = termbox.ColorBlack
			bg = termbox.ColorBlue
		case "E":
			fg = termbox.ColorRed
			bg = termbox.ColorDefault
		case "W":
			fg = termbox.ColorYellow
			bg = termbox.ColorDefault
		}
	} else {
		panic("Error buffer is empty, unable to to display error messages")
	}
	return fg, bg
}

func clearErrorMessage() {
	error_buffer = nil
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
		modified_status = " [+]"
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
	if error_buffer != nil {
		error_status := getErrorMessage(len(error_buffer) - 1)
		fg, bg := getErrorColor(error_status)
		for i, ch := range error_status {
			termbox.SetCell(int((COLS/2)-len(error_status)+used_space)+i, ROWS, ch, fg, bg)
		}
	}
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

func isDigit(ch rune) bool {
	if ch < '0' || ch > '9' {
		return false
	}
	return true
}

func commandMotions(ch rune) {
	switch ch {
	case 'j':
		moveCursor("down", command_positions)
		command_positions = 1
	case 'k':
		moveCursor("up", command_positions)
		command_positions = 1
	case 'l':
		moveCursor("right", command_positions)
		command_positions = 1
	case 'h':
		moveCursor("left", command_positions)
		command_positions = 1
	}
}

func commandMotionsArrows(key termbox.Key) {
	switch key {
	case termbox.KeyArrowDown:
		moveCursor("down", command_positions)
		command_positions = 1
	case termbox.KeyArrowUp:
		moveCursor("up", command_positions)
		command_positions = 1
	case termbox.KeyArrowLeft:
		moveCursor("left", command_positions)
		command_positions = 1
	case termbox.KeyArrowRight:
		moveCursor("right", command_positions)
		command_positions = 1
	}
}

var str_positions strings.Builder

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
			if isDigit(ch) {
				_, err := str_positions.WriteString(string(ch))
				if err == nil {
					positions, err2 := strconv.Atoi(str_positions.String())
					if err2 == nil {
						command_positions = positions
					} else {
						termbox.Close()
						panic(err2)
					}
				} else {
					termbox.Close()
					panic(err)
				}
			} else {
				if ch == 'j' || ch == 'k' || ch == 'l' || ch == 'h' {
					commandMotions(ch)
					str_positions.Reset()
				}
				switch ch {
				case '!':
					key_event = getKey()
					if key_event.Ch == 'Q' {
						error_buffer = nil
						termbox.Close()
						os.Exit(0)
					}
				case 'Q':
					if !modified {
						clearErrorMessage()
						termbox.Close()
						os.Exit(0)
					} else {
						setErrorMessage("Write changes before quitting or force quit <!Q>", 2)
					}
				case 'i':
					if current_col != 0 {
						current_col--
					}
					mode = 1
				case 'I':
					moveCursor("home", 0)
					mode = 1
				case 'a':
					if current_col < len(text_buffer[current_row]) {
						current_col++
					}
					mode = 1
				case 'A':
					moveCursor("end", 0)
					mode = 1
				case 'g':
					key_event = getKey()
					if key_event.Ch == 'g' {
						moveCursor("top", 0)
					}
				case 'G':
					moveCursor("bottom", 0)
				case 'o':
					insertNewLine(false)
					modified = true
					mode = 1
				case 'O':
					insertNewLine(true)
					modified = true
					mode = 1
				case 'w':
					writeFile(source_file)
					if error_buffer != nil {
						error_buffer = nil
					}
				}
			}
		}
	} else {
		if key == termbox.KeyArrowDown || key == termbox.KeyArrowUp || key == termbox.KeyArrowLeft || key == termbox.KeyArrowRight {
			commandMotionsArrows(key)
		}
		switch key {
		case termbox.KeyEnter:
			if mode == 1 {
				insertLine()
				modified = true
			} else {
				moveCursor("down", 1)
			}
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
			moveCursor("home", 0)
		case termbox.KeyEnd:
			moveCursor("end", 0)
		case termbox.KeyPgup:
			moveCursor("pageUp", 0)
		case termbox.KeyPgdn:
			moveCursor("pageDown", 0)
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

func insertLine() {
	right_line := make([]rune, len(text_buffer[current_row][current_col:]))
	copy(right_line, text_buffer[current_row][current_col:])
	left_line := make([]rune, len(text_buffer[current_row][:current_col]))
	copy(left_line, text_buffer[current_row][:current_col])
	text_buffer[current_row] = left_line
	current_row++
	current_col = 0
	new_text_buffer := make([][]rune, len(text_buffer)+1)
	copy(new_text_buffer, text_buffer[:current_row])
	new_text_buffer[current_row] = right_line
	copy(new_text_buffer[current_row+1:], text_buffer[current_row:])
	text_buffer = new_text_buffer
}

func insertNewLine(reverse bool) {
	if !reverse {
		current_row++
	}
	current_col = 0
	new_text_buffer := make([][]rune, len(text_buffer)+1)
	copy(new_text_buffer, text_buffer[:current_row])
	new_text_buffer[current_row] = []rune{' '}
	copy(new_text_buffer[current_row+1:], text_buffer[current_row:])
	text_buffer = new_text_buffer
}

func writeFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for row, line := range text_buffer {
		new_line := "\n"
		if row == len(text_buffer)-1 {
			new_line = ""
		}
		write_line := string(line) + new_line
		_, err = writer.WriteString(write_line)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		writer.Flush()
		modified = false
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
		// displayLineNumber()
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
