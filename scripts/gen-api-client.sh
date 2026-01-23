#!/bin/bash
set -e

# get the directory of the currently currently script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$DIR/.."

echo "🔄 Step 1: Generating Swagger documentation in Backend..."
cd "$PROJECT_ROOT/backend"
make swagger

echo "🔄 Step 2: Updating Shared API specs..."
cd "$PROJECT_ROOT"
# Ensure shared/api exists
mkdir -p shared/api
cp backend/docs/swagger.json shared/api/swagger.json
cp backend/docs/swagger.yaml shared/api/swagger.yaml
echo "   ✅ Copied swagger.json/yaml to shared/api/"

echo "🔄 Step 3: Generating TypeScript client into packages/web/api-client..."
# Generate into ../../packages/web/api-client
# Generate into ../../packages/web/api-client
pushd "$PROJECT_ROOT/frontend"
pnpm gen:api
popd

echo "   ✅ Generated types in packages/web/api-client/"

echo "✅ API utilization workflow complete!"
echo "   - Spec: shared/api/swagger.json"
echo "   - Types: packages/web/api-client/ (@piemdm/api-client)"
echo "   - Frontend uses workspace reference: @piemdm/api-client"
