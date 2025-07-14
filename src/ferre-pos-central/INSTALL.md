# Guía de Instalación - Ferre-POS API

Esta guía te ayudará a instalar y configurar la API REST del sistema Ferre-POS paso a paso.

## 📋 Requisitos Previos

### Software Requerido

1. **Node.js** (versión 18 o superior)
   ```bash
   # Verificar versión
   node --version
   
   # Si no tienes Node.js, descárgalo desde:
   # https://nodejs.org/
   ```

2. **PostgreSQL** (versión 12 o superior)
   ```bash
   # Verificar versión
   psql --version
   
   # Ubuntu/Debian
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   
   # CentOS/RHEL
   sudo yum install postgresql-server postgresql-contrib
   
   # macOS (con Homebrew)
   brew install postgresql
   ```

3. **Git** (para clonar el repositorio)
   ```bash
   git --version
   ```

## 🚀 Instalación Paso a Paso

### Paso 1: Clonar el Repositorio

```bash
git clone https://github.com/ferre-pos/api.git
cd ferre-pos-api
```

### Paso 2: Instalar Dependencias

```bash
# Instalar dependencias de producción y desarrollo
npm install

# O si prefieres yarn
yarn install
```

### Paso 3: Configurar PostgreSQL

#### 3.1 Crear Usuario y Base de Datos

```bash
# Conectar como usuario postgres
sudo -u postgres psql

# Crear usuario para la aplicación
CREATE USER ferre_pos_app WITH PASSWORD 'tu_password_seguro';

# Crear base de datos
CREATE DATABASE ferre_pos OWNER ferre_pos_app;

# Otorgar permisos
GRANT ALL PRIVILEGES ON DATABASE ferre_pos TO ferre_pos_app;

# Salir de psql
\q
```

#### 3.2 Ejecutar Script de Esquema

```bash
# Ejecutar el script SQL del esquema
psql -h localhost -U ferre_pos_app -d ferre_pos -f ferre_pos_servidor_central_schema.sql
```

### Paso 4: Configurar Variables de Entorno

```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar archivo .env
nano .env
```

**Configuración mínima requerida:**

```env
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ferre_pos
DB_USER=ferre_pos_app
DB_PASSWORD=tu_password_seguro

# JWT (generar una clave segura)
JWT_SECRET=tu_jwt_secret_muy_seguro_y_largo

# Servidor
PORT=3000
NODE_ENV=development
HOST=0.0.0.0
```

### Paso 5: Inicializar Base de Datos

```bash
# Inicializar con datos básicos
node src/utils/dbInit.js init
```

Este comando creará:
- Usuario administrador (RUT: 11111111-1, Password: admin123)
- Sucursal principal
- Terminal principal
- Categorías básicas
- Productos de ejemplo
- Configuraciones del sistema

### Paso 6: Verificar Instalación

```bash
# Iniciar servidor en modo desarrollo
npm run dev

# El servidor debería iniciar en http://localhost:3000
```

**Verificar endpoints:**
- Health Check: http://localhost:3000/health
- Documentación: http://localhost:3000/docs
- Info del Sistema: http://localhost:3000/info

### Paso 7: Probar Autenticación

```bash
# Probar login con usuario administrador
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "rut": "11111111-1",
    "password": "admin123"
  }'
```

## 🔧 Configuración Avanzada

### Configuración de Producción

#### Variables de Entorno Adicionales

```env
# Producción
NODE_ENV=production
LOG_LEVEL=info
LOG_FILE=logs/ferre-pos-api.log

# Seguridad
RATE_LIMIT_MAX=1000
RATE_LIMIT_WINDOW=60000

# Base de datos (producción)
DB_SSL=true
DB_POOL_MIN=5
DB_POOL_MAX=20

# CORS (ajustar según tu frontend)
CORS_ORIGIN=https://tu-frontend.com
CORS_CREDENTIALS=true
```

#### Configuración con PM2

```bash
# Instalar PM2 globalmente
npm install -g pm2

# Crear archivo de configuración PM2
cat > ecosystem.config.js << EOF
module.exports = {
  apps: [{
    name: 'ferre-pos-api',
    script: 'src/server.js',
    instances: 'max',
    exec_mode: 'cluster',
    env: {
      NODE_ENV: 'development'
    },
    env_production: {
      NODE_ENV: 'production',
      PORT: 3000
    }
  }]
}
EOF

# Iniciar en producción
pm2 start ecosystem.config.js --env production

# Configurar inicio automático
pm2 startup
pm2 save
```

### Configuración con Docker

#### Dockerfile

```dockerfile
FROM node:18-alpine

# Crear directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY package*.json ./

# Instalar dependencias
RUN npm ci --only=production

# Copiar código fuente
COPY src/ ./src/

# Crear usuario no-root
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nodejs -u 1001

# Cambiar propietario de archivos
RUN chown -R nodejs:nodejs /app
USER nodejs

# Exponer puerto
EXPOSE 3000

# Comando de inicio
CMD ["npm", "start"]
```

#### Docker Compose

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DB_HOST=postgres
      - DB_NAME=ferre_pos
      - DB_USER=ferre_pos_app
      - DB_PASSWORD=secure_password
      - JWT_SECRET=very_secure_jwt_secret
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:14-alpine
    environment:
      - POSTGRES_DB=ferre_pos
      - POSTGRES_USER=ferre_pos_app
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./ferre_pos_servidor_central_schema.sql:/docker-entrypoint-initdb.d/schema.sql
    restart: unless-stopped

volumes:
  postgres_data:
```

### Configuración de SSL/HTTPS

#### Con Nginx (Recomendado)

```nginx
server {
    listen 80;
    server_name tu-dominio.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name tu-dominio.com;

    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## 🧪 Testing

### Ejecutar Tests

```bash
# Tests unitarios
npm test

# Tests con coverage
npm run test:coverage

# Tests en modo watch
npm run test:watch
```

### Tests de Integración

```bash
# Configurar base de datos de test
createdb ferre_pos_test
psql -d ferre_pos_test -f ferre_pos_servidor_central_schema.sql

# Ejecutar tests de integración
NODE_ENV=test npm test
```

## 📊 Monitoreo

### Logs

```bash
# Ver logs en tiempo real
tail -f logs/ferre-pos-api.log

# Con PM2
pm2 logs ferre-pos-api

# Logs por nivel
grep "ERROR" logs/ferre-pos-api.log
grep "WARN" logs/ferre-pos-api.log
```

### Health Checks

```bash
# Health check básico
curl http://localhost:3000/health

# Health check con detalles
curl http://localhost:3000/health | jq
```

### Métricas con PM2

```bash
# Monitor en tiempo real
pm2 monit

# Estadísticas
pm2 show ferre-pos-api
```

## 🔒 Seguridad

### Configuraciones Recomendadas

1. **Cambiar credenciales por defecto**
   ```bash
   # Cambiar password del admin
   curl -X POST http://localhost:3000/api/auth/reset-password \
     -H "Authorization: Bearer <admin_token>" \
     -H "Content-Type: application/json" \
     -d '{
       "userId": "<admin_user_id>",
       "newPassword": "nueva_password_segura"
     }'
   ```

2. **Configurar HTTPS en producción**
3. **Usar variables de entorno seguras**
4. **Configurar firewall**
5. **Mantener dependencias actualizadas**

### Backup de Base de Datos

```bash
# Backup completo
pg_dump -h localhost -U ferre_pos_app ferre_pos > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup comprimido
pg_dump -h localhost -U ferre_pos_app ferre_pos | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz

# Restaurar backup
psql -h localhost -U ferre_pos_app ferre_pos < backup_file.sql
```

## 🆘 Solución de Problemas

### Problemas Comunes

#### Error de Conexión a Base de Datos

```bash
# Verificar que PostgreSQL esté ejecutándose
sudo systemctl status postgresql

# Verificar conexión
psql -h localhost -U ferre_pos_app -d ferre_pos -c "SELECT 1;"
```

#### Puerto en Uso

```bash
# Verificar qué proceso usa el puerto 3000
lsof -i :3000

# Cambiar puerto en .env
PORT=3001
```

#### Permisos de Base de Datos

```sql
-- Conectar como postgres y otorgar permisos
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ferre_pos_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ferre_pos_app;
```

### Logs de Debug

```env
# Activar logs detallados
LOG_LEVEL=debug
NODE_ENV=development
```

## 📞 Soporte

Si encuentras problemas durante la instalación:

1. Revisa los logs: `logs/ferre-pos-api.log`
2. Verifica la configuración: `.env`
3. Consulta la documentación: http://localhost:3000/docs
4. Contacta soporte: soporte@ferre-pos.cl

---

¡Felicitaciones! Tu API Ferre-POS está lista para usar. 🎉

