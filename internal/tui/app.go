package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/nexlycode/nexly/internal/config"
	"github.com/nexlycode/nexly/internal/handlers"
	"github.com/nexlycode/nexly/internal/providers"
	"github.com/nexlycode/nexly/internal/utils"
)

var (
	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	
	userBubbleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)
	
	assistantBubbleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("63")).
				Padding(0, 1)
)

type model struct {
	messages     []Message
	input        string
	provider     string
	model        string
	streaming    bool
	spinner      bool
	spinnerFrame int
	spinnerMutex sync.Mutex
	width        int
	height       int
	commandView  bool
	commands     []Command
	selectedCmd  int
	commandInput string
	errMsg       string
}

type Message struct {
	Role    string
	Content string
}

type Command struct {
	Name        string
	Description string
	Action      func(*model) (tea.Model, tea.Cmd)
}

func Run(cfg config.Config) {
	initialModel := model{
		provider:    cfg.Provider,
		model:      cfg.Model,
		messages:   []Message{},
		commands:   getCommands(),
		commandView: false,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getCommands() []Command {
	return []Command{
		{Name: "/provider", Description: "Switch AI provider", Action: switchProviderCmd},
		{Name: "/model", Description: "Switch AI model", Action: switchModelCmd},
		{Name: "/clear", Description: "Clear chat history", Action: clearChatCmd},
		{Name: "/help", Description: "Show help", Action: helpCmd},
		{Name: "/config", Description: "Configure API keys", Action: configCmd},
		{Name: "/exit", Description: "Exit Nexly", Action: exitCmd},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.HideCursor,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.commandView {
			return m.updateCommandPalette(msg)
		}

		if msg.String() == "ctrl+p" {
			m.commandView = true
			m.commandInput = ""
			m.selectedCmd = 0
			return m, nil
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "enter" && !m.streaming {
			if m.input == "" {
				return m, nil
			}
			if strings.HasPrefix(m.input, "/") {
				newModel, newCmd := m.handleCommand(m.input)
				return newModel, newCmd
			}
			return m.sendMessage()
		}

		if msg.String() == "ctrl+u" {
			m.input = ""
			return m, nil
		}

		if msg.String() == "backspace" && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
			return m, nil
		}

		if msg.Runes != nil {
			m.input += string(msg.Runes)
			return m, nil
		}

		return m, nil

	case spinnerTick:
		m.spinnerMutex.Lock()
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		m.spinnerMutex.Unlock()
		return m, nil

	case streamingComplete:
		m.streaming = false
		m.spinner = false
		return m, nil

	case streamingError:
		m.streaming = false
		m.spinner = false
		m.errMsg = msg.err.Error()
		return m, nil
	}

	return m, nil
}

func (m *model) handleCommand(input string) (tea.Model, tea.Cmd) {
	for _, cmd := range m.commands {
		if input == cmd.Name {
			return cmd.Action(m)
		}
	}

	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: fmt.Sprintf("Unknown command: %s", input),
	})
	m.input = ""
	return m, nil
}

func (m *model) sendMessage() (tea.Model, tea.Cmd) {
	userInput := m.input
	m.messages = append(m.messages, Message{
		Role:    "user",
		Content: userInput,
	})
	m.input = ""
	m.streaming = true
	m.spinner = true

	return m, tea.Batch(
		tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
			return spinnerTick{}
		}),
		func() tea.Msg {
			return m.streamResponse(userInput)
		},
	)
}

func (m *model) streamResponse(userInput string) tea.Msg {
	ctx := context.Background()
	
	projectContext := handlers.GetProjectContext()
	
	apiKey := config.GetAPIKey(m.provider)
	if apiKey == "" {
		return streamingError{fmt.Errorf("API key not set for provider: %s", m.provider)}
	}
	
	provider := providers.NewSimpleProvider(m.provider, apiKey, m.model)
	
	systemPrompt := `You are Nexly, a helpful AI coding assistant. You can read, write, and edit files. 
When asked to edit files, provide the complete updated file content. 
Be concise and helpful. Always provide code in markdown code blocks.`
	
	fullPrompt := fmt.Sprintf("Project context:\n%s\n\nUser: %s", projectContext, userInput)
	
	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fullPrompt},
	}

	var response strings.Builder
	mu := sync.Mutex{}

	err := provider.SendMessage(ctx, messages, func(content string) {
		mu.Lock()
		response.WriteString(content)
		mu.Unlock()
	})

	if err != nil {
		return streamingError{err}
	}

	result := response.String()
	
	config.AddMessage("user", userInput)
	config.AddMessage("assistant", result)

	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: result,
	})

	return streamingComplete{}
}

func (m *model) updateCommandPalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "escape" {
		m.commandView = false
		m.commandInput = ""
		return m, nil
	}

	if msg.String() == "enter" {
		if m.selectedCmd < len(m.commands) {
			return m.commands[m.selectedCmd].Action(m)
		}
	}

	if msg.String() == "up" && m.selectedCmd > 0 {
		m.selectedCmd--
		return m, nil
	}

	if msg.String() == "down" && m.selectedCmd < len(m.commands)-1 {
		m.selectedCmd++
		return m, nil
	}

	if msg.String() == "backspace" && len(m.commandInput) > 0 {
		m.commandInput = m.commandInput[:len(m.commandInput)-1]
		return m, nil
	}

	if msg.Runes != nil {
		m.commandInput += string(msg.Runes)
		for i, cmd := range m.commands {
			if strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(m.commandInput)) {
				m.selectedCmd = i
				break
			}
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var output strings.Builder

	if m.commandView {
		output.WriteString(m.renderCommandPalette())
	} else {
		output.WriteString(m.renderChat())
	}

	output.WriteString("\n")
	output.WriteString(renderInput(m.input, m.streaming))

	if m.errMsg != "" {
		output.WriteString("\n")
		output.WriteString(errorStyle.Render("Error: "+m.errMsg))
	}

	return output.String()
}

func (m model) renderChat() string {
	var output strings.Builder

	for _, msg := range m.messages {
		output.WriteString(renderMessage(msg))
		output.WriteString("\n")
	}

	if m.streaming {
		m.spinnerMutex.Lock()
		frame := spinnerFrames[m.spinnerFrame]
		m.spinnerMutex.Unlock()
		output.WriteString(assistantBubbleStyle.Render("Nexly") + " " + frame + "\n")
	}

	return output.String()
}

func renderMessage(msg Message) string {
	var bubble string
	if msg.Role == "user" {
		bubble = userBubbleStyle.Render("You")
	} else {
		bubble = assistantBubbleStyle.Render("Nexly")
	}

	content := utils.FormatMarkdown(msg.Content)
	lines := strings.Split(content, "\n")
	
	var contentStr strings.Builder
	for i, line := range lines {
		if i > 0 {
			contentStr.WriteString(strings.Repeat(" ", runewidth.StringWidth(bubble)-1))
		}
		contentStr.WriteString(" " + line + "\n")
	}

	return bubble + "\n" + contentStr.String()
}

func renderInput(input string, disabled bool) string {
	prompt := primaryStyle.Render("> ")
	if disabled {
		prompt = secondaryStyle.Render("... ")
	}
	return prompt + input + "_"
}

func (m model) renderCommandPalette() string {
	var output strings.Builder

	output.WriteString(primaryStyle.Render("Command Palette"))
	output.WriteString(" (Press Esc to close)\n")
	output.WriteString(secondaryStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	filtered := m.commands
	if m.commandInput != "" {
		filtered = []Command{}
		for _, cmd := range m.commands {
			if strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(m.commandInput)) {
				filtered = append(filtered, cmd)
			}
		}
	}

	for i, cmd := range filtered {
		prefix := "  "
		if i == m.selectedCmd {
			prefix = primaryStyle.Render("> ")
		}
		output.WriteString(fmt.Sprintf("%s%s %s\n", prefix, cmd.Name, secondaryStyle.Render(cmd.Description)))
	}

	output.WriteString("\n")
	output.WriteString(secondaryStyle.Render("Press Enter to execute, Esc to close, ↑↓ to navigate"))

	return output.String()
}

type spinnerTick struct{}

type streamingComplete struct{}

type streamingError struct {
	err error
}

var spinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

func switchProviderCmd(m *model) (tea.Model, tea.Cmd) {
	providersList := config.GetProviders()
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: "Available providers:\n" + strings.Join(providersList, "\n") + "\n\nUse 'nexly provider set <provider>' to switch.",
	})
	m.commandView = false
	m.commandInput = ""
	return m, nil
}

func switchModelCmd(m *model) (tea.Model, tea.Cmd) {
	models := config.GetModels(m.provider)
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: fmt.Sprintf("Available models for %s:\n%s\n\nUse 'nexly model set <model>' to switch.", m.provider, strings.Join(models, "\n")),
	})
	m.commandView = false
	m.commandInput = ""
	return m, nil
}

func clearChatCmd(m *model) (tea.Model, tea.Cmd) {
	config.ClearHistory()
	m.messages = []Message{}
	m.commandView = false
	m.commandInput = ""
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: "Chat history cleared.",
	})
	return m, nil
}

func helpCmd(m *model) (tea.Model, tea.Cmd) {
	m.commandView = false
	m.commandInput = ""
	helpText := `
Nexly - AI Coding Assistant
============================

Commands:
  /provider    - Switch AI provider
  /model      - Switch AI model
  /clear      - Clear chat history
  /config     - Configure API keys
  /help       - Show this help
  /exit       - Exit Nexly

Keyboard Shortcuts:
  Ctrl+P      - Open command palette
  Ctrl+C      - Exit Nexly
  Ctrl+U      - Clear input
`
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: helpText,
	})
	return m, nil
}

func configCmd(m *model) (tea.Model, tea.Cmd) {
	m.commandView = false
	m.commandInput = ""
	configText := `
Nexly Configuration
====================

To set API keys, edit ~/.nexly/config.json:

{
  "provider": "openai",
  "model": "gpt-4",
  "api_keys": {
    "openai": "sk-...",
    "anthropic": "sk-ant-...",
    "google": "AIza...",
    "openrouter": "sk-or-...",
    "nvidia": "nvapi-..."
  }
}

Available providers: openai, anthropic, google, openrouter, nvidia
`
	m.messages = append(m.messages, Message{
		Role:    "assistant",
		Content: configText,
	})
	return m, nil
}

func exitCmd(m *model) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}
