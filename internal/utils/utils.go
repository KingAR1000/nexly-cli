package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

func FormatMarkdown(text string) string {
	var output strings.Builder
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	codeLang := ""
	codeContent := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				output.WriteString(formatCodeBlock(codeLang, codeContent))
				codeContent = []string{}
				inCodeBlock = false
				codeLang = ""
			} else {
				inCodeBlock = true
				parts := strings.Split(strings.TrimPrefix(line, "```"), " ")
				if len(parts) > 1 {
					codeLang = parts[1]
				}
			}
			continue
		}

		if inCodeBlock {
			codeContent = append(codeContent, line)
			continue
		}

		line = formatInlineCode(line)
		line = formatBold(line)
		line = formatItalic(line)
		line = formatHeaders(line)
		line = formatLists(line)
		line = formatLinks(line)

		output.WriteString(line)
		output.WriteString("\n")
	}

	return output.String()
}

func formatHeaders(line string) string {
	if strings.HasPrefix(line, "### ") {
		return fmt.Sprintf("\033[1;36m%s\033[0m", strings.TrimPrefix(line, "### "))
	}
	if strings.HasPrefix(line, "## ") {
		return fmt.Sprintf("\033[1;35m%s\033[0m", strings.TrimPrefix(line, "## "))
	}
	if strings.HasPrefix(line, "# ") {
		return fmt.Sprintf("\033[1;34m%s\033[0m", strings.TrimPrefix(line, "# "))
	}
	return line
}

func formatBold(line string) string {
	re := regexp.MustCompile(`\*\*(.+?)\*\*`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "**")
		return fmt.Sprintf("\033[1m%s\033[0m", content)
	})
}

func formatItalic(line string) string {
	re := regexp.MustCompile(`\*(.+?)\*`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "*")
		return fmt.Sprintf("\033[3m%s\033[0m", content)
	})
}

func formatInlineCode(line string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(line, func(match string) string {
		content := strings.Trim(match, "`")
		return fmt.Sprintf("\033[32m%s\033[0m", content)
	})
}

func formatCodeBlock(lang string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("\033[33m```%s\033[0m\n", lang))
	for _, line := range lines {
		output.WriteString(fmt.Sprintf("\033[90m%s\033[0m\n", line))
	}
	output.WriteString("\033[33m```\033[0m\n")
	return output.String()
}

func formatLists(line string) string {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return "  • " + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
	}
	if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
		return "  " + line
	}
	return line
}

func formatLinks(line string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		return match
	})
}

func Truncate(s string, maxLen int) string {
	if runewidth.StringWidth(s) <= maxLen {
		return s
	}
	truncated := runewidth.Truncate(s, maxLen-3, "...")
	return truncated
}

func PadRight(s string, width int) string {
	currentWidth := runewidth.StringWidth(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

func PadLeft(s string, width int) string {
	currentWidth := runewidth.StringWidth(s)
	if currentWidth >= width {
		return s
	}
	return strings.Repeat(" ", width-currentWidth) + s
}

func Center(s string, width int) string {
	currentWidth := runewidth.StringWidth(s)
	if currentWidth >= width {
		return s
	}
	left := (width - currentWidth) / 2
	right := width - currentWidth - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func WordWrap(text string, width int) string {
	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if runewidth.StringWidth(currentLine+" "+word) <= width {
			currentLine += " " + word
		} else {
			result.WriteString(currentLine)
			result.WriteString("\n")
			currentLine = word
		}
	}

	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}

func GetColorForFile(filename string) string {
	ext := strings.TrimPrefix(filename, ".")
	switch ext {
	case "go":
		return "\033[32m"
	case "js", "ts", "jsx", "tsx":
		return "\033[33m"
	case "py":
		return "\033[34m"
	case "rs":
		return "\033[31m"
	case "java":
		return "\033[36m"
	case "c", "cpp", "h":
		return "\033[35m"
	default:
		return "\033[37m"
	}
}
