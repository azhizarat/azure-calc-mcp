package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func runSelfInstall() {
	fmt.Println("==========================================================")
	fmt.Println("  Azure Pricing Calculator MCP - Native Auto-Installer")
	fmt.Println("==========================================================")

	targetDir := getTargetInstallDir()
	targetExe := filepath.Join(targetDir, getBinaryName())

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Erro ao criar diretorio de instalacao: %v\n", err)
		os.Exit(1)
	}

	currentExe, err := os.Executable()
	if err == nil {
		currentAbs, _ := filepath.Abs(currentExe)
		targetAbs, _ := filepath.Abs(targetExe)
		if currentAbs != targetAbs {
			if err := copyFile(currentAbs, targetAbs); err == nil {
				fmt.Printf("[1/4] Binario instalado em: %s\n", targetAbs)
			} else {
				fmt.Printf("[1/4] Usando binario em: %s\n", currentAbs)
				targetExe = currentAbs
			}
		} else {
			fmt.Printf("[1/4] Binario ja esta no caminho de instalacao: %s\n", targetAbs)
		}
	}

	// 1. Claude Desktop (Traditional + Windows Store / MSIX)
	claudePaths := getClaudeDesktopConfigPaths()
	configuredClaude := false
	for _, p := range claudePaths {
		if err := updateMCPConfig(p, targetExe); err == nil {
			fmt.Printf("[2/4] Claude Desktop configurado com sucesso: %s\n", p)
			configuredClaude = true
		}
	}
	if !configuredClaude && len(claudePaths) > 0 {
		_ = updateMCPConfig(claudePaths[0], targetExe)
		fmt.Printf("[2/4] Claude Desktop configurado: %s\n", claudePaths[0])
	}

	// 2. Claude CLI (~/.claude.json)
	claudeCLIPath := getClaudeCLIConfigPath()
	if _, err := os.Stat(claudeCLIPath); err == nil {
		if err := updateMCPConfig(claudeCLIPath, targetExe); err == nil {
			fmt.Printf("[3/4] Claude CLI configurado com sucesso: %s\n", claudeCLIPath)
		}
	} else {
		fmt.Println("[3/4] Claude CLI (~/.claude.json) nao encontrado.")
	}

	// 3. Cursor
	cursorPath := getCursorConfigPath()
	if _, err := os.Stat(filepath.Dir(cursorPath)); err == nil {
		if err := updateMCPConfig(cursorPath, targetExe); err == nil {
			fmt.Printf("[4/4] Cursor configurado com sucesso: %s\n", cursorPath)
		}
	} else {
		fmt.Println("[4/4] Pasta do Cursor nao encontrada (pode configurar manualmente nas Settings).")
	}

	// 4. Windsurf
	windsurfPath := getWindsurfConfigPath()
	if _, err := os.Stat(filepath.Dir(windsurfPath)); err == nil {
		if err := updateMCPConfig(windsurfPath, targetExe); err == nil {
			fmt.Printf("[+] Windsurf configurado com sucesso: %s\n", windsurfPath)
		}
	}

	// 5. VS Code Cline & Roo Code
	for _, extPath := range getVSCodeExtensionPaths() {
		if _, err := os.Stat(filepath.Dir(extPath)); err == nil {
			if err := updateMCPConfig(extPath, targetExe); err == nil {
				fmt.Printf("[+] Extensao configurada com sucesso: %s\n", extPath)
			}
		}
	}

	fmt.Println("")
	fmt.Println("==========================================================")
	fmt.Println("  Instalacao concluida com sucesso!")
	fmt.Println("  Reinicie o Claude Desktop, Claude CLI, Cursor ou Windsurf.")
	fmt.Println("==========================================================")
}

func getTargetInstallDir() string {
	if runtime.GOOS == "windows" {
		localApp := os.Getenv("LOCALAPPDATA")
		if localApp == "" {
			localApp = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localApp, "azure-calc-mcp")
	}
	home := os.Getenv("HOME")
	return filepath.Join(home, ".azure-calc-mcp", "bin")
}

func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "azure-calc-mcp.exe"
	}
	return "azure-calc-mcp"
}

func getClaudeDesktopConfigPaths() []string {
	var paths []string
	if runtime.GOOS == "windows" {
		// Standard Win32 roaming path
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		standardPath := filepath.Join(appData, "Claude", "claude_desktop_config.json")
		paths = append(paths, standardPath)

		// Microsoft Store / MSIX virtualized package path
		localApp := os.Getenv("LOCALAPPDATA")
		if localApp == "" {
			localApp = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		pattern := filepath.Join(localApp, "Packages", "Claude_*", "LocalCache", "Roaming", "Claude", "claude_desktop_config.json")
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			paths = append(paths, matches...)
		}
	} else if runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		paths = append(paths, filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"))
	} else {
		home := os.Getenv("HOME")
		paths = append(paths, filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"))
	}
	return paths
}

func getClaudeCLIConfigPath() string {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude.json")
}

func getCursorConfigPath() string {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".cursor", "mcp.json")
}

func getWindsurfConfigPath() string {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
}

func getVSCodeExtensionPaths() []string {
	var paths []string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			paths = append(paths,
				filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"),
				filepath.Join(appData, "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "settings", "cline_mcp_settings.json"),
			)
		}
	}
	return paths
}

func updateMCPConfig(filePath, targetExe string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	config := map[string]interface{}{}
	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &config)
	}

	servers, ok := config["mcpServers"].(map[string]interface{})
	if !ok || servers == nil {
		servers = map[string]interface{}{}
	}

	servers["azure-calc"] = map[string]interface{}{
		"type":    "stdio",
		"command": targetExe,
	}
	config["mcpServers"] = servers

	outData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, outData, 0644)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
