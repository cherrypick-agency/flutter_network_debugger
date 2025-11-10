#!/bin/bash
set -e

echo "Running Go tests with coverage..."

# Run tests with coverage (excluding integration and e2e tests)
# Use -short flag to skip long-running load/stress tests
go test -short -covermode=atomic -coverprofile=coverage.out $(go list ./... | grep -v /internal/integration | grep -v /internal/e2e)

# Extract coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
COVERAGE_NUM=$(echo "$COVERAGE" | sed 's/%//')

echo "Total coverage: $COVERAGE"

# Determine badge color based on coverage percentage
if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
    COLOR="brightgreen"
elif (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
    COLOR="green"
elif (( $(echo "$COVERAGE_NUM >= 40" | bc -l) )); then
    COLOR="yellow"
else
    COLOR="red"
fi

# Generate shields.io endpoint JSON
cat > coverage.json << EOF
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "$COVERAGE",
  "color": "$COLOR"
}
EOF

echo "Generated coverage.json with color: $COLOR"
