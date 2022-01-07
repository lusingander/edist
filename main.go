package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sticky struct {
	path     string
	name     string
	contents []fs.DirEntry
}

func (s *sticky) Title() string {
	return s.name
}

func (s *sticky) Description() string {
	return s.path
}

func (s *sticky) FilterValue() string {
	return s.name
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

func isRTFTextData(f fs.DirEntry) bool {
	return f.Name() == "TXT.rtf"
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
			name:     file.Name(),
			contents: rtfdFiles,
		}
		stickies = append(stickies, sticky)
	}
	return stickies, nil
}

func openEditor(filepath string) error {
	cmd := exec.Command("vi", "./foo")
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

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, nil
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func run(args []string) error {
	if err := checkOS(); err != nil {
		return err
	}
	stickies, err := listStickies()
	if err != nil {
		return err
	}

	items := make([]list.Item, len(stickies))
	for i, s := range stickies {
		items[i] = s
	}
	m := model{list: list.NewModel(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "EDIST"

	p := tea.NewProgram(m)
	p.EnterAltScreen()
	return p.Start()
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
