<#
.SYNOPSIS
    Installs Azure Pricing Calculator MCP Server and automatically configures Claude Desktop & Cursor.
.DESCRIPTION
    Downloads the precompiled Go binary for your architecture (or uses local build),
    places it in %LOCALAPPDATA%\azure-calc-mcp, and configures Claude Desktop and Cursor.
.EXAMPLE
    irm https://raw.githubusercontent.com/kandiesky/azure-calc-mcp/main/install.ps1 | iex
#>

[CmdletBinding()]
param(
    [string]$Version = "latest",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  Azure Pricing Calculator MCP Server - Installer (Windows)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Determine Architecture
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit Windows is not supported. Please use a 64-bit OS."
    exit 1
}

Write-Host "[1/4] Detected Architecture: windows-$arch" -ForegroundColor Gray

# 2. Target Directory
$installDir = Join-Path $env:LOCALAPPDATA "azure-calc-mcp"
$exePath = Join-Path $installDir "azure-calc-mcp.exe"

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# 3. Acquire Binary
$localExe = Join-Path $PSScriptRoot "azure-calc-mcp.exe"
if (Test-Path $localExe) {
    Write-Host "[2/4] Installing from local build..." -ForegroundColor Green
    Copy-Item -Path $localExe -Destination $exePath -Force
} else {
    Write-Host "[2/4] Downloading latest release from GitHub..." -ForegroundColor Green
    $downloadUrl = if ($Version -eq "latest") {
        "https://github.com/kandiesky/azure-calc-mcp/releases/latest/download/azure-calc-mcp-windows-$arch.exe"
    } else {
        "https://github.com/kandiesky/azure-calc-mcp/releases/download/$Version/azure-calc-mcp-windows-$arch.exe"
    }
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing
    } catch {
        # Fallback to general windows-amd64 if specific architecture fails
        Write-Warning "Could not download $downloadUrl. Falling back to default binary release..."
        $fallbackUrl = "https://github.com/kandiesky/azure-calc-mcp/releases/latest/download/azure-calc-mcp.exe"
        Invoke-WebRequest -Uri $fallbackUrl -OutFile $exePath -UseBasicParsing
    }
}

# Verify Binary Exists and Runs
if (-not (Test-Path $exePath)) {
    Write-Error "Failed to install binary at $exePath"
    exit 1
}
Write-Host "      Installed to: $exePath" -ForegroundColor Gray

# 4. Auto-Configure Claude Desktop
Write-Host "[3/4] Configuring Claude Desktop..." -ForegroundColor Green
$claudeConfigDir = Join-Path $env:APPDATA "Claude"
$claudeConfigFile = Join-Path $claudeConfigDir "claude_desktop_config.json"

if (-not (Test-Path $claudeConfigDir)) {
    New-Item -ItemType Directory -Path $claudeConfigDir -Force | Out-Null
}

$claudeConfig = @{
    mcpServers = @{}
}

if (Test-Path $claudeConfigFile) {
    try {
        $raw = Get-Content -Path $claudeConfigFile -Raw -Encoding UTF8
        if ($raw.Trim().Length -gt 0) {
            $claudeConfig = $raw | ConvertFrom-Json -AsHashtable
            if (-not $claudeConfig.ContainsKey("mcpServers") -or ($null -eq $claudeConfig["mcpServers"])) {
                $claudeConfig["mcpServers"] = @{}
            }
        }
    } catch {
        Write-Warning "Could not parse existing claude_desktop_config.json, backing up to .bak"
        Copy-Item $claudeConfigFile "$claudeConfigFile.bak" -Force
        $claudeConfig = @{ mcpServers = @{} }
    }
}

$claudeConfig["mcpServers"]["azure-calc"] = @{
    command = $exePath
}

$jsonOut = $claudeConfig | ConvertTo-Json -Depth 10
Set-Content -Path $claudeConfigFile -Value $jsonOut -Encoding UTF8
Write-Host "      Configured: $claudeConfigFile" -ForegroundColor Gray

# 5. Auto-Configure Cursor (if installed or directory exists)
Write-Host "[4/4] Checking Cursor / Codex..." -ForegroundColor Green
$cursorDir = Join-Path $env:USERPROFILE ".cursor"
$cursorConfigFile = Join-Path $cursorDir "mcp.json"

if (Test-Path $cursorDir) {
    $cursorConfig = @{ mcpServers = @{} }
    if (Test-Path $cursorConfigFile) {
        try {
            $craw = Get-Content -Path $cursorConfigFile -Raw -Encoding UTF8
            if ($craw.Trim().Length -gt 0) {
                $cursorConfig = $craw | ConvertFrom-Json -AsHashtable
                if (-not $cursorConfig.ContainsKey("mcpServers") -or ($null -eq $cursorConfig["mcpServers"])) {
                    $cursorConfig["mcpServers"] = @{}
                }
            }
        } catch {
            $cursorConfig = @{ mcpServers = @{} }
        }
    }
    $cursorConfig["mcpServers"]["azure-calc"] = @{
        command = $exePath
    }
    Set-Content -Path $cursorConfigFile -Value ($cursorConfig | ConvertTo-Json -Depth 10) -Encoding UTF8
    Write-Host "      Configured Cursor: $cursorConfigFile" -ForegroundColor Gray
} else {
    Write-Host "      Cursor directory not detected (skipping). You can configure it manually in Cursor Settings." -ForegroundColor Gray
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Green
Write-Host "  Installation Completed Successfully!" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
Write-Host "  Binary Location:  $exePath"
Write-Host "  Claude Config:    $claudeConfigFile"
Write-Host ""
Write-Host "  Next steps: Restart Claude Desktop or Cursor to activate." -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Green
