# ned

A CLI note-taking application.

## Installation

```bash
go install github.com/julianshen/ned@latest
```

## Usage

```
ned <command> [options]
```

## Commands

- `new` or `n`: Create a new note.
- `edit` or `e`: Edit an existing note.
- `list` or `l`: List all notes.
- `delete` or `d`: Delete a note.
- `view` or `v`: View a note in the browser.
- `image`: Manage images in notes
  - `image list [folder]`: List images in a folder's ._images_ directory. If no folder is specified, lists images in the root ._images_ directory.
  - `image show [image]`: Show an image using the system's default viewer. The image path can be either a filename for root images (e.g., `image.jpg`) or include a folder path (e.g., `folder/image.jpg`).
- `help` or `h`: Show help for a command.
- `config`: Manage configuration
  - `config set [key] [value]`: Set a configuration value
  - `config show`: Show all configuration values
- `clip`: Clip webpage content to a note
  - Usage: `clip [note] [url]`
  - If ANTHROPIC_API_KEY is set in config, downloads and summarizes the webpage content
  - If no API key is set, saves just the URL to the note

All notes are stored in `$HOME/.mynotes` directory.

## Configuration

Configuration is stored in `$HOME/.config/ned/config.toml` in TOML format. Available configuration options:

- `ANTHROPIC_API_KEY`: API key for Claude.ai integration

## Features

- Markdown notes with `.md` extension (using [goldmark](https://github.com/yuin/goldmark) parser)
- Tree-style note listing
- Browser-based note viewing
- Mermaid diagram support in markdown files
- Image support in notes
  - Images are stored in `._images_` directories alongside notes
  - Support both relative paths (e.g., `![](image.png)`) and folder paths (e.g., `![](folder/image.png)`)
  - Images in notes inherit their note's folder structure (e.g., for a note in `folder1/note.md`, an image `![](image.png)` is stored in `folder1/._images_/image.png`)

Example of a Mermaid diagram:
```mermaid
graph TD
    A[Start] --> B{Is it?}
    B -->|Yes| C[OK]
    B -->|No| D[Not OK]
```
