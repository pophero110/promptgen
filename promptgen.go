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
	"strconv"
	"strings"
	"text/template"

	"github.com/atotto/clipboard"
)

const templateDir = ".promptgen/templates"

type PromptTemplate struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
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
	case "review":
		reviewTemplate()
	case "completion":
		completion()
	case "history":
		showHistory()
	case "versions":
		listVersions()
	case "view":
		viewVersion()
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Usage: promptgen <command> [args]

Commands:
  add NAME                 Add a new prompt template
  list                     List all prompt templates with versions
  delete NAME              Delete a prompt template and all versions
  update NAME              Update a prompt template by name (increments version)
  generate NAME [TEXT_INPUT | --clip]
                           Generate prompt from template; if TEXT_INPUT omitted, opens editor, or use --clip for clipboard input
  review NAME              Show latest version content of a prompt template
  versions NAME            List all versions of a template
  view NAME VERSION        View a specific version of a template
  completion SHELL         Output shell completion script (bash or zsh)
  history                  Show prompt generation history
`)
}

// Helpers

func getTemplatePath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, templateDir, name+".json")
}

func getTemplateVersionPath(name string, version int) string {
	home, _ := os.UserHomeDir()
	filename := fmt.Sprintf("%s_v%d.json", name, version)
	return filepath.Join(home, templateDir, filename)
}

func ensureTemplateDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, templateDir)
	return os.MkdirAll(dir, os.ModePerm)
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

func loadTemplateVersion(name string, version int) (PromptTemplate, error) {
	var t PromptTemplate
	data, err := os.ReadFile(getTemplateVersionPath(name, version))
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(data, &t)
	return t, err
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

	// Check if template exists already
	if _, err := os.Stat(getTemplatePath(name)); err == nil {
		fmt.Println("Template already exists. Use update command to modify it.")
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
		Version:  1,
		Template: string(content),
	}

	data, _ := json.MarshalIndent(tpl, "", "  ")

	// Save versioned file
	versionPath := getTemplateVersionPath(name, tpl.Version)
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		fmt.Println("Error saving versioned template:", err)
		return
	}

	// Save pointer to latest
	if err := os.WriteFile(getTemplatePath(name), data, 0644); err != nil {
		fmt.Println("Warning: failed to save main template pointer:", err)
	}

	fmt.Println("Template saved with version 1.")
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

	fmt.Println("Templates:")
	for _, f := range files {
		base := filepath.Base(f)
		// Skip versioned files (containing _v)
		if strings.Contains(base, "_v") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var t PromptTemplate
		if err := json.Unmarshal(data, &t); err == nil {
			fmt.Printf(" - %s (version %d)\n", t.Name, t.Version)
		}
	}
}

func deleteTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen delete NAME")
		return
	}
	name := os.Args[2]

	if err := ensureTemplateDir(); err != nil {
		fmt.Println("Error accessing templates:", err)
		return
	}

	home, _ := os.UserHomeDir()
	pattern := filepath.Join(home, templateDir, fmt.Sprintf("%s*.json", name))
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println("Error listing template files:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No such template found:", name)
		return
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Println("Error deleting file:", f, err)
		}
	}

	fmt.Println("Deleted template and all versions:", name)
}

func updateTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen update NAME")
		return
	}
	name := os.Args[2]

	// Load current template to get version
	tpl, err := loadTemplate(name)
	if err != nil {
		fmt.Println("Template not found:", name)
		return
	}

	fmt.Println("Enter new template content (end with EOF/Ctrl+D):")
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Error reading content:", err)
		return
	}

	tpl.Template = string(content)
	tpl.Version++ // increment version

	data, _ := json.MarshalIndent(tpl, "", "  ")

	if err := ensureTemplateDir(); err != nil {
		fmt.Println("Error creating template directory:", err)
		return
	}

	versionPath := getTemplateVersionPath(name, tpl.Version)
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		fmt.Println("Error saving versioned template:", err)
		return
	}

	// Update latest pointer
	if err := os.WriteFile(getTemplatePath(name), data, 0644); err != nil {
		fmt.Println("Warning: failed to update main template pointer:", err)
	}

	fmt.Printf("Template %q updated to version %d.\n", name, tpl.Version)
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
	fmt.Printf("\nGenerated Prompt (from template version %d):\n", tpl.Version)
	fmt.Println(promptStr)

	if err := clipboard.WriteAll(promptStr); err != nil {
		fmt.Println("Warning: failed to copy to clipboard:", err)
	} else {
		fmt.Println("\nPrompt copied to clipboard!")
	}
}

func reviewTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen review NAME")
		return
	}
	name := os.Args[2]

	tpl, err := loadTemplate(name)
	if err != nil {
		fmt.Printf("Template %q not found.\n", name)
		return
	}

	fmt.Printf("Template %q (version %d) content:\n\n", name, tpl.Version)
	fmt.Println(tpl.Template)
}

func completion() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen completion SHELL")
		fmt.Println("Supported shells: bash, zsh")
		return
	}

	shell := os.Args[2]

	switch shell {
	case "bash":
		fmt.Println(bashCompletionScript())
	case "zsh":
		fmt.Println(zshCompletionScript())
	default:
		fmt.Println("Unsupported shell:", shell)
	}
}

func bashCompletionScript() string {
	return `# bash completion for promptgen

_promptgen_completions() {
	local cur prev cmds templates
	COMPREPLY=()
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[COMP_CWORD-1]}"
	cmds="list add update delete generate review versions view completion history"

	# load templates from your data directory
	templates="$(promptgen list | tail -n +2 | awk '{print $2}')"

	if [[ $COMP_CWORD == 1 ]]; then
		COMPREPLY=( $(compgen -W "$cmds" -- "$cur") )
		return 0
	fi

	case "${COMP_WORDS[1]}" in
		generate|update|delete|review|versions|view)
			COMPREPLY=( $(compgen -W "$templates" -- "$cur") )
			return 0
			;;
		completion)
			COMPREPLY=( $(compgen -W "bash zsh" -- "$cur") )
			return 0
			;;
	esac
}

complete -F _promptgen_completions promptgen
`
}

func zshCompletionScript() string {
	return `#compdef promptgen

_arguments \
  '1:command:(list add update delete generate review versions view completion history)' \
  '2:template:->templates' \
  '3:arg:->args'

_templates() {
  reply=(${(f)"$(promptgen list | tail -n +2 | awk '{print $2}')"})
}

case $state in
  templates)
    _templates
    ;;
  args)
    ;;
esac
`
}

func showHistory() {
	historyPath := getHistoryPath()
	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No history found.")
		} else {
			fmt.Println("Error reading history:", err)
		}
		return
	}

	fmt.Println("Prompt Generation History:\n")
	fmt.Println(string(data))
}

func listVersions() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: promptgen versions NAME")
		return
	}
	name := os.Args[2]

	home, _ := os.UserHomeDir()
	pattern := filepath.Join(home, templateDir, fmt.Sprintf("%s_v*.json", name))
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println("Error listing versions:", err)
		return
	}
	if len(files) == 0 {
		fmt.Printf("No versions found for template %q\n", name)
		return
	}

	fmt.Printf("Versions for template %q:\n", name)
	for _, f := range files {
		base := filepath.Base(f) // e.g. "example_v3.json"
		verStr := strings.TrimSuffix(base, ".json")
		verParts := strings.Split(verStr, "_v")
		if len(verParts) != 2 {
			continue
		}
		fmt.Println("Version", verParts[1])
	}
}

func viewVersion() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: promptgen view NAME VERSION")
		return
	}
	name := os.Args[2]
	versionStr := os.Args[3]
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		fmt.Println("Invalid version number:", versionStr)
		return
	}

	tpl, err := loadTemplateVersion(name, version)
	if err != nil {
		fmt.Printf("Version %d of template %q not found.\n", version, name)
		return
	}

	fmt.Printf("Template %q version %d content:\n\n%s\n", name, version, tpl.Template)
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

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, templateDir, "history.log")
}
