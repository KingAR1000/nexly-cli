package handlers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetProjectContext() string {
	var context strings.Builder

	context.WriteString("Current Directory:\n")
	dir, _ := os.Getwd()
	context.WriteString("  " + dir + "\n\n")

	context.WriteString("Project Files:\n")
	entries, err := os.ReadDir(dir)
	if err == nil {
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				context.WriteString("  " + entry.Name() + "\n")
				count++
				if count > 20 {
					context.WriteString("  ... and more\n")
					break
				}
			}
		}
	}

	context.WriteString("\nDirectory Structure:\n")
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		rel, _ := filepath.Rel(dir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 1 || (len(parts) == 2 && parts[1] != "") {
			if !strings.HasPrefix(rel, ".") && rel != "node_modules" && rel != "vendor" && rel != ".git" {
				if info.IsDir() {
					context.WriteString("  📁 " + rel + "/\n")
				} else {
					context.WriteString("  📄 " + rel + "\n")
				}
			}
		}
		return nil
	})

	return context.String()
}

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func EditFile(path string, edits []FileEdit) error {
	content, err := ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(content, "\n")
	
	for _, edit := range edits {
		if edit.LineNumber > 0 && edit.LineNumber <= len(lines) {
			if edit.NewContent == "" {
				lines = append(lines[:edit.LineNumber-1], lines[edit.LineNumber:]...)
			} else {
				lines[edit.LineNumber-1] = edit.NewContent
			}
		}
	}

	newContent := strings.Join(lines, "\n")
	return WriteFile(path, newContent)
}

type FileEdit struct {
	LineNumber int
	NewContent string
}

func ShowDiff(original, new string) string {
	origLines := strings.Split(original, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString("--- Original\n")
	diff.WriteString("+++ Modified\n")

	maxLines := len(origLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		orig := ""
		if i < len(origLines) {
			orig = origLines[i]
		}
		new := ""
		if i < len(newLines) {
			new = newLines[i]
		}

		if orig == new {
			diff.WriteString(fmt.Sprintf("  %d: %s\n", i+1, orig))
		} else {
			if i < len(origLines) {
				diff.WriteString(fmt.Sprintf("- %d: %s\n", i+1, orig))
			}
			if i < len(newLines) {
				diff.WriteString(fmt.Sprintf("+ %d: %s\n", i+1, new))
			}
		}
	}

	return diff.String()
}

func SearchFiles(pattern string) ([]string, error) {
	cmd := exec.Command("grep", "-r", "-l", pattern, ".")
	cmd.Dir, _ = os.Getwd()
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	files := strings.Split(string(output), "\n")
	var result []string
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}

func GetGitInfo() string {
	var info strings.Builder

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir, _ = os.Getwd()
	if output, err := cmd.Output(); err == nil {
		info.WriteString("Branch: " + strings.TrimSpace(string(output)) + "\n")
	}

	cmd = exec.Command("git", "status", "--porcelain")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		modified := 0
		for _, line := range lines {
			if strings.HasPrefix(line, " M") || strings.HasPrefix(line, "??") {
				modified++
			}
		}
		if modified > 0 {
			info.WriteString(fmt.Sprintf("Modified files: %d\n", modified))
		}
	}

	return info.String()
}

func ParseFileEdits(content string) []FileEdit {
	var edits []FileEdit
	lines := strings.Split(content, "\n")
	
	currentEdit := FileEdit{}
	inEdit := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inEdit {
				if currentEdit.NewContent != "" {
					edits = append(edits, currentEdit)
				}
				currentEdit = FileEdit{}
				inEdit = false
			}
			continue
		}

		if strings.HasPrefix(line, "Edit file:") {
			inEdit = true
			continue
		}

		if strings.HasPrefix(line, "Line:") {
			fmt.Sscanf(line, "Line: %d", &currentEdit.LineNumber)
			continue
		}

		if inEdit && currentEdit.LineNumber > 0 {
			currentEdit.NewContent += line + "\n"
		}
	}

	return edits
}

func RunCommand(cmdStr string) (string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir, _ = os.Getwd()
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v - %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}
