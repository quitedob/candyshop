# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application
ARG BACKEND_URL=http://api:8080
ENV BACKEND_URL=${BACKEND_URL}
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# dumb-init 用于正确转发 SIGTERM
RUN apk --no-cache add dumb-init wget

# R4-19: Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy built application
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./

RUN chown -R appuser:appgroup /app
USER appuser

# Expose port
EXPOSE 3000

# Set environment
ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start the application (dumb-init forwards signals)
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", ".output/server/index.mjs"]
