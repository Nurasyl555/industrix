#!/bin/bash
# Healthcheck Script
# Checks if all services are running and healthy
# Usage: ./healthcheck.sh

set -e

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service endpoints to check
declare -A SERVICES=(
    ["Frontend"]="http://localhost:3000"
    ["Gateway"]="http://localhost:8080/health"
    ["PostgreSQL"]="localhost:5432"
    ["Redis"]="localhost:6379"
    ["Kafka"]="localhost:9092"
    ["OpenSearch"]="localhost:9200"
    ["MinIO"]="localhost:9000"
)

# Counters
TOTAL=0
UP=0
DOWN=0

echo "========================================"
echo "  Industrix Health Check"
echo "========================================"
echo ""

check_port() {
    local host=$1
    local port=$2
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

check_http() {
    local url=$1
    if curl -sf --max-time 5 "$url" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

for service in "${!SERVICES[@]}"; do
    TOTAL=$((TOTAL + 1))
    endpoint="${SERVICES[$service]}"
    
    printf "%-20s " "$service"
    
    if [[ $endpoint =~ ^http ]]; then
        # HTTP health check
        if check_http "$endpoint"; then
            echo -e "${GREEN}✓ UP${NC}"
            UP=$((UP + 1))
        else
            echo -e "${RED}✗ DOWN${NC}"
            DOWN=$((DOWN + 1))
        fi
    else
        # Port check
        IFS=':' read -r host port <<< "$endpoint"
        if check_port "$host" "$port"; then
            echo -e "${GREEN}✓ UP${NC}"
            UP=$((UP + 1))
        else
            echo -e "${RED}✗ DOWN${NC}"
            DOWN=$((DOWN + 1))
        fi
    fi
done

echo ""
echo "========================================"
echo "  Summary: $UP/$TOTAL services up"
echo "========================================"

if [ $DOWN -gt 0 ]; then
    echo -e "${RED}Warning: $DOWN service(s) are down!${NC}"
    exit 1
else
    echo -e "${GREEN}All services are healthy!${NC}"
    exit 0
fi
