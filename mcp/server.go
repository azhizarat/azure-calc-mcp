package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"azure-calc-mcp/azure"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 message
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 message
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolDefinition defines a tool exposed via MCP
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// Server is the stdio MCP server
type Server struct {
	azureClient *azure.Client
	logger      *log.Logger
}

// NewServer creates a new MCP server instance
func NewServer() (*Server, error) {
	client, err := azure.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure client: %w", err)
	}

	logger := log.New(os.Stderr, "[azure-calc-mcp] ", log.LstdFlags)

	return &Server{
		azureClient: client,
		logger:      logger,
	}, nil
}

// Run listens on stdin and writes to stdout
func (s *Server) Run() error {
	s.logger.Println("Azure Calc MCP Server started on stdio")
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Client disconnected (EOF)")
				return nil
			}
			return fmt.Errorf("error reading stdin: %w", err)
		}

		if len(line) == 0 || (len(line) == 1 && line[0] == '\n') {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.logger.Printf("Failed to unmarshal request: %v\n", err)
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		s.handleRequest(&req)
	}
}

func (s *Server) handleRequest(req *JSONRPCRequest) {
	// Notifications (no ID)
	if req.ID == nil {
		if req.Method == "notifications/initialized" {
			s.logger.Println("Client initialized notification received")
		}
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "ping":
		s.sendResult(req.ID, map[string]interface{}{})
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolCall(req)
	default:
		s.logger.Printf("Method not found: %s\n", req.Method)
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) {
	res := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "azure-calc-mcp",
			"version": "0.1.0",
		},
	}
	s.sendResult(req.ID, res)
}

func (s *Server) handleToolsList(req *JSONRPCRequest) {
	tools := []ToolDefinition{
		{
			Name: "create_azure_estimate",
			Description: "Gera um link oficial compartilhável da Calculadora de Preços do Azure (azure.com/e/...) " +
				"com recursos como AKS (Kubernetes), PostgreSQL Flexible Server, Storage Accounts (Blob), Máquinas Virtuais (VMs), Banco de Dados Azure SQL, App Service, Key Vault, Azure Monitor, etc.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nome descritivo da estimativa (ex: 'Produção Sistema XYZ', 'POC Migração', 'TIM - DEV v1')",
					},
					"currency": map[string]interface{}{
						"type":        "string",
						"description": "Moeda para cálculo ('BRL', 'USD', 'EUR', etc.). Padrão: 'BRL'",
						"default":     "BRL",
					},
					"resources": map[string]interface{}{
						"type":        "array",
						"description": "Lista de recursos do Azure a serem adicionados na calculadora",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"service": map[string]interface{}{
									"type":        "string",
									"description": "Slug do serviço: 'aks' (ou 'kubernetes-service'), 'postgresql' (ou 'flexible-server'), 'storage' (ou 'blob'), 'virtual-machines' (ou 'vm'), 'azure-sql-database' (ou 'sql'), 'app-service', 'key-vault', 'monitor' (ou 'log-analytics')",
								},
								"name": map[string]interface{}{
									"type":        "string",
									"description": "Nome da instância (ex: 'AKS - 2x D2s_v5', 'PostgreSQL B1ms 32GB')",
								},
								"region": map[string]interface{}{
									"type":        "string",
									"description": "Região do Azure (ex: 'brazilsouth', 'brazil-south', 'eastus', 'us-east', 'westeurope'). Padrão: 'brazil-south'",
									"default":     "brazil-south",
								},
								"quantity": map[string]interface{}{
									"type":        "integer",
									"description": "Quantidade de instâncias/nós/servidores. Padrão: 1",
									"default":     1,
								},
								"hours": map[string]interface{}{
									"type":        "integer",
									"description": "Horas por mês (730 para 24/7). Padrão: 730",
									"default":     730,
								},
								"billingOption": map[string]interface{}{
									"type":        "string",
									"description": "Opção de faturamento: 'one-hour' (Pay-as-you-go), 'one-year' (Reservado 1 ano), 'three-year' (Reservado 3 anos), 'savings-plan'",
									"default":     "one-hour",
								},
								"os": map[string]interface{}{
									"type":        "string",
									"description": "Sistema Operacional ('linux' ou 'windows'). Para VMs, AKS e App Service.",
								},
								"sku": map[string]interface{}{
									"type":        "string",
									"description": "Tamanho/SKU da VM ou nó AKS (ex: 'D2s_v5', 'D4s_v5', 'B2s').",
								},
								"diskTier": map[string]interface{}{
									"type":        "string",
									"description": "Tipo de disco gerenciado/OS: 'premiumssd', 'standardssd', 'standardhdd'.",
								},
								"diskSize": map[string]interface{}{
									"type":        "string",
									"description": "Tamanho do disco: 'E10' (128GB Standard SSD), 'P10' (128GB Premium), 'P15', 'P20', etc.",
								},
								"tier": map[string]interface{}{
									"type":        "string",
									"description": "Tier de serviço (ex: 'burstable', 'general-purpose' para PostgreSQL/SQL; 'standard' para AKS).",
								},
								"computeType": map[string]interface{}{
									"type":        "string",
									"description": "SKU de computação do PostgreSQL Flexible Server (ex: 'flexible-server-burstable-compute-b1ms').",
								},
								"computeTier": map[string]interface{}{
									"type":        "string",
									"description": "Tipo de computação SQL: 'provisioned' ou 'serverless'. Apenas para SQL.",
								},
								"vcores": map[string]interface{}{
									"type":        "integer",
									"description": "Quantidade de vCores (ex: 2, 4, 8, 16).",
								},
								"storageGb": map[string]interface{}{
									"type":        "integer",
									"description": "Armazenamento em GB/GiB (para PostgreSQL, SQL ou Blob Storage).",
								},
								"redundancy": map[string]interface{}{
									"type":        "string",
									"description": "Redundância de Storage: 'lrs', 'zrs', 'grs'. Padrão: 'lrs'.",
								},
								"accessTier": map[string]interface{}{
									"type":        "string",
									"description": "Nível de acesso: 'hot', 'cool', 'cold', 'archive'. Apenas para Storage.",
								},
								"operations": map[string]interface{}{
									"type":        "number",
									"description": "Operações do Key Vault em múltiplos de 10.000 (ex: 1.0 para 10k ops).",
								},
								"dailyLogsIngested": map[string]interface{}{
									"type":        "number",
									"description": "Volume diário de logs do Azure Monitor/Log Analytics em GB/dia (ex: 0.1 para 100 MB/dia).",
								},
								"clusterCount": map[string]interface{}{
									"type":        "integer",
									"description": "Número de clusters AKS. Padrão: 1",
								},
								"slaOption": map[string]interface{}{
									"type":        "string",
									"description": "Opção SLA do AKS: 'no-sla-free-non-production' ou 'sla'.",
								},
								"appServicePlan": map[string]interface{}{
									"type":        "string",
									"description": "Plano do App Service (ex: 'B1', 'S1', 'P1v3', 'P2v3'). Apenas para App Service.",
								},
								"extra": map[string]interface{}{
									"type":        "object",
									"description": "Parâmetros adicionais arbitrários a serem injetados diretamente no payload do serviço.",
								},
							},
							"required": []string{"service"},
						},
					},
				},
				"required": []string{"resources"},
			},
		},
		{
			Name: "list_azure_services",
			Description: "Retorna a lista de serviços do Azure com schemas pré-compilados e suporte nativo, " +
				"além de orientações de como usar qualquer um dos 200+ serviços do catálogo do Azure via schema dinâmico.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	s.sendResult(req.ID, map[string]interface{}{"tools": tools})
}

func (s *Server) handleToolCall(req *JSONRPCRequest) {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &callParams); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Invalid params: %v", err))
		return
	}

	switch callParams.Name {
	case "create_azure_estimate":
		s.handleCreateEstimateCall(req, callParams.Arguments)
	case "list_azure_services":
		s.handleListServicesCall(req)
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Unknown tool: %s", callParams.Name))
	}
}

func (s *Server) handleListServicesCall(req *JSONRPCRequest) {
	embedded := azure.ListEmbeddedServices()
	servicesText := "### 🌐 Catálogo de Serviços do Azure MCP\n\n" +
		"O servidor possui **resolução universal de schemas**: ele baixa e monta automaticamente qualquer serviço dos mais de 200 do Azure, com os seguintes pré-compilados e otimizados:\n\n"

	servicesText += "**Serviços com Otimização Instantânea (0ms):**\n"
	for _, s := range embedded {
		servicesText += fmt.Sprintf("- `%s`\n", s)
	}
	servicesText += "- `virtual-machines` (VMs)\n- `azure-sql-database` (SQL)\n\n"
	servicesText += "**Qualquer Outro Serviço do Azure:**\n" +
		"Para usar qualquer outro serviço não listado acima (ex: `databricks`, `synapse-analytics`, `purview`), " +
		"basta informar o slug oficial no campo `service` e enviar os parâmetros específicos via `extra: { ... }`. " +
		"O servidor busca o schema oficial em milissegundos e monta o payload completo automaticamente!"

	s.sendResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": servicesText,
			},
		},
	})
}

func (s *Server) handleCreateEstimateCall(req *JSONRPCRequest, args map[string]interface{}) {
	argBytes, err := json.Marshal(args)
	if err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Failed to marshal arguments: %v", err))
		return
	}

	var opts azure.RequestOptions
	if err := json.Unmarshal(argBytes, &opts); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Failed to parse request options: %v", err))
		return
	}

	s.logger.Printf("Creating estimate '%s' with %d resources...\n", opts.Name, len(opts.Resources))
	result, err := s.azureClient.CreateEstimate(opts)
	if err != nil {
		s.logger.Printf("Error creating estimate: %v\n", err)
		s.sendToolError(req.ID, fmt.Sprintf("Erro ao gerar calculadora Azure: %v", err))
		return
	}

	output := formatEstimateResponse(result, opts)

	s.sendResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output,
			},
		},
	})
}

func formatEstimateResponse(res *azure.EstimateResult, opts azure.RequestOptions) string {
	var sb fmt.Stringer
	str := fmt.Sprintf("### ✅ Estimativa Oficial do Azure Criada com Sucesso!\n\n"+
		"- **Nome da Estimativa:** %s\n"+
		"- **Link Oficial Compartilhável:** [%s](%s)\n"+
		"- **Link Direto da Calculadora:** [Abrir Calculadora](%s)\n"+
		"- **Moeda:** %s\n"+
		"- **Total de Recursos Configurados:** %d\n\n"+
		"#### 📋 Recursos Incluídos:\n",
		res.Name, res.ShareURL, res.ShareURL, res.CalculatorURL, res.Currency, res.ItemCount)

	for i, r := range opts.Resources {
		name := r.Name
		if name == "" {
			name = fmt.Sprintf("Item %d", i+1)
		}
		region := r.Region
		if region == "" {
			region = "eastus"
		}
		qty := r.Quantity
		if qty <= 0 {
			qty = 1
		}

		var details []string
		if r.OS != "" {
			details = append(details, fmt.Sprintf("SO: %s", r.OS))
		}
		if r.SKU != "" {
			details = append(details, fmt.Sprintf("SKU: %s", r.SKU))
		}
		if r.DiskSize != "" || r.DiskTier != "" {
			details = append(details, fmt.Sprintf("Disco: %s %s", r.DiskTier, r.DiskSize))
		}
		if r.Tier != "" {
			details = append(details, fmt.Sprintf("Tier: %s", r.Tier))
		}
		if r.VCores > 0 {
			details = append(details, fmt.Sprintf("vCores: %d", r.VCores))
		}
		if r.StorageGB > 0 {
			details = append(details, fmt.Sprintf("Armazenamento: %d GB", r.StorageGB))
		}
		if r.AppServicePlan != "" {
			details = append(details, fmt.Sprintf("Plano: %s", r.AppServicePlan))
		}

		detailStr := ""
		if len(details) > 0 {
			detailStr = fmt.Sprintf(" (%s)", fmt.Sprint(details))
		}

		str += fmt.Sprintf("%d. **%s** — %s | Qtd: %d | Região: `%s`%s\n",
			i+1, name, r.Service, qty, region, detailStr)
	}

	str += "\n> 💡 *Você pode abrir o link oficial acima no seu navegador para visualizar valores detalhados, ajustar métricas ou exportar para Excel.*"
	_ = sb
	return str
}

func (s *Server) sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeResponse(resp)
}

func (s *Server) sendError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	s.writeResponse(resp)
}

func (s *Server) sendToolError(id interface{}, message string) {
	s.sendResult(id, map[string]interface{}{
		"isError": true,
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	})
}

func (s *Server) writeResponse(resp JSONRPCResponse) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		s.logger.Printf("Failed to marshal response: %v\n", err)
		return
	}
	bytes = append(bytes, '\n')
	os.Stdout.Write(bytes)
}
