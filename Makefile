SHELL := /bin/sh

.PHONY: tui-demo

tui-demo:
	@set -eu; \
	LIVE_LOG=$$(mktemp -t specter-live.XXXXXX.log); \
	SHADOW_LOG=$$(mktemp -t specter-shadow.XXXXXX.log); \
	PROXY_LOG=$$(mktemp -t specter-proxy.XXXXXX.log); \
	echo "Starting live server (:3000)..."; \
	go run ./cmd/testserver --port 3000 --mode live >"$$LIVE_LOG" 2>&1 & LIVE_PID=$$!; \
	echo "Starting shadow server (:3001)..."; \
	go run ./cmd/testserver --port 3001 --mode shadow >"$$SHADOW_LOG" 2>&1 & SHADOW_PID=$$!; \
	echo "Starting Specter proxy (:8080)..."; \
	go run ./cmd/specter --config internal/config/specter.yaml --ui proxy >"$$PROXY_LOG" 2>&1 & PROXY_PID=$$!; \
	cleanup() { \
		echo ""; \
		echo "Stopping demo processes..."; \
		kill $$PROXY_PID $$SHADOW_PID $$LIVE_PID 2>/dev/null || true; \
		wait $$PROXY_PID $$SHADOW_PID $$LIVE_PID 2>/dev/null || true; \
		echo "Logs:"; \
		echo "  live:   $$LIVE_LOG"; \
		echo "  shadow: $$SHADOW_LOG"; \
		echo "  proxy:  $$PROXY_LOG"; \
	}; \
	trap cleanup INT TERM EXIT; \
	sleep 1; \
	echo "Generating sample traffic..."; \
	for i in $$(seq 1 20); do \
		curl -s -H "X-User-ID: user-$$i" http://127.0.0.1:8080/profile >/dev/null || true; \
	done; \
	echo "Launching TUI (press q to quit)..."; \
	TERM=xterm-256color go run ./cmd/specter --config internal/config/specter.yaml --ui tui
