package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"strings"
	"bufio"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type ConnectionInfo struct {
	IP   string
	Port int
	Conn net.Conn
}

type message struct {
	Text string
}

type model struct {
	connections []ConnectionInfo
	selected    int
	chat        []string
	input       textinput.Model
	chatView    viewport.Model
	sidebar     list.Model
	currentConn *ConnectionInfo
}

var chatStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFA500")). // Orange
	Background(lipgloss.Color("#2f2f2f")). // Funken-Grau
	Padding(1, 1)

var inputStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#39FF14")). // Neon-Grün
	Background(lipgloss.Color("#000000"))  // Schwarz
	Padding(0, 1)

func askTarget() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ziel-IP: ")
	ip, _ := reader.ReadString('\n')
	fmt.Print("Port: ")
	port, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)
	port = strings.TrimSpace(port)
	if ip == "" {
		ip = "127.0.0.1"
	}
	if port == "" {
		port = "5555"
	}
	return ip, port
}

func startP2PConnection(ip string, port string) (*net.Conn, error) {
	address := ip + ":" + port
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func initialModel() model {
	ip, portStr := askTarget()
	port, _ := strconv.Atoi(portStr)

	// Sidebar Liste
	items := []list.Item{}
	sidebar := list.New(items, list.NewDefaultDelegate(), 20, 10)
	sidebar.Title = "Verbindungen"

	// Eingabefeld
	ti := textinput.New()
	ti.Placeholder = "Nachricht eingeben..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Chat-Viewport
	cv := viewport.New(50, 15)
	cv.SetContent("Willkommen zum P2P-Chat!\n")

	conn, err := startP2PConnection(ip, portStr)
	var currentConn *ConnectionInfo
	if err == nil {
		currentConn = &ConnectionInfo{
			IP:   ip,
			Port: port,
			Conn: *conn,
		}
		sidebar.InsertItem(0, list.NewItem(fmt.Sprintf("%s:%s", ip, portStr), "", 0, nil))
	}

	return model{
		connections: []ConnectionInfo{},
		selected:    0,
		chat:        []string{},
		input:       ti,
		chatView:    cv,
		sidebar:     sidebar,
		currentConn: currentConn,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input.Value() != "" && m.currentConn != nil {
				text := m.input.Value()
				m.chat = append(m.chat, "Ich: "+text)
				m.chatView.SetContent(strings.Join(m.chat, "\n"))
				m.input.SetValue("")
				// Nachricht senden
				go func(c net.Conn, t string) {
					c.Write([]byte(t + "\n"))
				}(m.currentConn.Conn, text)
			}
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Eingabefeld updaten
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	sidebarView := m.sidebar.View()
	chatView := chatStyle.Render(m.chatView.View())
	inputView := inputStyle.Render(m.input.View())
	return lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarView,
		lipgloss.JoinVertical(lipgloss.Top, chatView, inputView),
	)
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Fehler:", err)
		os.Exit(1)
	}
}
