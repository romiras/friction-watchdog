import subprocess
import json
import threading
import time
import sys

def read_output(process):
    for line in iter(process.stdout.readline, b''):
        print(f"\n[SERVER] {line.decode().strip()}")

def main():
    # Start the server
    process = subprocess.Popen(
        ['./friction-watchdog'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=sys.stderr, # route stderr directly to console
    )

    # Thread to read stdout from server
    reader_thread = threading.Thread(target=read_output, args=(process,), daemon=True)
    reader_thread.start()

    def send_request(req):
        print(f"\n[CLIENT] Sending: {json.dumps(req)}")
        process.stdin.write((json.dumps(req) + '\n').encode())
        process.stdin.flush()
        time.sleep(1) # wait for response

    # 1. Handshake
    send_request({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test-agent", "version": "1.0.0"}
        }
    })

    send_request({
        "jsonrpc": "2.0",
        "method": "notifications/initialized"
    })

    # 2. Check tools
    send_request({
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list"
    })

    # 3. Read prompt
    send_request({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "prompts/get",
        "params": {"name": "coaching_mode"}
    })

    time.sleep(2)
    process.terminate()

if __name__ == "__main__":
    main()
