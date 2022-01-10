package edist

import (
	"crypto/md5"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func edit(filepath string) (bool, error) {
	before, err := getMD5(filepath)
	if err != nil {
		return false, err
	}
	err = openEditor(filepath)
	if err != nil {
		return false, err
	}
	after, err := getMD5(filepath)
	if err != nil {
		return false, err
	}
	return before != after, nil
}

func getMD5(filepath string) ([16]byte, error) {
	binary, err := os.ReadFile(filepath)
	if err != nil {
		return [16]byte{}, err
	}
	return md5.Sum(binary), nil
}

func openEditor(filepath string) error {
	cmd := exec.Command("vi", filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restartStickies() error {
	if err := exec.Command("killall", "Stickies").Run(); err != nil {
		return err
	}
	if err := exec.Command("open", "-a", "stickies").Run(); err != nil {
		return err
	}
	return nil
}

func checkOS() error {
	if runtime.GOOS != "darwin" {
		return errors.New("unsupported os")
	}
	return nil
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
