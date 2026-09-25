# Azure Pricing Calculator MCP Server ⚡

Servidor **MCP (Model Context Protocol)** em **Go (Golang)** de binário único que permite que assistentes de IA (Claude Desktop, Claude Code, Cursor, Codex, etc.) gerem **links oficiais e permanentes da Calculadora de Preços do Azure (`https://azure.com/e/...`)** a partir de linguagem natural ou especificações técnicas de arquitetura.

---

## 🚀 Como Funciona

1. Você pede para a IA no chat:
   > *"Gere uma estimativa da Azure para 2 VMs D4s_v5 Linux no Brasil South com SSD Premium 256GB, 1 Azure SQL General Purpose 4 vCores e 1 Storage Account Hot de 1TB em BRL."*
2. O LLM chama a tool `create_azure_estimate` do servidor MCP.
3. O binário obtém a sessão da Microsoft, monta os recursos e gera a estimativa via API.
4. Você recebe o link oficial oficial `https://azure.com/e/<hash>` pronto para enviar ao cliente.

---

## 🛠️ Configuração nos Clientes MCP

### 1. Claude Desktop (`claude_desktop_config.json`)
No Windows, abra `%APPDATA%\Claude\claude_desktop_config.json` e adicione:

```json
{
  "mcpServers": {
    "azure-calc": {
      "command": "C:\\Users\\Hylex\\.gemini\\antigravity\\scratch\\azure-calc-mcp\\azure-calc-mcp.exe"
    }
  }
}
```

### 2. Cursor / Codex
Adicione nas configurações de MCP Tools do editor:
- **Type:** `command`
- **Command:** `C:\Users\Hylex\.gemini\antigravity\scratch\azure-calc-mcp\azure-calc-mcp.exe`

---

## 🧪 Testes via Linha de Comando (CLI)

O binário também pode ser testado diretamente no terminal:

```powershell
# Teste rápido automático (gera estimativa de 4 recursos e devolve o link)
.\azure-calc-mcp.exe -test

# Gerar a partir de um arquivo JSON
.\azure-calc-mcp.exe -cli -file arquitetura.json
```

---

## 📋 Recursos Suportados

- **AKS / Kubernetes Service (`aks`, `kubernetes-service`):** Nós/VMs, SO (Linux/Windows), Discos OS gerenciados (Standard SSD, Premium SSD), contagem de clusters, SLA tier.
- **Azure Database for PostgreSQL Flexible Server (`postgresql`, `flexible-server`):** Camadas (Burstable, General Purpose), computação (ex: `B1ms`), armazenamento SSD Premium, backup LRS.
- **Storage Accounts (`storage`, `blob`):** General Purpose v2, Block Blob, Redundância (`LRS`, `ZRS`, `GRS`), Camada de acesso (`Hot`, `Cool`, `Archive`), transações e capacidade em GB.
- **Key Vault (`key-vault`):** Operações normais e avançadas, chaves protegidas, pools HSM.
- **Azure Monitor / Log Analytics (`monitor`):** Volume de ingestão de logs em GB/dia, retenção interativa.
- **Virtual Machines (`vm`, `virtual-machines`):** SO (Linux/Windows), SKUs (ex: `D4s_v5`, `B2s`, etc.), Discos (`premiumssd`, `standardssd`, tamanhos `P10`, `P15`, `P20`, etc.), billing options.
- **Azure SQL Database (`azure-sql-database`, `sql`):** Tiers (`general-purpose`, `business-critical`), Provisioned / Serverless, contagem de vCores, Storage GB.
- **App Service (`app-service`):** Linux/Windows, Planos (`B1`, `S1`, `P1v3`, `P2v3`, etc.).
- **Recursos Genéricos do Azure:** Aceita qualquer `serviceSlug` e parâmetros arbitrários via `extra`.
- **Tratamento Automático de Compatibilidade:** Normalização de regiões (`brazilsouth` -> `brazil-south`), programa de licenciamento `MCA` sem necessidade de login prévio e schemas oficiais embutidos sem alertas de indisponibilidade.
