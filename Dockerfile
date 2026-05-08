FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -S sentra && adduser -S sentra -G sentra
COPY --from=builder --chown=sentra:sentra /app/.next/standalone ./
COPY --from=builder --chown=sentra:sentra /app/.next/static ./.next/static
COPY --from=builder --chown=sentra:sentra /app/public ./public
USER sentra
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
CMD ["node", "server.js"]
