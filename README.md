# glyph

> Small CLI tools to capture, retrieve and understand what happens in your terminal — because your memory deserves better infrastructure.

---

## Tools

| Name    | Description                                         | Status         |
|---------|-----------------------------------------------------|----------------|
| `pin`   | Clipboard for URLs, commands, paths and notes       | in development |
| `ask`   | Ask an AI with automatic directory context          | in development |
| `diff`  | Explain a git diff in plain English via AI          | in development |
| `stand` | Generate a standup update from recent git commits   | in development |

---

## Installation

### macOS / Linux

```sh
curl -fsSL https://github.com/reky0/glyph/releases/latest/download/install.sh | bash
```

The installer auto-detects your platform, presents a selection menu, downloads only what you choose, and offers to update your `$PATH`.

Install specific tools non-interactively:

```sh
curl -fsSL https://github.com/reky0/glyph/releases/latest/download/install.sh | bash -s -- pin ask
# or all at once:
curl -fsSL https://github.com/reky0/glyph/releases/latest/download/install.sh | bash -s -- --all
```

Override the install directory (default: `~/.local/bin`):

```sh
GLYPH_INSTALL_DIR=/usr/local/bin curl -fsSL .../install.sh | bash -s -- --all
```

### Windows (PowerShell)

```powershell
irm https://github.com/reky0/glyph/releases/latest/download/install.ps1 | iex
```

Install specific tools:

```powershell
.\install.ps1 -Tools pin,ask
# or all at once:
.\install.ps1 -All
```

Override the install directory (default: `%LOCALAPPDATA%\glyph\bin`):

```powershell
.\install.ps1 -All -InstallDir "C:\tools"
```

### From source

Requires Go 1.24+.

```sh
git clone https://github.com/reky0/glyph.git
cd glyph
make build          # → dist/pin  dist/ask  dist/diff  dist/stand
```

---

## Configuration

On first run, create `~/.config/glyph/config.toml`:

```toml
ai_provider = "groq"                      # groq | ollama
ai_model    = "llama-3.3-70b-versatile"
api_key     = "YOUR_KEY_HERE"             # ignored when provider is ollama
ollama_host = "http://localhost:11434"    # only used when provider is ollama
default_style = "rounded"
```

### Providers

- **groq** — cloud inference via [Groq](https://console.groq.com). Requires `api_key`.
- **ollama** — local inference via [Ollama](https://ollama.ai). Set `ollama_host` and leave `api_key` empty.

---

## `--style` flag

Every tool accepts a `--style` flag that controls output rendering:

| Value     | Description                                         |
|-----------|-----------------------------------------------------|
| `rounded` | Rounded borders, accent color `#7C6AF7` (default)  |
| `ascii`   | ASCII borders (`+`, `-`, `|`), muted grey only      |
| `minimal` | No borders, aligned columns, accent `#A8A8A8`       |

```sh
pin list --style ascii
ask "what is a goroutine?" --style minimal
```

---

## Data storage

Each tool stores its data under `~/.local/share/glyph/<toolname>/` following the XDG Base Directory specification.

| Tool    | Data path                           |
|---------|-------------------------------------|
| `pin`   | `~/.local/share/glyph/pin/pins.json` |
| `ask`   | _(no persistent state)_             |
| `diff`  | _(no persistent state)_             |
| `stand` | _(no persistent state)_             |

---

## Quick usage

```sh
# pin — save and retrieve things
pin add "https://pkg.go.dev/net/http" --tag go
pin add "kubectl get pods -n default" --cmd
pin list
pin search "kubectl"
pin get <id> | pbcopy
pin rm <id>

# ask — AI assistant
ask "how do I reverse a slice in Go?"
cat error.log | ask "what caused this?"
ask "explain this function" --no-context

# diff — explain changes
diff                         # git diff HEAD
diff --staged                # git diff --cached
diff --commit abc1234        # git show abc1234

# stand — standup generator
stand                        # commits since midnight
stand --since yesterday
stand --since "2 days ago"
```
