package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"gotodo/functions"
)

type screen int

const (
	menuScreen screen = iota
	listScreen
	inputScreen
	confirmScreen
)

type menuItem struct {
	label  string
	filter functions.Status // used by list entries; "" means all tasks
	action string           // "list", "add" or "quit"
}

var menuItems = []menuItem{
	{label: "View all tasks", action: "list"},
	{label: "View todo", filter: functions.StatusTodo, action: "list"},
	{label: "View in progress", filter: functions.StatusInProgress, action: "list"},
	{label: "View done", filter: functions.StatusDone, action: "list"},
	{label: "Add a new task", action: "add"},
	{label: "Quit", action: "quit"},
}

// Filters cycled with tab on the task list.
var filters = []functions.Status{"", functions.StatusTodo, functions.StatusInProgress, functions.StatusDone}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#5A56E0")).Padding(0, 2)
	subtitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6A3FF"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	faintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#777777"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	warnStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5A56E0")).Padding(1, 2)

	statusStyles = map[functions.Status]lipgloss.Style{
		functions.StatusTodo:       lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
		functions.StatusInProgress: lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")),
		functions.StatusDone:       lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
	}
	statusIcons = map[functions.Status]string{
		functions.StatusTodo:       "[ ]",
		functions.StatusInProgress: "[~]",
		functions.StatusDone:       "[x]",
	}
)

type model struct {
	screen     screen
	menuCursor int

	allTasks []functions.Task
	tasks    []functions.Task // allTasks narrowed by filter
	filter   functions.Status
	cursor   int
	offset   int

	input     textinput.Model
	editingID string // empty when adding a new task
	returnTo  screen // where esc on the input screen goes back to

	message string
	isErr   bool

	width, height int
}

// Run starts the interactive app and blocks until the user quits.
func Run() error {
	_, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run()
	return err
}

func newModel() model {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.Width = 50
	ti.Prompt = "> "

	m := model{input: ti, width: 80, height: 24}
	m.reload()
	return m
}

func (m *model) reload() {
	all, err := functions.GetTasks()
	if err != nil {
		m.setError(err)
		return
	}
	m.allTasks = all
	m.tasks = nil
	for _, t := range all {
		if m.filter == "" || t.Status == m.filter {
			m.tasks = append(m.tasks, t)
		}
	}
	if m.cursor >= len(m.tasks) {
		m.cursor = len(m.tasks) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) setMessage(format string, args ...any) {
	m.message = fmt.Sprintf(format, args...)
	m.isErr = false
}

func (m *model) setError(err error) {
	m.message = err.Error()
	m.isErr = true
}

func (m *model) openList(filter functions.Status) {
	m.filter = filter
	m.cursor, m.offset = 0, 0
	m.screen = listScreen
	m.reload()
}

func (m *model) openInput(editing *functions.Task) tea.Cmd {
	m.editingID = ""
	m.returnTo = m.screen
	m.input.Reset()
	m.input.Placeholder = "What needs to be done?"
	if editing != nil {
		m.editingID = editing.Id
		m.input.SetValue(editing.Description)
		m.input.CursorEnd()
	}
	m.screen = inputScreen
	return m.input.Focus()
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.scrollToCursor()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		switch m.screen {
		case menuScreen:
			m, cmd = m.updateMenu(msg)
		case listScreen:
			m, cmd = m.updateList(msg)
		case inputScreen:
			m, cmd = m.updateInput(msg)
		case confirmScreen:
			m, cmd = m.updateConfirm(msg)
		}
		m.scrollToCursor()
		return m, cmd
	}
	if m.screen == inputScreen {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.menuCursor = (m.menuCursor - 1 + len(menuItems)) % len(menuItems)
	case "down", "j":
		m.menuCursor = (m.menuCursor + 1) % len(menuItems)
	case "q", "esc":
		return m, tea.Quit
	case "enter", " ":
		m.message = ""
		item := menuItems[m.menuCursor]
		switch item.action {
		case "list":
			m.openList(item.filter)
		case "add":
			m.filter = ""
			return m, m.openInput(nil)
		case "quit":
			return m, tea.Quit
		}
	default:
		// Number keys jump straight to a menu entry.
		if s := msg.String(); len(s) == 1 && s[0] >= '1' && int(s[0]-'0') <= len(menuItems) {
			m.menuCursor = int(s[0] - '1')
			return m.updateMenu(tea.KeyMsg{Type: tea.KeyEnter})
		}
	}
	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (model, tea.Cmd) {
	current := m.currentTask()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = max(len(m.tasks)-1, 0)
	case "tab", "shift+tab":
		step := 1
		if msg.String() == "shift+tab" {
			step = len(filters) - 1
		}
		for i, f := range filters {
			if f == m.filter {
				m.openList(filters[(i+step)%len(filters)])
				break
			}
		}
	case "a":
		return m, m.openInput(nil)
	case "e":
		if current != nil {
			return m, m.openInput(current)
		}
	case "d", "delete":
		if current != nil {
			m.screen = confirmScreen
		}
	case "enter", " ":
		if current != nil {
			m.setStatus(current, nextStatus(current.Status))
		}
	case "1":
		if current != nil {
			m.setStatus(current, functions.StatusTodo)
		}
	case "2":
		if current != nil {
			m.setStatus(current, functions.StatusInProgress)
		}
	case "3":
		if current != nil {
			m.setStatus(current, functions.StatusDone)
		}
	case "esc", "backspace", "m":
		m.screen = menuScreen
		m.message = ""
		m.reload()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) setStatus(t *functions.Task, s functions.Status) {
	if err := functions.UpdateTask(t.Id, "", s); err != nil {
		m.setError(err)
		return
	}
	m.setMessage("Task #%s marked as %s", t.Id, statusLabel(s))
	m.reload()
}

func (m model) updateInput(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.input.Blur()
		m.message = ""
		m.screen = m.returnTo
		m.reload()
		return m, nil
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			m.message, m.isErr = "Description cannot be empty", true
			return m, nil
		}
		if m.editingID == "" {
			if err := functions.AddTask(text); err != nil {
				m.setError(err)
				return m, nil
			}
			m.setMessage("Added: %s", text)
			m.filter = ""
			m.reload()
			m.cursor = max(len(m.tasks)-1, 0)
		} else {
			if err := functions.UpdateTask(m.editingID, text, ""); err != nil {
				m.setError(err)
				return m, nil
			}
			m.setMessage("Updated task #%s", m.editingID)
			m.reload()
		}
		m.input.Blur()
		m.screen = listScreen
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateConfirm(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if t := m.currentTask(); t != nil {
			if err := functions.DeleteTaskById(t.Id); err != nil {
				m.setError(err)
			} else {
				m.setMessage("Deleted task #%s", t.Id)
			}
		}
		m.reload()
		m.screen = listScreen
	case "n", "N", "esc", "q":
		m.message = ""
		m.screen = listScreen
	}
	return m, nil
}

func (m model) currentTask() *functions.Task {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	t := m.tasks[m.cursor]
	return &t
}

func nextStatus(s functions.Status) functions.Status {
	switch s {
	case functions.StatusTodo:
		return functions.StatusInProgress
	case functions.StatusInProgress:
		return functions.StatusDone
	default:
		return functions.StatusTodo
	}
}

func statusLabel(s functions.Status) string {
	switch s {
	case functions.StatusTodo:
		return "todo"
	case functions.StatusInProgress:
		return "in progress"
	case functions.StatusDone:
		return "done"
	}
	return "all"
}

func tabLabel(s functions.Status) string {
	l := statusLabel(s)
	return strings.ToUpper(l[:1]) + l[1:]
}

// ---------- views ----------

func (m model) View() string {
	var body string
	switch m.screen {
	case menuScreen:
		body = m.menuView()
	case listScreen:
		body = m.listView()
	case inputScreen:
		body = m.inputView()
	case confirmScreen:
		body = m.confirmView()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("GoTodo"))
	b.WriteString("  ")
	b.WriteString(m.summary())
	b.WriteString("\n\n")
	b.WriteString(body)
	if m.message != "" {
		b.WriteString("\n\n")
		if m.isErr {
			b.WriteString(errStyle.Render("✗ " + m.message))
		} else {
			b.WriteString(okStyle.Render("✓ " + m.message))
		}
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m model) summary() string {
	counts := map[functions.Status]int{}
	for _, t := range m.allTasks {
		counts[t.Status]++
	}
	return faintStyle.Render(fmt.Sprintf("%d tasks · ", len(m.allTasks))) +
		statusStyles[functions.StatusTodo].Render(fmt.Sprintf("%d todo", counts[functions.StatusTodo])) +
		faintStyle.Render(" · ") +
		statusStyles[functions.StatusInProgress].Render(fmt.Sprintf("%d in progress", counts[functions.StatusInProgress])) +
		faintStyle.Render(" · ") +
		statusStyles[functions.StatusDone].Render(fmt.Sprintf("%d done", counts[functions.StatusDone]))
}

func (m model) menuView() string {
	var b strings.Builder
	b.WriteString(subtitleStyle.Render("What would you like to do?"))
	b.WriteString("\n\n")
	for i, item := range menuItems {
		line := fmt.Sprintf("%d. %s", i+1, item.label)
		if i == m.menuCursor {
			b.WriteString(selectedStyle.Render("› " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ move • enter select • 1-6 jump • q quit"))
	return b.String()
}

func (m model) visibleRows() int {
	// Leave room for header, tabs, help and message lines.
	return max(m.height-14, 3)
}

// scrollToCursor keeps the selected task inside the visible window.
func (m *model) scrollToCursor() {
	rows := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	m.offset = max(min(m.offset, len(m.tasks)-rows), 0)
}

func (m model) listView() string {
	var b strings.Builder

	// Filter tabs
	var tabs []string
	for _, f := range filters {
		label := " " + tabLabel(f) + " "
		if f == m.filter {
			tabs = append(tabs, selectedStyle.Underline(true).Render(label))
		} else {
			tabs = append(tabs, faintStyle.Render(label))
		}
	}
	b.WriteString(strings.Join(tabs, faintStyle.Render("│")))
	b.WriteString("\n\n")

	if len(m.tasks) == 0 {
		if m.filter == "" {
			b.WriteString(faintStyle.Render("  No tasks yet. Press 'a' to add one."))
		} else {
			b.WriteString(faintStyle.Render(fmt.Sprintf("  No %s tasks.", statusLabel(m.filter))))
		}
		b.WriteString("\n")
	} else {
		offset := m.offset
		end := min(offset+m.visibleRows(), len(m.tasks))

		descWidth := max(m.width-40, 15)
		if offset > 0 {
			b.WriteString(faintStyle.Render(fmt.Sprintf("  ↑ %d more", offset)) + "\n")
		}
		for i := offset; i < end; i++ {
			t := m.tasks[i]
			style := statusStyles[t.Status]
			desc := runewidth.FillRight(runewidth.Truncate(t.Description, descWidth, "…"), descWidth)
			line := fmt.Sprintf("%s %-4s %s  %s",
				style.Render(statusIcons[t.Status]),
				"#"+t.Id,
				desc,
				faintStyle.Render(formatTime(t.UpdatedAt)))
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("› ") + line)
			} else {
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
		}
		if end < len(m.tasks) {
			b.WriteString(faintStyle.Render(fmt.Sprintf("  ↓ %d more", len(m.tasks)-end)) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ move • enter cycle status • 1 todo • 2 in progress • 3 done\n" +
		"a add • e edit • d delete • tab switch filter • esc menu • q quit"))
	return b.String()
}

func (m model) inputView() string {
	title := "Add a new task"
	if m.editingID != "" {
		title = "Edit task #" + m.editingID
	}
	content := subtitleStyle.Render(title) + "\n\n" + m.input.View()
	return boxStyle.Render(content) + "\n\n" + helpStyle.Render("enter save • esc cancel")
}

func (m model) confirmView() string {
	t := m.currentTask()
	if t == nil {
		return ""
	}
	content := warnStyle.Render("Delete this task?") + "\n\n" +
		fmt.Sprintf("#%s  %s", t.Id, t.Description)
	return boxStyle.BorderForeground(lipgloss.Color("#FF5555")).Render(content) + "\n\n" +
		helpStyle.Render("y delete • n cancel")
}

func formatTime(s string) string {
	t, err := functions.ParseTime(s)
	if err != nil {
		return s
	}
	return t.Local().Format("Jan 2 15:04")
}
