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
				fmt.Printf("[1/3] Binario instalado em: %s\n", targetAbs)
			} else {
				fmt.Printf("[1/3] Usando binario em: %s\n", currentAbs)
				targetExe = currentAbs
			}
		} else {
			fmt.Printf("[1/3] Binario ja esta no caminho de instalacao: %s\n", targetAbs)
		}
	}

	// Claude Desktop
	claudePath := getClaudeConfigPath()
	if err := updateMCPConfig(claudePath, targetExe); err == nil {
		fmt.Printf("[2/3] Claude Desktop configurado com sucesso: %s\n", claudePath)
	} else {
		fmt.Printf("[2/3] Aviso ao configurar Claude Desktop: %v\n", err)
	}

	// Cursor
	cursorPath := getCursorConfigPath()
	if _, err := os.Stat(filepath.Dir(cursorPath)); err == nil {
		if err := updateMCPConfig(cursorPath, targetExe); err == nil {
			fmt.Printf("[3/3] Cursor configurado com sucesso: %s\n", cursorPath)
		}
	} else {
		fmt.Println("[3/3] Pasta do Cursor nao encontrada (pode configurar manualmente nas Settings).")
	}

	// Windsurf
	windsurfPath := getWindsurfConfigPath()
	if _, err := os.Stat(filepath.Dir(windsurfPath)); err == nil {
		if err := updateMCPConfig(windsurfPath, targetExe); err == nil {
			fmt.Printf("[+] Windsurf configurado com sucesso: %s\n", windsurfPath)
		}
	}

	// VS Code Cline & Roo Code
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
	fmt.Println("  Reinicie o Claude Desktop, Cursor ou seu assistente de IA.")
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

func getClaudeConfigPath() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	} else if runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	}
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
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
