package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"azure-calc-mcp/azure"
	"azure-calc-mcp/mcp"
)

func main() {
	// Check if install command is requested
	for _, arg := range os.Args[1:] {
		if arg == "install" || arg == "--install" || arg == "-install" {
			runSelfInstall()
			return
		}
	}

	cliMode := flag.Bool("cli", false, "Executa em modo linha de comando (CLI) em vez de servidor MCP")
	testMode := flag.Bool("test", false, "Gera uma estimativa de teste rápida")
	inputFile := flag.String("file", "", "Caminho para arquivo JSON de especificação de arquitetura")
	flag.Parse()

	if *testMode {
		runTest()
		return
	}

	if *cliMode && *inputFile != "" {
		runFromFile(*inputFile)
		return
	}

	// Default: Run as MCP server over stdio
	server, err := mcp.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar servidor MCP: %v\n", err)
		os.Exit(1)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro na execução do servidor MCP: %v\n", err)
		os.Exit(1)
	}
}

func runTest() {
	fmt.Println("🚀 Iniciando teste rápido do Azure Calc MCP...")

	client, err := azure.NewClient()
	if err != nil {
		fmt.Printf("Erro ao criar cliente: %v\n", err)
		os.Exit(1)
	}

	opts := azure.RequestOptions{
		Name:     "Arquitetura Teste Go MCP",
		Currency: "BRL",
		Resources: []azure.ResourceInput{
			{
				Service:       "virtual-machines",
				Name:          "Web Server Cluster",
				Region:        "brazilsouth",
				Quantity:      2,
				OS:            "linux",
				SKU:           "Standard_D4s_v5",
				DiskTier:      "premiumssd",
				DiskSize:      "P15",
				BillingOption: "one-hour",
			},
			{
				Service:     "azure-sql-database",
				Name:        "Banco de Produção",
				Region:      "brazilsouth",
				Tier:        "general-purpose",
				ComputeTier: "provisioned",
				VCores:      4,
				StorageGB:   128,
			},
			{
				Service:        "app-service",
				Name:           "API Gateway App",
				Region:         "brazilsouth",
				OS:             "linux",
				AppServicePlan: "P1v3",
				Quantity:       1,
			},
			{
				Service:    "storage",
				Name:       "Logs and Media Storage",
				Region:     "brazilsouth",
				StorageGB:  1500,
				Redundancy: "lrs",
				AccessTier: "hot",
			},
		},
	}

	fmt.Printf("📦 Enviando %d recursos para a Calculadora do Azure (Moeda: %s)...\n", len(opts.Resources), opts.Currency)
	res, err := client.CreateEstimate(opts)
	if err != nil {
		fmt.Printf("❌ Erro ao criar estimativa: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 Sucesso!")
	fmt.Printf("🔗 Link Oficial Compartilhável: %s\n", res.ShareURL)
	fmt.Printf("🌐 Link Direto da Calculadora: %s\n", res.CalculatorURL)
	fmt.Printf("📝 ID do Snapshot: %s\n", res.ID)
}

func runFromFile(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo %s: %v\n", filepath, err)
		os.Exit(1)
	}

	var opts azure.RequestOptions
	if err := json.Unmarshal(data, &opts); err != nil {
		fmt.Printf("Erro ao decodificar JSON: %v\n", err)
		os.Exit(1)
	}

	client, err := azure.NewClient()
	if err != nil {
		fmt.Printf("Erro ao criar cliente: %v\n", err)
		os.Exit(1)
	}

	res, err := client.CreateEstimate(opts)
	if err != nil {
		fmt.Printf("Erro ao criar estimativa: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 Estimativa gerada com sucesso!")
	fmt.Printf("🔗 Link Oficial: %s\n", res.ShareURL)
	fmt.Printf("🌐 Link Calculadora: %s\n", res.CalculatorURL)
}
