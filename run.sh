CONTAINER_NAME="customer_app"
HOST_PORT=8080
CONTAINER_PORT=8080
ENV_FILE=".env"           # optional env file to pass into container
PULL_IMAGE=false          # set to true to docker pull before running (still does NOT use Dockerfile)
RESTART_POLICY="unless-stopped"
MEMORY_RESERVATION="128m"  # memory reservation (soft limit)
MEMORY_LIMIT="256m"        # memory limit (hard limit)
# ------------------------------------------

log() { printf "%s %s\n" "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$*"; }

# Optionally pull latest image from registry (does not rebuild)
if [ "${PULL_IMAGE}" = true ]; then
  log "Pulling image ${IMAGE} ..."
  docker pull "${IMAGE}"
fi

# Stop & remove existing container if present
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  log "Stopping and removing existing container ${CONTAINER_NAME} ..."
  docker rm -f "${CONTAINER_NAME}" || true
fi

# Build run command
RUN_CMD=(docker run --platform linux/amd64 -d --name "${CONTAINER_NAME}" -p "${HOST_PORT}:${CONTAINER_PORT}" --restart "${RESTART_POLICY}")

if [ -f "${ENV_FILE}" ]; then
  log "Using env file ${ENV_FILE}"
  RUN_CMD+=(--env-file "${ENV_FILE}")
fi

# Add memory reservation and limit
RUN_CMD+=(--memory-reservation "${MEMORY_RESERVATION}" --memory "${MEMORY_LIMIT}")

RUN_CMD+=("${IMAGE}")

log "Running container: ${RUN_CMD[*]}"
"${RUN_CMD[@]}"

log "Container ${CONTAINER_NAME} started and mapped to host port ${HOST_PORT}"
