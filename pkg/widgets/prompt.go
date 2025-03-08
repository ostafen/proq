package widgets

import (
	"fmt"
	"strings"

	ui "github.com/ostafen/termui/v3"
	"github.com/ostafen/termui/v3/widgets"
)

type CmdHandler func(cmd string, args ...string) error

type Prompt struct {
	cmds map[string]CmdHandler
	*widgets.Paragraph
	hasError bool
}

const (
	promptInitialText = "> "
	promptCursor      = "[ ](bg:white)"
)

func NewPrompt() *Prompt {
	p := widgets.NewParagraph()
	p.Title = "Prompt"
	p.Text = promptInitialText + promptCursor
	p.TextStyle = ui.NewStyle(ui.ColorWhite)
	p.BorderStyle.Fg = ui.ColorWhite

	return &Prompt{
		Paragraph: p,
	}
}

func (p *Prompt) SetHandlers(cmds map[string]CmdHandler) {
	p.cmds = cmds
}

func (p *Prompt) OnKeyPressed(key string) bool {
	if p.clearError() && key == "<Enter>" {
		return true
	}

	switch key {
	case "<Up>", "<Down>", "<Left>", "<Right>":
		return false
	case "<Backspace>":
		if len(p.Text) > len(promptInitialText)+len(promptCursor) {
			p.Text = p.Text[:len(p.Text)-len(promptCursor)-1] + promptCursor
		}
	case "<Space>":
		p.updateText(" ")
	case "<Enter>":
		p.runCommand()
	case "<C-c>", "<Escape>":
		p.showExitHint("")
	default:
		p.updateText(key)
	}
	return true
}

func (p *Prompt) showExitHint(text string) {
	p.setError(fmt.Errorf("%stype \":q\" to exit", text))
}

func (p *Prompt) runCommand() {
	line := p.line()
	if line == "" {
		p.setError(fmt.Errorf("empty line"))
		return
	}

	cmdName, args, ok := tryParseCmd(line)
	if !ok {
		p.setError(fmt.Errorf("\"%s\" is not a valid command", line))
		return
	}

	for key, handler := range p.cmds {
		if strings.HasPrefix(line, ":"+key) {
			if err := handler(cmdName, args...); err != nil {
				p.setError(err)
				return
			}
			return
		}
	}

	p.setError(fmt.Errorf("\"%s\" is not a valid command", cmdName))
}

func (p *Prompt) setError(err error) {
	p.Text = fmt.Sprintf(" [%s](bg:red)", err)
	p.hasError = true
}

func (p *Prompt) clearError() bool {
	if !p.hasError {
		return false
	}

	p.Paragraph.Text = promptInitialText + promptCursor
	p.hasError = false
	return true
}

func (p *Prompt) line() string {
	return p.Text[len(promptInitialText) : len(p.Text)-len(promptCursor)]
}

func tryParseCmd(line string) (string, []string, bool) {
	if !strings.HasPrefix(line, ":") {
		return "", nil, false
	}

	parts := strings.Fields(line[1:])
	if len(parts) == 0 {
		return "", nil, false
	}

	cmdName := parts[0]
	return cmdName, parts[1:], true
}

func (p *Prompt) updateText(s string) {
	p.Text = promptInitialText + p.line() + s + promptCursor
}

func (p *Prompt) Resize(width, height int) {}
