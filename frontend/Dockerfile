# syntax=docker/dockerfile:1

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

FROM node:22-alpine AS builder
WORKDIR /app
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npx next build \
    && rm -rf \
        .next/standalone/node_modules/typescript \
        .next/standalone/node_modules/caniuse-lite \
        .next/standalone/node_modules/sharp \
        .next/standalone/node_modules/@img

FROM alpine:3.22 AS runner
WORKDIR /app
ENV NODE_ENV=production HOSTNAME=0.0.0.0 PORT=3001 NEXT_TELEMETRY_DISABLED=1
RUN apk add --no-cache ca-certificates libstdc++ \
    && addgroup -S -g 10001 clubmax \
    && adduser -S -D -H -u 10001 -G clubmax clubmax
COPY --from=builder /usr/local/bin/node /usr/local/bin/node
COPY --from=builder --chown=clubmax:clubmax /app/public ./public
COPY --from=builder --chown=clubmax:clubmax /app/.next/standalone ./
COPY --from=builder --chown=clubmax:clubmax /app/.next/static ./.next/static
USER clubmax
EXPOSE 3001
CMD ["node", "server.js"]
