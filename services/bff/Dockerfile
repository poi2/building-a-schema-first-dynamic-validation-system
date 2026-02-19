# Placeholder for BFF service Dockerfile
# To be implemented in Milestone 4

FROM node:20-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY package.json package-lock.json ./
COPY services/bff services/bff

# Install dependencies
RUN npm ci

# Build the service
WORKDIR /app/services/bff
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/services/bff/dist ./dist
COPY --from=builder /app/services/bff/package.json ./

RUN npm ci --production

EXPOSE 3001
CMD ["node", "dist/index.js"]
