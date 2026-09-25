#!/usr/bin/env node

/**
 * Azure Pricing Calculator MCP Server - Universal Runner & Installer
 * Compatible with npx, Claude Desktop, Cursor, and terminal usage.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawn } = require('child_process');
const https = require('https');

const REPO = 'azhizarat/azure-calc-mcp';

// 1. Resolve Platform and Arch
function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  let osName = '';
  let ext = '';
  if (platform === 'win32') {
    osName = 'windows';
    ext = '.exe';
  } else if (platform === 'darwin') {
    osName = 'darwin';
  } else if (platform === 'linux') {
    osName = 'linux';
  } else {
    throw new Error(`Unsupported OS: ${platform}`);
  }

  let archName = '';
  if (arch === 'x64') {
    archName = 'amd64';
  } else if (arch === 'arm64') {
    archName = 'arm64';
  } else {
    throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { osName, archName, ext };
}

// 2. Resolve Binary Destination
function getBinaryPath() {
  const { osName, ext } = getPlatformInfo();
  if (osName === 'windows') {
    const base = process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local');
    return path.join(base, 'azure-calc-mcp', `azure-calc-mcp${ext}`);
  } else {
    return path.join(os.homedir(), '.azure-calc-mcp', 'bin', `azure-calc-mcp${ext}`);
  }
}

// 3. Download helper with redirect support
function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'User-Agent': 'azure-calc-mcp-installer' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return downloadFile(res.headers.location, destPath).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Download failed with HTTP ${res.statusCode} from ${url}`));
      }
      const dir = path.dirname(destPath);
      if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

      const file = fs.createWriteStream(destPath);
      res.pipe(file);
      file.on('finish', () => {
        file.close(() => resolve());
      });
      file.on('error', (err) => {
        fs.unlink(destPath, () => {});
        reject(err);
      });
    }).on('error', reject);
  });
}

// 4. Ensure Binary is Ready
async function ensureBinary() {
  const targetPath = getBinaryPath();

  // Check if binary already exists
  if (fs.existsSync(targetPath)) {
    return targetPath;
  }

  // Check if running from local repo build
  const localRepoBin = path.join(__dirname, '..', `azure-calc-mcp${getPlatformInfo().ext}`);
  if (fs.existsSync(localRepoBin)) {
    const dir = path.dirname(targetPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    fs.copyFileSync(localRepoBin, targetPath);
    if (process.platform !== 'win32') fs.chmodSync(targetPath, 0o755);
    return targetPath;
  }

  // Download from GitHub Release
  const { osName, archName, ext } = getPlatformInfo();
  const binaryName = `azure-calc-mcp-${osName}-${archName}${ext}`;
  const url = `https://github.com/${REPO}/releases/latest/download/${binaryName}`;

  process.stderr.write(`[azure-calc-mcp] Downloading native binary for ${osName}-${archName}...\n`);
  try {
    await downloadFile(url, targetPath);
  } catch (err) {
    // Fallback to legacy default name
    const fallbackUrl = `https://github.com/${REPO}/releases/latest/download/azure-calc-mcp${ext}`;
    await downloadFile(fallbackUrl, targetPath);
  }

  if (process.platform !== 'win32') {
    fs.chmodSync(targetPath, 0o755);
  }

  process.stderr.write(`[azure-calc-mcp] Binary ready at ${targetPath}\n`);
  return targetPath;
}

// 5. Config Helper for Claude Desktop & Cursor
function getClaudeDesktopConfigPaths() {
  const platform = os.platform();
  const paths = [];
  if (platform === 'win32') {
    // Win32 Roaming
    paths.push(path.join(process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming'), 'Claude', 'claude_desktop_config.json'));
    // Windows Store / MSIX package
    const localPackages = path.join(process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'), 'Packages');
    if (fs.existsSync(localPackages)) {
      try {
        const matches = fs.readdirSync(localPackages).filter(d => d.startsWith('Claude_'));
        for (const m of matches) {
          const storeConfig = path.join(localPackages, m, 'LocalCache', 'Roaming', 'Claude', 'claude_desktop_config.json');
          if (fs.existsSync(path.dirname(storeConfig))) {
            paths.push(storeConfig);
          }
        }
      } catch (e) {}
    }
  } else if (platform === 'darwin') {
    paths.push(path.join(os.homedir(), 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json'));
  } else {
    paths.push(path.join(os.homedir(), '.config', 'Claude', 'claude_desktop_config.json'));
  }
  return paths;
}

function getClaudeCLIConfigPath() {
  return path.join(os.homedir(), '.claude.json');
}

function getCursorConfigPath() {
  return path.join(os.homedir(), '.cursor', 'mcp.json');
}

function updateConfigFile(filePath, serverConfig) {
  const dir = path.dirname(filePath);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

  let config = { mcpServers: {} };
  if (fs.existsSync(filePath)) {
    try {
      config = JSON.parse(fs.readFileSync(filePath, 'utf8'));
    } catch (e) {
      fs.copyFileSync(filePath, `${filePath}.bak`);
    }
  }

  if (!config.mcpServers) config.mcpServers = {};
  config.mcpServers['azure-calc'] = serverConfig;

  fs.writeFileSync(filePath, JSON.stringify(config, null, 2), 'utf8');
}

// 6. Installer Mode
async function runInstaller() {
  console.log('==========================================================');
  console.log('  Azure Pricing Calculator MCP - Auto-Installer');
  console.log('==========================================================');

  const binaryPath = await ensureBinary();
  console.log(`[1/4] Native binary installed: ${binaryPath}`);

  // Claude Desktop (Win32 + Microsoft Store)
  const claudePaths = getClaudeDesktopConfigPaths();
  for (const cp of claudePaths) {
    updateConfigFile(cp, { type: 'stdio', command: binaryPath });
    console.log(`[2/4] Claude Desktop configured: ${cp}`);
  }

  // Claude CLI (~/.claude.json)
  const cliPath = getClaudeCLIConfigPath();
  if (fs.existsSync(cliPath)) {
    updateConfigFile(cliPath, { type: 'stdio', command: binaryPath });
    console.log(`[3/4] Claude CLI configured: ${cliPath}`);
  }

  // Cursor
  const cursorDir = path.dirname(getCursorConfigPath());
  if (fs.existsSync(cursorDir)) {
    const cursorPath = getCursorConfigPath();
    updateConfigFile(cursorPath, { command: binaryPath });
    console.log(`[3/4] Cursor configured: ${cursorPath}`);
  }

  // Windsurf
  const windsurfDir = path.join(os.homedir(), '.codeium', 'windsurf');
  if (fs.existsSync(windsurfDir)) {
    const windsurfConfig = path.join(windsurfDir, 'mcp_config.json');
    updateConfigFile(windsurfConfig, { command: binaryPath });
    console.log(`[+] Windsurf configured: ${windsurfConfig}`);
  }

  // VS Code Cline & Roo Code
  const appData = process.env.APPDATA || path.join(os.homedir(), '.config');
  const clinePath = path.join(appData, 'Code', 'User', 'globalStorage', 'saoudrizwan.claude-dev', 'settings', 'cline_mcp_settings.json');
  const rooPath = path.join(appData, 'Code', 'User', 'globalStorage', 'rooveterinaryinc.roo-cline', 'settings', 'cline_mcp_settings.json');
  
  if (fs.existsSync(path.dirname(clinePath))) {
    updateConfigFile(clinePath, { command: binaryPath });
    console.log(`[+] VS Code Cline configured: ${clinePath}`);
  }
  if (fs.existsSync(path.dirname(rooPath))) {
    updateConfigFile(rooPath, { command: binaryPath });
    console.log(`[+] VS Code Roo Code configured: ${rooPath}`);
  }

  console.log('');
  console.log('==========================================================');
  console.log('  Installation completed successfully!');
  console.log('  Compatible with Claude Desktop, Cursor, Windsurf, VS Code (Copilot/Cline), etc.');
  console.log('  Restart your AI tool to start using.');
  console.log('==========================================================');
}

// 7. Main Execution (Stdio Proxy or Installer)
async function main() {
  const args = process.argv.slice(2);

  if (args.includes('install') || args.includes('--install')) {
    await runInstaller();
    return;
  }

  // MCP Server Mode (Claude Desktop or Cursor runs npx azure-calc-mcp)
  try {
    const binaryPath = await ensureBinary();
    const child = spawn(binaryPath, args, {
      stdio: 'inherit',
      env: process.env
    });

    child.on('error', (err) => {
      process.stderr.write(`[azure-calc-mcp] Process error: ${err.message}\n`);
      process.exit(1);
    });

    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code || 0);
      }
    });

    // Forward termination signals
    process.on('SIGINT', () => child.kill('SIGINT'));
    process.on('SIGTERM', () => child.kill('SIGTERM'));
  } catch (err) {
    process.stderr.write(`[azure-calc-mcp] Fatal error: ${err.message}\n`);
    process.exit(1);
  }
}

main();
