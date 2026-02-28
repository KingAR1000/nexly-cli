#!/bin/bash

set -e

NEXLY_VERSION="1.0.0"
NEXLY_CONFIG_DIR="$HOME/.nexly"
NEXLY_BIN_DIR="$HOME/.local/bin"

echo "Installing Nexly v${NEXLY_VERSION}..."

if [ "$(id -u)" -eq 0 ]; then
    echo "Error: Do NOT run as root or with sudo!"
    echo "Run as regular user: curl -fsSL https://nexlycode.vercel.app/install.sh | bash"
    exit 1
fi

echo "Creating config directory..."
mkdir -p "$NEXLY_CONFIG_DIR"

if [ ! -f "$NEXLY_CONFIG_DIR/config.json" ]; then
    echo "Creating default config..."
    cat > "$NEXLY_CONFIG_DIR/config.json" << 'EOF'
{
  "provider": "openai",
  "model": "gpt-4",
  "temperature": 0.7,
  "max_tokens": 4096,
  "api_keys": {},
  "history": []
}
EOF
    chmod 600 "$NEXLY_CONFIG_DIR/config.json"
fi

echo "Creating bin directory..."
mkdir -p "$NEXLY_BIN_DIR"

echo "Downloading Nexly binary..."
curl -fsSL "https://nexlycode.vercel.app/nexly" -o "$NEXLY_BIN_DIR/nexly"
chmod +x "$NEXLY_BIN_DIR/nexly"

echo ""
echo "Installation complete!"
echo ""
echo "Add to your PATH if needed:"
echo "  export PATH=\$PATH:$NEXLY_BIN_DIR"
echo ""
echo "Then run: nexly"
echo ""
echo "To configure your API key, edit ~/.nexly/config.json"
