package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Struct für gespeicherte Verbindungsinfos
type ConnectionInfo struct {
	Hostname string
	IP       string
	Port     int
	Conn     net.Conn
}

// TCP-Verbindung starten
func startP2PConnection(ip, port string) (*net.Conn, error) {
	address := ip + ":" + port
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	fmt.Println("Verbindung hergestellt zu", address)
	return &conn, nil
}

// IP/Port abfragen
func askTarget() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Bitte Ziel-Computer eingeben (z.B. 192.168.1.10): ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	fmt.Print("Bitte Port eingeben (z.B. 5555): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	if target == "" {
		target = "127.0.0.1"
	}
	if port == "" {
		port = "5555"
	}

	return target, port
}

// Bubble Tea Model
type model struct {
	choices  []string
	cursor   int
	connData ConnectionInfo
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			switch m.choices[m.cursor] {
			case "Verbinden":
				if m.connData.Conn == nil {
					conn, err := startP2PConnection(m.connData.IP, strconv.Itoa(m.connData.Port))
					if err != nil {
						fmt.Println("Verbindung fehlgeschlagen:", err)
					} else {
						m.connData.Conn = *conn
					}
				} else {
					fmt.Println("Bereits verbunden!")
				}
			case "Trennen":
				if m.connData.Conn != nil {
					m.connData.Conn.Close()
					m.connData.Conn = nil
					fmt.Println("Verbindung getrennt")
				} else {
					fmt.Println("Keine Verbindung aktiv")
				}
			case "Beenden":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Wähle eine Option:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	return s + "\nBenutze Pfeiltasten und Enter\n"
}

func main() {
	ip, portStr := askTarget()
	port, _ := strconv.Atoi(portStr)

	m := model{
		choices: []string{"Verbinden", "Trennen", "Beenden"},
		cursor:  0,
		connData: ConnectionInfo{
			IP:   ip,
			Port: port,
		},
	}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Fehler:", err)
		os.Exit(1)
	}

	// Am Ende Verbindungsinfos ausgeben
	if m.connData.Conn != nil {
		fmt.Println("Verbindung aktiv zu:", m.connData.IP, "Port:", m.connData.Port)
		m.connData.Conn.Close()
	}
}
