# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

COPY . .
RUN npm install
RUN npm run build

# Runtime stage
FROM node:22-alpine AS runner

WORKDIR /app

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package*.json ./

RUN npm ci --omit=dev

EXPOSE 4321

CMD ["npx", "astro", "preview", "--host", "0.0.0.0", "--port", "4321"]
