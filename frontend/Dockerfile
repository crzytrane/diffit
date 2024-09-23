# Build stage
FROM oven/bun AS builder

WORKDIR /app

COPY bun.lockb package.json ./
RUN bun install --frozen-lockfile

# Copy the source code and build the application
COPY . ./
RUN bun run build

FROM nginx:1.25.4-alpine-slim AS production

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 9000

CMD ["nginx", "-g", "daemon off;"]

