#!/bin/bash
set -e

# Deploy to remote workers
# Usage: ./scripts/deploy-workers.sh [--binary-only] [--restart-only] [--env KEY=VALUE]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default values
BINARY_ONLY=false
RESTART_ONLY=false
ENV_UPDATES=()

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --binary-only)
      BINARY_ONLY=true
      shift
      ;;
    --restart-only)
      RESTART_ONLY=true
      shift
      ;;
    --env)
      ENV_UPDATES+=("$2")
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--binary-only] [--restart-only] [--env KEY=VALUE]"
      exit 1
      ;;
  esac
done

# Worker configurations
declare -A WORKERS
WORKERS[gcp-flights-worker]="gcp:flights-worker:us-west1-a"
WORKERS[azure-flights-worker]="azure:flights-worker"

# Try SSH via Tailscale first, fallback to cloud CLI
ssh_worker() {
  local name=$1
  local command=$2
  local provider=$3
  local vm_name=$4
  local zone=$5
  
  echo "→ Attempting Tailscale SSH to $name..."
  if timeout 5 ssh -o ConnectTimeout=5 "$name" "$command" 2>/dev/null; then
    echo "✓ $name: Success via Tailscale"
    return 0
  fi
  
  echo "⚠ Tailscale failed, trying $provider CLI..."
  
  case $provider in
    gcp)
      # Try IAP tunnel, which requires temporary SSH access
      echo "→ Using GCP IAP tunnel (requires port 22 open via IAP)"
      gcloud compute ssh "$vm_name" --zone="$zone" --tunnel-through-iap --command="$command"
      ;;
    azure)
      # Use run-command API
      echo "→ Using Azure run-command API"
      local rg=$(az vm list --query "[?name=='$vm_name'].resourceGroup" -o tsv)
      az vm run-command invoke \
        --resource-group "$rg" \
        --name "$vm_name" \
        --command-id RunShellScript \
        --scripts "$command" \
        --query 'value[0].message' -o tsv | sed 's/\\n/\n/g'
      ;;
    *)
      echo "✗ Unknown provider: $provider"
      return 1
      ;;
  esac
}

# Deploy to a single worker
deploy_worker() {
  local worker_key=$1
  local config=${WORKERS[$worker_key]}
  
  IFS=':' read -r provider vm_name zone <<< "$config"
  
  echo ""
  echo "==================================="
  echo "Deploying to: $worker_key"
  echo "Provider: $provider"
  echo "==================================="
  
  # Build command
  local cmd=""
  
  if [[ $RESTART_ONLY == false ]]; then
    # Detect architecture and download binary
    cmd+='ARCH=$(uname -m); '
    cmd+='if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then '
    cmd+='  URL="https://github.com/gilby125/google-flights-api/releases/latest/download/flight-api-linux-arm64"; '
    cmd+='else '
    cmd+='  URL="https://github.com/gilby125/google-flights-api/releases/latest/download/flight-api-linux-amd64"; '
    cmd+='fi; '
    cmd+='echo "Downloading $URL..."; '
    cmd+='curl -fsSL -o /tmp/google-flights-api "$URL"; '
    cmd+='chmod +x /tmp/google-flights-api; '
    
    if [[ $BINARY_ONLY == false ]]; then
      # Update env vars if provided
      for env_update in "${ENV_UPDATES[@]}"; do
        IFS='=' read -r key value <<< "$env_update"
        cmd+="sudo sed -i 's|^${key}=.*|${key}=${value}|' /etc/google-flights/worker.env; "
      done
      
      # Stop, install, start
      cmd+='sudo systemctl stop google-flights-worker@1 || true; '
      cmd+='sudo mv /tmp/google-flights-api /opt/google-flights/google-flights-api; '
      cmd+='sudo chown flights:flights /opt/google-flights/google-flights-api 2>/dev/null || true; '
      cmd+='sudo systemctl start google-flights-worker@1; '
      cmd+='sleep 2; '
      cmd+='sudo systemctl status google-flights-worker@1 --no-pager | head -10'
    else
      # Just install binary
      cmd+='sudo mv /tmp/google-flights-api /opt/google-flights/google-flights-api; '
      cmd+='sudo chown flights:flights /opt/google-flights/google-flights-api 2>/dev/null || true; '
      cmd+='echo "Binary installed. Use --restart-only to restart worker."'
    fi
  else
    # Just restart
    if [[ ${#ENV_UPDATES[@]} -gt 0 ]]; then
      for env_update in "${ENV_UPDATES[@]}"; do
        IFS='=' read -r key value <<< "$env_update"
        cmd+="sudo sed -i 's|^${key}=.*|${key}=${value}|' /etc/google-flights/worker.env; "
      done
    fi
    cmd+='sudo systemctl restart google-flights-worker@1; '
    cmd+='sleep 2; '
    cmd+='sudo systemctl status google-flights-worker@1 --no-pager | head -10'
  fi
  
  # Execute
  ssh_worker "$worker_key" "$cmd" "$provider" "$vm_name" "$zone"
}

# Main execution
echo "Worker Deployment Script"
echo "========================="
echo "Options:"
echo "  Binary only: $BINARY_ONLY"
echo "  Restart only: $RESTART_ONLY"
echo "  Env updates: ${ENV_UPDATES[*]:-none}"

for worker in "${!WORKERS[@]}"; do
  deploy_worker "$worker"
done

echo ""
echo "✓ All workers deployed successfully!"
