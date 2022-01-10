package edist

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type sticky struct {
	path     string
	rtfd     fs.DirEntry
	contents []fs.DirEntry
	rtf      string
}

func (s *sticky) Title() string {
	return s.rtfd.Name()
}

func (s *sticky) Description() string {
	info, err := s.rtfd.Info()
	if err != nil {
		return "***"
	}
	return info.ModTime().Format("2006/01/02 03:04:56")
}

func (s *sticky) FilterValue() string {
	return s.rtfd.Name()
}

func stickiesDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	ret := filepath.Join(home, "Library/Containers/com.apple.Stickies/Data/Library/Stickies")
	return ret, nil
}

func isRTFD(f fs.DirEntry) bool {
	return f.IsDir() && filepath.Ext(f.Name()) == ".rtfd"
}

func listStickies() ([]*sticky, error) {
	stickies := make([]*sticky, 0)
	baseDir, err := stickiesDataDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !isRTFD(file) {
			continue
		}
		rtfdDir := filepath.Join(baseDir, file.Name())
		rtfdFiles, err := os.ReadDir(rtfdDir)
		if err != nil {
			return nil, err
		}
		sticky := &sticky{
			path:     rtfdDir,
			rtfd:     file,
			contents: rtfdFiles,
			rtf:      filepath.Join(rtfdDir, "TXT.rtf"),
		}
		stickies = append(stickies, sticky)
	}
	return stickies, nil
}

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "201", Dark: "201"}).
				Render
)

type errorMsg struct {
	e error
}

func errorCmd(e error) tea.Cmd {
	return func() tea.Msg { return errorMsg{e} }
}

type redrawMsg struct{}

func redrawCmd() tea.Msg {
	return redrawMsg{}
}

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		item, ok := m.SelectedItem().(*sticky)
		if !ok {
			return nil
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.edit):
				editted, err := edit(item.rtf)

				// Hack: re-enable the altscreen by printing directly to
				// stdout.
				//
				// Vim also runs in the altscreen and exits the altsceen on
				// exit. As a result, we need to manually jump back in after
				// Vim closes.
				termenv.AltScreen()

				if err != nil {
					return errorCmd(err)
				}
				if editted {
					if err := restartStickies(); err != nil {
						return errorCmd(err)
					}
				}
				return tea.Batch(redrawCmd, tea.HideCursor)
			}
		}
		return nil
	}

	help := []key.Binding{keys.edit}
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	edit key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		edit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
	}
}

type model struct {
	list list.Model

	currentSize tea.WindowSizeMsg
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.currentSize = msg
	case redrawMsg:
		cmd := func() tea.Msg { return m.currentSize }
		return m, cmd
	case errorMsg:
		errorMessage := statusMessageStyle(msg.e.Error())
		cmd := m.list.NewStatusMessage(errorMessage)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func initModel(stickies []*sticky) model {
	items := make([]list.Item, len(stickies))
	for i, s := range stickies {
		items[i] = s
	}
	delegateKeys := newDelegateKeyMap()
	delegate := newItemDelegate(delegateKeys)
	m := model{list: list.NewModel(items, delegate, 0, 0)}
	m.list.Title = "EDIST"
	return m
}

func Start() error {
	if err := checkOS(); err != nil {
		return err
	}
	stickies, err := listStickies()
	if err != nil {
		return err
	}
	m := initModel(stickies)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Start()
}
