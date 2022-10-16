package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss" // package for the future, for style changes.
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var db = dbStartUp()
var TODAY string = time.Now().Format("02-01-2006")

type model struct {
	choices   []string         // items on the to-do list
	cursor    int              // which list item our cursor is pointing at
	selected  map[int]struct{} // which items are selected
	addingNew textinput.Model  // used to add new values
}

func initialModel() model {
	items := model{selected: make(map[int]struct{})}
	values := getValues()
	insertModel := textinput.New()
	insertModel.Placeholder = "Add New Item"
	insertModel.Prompt = ""    // we do not want an additional prompt for the "add new item"
	insertModel.CharLimit = 20 // no real reason for this limitation, can experiment
	insertModel.Width = 10     // no real reason for this limitation, can experiment

	if len(values) == 0 {
		items.choices = []string{"Protein Shake", "Creatine", "Vitamin D", "Cheese", "Pizza", "Chocolate", "Bananas"}
	} else {
		items.choices = values
	}
	items.addingNew = insertModel

	for i := 0; i < len(items.choices); i++ {
		items.addItems(items.choices[i], false)
	}

	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" key moves the cursor up
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" key moves the cursor down
		case "down":
			if m.cursor < len(m.choices) { //-1 {
				m.cursor++
			}

		// The "enter" key toggles the selected state for the item that the cursor is pointing at.
		// if pointing at addNew, this then allows a new value to be typed in.
		// de-focus to stop input and update DB selection boolean.
		case "enter":
			if m.cursor < len(m.choices) {
				m.addingNew.Blur()
				m.checkItem(m.cursor)
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
				// if out of range, then we are on the "addNew selection."
			} else {
				_, ok := m.selected[m.cursor]
				// logic to toggle the selection and also "focus" when selected to allow for input
				if ok {
					m.addNewItem()
				} else {
					m.selected[m.cursor] = struct{}{}
					m.addingNew.Focus()
				}
			}
		case "delete":
			m.deleteItems(m.choices[m.cursor])
			m.choices = deleteChoice(m.choices, m.cursor)
		}
	}
	cmd = m.updateInputs(msg) // used to drive the addNewItem functionality.
	return m, cmd
}

// remove item from array by appending the other elements of the array.
func deleteChoice(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// insert new item to choices, from the addingNew model
// unfocus to stop further input
// reset addingNew back to default state
// delete selection of addNew, reselect newly inserted item
// add new item to DB
func (m *model) addNewItem() {
	newItem := m.addingNew.Value()
	m.choices = append(m.choices, newItem)
	m.addingNew.Blur()
	m.addingNew.Reset()
	delete(m.selected, m.cursor)
	m.selected[len(m.choices)-1] = struct{}{}
	m.addItems(newItem, true)
}

// handles typing once focused.
func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.addingNew, cmd = m.addingNew.Update(msg)
	return cmd
}

func (m model) View() string {
	// The header
	s := "What do you need?\n\n"

	// Iterate over our choices
	for i := 0; i <= len(m.choices); i++ {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		if i < len(m.choices) {
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, m.choices[i])
		} else {
			// Render the addingNew row, both in "adding" state and in default state.
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, m.addingNew.View())
		}
	}
	// The footer
	s += "\nPress Delete to delete an item. \nPress q to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// create a basic database table if it doesn't exist.
func dbStartUp() *sql.DB {
	database, _ :=
		sql.Open("sqlite3", "./list.db")
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS list (id INTEGER PRIMARY KEY, item VARCHAR NOT NULL UNIQUE, checked BOOL NOT NULL, date TEXT NOT NULL)")
	statement.Exec()

	return database
}

func (m model) addItems(item string, checked bool) {
	statement, _ :=
		db.Prepare("INSERT INTO list (item, date, checked) VALUES (?, ?, ?)")
	statement.Exec(item, TODAY, checked)
}

func (m model) deleteItems(item string) {
	statement, _ :=
		db.Prepare("DELETE FROM list WHERE item = ?")
	statement.Exec(item)
}

func (m model) checkItem(i int) {
	statement, _ :=
		db.Prepare("UPDATE list SET checked = NOT checked WHERE item = ?")
	statement.Exec(m.choices[i])
}

// prepopulate the choices if there is an existing database, rather than using default values
// i.e. memory between sessions.
func getValues() []string {
	var item string
	itemArray := []string{}

	rows, _ :=
		db.Query("SELECT item FROM list")
	for rows.Next() {
		rows.Scan(&item)
		itemArray = append(itemArray, item)
	}
	return itemArray
}
