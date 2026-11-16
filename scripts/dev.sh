#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

if [[ ! -f .env ]]; then
	echo 'Missing .env. Copy .env.example to .env and configure it first.' >&2
	exit 1
fi
set -a
source .env
set +a

for command in go npm setsid; do
	command -v "$command" >/dev/null || { echo "Missing required command: $command" >&2; exit 1; }
done

# PostgreSQL server binaries are not always on PATH (for example, on Ubuntu).
if ! command -v pg_ctl >/dev/null && command -v pg_config >/dev/null; then
	export PATH="$(pg_config --bindir):$PATH"
fi
for command in initdb pg_ctl psql createdb; do
	command -v "$command" >/dev/null || { echo "Install PostgreSQL server binaries; missing: $command" >&2; exit 1; }
done

: "${POSTGRES_USER:?Set POSTGRES_USER in .env}"
: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in .env}"
: "${POSTGRES_DB:?Set POSTGRES_DB in .env}"
export DEBUG=true
export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT="${DEV_POSTGRES_PORT:-28473}"
export BACKEND_PORT="${BACKEND_PORT:-28472}"
export FRONTEND_PORT="${FRONTEND_PORT:-28471}"
export BACKEND_DOMAIN="http://127.0.0.1:$BACKEND_PORT"
export PGPASSWORD="$POSTGRES_PASSWORD"

data="$root/.dev/postgres"
mkdir -p "$root/.dev"
chmod 700 "$root/.dev"

pids=()
started_postgres=false
cleanup() {
	trap - EXIT INT TERM
	for pid in "${pids[@]}"; do
		kill -TERM -- "-$pid" 2>/dev/null || true
	done
	for pid in "${pids[@]}"; do
		wait "$pid" 2>/dev/null || true
	done
	if "$started_postgres"; then
		pg_ctl -D "$data" -m fast -w stop
	fi
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if [[ ! -f "$data/PG_VERSION" ]]; then
	initdb -D "$data" -U "$POSTGRES_USER" --auth-local=trust --auth-host=scram-sha-256 \
		--pwfile=<(printf '%s\n' "$POSTGRES_PASSWORD")
fi
if ! pg_ctl -D "$data" status >/dev/null 2>&1; then
	pg_ctl -D "$data" -l "$root/.dev/postgres.log" \
		-o "-h 127.0.0.1 -p $POSTGRES_PORT -k '$root/.dev'" -w start
	started_postgres=true
fi

if [[ "$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
	-v ON_ERROR_STOP=1 -v db="$POSTGRES_DB" -At <<'SQL'
SELECT 1 FROM pg_database WHERE datname = :'db';
SQL
)" != 1 ]]; then
	createdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$POSTGRES_DB"
fi

cd "$root/frontend"
if [[ ! -f node_modules/.package-lock.json || package-lock.json -nt node_modules/.package-lock.json || package.json -nt node_modules/.package-lock.json ]]; then
	npm ci
fi

cd "$root/backend"
setsid go run github.com/air-verse/air@v1.61.7 &
pids+=("$!")
cd "$root/frontend"
setsid npm run dev -- --port "$FRONTEND_PORT" --strictPort &
pids+=("$!")

echo "Local development: frontend http://localhost:$FRONTEND_PORT, backend $BACKEND_DOMAIN, PostgreSQL $POSTGRES_HOST:$POSTGRES_PORT"
wait -n "${pids[@]}"
