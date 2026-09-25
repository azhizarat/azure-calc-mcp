import subprocess
import json

proc = subprocess.Popen(
    [r"C:\Users\Hylex\.gemini\antigravity\scratch\azure-calc-mcp\azure-calc-mcp.exe"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 1. initialize
init_req = {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05"}}
proc.stdin.write(json.dumps(init_req) + "\n")
proc.stdin.flush()
init_res = json.loads(proc.stdout.readline())
print("1. Initialize response:", init_res)

# 2. tools/list
tools_req = {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}
proc.stdin.write(json.dumps(tools_req) + "\n")
proc.stdin.flush()
tools_res = json.loads(proc.stdout.readline())
print("\n2. Tools count:", len(tools_res["result"]["tools"]))
print("   Tool name:", tools_res["result"]["tools"][0]["name"])

# 3. tools/call
call_req = {
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "create_azure_estimate",
        "arguments": {
            "name": "Teste MCP Stdio",
            "currency": "USD",
            "resources": [
                {
                    "service": "vm",
                    "name": "Frontend VM",
                    "region": "eastus",
                    "sku": "B2s",
                    "quantity": 1,
                    "os": "linux"
                }
            ]
        }
    }
}
proc.stdin.write(json.dumps(call_req) + "\n")
proc.stdin.flush()
call_res = json.loads(proc.stdout.readline())
print("\n3. Tool call result:")
print(call_res["result"]["content"][0]["text"])

proc.terminate()
