# Multi-stage build for Vue.js application
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Generate API Client (Pre-install to satisfy workspace dependencies)
# We use npx to generate the client before pnpm install because pnpm install
# requires the workspace package (api-client) to exist with a package.json.
COPY shared/api/swagger.json ./shared/api/swagger.json
RUN npx @openapitools/openapi-generator-cli generate \
    -i shared/api/swagger.json \
    -g typescript-axios \
    -o packages/web/api-client \
    --additional-properties=supportsES6=true,npmName=@piemdm/api-client,npmVersion=1.0.0,withSeparateModelsAndApi=true,apiPackage=api,modelPackage=models,useSingleRequestParameter=false,skipValidateSpec=true


# Copy workspace configs
COPY pnpm-lock.yaml pnpm-workspace.yaml ./

# Copy package definitions
COPY frontend/package.json ./frontend/
COPY packages/web/workflow-vue/package.json ./packages/web/workflow-vue/

# Install dependencies
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# Copy source code
COPY frontend/ ./frontend/
COPY packages/web/workflow-vue/ ./packages/web/workflow-vue/

# Build api-client (Generated code needs build?)
# typescript-axios generator produces .ts files and a tsconfig.
# Usually needs 'tsc' to build to dist.
# The generated package.json usually has a build script?
# Let's assume we need to build it.
WORKDIR /app/packages/web/api-client
RUN pnpm install && pnpm build

# Build workflow-vue
WORKDIR /app/packages/web/workflow-vue
RUN pnpm build

# Build frontend
WORKDIR /app/frontend
RUN pnpm build

# Production stage
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/frontend/dist /usr/share/nginx/html

# Copy nginx configuration
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Create nginx user and set permissions
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
