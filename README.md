# promptgen

A simple, expressive, and predictable command-line tool to manage and generate prompts from templates.

## Additional Useful Features to Consider

- Template preview: Preview a template’s content without generating to confirm before use.

- Search templates: Quickly find templates by name or keyword.

- Template variables: Support multiple placeholders (e.g. <input>, <context>) with flexible input mapping.

- Configurable editor: Allow users to specify their preferred editor via environment variable or config.

- Template import/export: Export templates to share with others or import shared templates.

- History/log: Keep a history of generated prompts for easy reference and reuse.

- JSON/YAML input support: Allow structured input for complex templates.

- Interactive mode: Guided prompt generation with step-by-step questions.

- Auto-update templates: Pull template updates from a remote repo or URL.

- Syntax validation: Check template syntax before saving or generating.

- Multi-language support: Localize UI messages and error prompts.

- Template versioning: Keep versions of templates and revert if needed.

- Custom output formats: Save generated prompts directly to files, or export as markdown, JSON, etc.

- Integration with other tools: For example, integrate with GPT APIs to generate completions directly.

Shell completions: Support bash/zsh/fish auto-completion for commands and template names.
---

## Installation

1. Make sure you have Go installed (version 1.16+ recommended).

2. Build the binary:

```bash
go build -o promptgen promptgen.go
````

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

```
promptgen COMMAND [ARGS...]
```

### Commands

* `list`
  Lists all saved prompt templates.

* `create NAME`
  Create a new prompt template with the given `NAME`. You will be prompted to enter the template content.

* `update NAME`
  Update an existing prompt template named `NAME`. Shows current content, then prompts for new content.

* `delete NAME`
  Delete a prompt template by `NAME`.

* `generate NAME [TEXT_INPUT | --clip]`
  Generate a prompt using the template `NAME`.

  * If `TEXT_INPUT` is provided, it is used as input.
  * If `--clip` flag is provided, input is read from the clipboard.
  * If neither, opens an editor to enter input, falling back to clipboard if empty.

---

## Template Syntax

Use `<input>` as a placeholder in your template. It will be replaced by the user input.

Example:

```
Summarize the following text:

<input>
```

---

## Examples

### Create a template

```bash
promptgen create summary
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

---

## Clipboard Support

Generated prompts are automatically copied to your clipboard for easy pasting.

---

## License

MIT License

---

Feel free to contribute or report issues!
If you want me to save this as a file or add any extra badges or sections, just say!

