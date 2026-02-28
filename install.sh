#!/bin/bash

set -e

NEXLY_VERSION="1.0.0"
NEXLY_INSTALL_DIR="/usr/local/bin"
NEXLY_CONFIG_DIR="$HOME/.nexly"

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

echo "Downloading Nexly binary..."
curl -fsSL "https://nexlycode.vercel.app/nexly" -o "$NEXLY_INSTALL_DIR/nexly"
chmod +x "$NEXLY_INSTALL_DIR/nexly"

echo ""
echo "Installation complete!"
echo ""
echo "Run 'nexly' to start."
echo ""
echo "To configure your API key, edit ~/.nexly/config.json"
echo ""
echo "Example config:"
echo '{'
echo '  "provider": "openai",'
echo '  "model": "gpt-4",'
echo '  "api_keys": {'
echo '    "openai": "sk-your-api-key-here"'
echo '  }'
echo '}'
echo ""
