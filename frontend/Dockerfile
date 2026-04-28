# Build stage
FROM node:22-alpine AS builder

ARG PUBLIC_API_URL=http://backend:8082

WORKDIR /app

# Create .env file at build time with the arg value (Vite prefix required for Astro)
RUN echo "VITE_PUBLIC_API_URL=${PUBLIC_API_URL}" > .env

# Copy source files
COPY . .

# Install dependencies and build
RUN npm install && npm run build

# Runtime stage
FROM node:22-alpine AS runner

WORKDIR /app

COPY --from=builder /app/src ./src
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/astro.config.mjs ./astro.config.mjs
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package*.json ./

RUN npm ci --omit=dev && npm install http-proxy

EXPOSE 4321

CMD ["node", "src/server/server.cjs"]
