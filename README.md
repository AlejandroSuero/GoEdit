# GoEdit - A bare bones text editor, written in Go

For this project I will be using [termbox-go](https://github.com/nsf/termbox-go),
a minimalistic API for text-based user interfaces.

My objective is to create a simple text editor, like [vi](https://en.wikipedia.org/wiki/Vi_(text_editor)),
with the base functionality

Index:

- [Installation](https://github.com/AlejandroSuero/GoEdit#installation)
- [Supported commands](https://github.com/AlejandroSuero/GoEdit#supported-commands)

## Installation

To start using this editor, first thing you need is to install **Go** -> [golang's official site](https://go.dev/).

Once **Go** is installed, go to your terminal clone this repository.

```bash
git clone https://github.com/AlejandroSuero/GoEdit
```

> Note: If you don't have git installed, go to -> [git's official site](https://git-scm.com/) or select at the top where it is a green button that says "Code ▼" and next "Download ZIP".

Next, go into the directory where you have it cloned.

```bash
# This will initialize the project
go init goedit

# This will download the dependencies
go mod tidy
```

Finally, you can build the project and start using it.

```bash
# For UNIX users
go build -o goedit editor.go
./goedit <path-to-the-file-or-leave-it-empty-for-a-welcome-message>

# For Windows users
go build -o goedit.exe editor.go
.\goedit.exe <path-to-the-file-or-leave-it-empty-for-a-welcome-message>
```

For now, check the commands and GoEdit those files 🫡

## Supported commands

These are the commands which are currently supported.

> Note: Commands are case sensitive

### Normal mode commands

| Command | Description |
|---|---|
| `Q` | Exits the editor|
| `!Q` | Forces the exit. (Useful if you don't want to write your changes) |
| `w` | Writes file |
| `i` | Enters insert mode one character before |
| `I` | Enters insert mode at the beginning of the line |
| `a` | Enters insert mode one character after |
| `A` | Enters insert mode at the end of the line |
| `o` | Inserts a new line bellow the cursor and enters insert mode |
| `O` | Inserts a new line on top of the cursor and enters insert mode |
| `j` or `ArrowDown` | Moves cursor down `n` times. Ex: If `j` is pressed it will move once, if `2` and then `j` are pressed it will move it twice. Any number is valid, and `2ArrowDown` as well |
| `k` or `ArrowUp` | Moves cursor up `n` times. Ex: If `k` is pressed it will move once, if `2` and then `k` are pressed it will move twice. Any number is valid, and `2ArrowUp` as well |
| `l` or `ArrowRight` | Moves cursor right `n` times. Ex: If `l` is pressed it will move once, if `2` and then `l` are pressed it will move twice. Any number is valid, and `2ArrowRight` as well |
| `h` or `ArrowLeft` | Moves cursor left `n` times. Ex: If `h` is pressed it will move once, if `2` and then `h` are pressed it will move twice. Any number is valid, and `2ArrowLeft` as well |
| `PageUp` | Moves the cursor half a page up |
| `PageDown` | Moves the cursor half a page down |
| `gg` | Moves the cursor to the beginning |
| `G` | Moves the cursor to then end of the file |

### Insert mode commands

| Command | Description |
|---|---|
| `<Esc>` or `<Ctrl>q` | Exits insert mode |
| Basic interactions | Write any character, insert spaces, tabs, delete characters and add new lines |
