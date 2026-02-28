# Nexly - AI Coding Assistant

Nexly is a powerful CLI coding assistant that replicates the functionality of Opencode, fully rebranded with its own identity.

## Features

- **Multi-Provider Support**: OpenAI, Anthropic, Google, NVIDIA, OpenRouter
- **Model Switching**: Switch between different models within each provider
- **Command Palette**: Press Ctrl+P to open the command palette
- **Terminal-First UI**: Beautiful terminal interface with syntax highlighting
- **Streaming Responses**: Real-time AI responses as they are generated
- **Project Context**: Automatically reads and understands your project files
- **File Editing**: Read, write, and edit files with diff previews

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://nexlycode.vercel.app/install.sh | bash
```

### Manual Installation

1. Download the latest release
2. Extract the zip file
3. Run `sudo cp nexly /usr/local/bin/nexly`
4. Run `nexly` to start

## Configuration

Nexly stores configuration in `~/.nexly/config.json`:

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "temperature": 0.7,
  "max_tokens": 4096,
  "api_keys": {
    "openai": "sk-your-api-key",
    "anthropic": "sk-ant-your-api-key",
    "google": "your-google-api-key",
    "openrouter": "sk-or-your-api-key",
    "nvidia": "nvapi-your-api-key"
  }
}
```

## Usage

### Basic Commands

- `nexly` - Start the interactive CLI
- `nexly provider set <provider>` - Switch AI provider
- `nexly model set <model>` - Switch AI model
- `nexly config` - Show current configuration
- `nexly version` - Show version

### Command Palette

Press `Ctrl+P` to open the command palette with these commands:
- `/provider` - Switch provider
- `/model` - Switch model
- `/clear` - Clear chat history
- `/config` - Configure API keys
- `/help` - Show help
- `/exit` - Exit Nexly

### Keyboard Shortcuts

- `Ctrl+P` - Open command palette
- `Ctrl+C` - Exit Nexly
- `Ctrl+U` - Clear input

## Supported Providers

| Provider | Models |
|----------|--------|
| OpenAI | gpt-4, gpt-4o, gpt-3.5-turbo, o1 |
| Anthropic | claude-3-5-sonnet, claude-3-opus |
| Google | gemini-1.5-pro, gemini-1.5-flash |
| OpenRouter | Various open-source models |
| NVIDIA | llama-3.1-nemotron, mixtral |

## License

MIT License
