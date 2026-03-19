#!/usr/bin/env bash
set -euo pipefail

TOTAL_REQUESTS="${TOTAL_REQUESTS:-20}"
USER_ID="${USER_ID:-loadtest-user-42}"
USER_HEADER_NAME="${USER_HEADER_NAME:-X-User-ID}"
NODE_A_URL="${NODE_A_URL:-http://localhost:8080/}"
NODE_B_URL="${NODE_B_URL:-http://localhost:8081/}"
METHOD="${METHOD:-GET}"

echo "Sending ${TOTAL_REQUESTS} requests alternating between:"
echo "  A: ${NODE_A_URL}"
echo "  B: ${NODE_B_URL}"
echo "Header: ${USER_HEADER_NAME}: ${USER_ID}"
echo

for i in $(seq 1 "${TOTAL_REQUESTS}"); do
  if (( i % 2 == 1 )); then
    target_url="${NODE_A_URL}"
    node_label="A"
  else
    target_url="${NODE_B_URL}"
    node_label="B"
  fi

  printf "[%02d/%02d] node=%s url=%s ... " "${i}" "${TOTAL_REQUESTS}" "${node_label}" "${target_url}"

  http_code="$(
    curl -sS -o /dev/null -w "%{http_code}" \
      -X "${METHOD}" \
      -H "${USER_HEADER_NAME}: ${USER_ID}" \
      -H "Content-Type: application/json" \
      --data "{\"request\":${i},\"user_id\":\"${USER_ID}\"}" \
      "${target_url}"
  )"

  echo "status=${http_code}"
done

echo
echo "Done. Sent ${TOTAL_REQUESTS} alternating requests with ${USER_HEADER_NAME}: ${USER_ID}."