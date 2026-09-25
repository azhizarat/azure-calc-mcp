import subprocess
import json

def run_test():
    proc = subprocess.Popen(
        [r".\azure-calc-mcp.exe"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        text=True,
        cwd=r"C:\Users\Hylex\.gemini\antigravity\scratch\azure-calc-mcp"
    )

    def send_rpc(msg):
        line = json.dumps(msg) + "\n"
        proc.stdin.write(line)
        proc.stdin.flush()
        resp_line = proc.stdout.readline()
        return json.loads(resp_line)

    # 1. Initialize
    init_req = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "clientInfo": {"name": "test-client", "version": "1.0.0"}
        }
    }
    r = send_rpc(init_req)
    print("Initialize:", r.get("result", {}).get("serverInfo"))

    # 2. Call tool create_azure_estimate for the TIM architecture
    tool_req = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/call",
        "params": {
            "name": "create_azure_estimate",
            "arguments": {
                "name": "TIM - AI Champion Education - DEV v1 (Go MCP)",
                "currency": "BRL",
                "resources": [
                    {
                        "service": "aks",
                        "name": "AKS - 2x D2s_v5 (Free tier)",
                        "region": "brazilsouth",
                        "sku": "d2sv5",
                        "quantity": 2,
                        "hours": 730,
                        "os": "linux",
                        "tier": "standard",
                        "diskTier": "standardssd",
                        "diskSize": "e10",
                        "clusterCount": 1,
                        "slaOption": "no-sla-free-non-production"
                    },
                    {
                        "service": "postgresql",
                        "name": "PostgreSQL B1ms 32GB",
                        "region": "brazilsouth",
                        "deploymentType": "flexibleserver",
                        "tier": "burstable",
                        "computeType": "flexible-server-burstable-compute-b1ms",
                        "quantity": 1,
                        "hours": 730,
                        "storageGb": 32
                    },
                    {
                        "service": "storage",
                        "name": "Blob 20 GB",
                        "region": "brazilsouth",
                        "storageGb": 20,
                        "accessTier": "hot",
                        "redundancy": "lrs"
                    },
                    {
                        "service": "key-vault",
                        "name": "Key Vault",
                        "region": "brazilsouth",
                        "operations": 1.0
                    },
                    {
                        "service": "monitor",
                        "name": "Log Analytics 100 MB/dia",
                        "region": "brazilsouth",
                        "dailyLogsIngested": 0.1
                    }
                ]
            }
        }
    }

    call_res = send_rpc(tool_req)
    print("Tool Call Result:")
    text_content = call_res.get("result", {}).get("content", [{}])[0].get("text", "")
    print(text_content)

    proc.terminate()

if __name__ == "__main__":
    run_test()
