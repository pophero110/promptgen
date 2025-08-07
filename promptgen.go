package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/atotto/clipboard"
)

const templateDir = ".promptgen/templates"

type PromptTemplate struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		addTemplate()
	case "list":
		listTemplates()
	case "delete":
		deleteTemplate()
	case "update":
		updateTemplate()
	case "generate":
		generatePrompt()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Usage: promptgen <command> [args]

Commands:
  add           Add a new prompt template
  list          List all prompt templates
  delete NAME   Delete a prompt template by name
  update NAME   Update a prompt template by name
  generate NAME [TEXT_INPUT]   Generate prompt from template; if TEXT_INPUT omitted, opens editor
`)
}

// Helpers
func getTemplatePath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, templateDir, name+".json")
}

func ensureTemplateDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, templateDir)
	return os.MkdirAll(dir, os.ModePerm)
}

// CRUD

func addTemplate() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Template name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Println("Template name cannot be empty")
		return
	}

	fmt.Println("Enter template content (end with EOF/Ctrl+D):")
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Error reading template content:", err)
		return
	}

	if err := ensureTemplateDir(); err != nil {
		fmt.Println("Error creating template directory:", err)
		return
	}

	tpl := PromptTemplate{
		Name:     name,
		Template: string(content),
	}

	data, _ := json.MarshalIndent(tpl, "", "  ")
	if err := os.WriteFile(getTemplatePath(name), data, 0644); err != nil {
		fmt.Println("Error saving template:", err)
		return
	}

	fmt.Println("Template saved.")
}

func listTemplates() {
	if err := ensureTemplateDir(); err != nil {
		fmt.Println("Error accessing templates:", err)
		return
	}
	home, _ := os.UserHomeDir()
	files, err := filepath.Glob(filepath.Join(home, templateDir, "*.json"))
	if err != nil {
		fmt.Println("Error listing templates:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No templates found.")
		return
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var t PromptTemplate
		if err := json.Unmarshal(data, &t); err == nil {
			fmt.Println("-", t.Name)
		}
	}
}

func deleteTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen delete NAME")
		return
	}
	name := os.Args[2]

	if err := os.Remove(getTemplatePath(name)); err != nil {
		fmt.Println("Error deleting template:", err)
		return
	}
	fmt.Println("Deleted template:", name)
}

func updateTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen update NAME")
		return
	}
	name := os.Args[2]

	path := getTemplatePath(name)
	if _, err := os.Stat(path); err != nil {
		fmt.Println("Template not found:", name)
		return
	}

	fmt.Println("Enter new template content (end with EOF/Ctrl+D):")
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Error reading content:", err)
		return
	}

	tpl := PromptTemplate{
		Name:     name,
		Template: string(content),
	}

	data, _ := json.MarshalIndent(tpl, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Println("Error updating template:", err)
		return
	}
	fmt.Println("Template updated.")
}

func generatePrompt() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen generate NAME [TEXT_INPUT | --clip]")
		return
	}
	name := os.Args[2]

	tpl, err := loadTemplate(name)
	if err != nil {
		fmt.Println("Error loading template:", err)
		return
	}

	var input string

	// Check for --clip flag
	if len(os.Args) >= 4 && os.Args[3] == "--clip" {
		input, err = clipboard.ReadAll()
		if err != nil {
			fmt.Println("Failed to read from clipboard:", err)
			return
		}
		fmt.Println("(Using input from clipboard)")
	} else if len(os.Args) >= 4 {
		input = os.Args[3]
	} else {
		input, err = openEditorForInput()
		if err != nil || strings.TrimSpace(input) == "" {
			input, err = clipboard.ReadAll()
			if err != nil {
				fmt.Println("Failed to read from clipboard:", err)
				return
			}
			fmt.Println("(Using input from clipboard as fallback)")
		} else {
			fmt.Println("(Using input from editor)")
		}
	}

	// Replace <input> with Go template syntax
	normalizedTpl := strings.ReplaceAll(tpl.Template, "<input>", "{{.Input}}")
	data := map[string]string{"Input": input}

	tmpl, err := template.New(tpl.Name).Parse(normalizedTpl)
	if err != nil {
		fmt.Println("Template parse error:", err)
		return
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		fmt.Println("Template execution error:", err)
		return
	}

	promptStr := output.String()
	fmt.Println("\nGenerated Prompt:")
	fmt.Println(promptStr)

	if err := clipboard.WriteAll(promptStr); err != nil {
		fmt.Println("Warning: failed to copy to clipboard:", err)
	} else {
		fmt.Println("\nPrompt copied to clipboard!")
	}
}

func loadTemplate(name string) (PromptTemplate, error) {
	var t PromptTemplate
	data, err := os.ReadFile(getTemplatePath(name))
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(data, &t)
	return t, err
}

func openEditorForInput() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	tmpfile, err := ioutil.TempFile("", "promptgen_input_*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}
