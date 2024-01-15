package main

import "os"
import "fmt"

import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

func print_message(col int, row int, fg termbox.Attribute, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	message := "GoEdit - A bare bones text editor, written in Go"
	for {
		col, row := termbox.Size()
		print_message(int((col-len(message))/2), int(row/2), termbox.ColorDefault, termbox.ColorDefault, message)
		termbox.Flush()
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}

func main() {
	run_editor()
}
