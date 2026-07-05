FROM oven/bun:1.3.14-alpine AS builder

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

RUN bun run build

FROM nginx:alpine AS runtime

RUN mkdir -p /tmp/nginx /var/cache/nginx \
  && chown -R nginx:nginx /tmp/nginx /var/cache/nginx /usr/share/nginx/html \
  && sed -i 's#pid /var/run/nginx.pid;#pid /tmp/nginx.pid;#' /etc/nginx/nginx.conf

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=nginx:nginx /app/dist /usr/share/nginx/html

EXPOSE 8080

USER nginx

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]