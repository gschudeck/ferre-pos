# Dockerfile para Sistema Ferre-POS API
# Imagen optimizada para producción con Node.js 18 LTS

# Etapa de construcción
FROM node:18-alpine AS builder

# Información del mantenedor
LABEL maintainer="Manus AI"
LABEL description="API REST para Sistema Ferre-POS"
LABEL version="2.0.0"

# Instalar dependencias del sistema
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    git \
    curl

# Crear directorio de trabajo
WORKDIR /app

# Copiar archivos de configuración
COPY package*.json ./
COPY .npmrc* ./

# Instalar dependencias
RUN npm ci --only=production && npm cache clean --force

# Etapa de producción
FROM node:18-alpine AS production

# Instalar dependencias de runtime
RUN apk add --no-cache \
    dumb-init \
    curl \
    tzdata

# Crear usuario no-root
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Crear directorios necesarios
RUN mkdir -p /app/logs /app/uploads /app/temp && \
    chown -R nodejs:nodejs /app

# Establecer directorio de trabajo
WORKDIR /app

# Copiar dependencias desde builder
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules

# Copiar código fuente
COPY --chown=nodejs:nodejs . .

# Crear directorios de logs y uploads
RUN mkdir -p logs uploads temp && \
    chown -R nodejs:nodejs logs uploads temp

# Configurar timezone
ENV TZ=America/Santiago
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Variables de entorno
ENV NODE_ENV=production
ENV PORT=3000
ENV NODE_OPTIONS="--max-old-space-size=2048"

# Exponer puerto
EXPOSE 3000

# Cambiar a usuario no-root
USER nodejs

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Comando de inicio con dumb-init
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "src/server.js"]

# Metadata
LABEL org.opencontainers.image.title="Ferre-POS API"
LABEL org.opencontainers.image.description="API REST para Sistema de Punto de Venta para Ferreterías"
LABEL org.opencontainers.image.version="2.0.0"
LABEL org.opencontainers.image.authors="Manus AI"
LABEL org.opencontainers.image.source="https://github.com/ferre-pos/api"
LABEL org.opencontainers.image.licenses="MIT"

