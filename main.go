package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	appStyle  = lipgloss.NewStyle().Margin(1, 2)
	keyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	quitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

var errCancelled = errors.New("cancelled")

// footerHeight is the number of rows the hotkey footer adds under the list.
const footerHeight = 1

// hotkey renders one "key description" pair for the footer.
func hotkey(k, desc string) string {
	return keyStyle.Render(k) + " " + descStyle.Render(desc)
}

var footer = strings.Join([]string{
	hotkey("↑/↓", "navigate"),
	hotkey("enter", "select"),
	hotkey("/", "filter"),
	hotkey("esc", "close"),
}, "   ")

var quitPrompt = quitStyle.Render("Close without selecting a store?") + "  " +
	hotkey("y", "yes") + "  " + hotkey("n", "no") + "  " + hotkey("esc", "cancel")

type picker struct {
	list       list.Model
	selected   list.Item
	confirming bool
}

type orgItem struct {
	org Org
}

func (i orgItem) Title() string       { return i.org.Name }
func (i orgItem) Description() string { return "Organization " + i.org.ID }
func (i orgItem) FilterValue() string { return i.org.Name }

type storeItem struct {
	store Store
}

func (i storeItem) Title() string       { return i.store.Name }
func (i storeItem) Description() string { return badge(i.store.Type) + " " + i.store.Domain }
func (i storeItem) FilterValue() string { return i.store.Name + " " + i.store.Domain }

func badge(t string) string {
	colors := map[string]string{
		"dev":             "42",
		"production":      "33",
		"client-transfer": "220",
		"collaborator":    "177",
	}
	if c, ok := colors[t]; ok {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(t)
	}
	return t
}

func newPicker(items []list.Item, title string) *picker {
	l := list.New(items, list.NewDefaultDelegate(), 40, 30)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)       // replaced by our own footer
	l.DisableQuitKeybindings() // "v" used to quit the app by accident
	return &picker{list: l}
}

// Init returns startup commands. We need none.
func (m picker) Init() tea.Cmd { return nil }

// Update is the whole app: every event is a tea.Msg, and we return the new
// model plus any command. This is your BOSS-2 state machine.
func (m picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Terminal resized (fires once at startup too). H/V = padding from
		// appStyle, +1 row for the footer.
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-footerHeight)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg) // everything else goes to the list
	return m, cmd
}

// handleKey routes key presses. ctrl+c is checked first so it stays an
// unconditional escape hatch even while the filter box has focus. While the
// quit prompt is up it swallows everything except the answers, so the list
// can't react behind it.
func (m picker) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.confirming {
		switch key {
		case "y", "Y":
			return m, tea.Quit
		case "n", "N", "esc", "q":
			m.confirming = false
		}
		return m, nil
	}

	// Never steal keys while the user is typing in the filter box.
	if m.list.FilterState() == list.Filtering {
		return m.toList(msg)
	}

	switch key {
	case "esc":
		// esc peels one layer at a time: cancel filter, clear filter, close.
		if m.list.IsFiltered() {
			return m.toList(msg)
		}
		m.confirming = true
		return m, nil

	case "enter":
		m.selected = m.list.SelectedItem()
		return m, tea.Quit // leave the alternate screen

	case "q":
		m.confirming = true
		return m, nil
	}

	return m.toList(msg)
}

// toList forwards a message to the list and keeps its command.
func (m picker) toList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current frame as a tea.View.
func (m picker) View() tea.View {
	v := tea.NewView(appStyle.Render(m.list.View() + "\n" + m.statusLine()))
	v.AltScreen = true
	return v
}

// statusLine is the hotkey footer, swapped for the quit prompt while asking.
func (m picker) statusLine() string {
	if m.confirming {
		return quitPrompt
	}
	return footer
}

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

// wantsVersion reports whether argv asks for the version.
func wantsVersion(args []string) bool {
	for _, a := range args {
		if a == "--version" || a == "-version" {
			return true
		}
	}
	return false
}

func main() {
	if wantsVersion(os.Args[1:]) {
		fmt.Println("hopify " + version)
		return
	}

	for {
		s, err := pickStore()
		if errors.Is(err, errCancelled) {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		runThemeDev(s)
	}
}

// runPicker shows a tea list picker; returns the chosen item, or errCancelled.
func runPicker(items []list.Item, title string) (list.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("nothing to pick")
	}
	if len(items) == 1 {
		return items[0], nil // skip the menu entirely
	}
	p := tea.NewProgram(newPicker(items, title))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}
	selected := model.(picker).selected
	if selected == nil {
		return nil, errCancelled
	}
	return selected, nil
}

// pickStore fetches orgs + stores, shows the pickers, returns the choice.
func pickStore() (Store, error) {
	orgs, err := runLoading[[]Org]("Fetching organizations…", fetchOrgs)
	if err != nil {
		return Store{}, err
	}
	if len(orgs) == 0 {
		return Store{}, errors.New("no organizations found — run 'shopify login' first")
	}

	orgItems := make([]list.Item, len(orgs))
	for i, o := range orgs {
		orgItems[i] = orgItem{o}
	}
	chosenOrg, err := runPicker(orgItems, "Select an organization")
	if err != nil {
		return Store{}, err
	}

	stores, err := runLoading[[]Store]("Fetching stores…", func() ([]Store, error) {
		return fetchStores(chosenOrg.(orgItem).org.ID)
	})
	if err != nil {
		return Store{}, err
	}

	storeItems := make([]list.Item, len(stores))
	for i, s := range stores {
		storeItems[i] = storeItem{s}
	}
	chosenStore, err := runPicker(storeItems, "Select a store")
	if err != nil {
		return Store{}, err
	}
	return chosenStore.(storeItem).store, nil
}

func runThemeDev(s Store) {
	lipgloss.Println(appStyle.Render("▶ theme dev on " + s.Domain + " (" + s.Type + ")"))
	cmd := exec.Command("shopify", "theme", "dev", "--store="+s.Domain)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("theme dev exited: %v", err)
	}
}
