package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var db = dbStartUp()
var TODAY string = time.Now().Format("02-01-2006")

type model struct {
	choices   []string            // items on the to-do list
	cursor    int              // which to-do list item our cursor is pointing at
	selected  map[int]struct{} // which to-do items are selected
	addingNew textinput.Model  //testing
}

func initialModel() model {
	items := model{selected: make(map[int]struct{})}
	values := getValues()
	new := textinput.New()
	new.Prompt = ""
	new.Placeholder = "Add New Item"
	//new.Focus() // does this lock it?
	new.CharLimit = 20 // was 156
	new.Width = 10     // was 20

	if len(values) == 0 {
		items.choices = []string{"Protein Shake", "Creatine", "Vitamin D", "Cuddles", "Pizza", "Chocolate", "Bananas"}
	} else {
		items.choices = values
	}
	// A map which indicates which choices are selected. We're using
	// the  map like a mathematical set. The keys refer to the indexes
	// of the `choices` slice, above.
	items.addingNew = new

	for i := 0; i < len(items.choices); i++ {
		items.addItems(items.choices[i], false)
	}

	return items
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return textinput.Blink // nil
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

		// The "up" and "k" keys move the cursor up
		case "up": //, "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down": //, "j":
			if m.cursor < len(m.choices) { //-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter": //, " ":
			comparison := textinput.New()
			comparison.Prompt = ""
			comparison.Placeholder = "Please type value here"
			// comparison.Focus()

			if m.cursor < len(m.choices) {
				m.addingNew.Blur()
				m.checkItem(m.cursor)
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			} else { //if m.choices[m.cursor-1] != comparison.View() {
				_, ok := m.selected[m.cursor]
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
	cmd = m.updateInputs(msg)
	// if m.cursor != len(m.choices) && m.addingNew.Value() != "Add New Item" {
	// 	m.choices = append(m.choices, m.addingNew.Value())
	// 	m.addingNew.Reset()
	// }

	return m, cmd
}

func deleteChoice(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func (m *model) addNewItem() {
	newItem := m.addingNew.Value()
	m.choices = append(m.choices, newItem)
	m.addingNew.Blur()
	m.addingNew.Reset()
	delete(m.selected, m.cursor)
	m.selected[len(m.choices)-1] = struct{}{}
	m.addItems(newItem, true)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.addingNew, cmd = m.addingNew.Update(msg)
	return cmd
}

func (m model) View() string {
	// The header
	s := "What do you have to have today?\n\n"

	// Iterate over our choices
	for i := 0; i <= len(m.choices); i++ { //, choice := range m.choices {

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
			//var typeInterface interface{} = m.choices
			//if _, ok := m.choices[i].(string); ok {
				s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, m.choices[i])
			//} else {
			//	s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, m.choices[i]) //m.choices[i].Value())
			//}
		} else {
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, m.addingNew.View())
		}
	}

	// } else {
	// 	s += "oops"
	// }
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
	if m.choices[i] == "Add New Item" {
		m.choices[i] = "lol"
		fmt.Scanln(&m.choices[i])
		// m.choices = append(m.choices, "Add New Item") // corrupts rest of it
	}

	statement, _ :=
		db.Prepare("UPDATE list SET checked = NOT checked WHERE item = ?")
	statement.Exec(m.choices[i])
}

func getValues() []string { //map[string]bool {
	// itemMap := make(map[string]bool)
	var item string
	// var status bool
	itemArray := []string{}

	rows, _ :=
		db.Query("SELECT item FROM list") //, checked FROM list")
	for rows.Next() {
		rows.Scan(&item) //, &status)
		//itemMap[item] = status
		itemArray = append(itemArray, item)
	}
	return itemArray
}
