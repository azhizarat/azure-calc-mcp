#!/usr/bin/env bash
set -e

# ==========================================================
#   Azure Pricing Calculator MCP Server - Installer (macOS/Linux)
# ==========================================================

REPO="azhizarat/azure-calc-mcp"
VERSION="${1:-latest}"

echo "=========================================================="
echo "  Azure Pricing Calculator MCP Server - Installer"
echo "=========================================================="

# 1. Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)
    echo "Error: Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "[1/4] Detected Platform: ${OS}-${ARCH}"

# 2. Setup install directory
INSTALL_DIR="$HOME/.azure-calc-mcp/bin"
mkdir -p "$INSTALL_DIR"
BINARY_PATH="$INSTALL_DIR/azure-calc-mcp"

# 3. Download or Copy Binary
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/azure-calc-mcp" ]; then
  echo "[2/4] Installing from local build..."
  cp "$SCRIPT_DIR/azure-calc-mcp" "$BINARY_PATH"
else
  echo "[2/4] Downloading latest release from GitHub..."
  if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/azure-calc-mcp-${OS}-${ARCH}"
  else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/azure-calc-mcp-${OS}-${ARCH}"
  fi
  
  if curl -fsSL -o "$BINARY_PATH" "$URL"; then
    echo "      Downloaded ${OS}-${ARCH} binary."
  else
    echo "      Falling back to default release asset..."
    curl -fsSL -o "$BINARY_PATH" "https://github.com/${REPO}/releases/latest/download/azure-calc-mcp"
  fi
fi

chmod +x "$BINARY_PATH"
echo "      Installed to: $BINARY_PATH"

# 4. Auto-Configure Claude Desktop
echo "[3/4] Configuring Claude Desktop..."
if [ "$OS" = "darwin" ]; then
  CLAUDE_CONFIG_DIR="$HOME/Library/Application Support/Claude"
else
  CLAUDE_CONFIG_DIR="$HOME/.config/Claude"
fi

CLAUDE_CONFIG_FILE="$CLAUDE_CONFIG_DIR/claude_desktop_config.json"
mkdir -p "$CLAUDE_CONFIG_DIR"

# Merge JSON safely using python3 or node
if command -v python3 >/dev/null 2>&1; then
  python3 -c "
import json, os
path = '$CLAUDE_CONFIG_FILE'
bin_path = '$BINARY_PATH'
data = {'mcpServers': {}}
if os.path.exists(path):
    try:
        with open(path, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception:
        pass
if 'mcpServers' not in data or not isinstance(data['mcpServers'], dict):
    data['mcpServers'] = {}
data['mcpServers']['azure-calc'] = {'command': bin_path}
with open(path, 'w', encoding='utf-8') as f:
    json.dump(data, f, indent=2)
"
  echo "      Configured: $CLAUDE_CONFIG_FILE"
elif command -v node >/dev/null 2>&1; then
  node -e "
const fs = require('fs');
const file = '$CLAUDE_CONFIG_FILE';
let data = { mcpServers: {} };
if (fs.existsSync(file)) {
  try { data = JSON.parse(fs.readFileSync(file, 'utf8')); } catch(e) {}
}
if (!data.mcpServers) data.mcpServers = {};
data.mcpServers['azure-calc'] = { command: '$BINARY_PATH' };
fs.writeFileSync(file, JSON.stringify(data, null, 2));
"
  echo "      Configured: $CLAUDE_CONFIG_FILE"
else
  echo "      Notice: Install python3 or node to auto-merge configuration."
  echo "      Or add manually to $CLAUDE_CONFIG_FILE:"
  echo '      {"mcpServers": {"azure-calc": {"command": "'$BINARY_PATH'"}}}'
fi

# 5. Finished
echo "[4/4] Finalizing setup..."
echo ""
echo "=========================================================="
echo "  Installation Completed Successfully!"
echo "=========================================================="
echo "  Binary Location:  $BINARY_PATH"
echo "  Claude Config:    $CLAUDE_CONFIG_FILE"
echo ""
echo "  Next steps: Restart Claude Desktop or Cursor to activate."
echo "=========================================================="
