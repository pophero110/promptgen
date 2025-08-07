# promptgen

A simple, expressive, and predictable command-line tool to manage and generate prompts from templates.

## Features

* Manage prompt templates: add, list, update, delete, and review.
* Generate prompts from templates with flexible input:

  * Inline text
  * Clipboard content
  * Editor input (with fallback to clipboard)
* Template preview: preview a template’s content without generating (via `review` command).
* Clipboard support: generated prompts automatically copied to clipboard.
* Shell completions for bash and zsh.
* And more features planned (see below).

## Additional Useful Features to Consider

* Search templates by name or keyword.
* Template variables: Support multiple placeholders (e.g., `<input>`, `<context>`) with flexible input mapping.
* Configurable editor: Allow users to specify preferred editor via environment variable or config.
* Template import/export: Export templates to share or import shared ones.
* History/log: Keep a history of generated prompts for easy reference and reuse.
* JSON/YAML input support: Allow structured input for complex templates.
* Interactive mode: Guided prompt generation with step-by-step questions.
* Auto-update templates: Pull template updates from a remote repository or URL.
* Syntax validation: Check template syntax before saving or generating.
* Multi-language support: Localize UI messages and error prompts.
* Custom output formats: Save generated prompts directly to files or export as markdown, JSON, etc.
* Integration with other tools: For example, GPT API integration for direct completions.

---

## Installation

1. Make sure you have Go installed (version 1.16+ recommended).

2. Build the binary:

```bash
go build -o promptgen promptgen.go
```

3. Optionally, move the binary to a directory in your `$PATH`:

```bash
mv promptgen ~/.local/bin/
# Make sure ~/.local/bin is in your PATH
```

Or install using:

```bash
go install
```

---

## Usage

```bash
promptgen COMMAND [ARGS...]
```

### Commands

* `list`
  List all saved prompt templates.

* `add NAME`
  Add a new prompt template named `NAME`. Enter content in the prompt.

* `update NAME`
  Update an existing prompt template named `NAME`. Shows current content, then prompts for new content.

* `delete NAME`
  Delete a prompt template by name.

* `generate NAME [TEXT_INPUT | --clip]`
  Generate a prompt from template `NAME`.

  * Use `TEXT_INPUT` as input if provided.
  * Use `--clip` flag to read input from clipboard.
  * Otherwise, open editor for input, fallback to clipboard if input empty.

* `review NAME`
  Preview the content of a prompt template without generating it.

* `completion SHELL`
  Output shell completion script for `bash` or `zsh`.

---

## Template Syntax

Use `<input>` as a placeholder in your template. It will be replaced by the user input when generating.

Example:

```
Summarize the following text:

<input>
```

---

## Examples

### Add a template

```bash
promptgen add summary
```

Enter template content (end with EOF / Ctrl+D):

```
Summarize the following text clearly:

<input>
```

### Generate prompt with inline input

```bash
promptgen generate summary "This is the text to summarize."
```

### Generate prompt using clipboard content as input

```bash
promptgen generate summary --clip
```

### Generate prompt by opening editor for input

```bash
promptgen generate summary
```

### Preview a template’s content

```bash
promptgen review summary
```

---

## Clipboard Support

Generated prompts are automatically copied to your clipboard for easy pasting.

---

## License

MIT License

---

Feel free to contribute or report issues!
If you want me to save this as a file or add badges or more sections, just say!
