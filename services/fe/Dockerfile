# Placeholder for Frontend service Dockerfile
# To be implemented in Milestone 4

FROM node:20-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY package.json package-lock.json ./
COPY services/fe services/fe

# Install dependencies
RUN npm ci

# Build the service
WORKDIR /app/services/bff
RUN npm run build

FROM node:20-alpine
RUN npm install -g serve
WORKDIR /app
COPY --from=builder /app/services/fe/dist ./dist

EXPOSE 3000
CMD ["serve", "-s", "dist", "-l", "3000"]
