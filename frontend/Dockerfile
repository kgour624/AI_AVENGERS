# ============================================================
# Stage 1: Build
# WHY node:20-alpine: LTS version, alpine = small image
# ============================================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files first for layer caching
# WHY: If only source changes (not deps), npm install is skipped
# npm install works even without package-lock.json
# WHY not npm ci: npm ci requires package-lock.json to exist.
# Once package-lock.json is committed, switch back to npm ci.
COPY package.json package-lock.json* ./
RUN npm install --silent

# Copy source and build
COPY . .

# Use production env (VITE_API_URL=/api — relative, nginx proxies)
COPY .env.production .env

RUN npm run build

# ============================================================
# Stage 2: Serve
# WHY nginx:alpine: ~25MB, serves static files efficiently
# node is not needed at runtime for a built SPA
# ============================================================
FROM nginx:alpine AS runner

# Copy built static files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config for SPA routing
# WHY custom config: default nginx config doesn't handle
# client-side routing (React Router). try_files $uri /index.html
# ensures all routes return index.html so React Router works.
COPY nginx.frontend.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
