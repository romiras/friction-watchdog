import subprocess
import json
import threading
import time
import sys
import urllib.request
import urllib.parse

# ---------------------------------------------------------------
# Friction Watchdog — E2E Test Client
# Simulates an AI agent: handshake, tools check, prompt check,
# and the Phase 3 causal scenario (error loop → coaching prompt).
# ---------------------------------------------------------------

received_messages = []
notification_event = threading.Event()


def read_output(process):
    """Background reader: captures server stdout JSON-RPC messages."""
    for line in iter(process.stdout.readline, b""):
        text = line.decode().strip()
        if not text:
            continue
        print(f"\n[SERVER] {text}")
        try:
            msg = json.loads(text)
            received_messages.append(msg)
            # Detect resource-updated notification
            if msg.get("method") == "notifications/resources/updated":
                notification_event.set()
        except json.JSONDecodeError:
            pass


def send_rpc(process, req):
    print(f"\n[CLIENT] Sending: {json.dumps(req)}")
    process.stdin.write((json.dumps(req) + "\n").encode())
    process.stdin.flush()
    time.sleep(0.5)


def post_terminal_event(exit_code: int, command: str = ""):
    """POST a terminal event to the watchdog HTTP sensor."""
    data = urllib.parse.urlencode({"exit_code": str(exit_code), "command": command}).encode()
    try:
        req = urllib.request.Request("http://localhost:8080/terminal-event", data=data, method="POST")
        urllib.request.urlopen(req, timeout=1)
        print(f"[CLIENT] POST terminal-event exit_code={exit_code} command={command!r}")
    except Exception as e:
        print(f"[CLIENT] POST failed: {e}")


def main():
    process = subprocess.Popen(
        ["./friction-watchdog"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=sys.stderr,
    )

    reader_thread = threading.Thread(target=read_output, args=(process,), daemon=True)
    reader_thread.start()

    # Allow server to start
    time.sleep(0.5)

    # ── 1. Handshake ────────────────────────────────────────────
    send_rpc(process, {
        "jsonrpc": "2.0", "id": 1, "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test-agent", "version": "1.0.0"},
        },
    })
    send_rpc(process, {"jsonrpc": "2.0", "method": "notifications/initialized"})

    # ── 2. Happy-path checks ─────────────────────────────────────
    send_rpc(process, {"jsonrpc": "2.0", "id": 2, "method": "tools/list"})
    send_rpc(process, {"jsonrpc": "2.0", "id": 3, "method": "prompts/get",
                       "params": {"name": "coaching_mode"}})

    # ── 3. Phase 3: error-loop causal scenario ───────────────────
    print("\n[SCENARIO] Posting 3× failures with command='go test ./...'")
    for _ in range(3):
        post_terminal_event(exit_code=1, command="go test ./...")
        time.sleep(0.2)

    # Wait up to 15 s for the watchdog idle-checker to fire
    print("[SCENARIO] Waiting for notifications/resources/updated ...")
    fired = notification_event.wait(timeout=15)
    if not fired:
        print("[FAIL] Notification never arrived within 15 s")
        process.terminate()
        sys.exit(1)
    print("[SCENARIO] Notification received. Calling prompts/get ...")

    # Capture the next message id to identify our response
    next_id = 10
    send_rpc(process, {"jsonrpc": "2.0", "id": next_id, "method": "prompts/get",
                       "params": {"name": "coaching_mode"}})
    time.sleep(1)

    # ── 4. Assert ────────────────────────────────────────────────
    causal_response = next(
        (m for m in received_messages if m.get("id") == next_id), None
    )
    if causal_response is None:
        print("[FAIL] No response for prompts/get after notification")
        process.terminate()
        sys.exit(1)

    prompt_text = ""
    try:
        prompt_text = causal_response["result"]["messages"][0]["content"]["text"]
    except (KeyError, IndexError, TypeError) as e:
        print(f"[FAIL] Could not extract prompt text: {e}")
        process.terminate()
        sys.exit(1)

    print(f"\n[ASSERT] Prompt text:\n{prompt_text}\n")

    failed = False
    if "error_loop" not in prompt_text:
        print('[FAIL] prompt text does not contain "error_loop"')
        failed = True
    if "go test ./..." not in prompt_text:
        print('[FAIL] prompt text does not contain "go test ./..."')
        failed = True

    if not failed:
        print("[PASS] All assertions passed.")
    else:
        process.terminate()
        sys.exit(1)

    time.sleep(1)
    process.terminate()


if __name__ == "__main__":
    main()
