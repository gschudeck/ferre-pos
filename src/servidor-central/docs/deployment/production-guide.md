# Guía de Deployment en Producción

## Introducción

Esta guía proporciona instrucciones detalladas para el deployment del Servidor Central Ferre-POS en un entorno de producción. Incluye configuraciones de seguridad, optimizaciones de rendimiento, monitoreo y mejores prácticas operacionales.

## Requisitos del Sistema

### Hardware Mínimo
- **CPU**: 4 cores (2.4 GHz o superior)
- **RAM**: 8 GB (16 GB recomendado)
- **Almacenamiento**: 100 GB SSD (500 GB recomendado)
- **Red**: 1 Gbps

### Hardware Recomendado para Alta Carga
- **CPU**: 8 cores (3.0 GHz o superior)
- **RAM**: 32 GB
- **Almacenamiento**: 1 TB NVMe SSD
- **Red**: 10 Gbps
- **Backup**: Almacenamiento adicional para backups

### Software
- **Sistema Operativo**: Ubuntu 20.04 LTS o superior / CentOS 8 / RHEL 8
- **Go**: 1.19 o superior
- **PostgreSQL**: 14 o superior
- **Redis**: 6.0 o superior (opcional, para cache)
- **Nginx**: 1.18 o superior (como reverse proxy)
- **Docker**: 20.10 o superior (opcional)

## Preparación del Entorno

### 1. Configuración del Sistema Operativo

#### Actualización del Sistema
```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade -y

# CentOS/RHEL
sudo yum update -y
```

#### Configuración de Límites del Sistema
```bash
# Editar /etc/security/limits.conf
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf
echo "* soft nproc 32768" >> /etc/security/limits.conf
echo "* hard nproc 32768" >> /etc/security/limits.conf
```

#### Configuración de Kernel
```bash
# Editar /etc/sysctl.conf
echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
echo "net.core.netdev_max_backlog = 5000" >> /etc/sysctl.conf
echo "net.ipv4.tcp_fin_timeout = 30" >> /etc/sysctl.conf

# Aplicar cambios
sudo sysctl -p
```

### 2. Instalación de PostgreSQL

#### Instalación
```bash
# Ubuntu/Debian
sudo apt install postgresql-14 postgresql-contrib-14

# CentOS/RHEL
sudo yum install postgresql14-server postgresql14-contrib
sudo postgresql-14-setup initdb
```

#### Configuración de PostgreSQL
```bash
# Editar postgresql.conf
sudo nano /etc/postgresql/14/main/postgresql.conf
```

```ini
# Configuración optimizada para producción
listen_addresses = 'localhost'
port = 5432
max_connections = 200
shared_buffers = 2GB                    # 25% de RAM
effective_cache_size = 6GB              # 75% de RAM
work_mem = 16MB
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1                  # Para SSD
effective_io_concurrency = 200          # Para SSD
min_wal_size = 1GB
max_wal_size = 4GB
```

#### Configuración de Autenticación
```bash
# Editar pg_hba.conf
sudo nano /etc/postgresql/14/main/pg_hba.conf
```

```ini
# Configuración de acceso
local   all             postgres                                peer
local   all             all                                     md5
host    all             all             127.0.0.1/32            md5
host    all             all             ::1/128                 md5
```

#### Creación de Base de Datos y Usuarios
```sql
-- Conectar como postgres
sudo -u postgres psql

-- Crear base de datos
CREATE DATABASE ferre_pos_central;

-- Crear usuarios específicos por API
CREATE USER ferre_pos_user WITH PASSWORD 'secure_password_pos_2024!';
CREATE USER ferre_sync_user WITH PASSWORD 'secure_password_sync_2024!';
CREATE USER ferre_labels_user WITH PASSWORD 'secure_password_labels_2024!';
CREATE USER ferre_reports_user WITH PASSWORD 'secure_password_reports_2024!';

-- Otorgar permisos
GRANT ALL PRIVILEGES ON DATABASE ferre_pos_central TO ferre_pos_user;
GRANT ALL PRIVILEGES ON DATABASE ferre_pos_central TO ferre_sync_user;
GRANT ALL PRIVILEGES ON DATABASE ferre_pos_central TO ferre_labels_user;
GRANT ALL PRIVILEGES ON DATABASE ferre_pos_central TO ferre_reports_user;

-- Configurar esquemas por API (opcional)
\c ferre_pos_central;
CREATE SCHEMA IF NOT EXISTS pos;
CREATE SCHEMA IF NOT EXISTS sync;
CREATE SCHEMA IF NOT EXISTS labels;
CREATE SCHEMA IF NOT EXISTS reports;

GRANT ALL ON SCHEMA pos TO ferre_pos_user;
GRANT ALL ON SCHEMA sync TO ferre_sync_user;
GRANT ALL ON SCHEMA labels TO ferre_labels_user;
GRANT ALL ON SCHEMA reports TO ferre_reports_user;
```

### 3. Instalación de Redis (Opcional)

```bash
# Ubuntu/Debian
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis
```

#### Configuración de Redis
```bash
# Editar /etc/redis/redis.conf
sudo nano /etc/redis/redis.conf
```

```ini
# Configuración de seguridad
bind 127.0.0.1
port 6379
requirepass your_redis_password_2024!
maxmemory 1gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### 4. Instalación de Nginx

```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo yum install nginx
```

#### Configuración de Nginx
```bash
# Crear configuración para Ferre-POS
sudo nano /etc/nginx/sites-available/ferre-pos
```

```nginx
upstream ferre_pos_backend {
    server 127.0.0.1:8080;
    # Para múltiples instancias:
    # server 127.0.0.1:8081;
    # server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name api.ferreteria.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.ferreteria.com;
    
    # SSL Configuration
    ssl_certificate /etc/ssl/certs/ferreteria.com.crt;
    ssl_certificate_key /etc/ssl/private/ferreteria.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # Security Headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    # Client Settings
    client_max_body_size 50M;
    client_body_timeout 60s;
    client_header_timeout 60s;
    
    # Proxy Settings
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
    proxy_buffer_size 4k;
    proxy_buffers 8 4k;
    proxy_busy_buffers_size 8k;
    
    # Logging
    access_log /var/log/nginx/ferre-pos-access.log;
    error_log /var/log/nginx/ferre-pos-error.log;
    
    # API Routes
    location /api/ {
        proxy_pass http://ferre_pos_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS Headers (si no se manejan en la aplicación)
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With";
        
        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With";
            add_header Access-Control-Max-Age 1728000;
            add_header Content-Type 'text/plain charset=UTF-8';
            add_header Content-Length 0;
            return 204;
        }
    }
    
    # Health Check
    location /health {
        proxy_pass http://ferre_pos_backend;
        access_log off;
    }
    
    # Metrics (protegido)
    location /metrics {
        proxy_pass http://ferre_pos_backend;
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        deny all;
    }
    
    # Static Files (si los hay)
    location /static/ {
        alias /var/www/ferre-pos/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

```bash
# Habilitar sitio
sudo ln -s /etc/nginx/sites-available/ferre-pos /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Deployment de la Aplicación

### 1. Preparación del Código

#### Compilación Optimizada
```bash
# Clonar repositorio
git clone https://github.com/tu-organizacion/ferre-pos-servidor-central.git
cd ferre-pos-servidor-central

# Compilar para producción
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags)" \
    -o bin/ferre-pos-server \
    cmd/server/main.go

# Verificar binario
file bin/ferre-pos-server
ldd bin/ferre-pos-server  # Debe mostrar "not a dynamic executable"
```

#### Estructura de Directorios
```bash
# Crear estructura de producción
sudo mkdir -p /opt/ferre-pos/{bin,configs,logs,data,backups}
sudo mkdir -p /var/lib/ferre-pos/{storage,labels,reports,temp}
sudo mkdir -p /var/log/ferre-pos

# Copiar archivos
sudo cp bin/ferre-pos-server /opt/ferre-pos/bin/
sudo cp -r configs/* /opt/ferre-pos/configs/
sudo cp schema/ferre_pos_servidor_central_schema_optimizado.sql /opt/ferre-pos/data/

# Configurar permisos
sudo useradd --system --shell /bin/false --home /opt/ferre-pos ferre-pos
sudo chown -R ferre-pos:ferre-pos /opt/ferre-pos
sudo chown -R ferre-pos:ferre-pos /var/lib/ferre-pos
sudo chown -R ferre-pos:ferre-pos /var/log/ferre-pos
sudo chmod +x /opt/ferre-pos/bin/ferre-pos-server
```

### 2. Configuración de Producción

#### Archivo de Configuración Principal
```bash
sudo nano /opt/ferre-pos/configs/config.yaml
```

```yaml
# Configuración de Producción
server:
  host: "127.0.0.1"
  port: 8080
  mode: "release"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  max_header_bytes: 1048576
  trusted_proxies:
    - "127.0.0.1"
    - "10.0.0.0/8"
  enable_profiling: false
  graceful_timeout: "30s"

database:
  pos:
    host: "localhost"
    port: 5432
    user: "ferre_pos_user"
    password: "secure_password_pos_2024!"
    database: "ferre_pos_central"
    ssl_mode: "require"
    max_open_conns: 50
    max_idle_conns: 10
    conn_max_lifetime: 15
    conn_max_idle_time: 5
    log_level: "error"
    slow_threshold: 100

security:
  jwt_secret: "your-super-secret-jwt-key-production-2024!"
  jwt_expiration: "8h"
  refresh_token_expiration: "168h"
  password_min_length: 12
  password_require_special: true
  max_login_attempts: 3
  login_lockout_duration: "30m"
  enable_two_factor: true
  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst_size: 10

logging:
  global:
    level: "warn"
    format: "json"
    output: "file"
    file_path: "/var/log/ferre-pos/app.log"
    max_size: 100
    max_backups: 10
    max_age: 30
    compress: true
    log_requests: true
    log_responses: false

cache:
  type: "redis"
  host: "localhost"
  port: 6379
  password: "your_redis_password_2024!"
  database: 0
  max_connections: 20
  default_ttl: "1h"

storage:
  type: "local"
  base_path: "/var/lib/ferre-pos/storage"
  max_file_size: 50
  
monitoring:
  enabled: true
  metrics_path: "/metrics"
  health_path: "/health"
  enable_pprof: false
  collect_interval: "30s"
```

#### Variables de Entorno
```bash
sudo nano /opt/ferre-pos/.env
```

```bash
# Variables de Entorno de Producción
FERRE_POS_CONFIG=/opt/ferre-pos/configs/config.yaml
FERRE_POS_ENV=production
FERRE_POS_LOG_LEVEL=warn
GOMAXPROCS=4
GOGC=100
```

### 3. Servicio Systemd

```bash
sudo nano /etc/systemd/system/ferre-pos.service
```

```ini
[Unit]
Description=Ferre-POS Servidor Central
Documentation=https://github.com/tu-organizacion/ferre-pos-servidor-central
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=ferre-pos
Group=ferre-pos
WorkingDirectory=/opt/ferre-pos
ExecStart=/opt/ferre-pos/bin/ferre-pos-server
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=always
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/ferre-pos /var/log/ferre-pos /opt/ferre-pos/configs
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Resource Limits
LimitNOFILE=65536
LimitNPROC=32768
MemoryMax=4G
CPUQuota=400%

# Environment
Environment=FERRE_POS_CONFIG=/opt/ferre-pos/configs/config.yaml
Environment=FERRE_POS_ENV=production
Environment=GOMAXPROCS=4
EnvironmentFile=-/opt/ferre-pos/.env

[Install]
WantedBy=multi-user.target
```

```bash
# Habilitar y iniciar servicio
sudo systemctl daemon-reload
sudo systemctl enable ferre-pos
sudo systemctl start ferre-pos
sudo systemctl status ferre-pos
```

## Configuración de Seguridad

### 1. Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 8080/tcp  # Bloquear acceso directo a la aplicación
sudo ufw enable

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. Certificados SSL

#### Usando Let's Encrypt
```bash
# Instalar Certbot
sudo apt install certbot python3-certbot-nginx

# Obtener certificado
sudo certbot --nginx -d api.ferreteria.com

# Configurar renovación automática
sudo crontab -e
# Agregar: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### Usando Certificados Propios
```bash
# Generar clave privada
sudo openssl genrsa -out /etc/ssl/private/ferreteria.com.key 4096

# Generar CSR
sudo openssl req -new -key /etc/ssl/private/ferreteria.com.key \
    -out /etc/ssl/certs/ferreteria.com.csr

# Instalar certificado firmado
sudo cp ferreteria.com.crt /etc/ssl/certs/
sudo chmod 644 /etc/ssl/certs/ferreteria.com.crt
sudo chmod 600 /etc/ssl/private/ferreteria.com.key
```

### 3. Hardening del Sistema

#### Configuración SSH
```bash
sudo nano /etc/ssh/sshd_config
```

```ini
# Configuración SSH segura
Port 22
Protocol 2
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
MaxAuthTries 3
ClientAliveInterval 300
ClientAliveCountMax 2
```

#### Fail2Ban
```bash
# Instalar Fail2Ban
sudo apt install fail2ban

# Configurar
sudo nano /etc/fail2ban/jail.local
```

```ini
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
```

## Monitoreo y Logging

### 1. Configuración de Logs

#### Logrotate
```bash
sudo nano /etc/logrotate.d/ferre-pos
```

```ini
/var/log/ferre-pos/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 ferre-pos ferre-pos
    postrotate
        systemctl reload ferre-pos
    endscript
}
```

#### Rsyslog (Opcional)
```bash
sudo nano /etc/rsyslog.d/50-ferre-pos.conf
```

```ini
# Ferre-POS Logging
:programname, isequal, "ferre-pos" /var/log/ferre-pos/syslog.log
& stop
```

### 2. Monitoreo con Prometheus

#### Instalación de Prometheus
```bash
# Crear usuario
sudo useradd --no-create-home --shell /bin/false prometheus

# Descargar Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.40.0/prometheus-2.40.0.linux-amd64.tar.gz
tar xvf prometheus-2.40.0.linux-amd64.tar.gz
sudo cp prometheus-2.40.0.linux-amd64/prometheus /usr/local/bin/
sudo cp prometheus-2.40.0.linux-amd64/promtool /usr/local/bin/
sudo chown prometheus:prometheus /usr/local/bin/prometheus
sudo chown prometheus:prometheus /usr/local/bin/promtool

# Crear directorios
sudo mkdir /etc/prometheus
sudo mkdir /var/lib/prometheus
sudo chown prometheus:prometheus /etc/prometheus
sudo chown prometheus:prometheus /var/lib/prometheus
```

#### Configuración de Prometheus
```bash
sudo nano /etc/prometheus/prometheus.yml
```

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "ferre_pos_rules.yml"

scrape_configs:
  - job_name: 'ferre-pos'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
    
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
      
  - job_name: 'postgres-exporter'
    static_configs:
      - targets: ['localhost:9187']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - localhost:9093
```

### 3. Alertas

#### Reglas de Alertas
```bash
sudo nano /etc/prometheus/ferre_pos_rules.yml
```

```yaml
groups:
  - name: ferre_pos_alerts
    rules:
      - alert: FerrePoSDown
        expr: up{job="ferre-pos"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Ferre-POS servidor está down"
          description: "El servidor Ferre-POS no responde por más de 1 minuto"
          
      - alert: HighResponseTime
        expr: http_request_duration_seconds{quantile="0.95"} > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Alto tiempo de respuesta"
          description: "El 95% de las requests toman más de 1 segundo"
          
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Alta tasa de errores"
          description: "Más del 10% de requests retornan errores 5xx"
          
      - alert: DatabaseConnectionsHigh
        expr: postgres_stat_activity_count > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Muchas conexiones a base de datos"
          description: "Más de 80 conexiones activas a PostgreSQL"
```

## Backup y Recuperación

### 1. Backup de Base de Datos

#### Script de Backup
```bash
sudo nano /opt/ferre-pos/scripts/backup-db.sh
```

```bash
#!/bin/bash

# Configuración
DB_NAME="ferre_pos_central"
DB_USER="postgres"
BACKUP_DIR="/opt/ferre-pos/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/ferre_pos_backup_$DATE.sql"
RETENTION_DAYS=30

# Crear directorio si no existe
mkdir -p $BACKUP_DIR

# Realizar backup
pg_dump -U $DB_USER -h localhost $DB_NAME > $BACKUP_FILE

# Comprimir backup
gzip $BACKUP_FILE

# Verificar backup
if [ $? -eq 0 ]; then
    echo "Backup exitoso: $BACKUP_FILE.gz"
    
    # Limpiar backups antiguos
    find $BACKUP_DIR -name "ferre_pos_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete
    
    # Log del backup
    echo "$(date): Backup exitoso - $BACKUP_FILE.gz" >> /var/log/ferre-pos/backup.log
else
    echo "Error en backup"
    echo "$(date): Error en backup" >> /var/log/ferre-pos/backup.log
    exit 1
fi
```

```bash
# Hacer ejecutable
sudo chmod +x /opt/ferre-pos/scripts/backup-db.sh

# Configurar cron para backup automático
sudo crontab -e
# Agregar: 0 2 * * * /opt/ferre-pos/scripts/backup-db.sh
```

### 2. Backup de Archivos

```bash
sudo nano /opt/ferre-pos/scripts/backup-files.sh
```

```bash
#!/bin/bash

# Configuración
SOURCE_DIRS="/opt/ferre-pos/configs /var/lib/ferre-pos"
BACKUP_DIR="/opt/ferre-pos/backups/files"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/ferre_pos_files_$DATE.tar.gz"

# Crear directorio
mkdir -p $BACKUP_DIR

# Realizar backup
tar -czf $BACKUP_FILE $SOURCE_DIRS

# Verificar y limpiar
if [ $? -eq 0 ]; then
    echo "Backup de archivos exitoso: $BACKUP_FILE"
    find $BACKUP_DIR -name "ferre_pos_files_*.tar.gz" -mtime +7 -delete
else
    echo "Error en backup de archivos"
    exit 1
fi
```

### 3. Procedimiento de Recuperación

#### Recuperación de Base de Datos
```bash
# Detener aplicación
sudo systemctl stop ferre-pos

# Restaurar base de datos
gunzip -c /opt/ferre-pos/backups/ferre_pos_backup_YYYYMMDD_HHMMSS.sql.gz | \
    psql -U postgres -h localhost ferre_pos_central

# Iniciar aplicación
sudo systemctl start ferre-pos
```

#### Recuperación de Archivos
```bash
# Extraer backup
tar -xzf /opt/ferre-pos/backups/files/ferre_pos_files_YYYYMMDD_HHMMSS.tar.gz -C /

# Verificar permisos
sudo chown -R ferre-pos:ferre-pos /opt/ferre-pos
sudo chown -R ferre-pos:ferre-pos /var/lib/ferre-pos

# Reiniciar servicio
sudo systemctl restart ferre-pos
```

## Optimización de Rendimiento

### 1. Optimización de Go

#### Variables de Entorno
```bash
# En /opt/ferre-pos/.env
GOMAXPROCS=4                    # Número de CPUs
GOGC=100                        # Frecuencia de GC (default)
GOMEMLIMIT=3GiB                 # Límite de memoria
```

#### Profiling en Producción
```bash
# Habilitar pprof solo cuando sea necesario
# En config.yaml:
monitoring:
  enable_pprof: true  # Solo temporalmente

# Obtener profiles
go tool pprof http://localhost:8080/debug/pprof/profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

### 2. Optimización de PostgreSQL

#### Configuración Avanzada
```sql
-- Configuraciones específicas para Ferre-POS
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET pg_stat_statements.track = 'all';
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- Log queries > 1s
ALTER SYSTEM SET log_checkpoints = on;
ALTER SYSTEM SET log_connections = on;
ALTER SYSTEM SET log_disconnections = on;
ALTER SYSTEM SET log_lock_waits = on;

-- Reiniciar PostgreSQL
SELECT pg_reload_conf();
```

#### Índices Optimizados
```sql
-- Índices específicos para consultas frecuentes
CREATE INDEX CONCURRENTLY idx_productos_activo_categoria 
ON productos(activo, categoria_id) WHERE activo = true;

CREATE INDEX CONCURRENTLY idx_ventas_fecha_sucursal 
ON ventas(fecha, sucursal_id);

CREATE INDEX CONCURRENTLY idx_stock_producto_sucursal 
ON stock_sucursal(producto_id, sucursal_id);

-- Índices para búsqueda de texto
CREATE INDEX CONCURRENTLY idx_productos_search 
ON productos USING gin(to_tsvector('spanish', nombre || ' ' || descripcion));
```

### 3. Monitoreo de Rendimiento

#### Métricas Clave
- **Response Time**: < 200ms para 95% de requests
- **Throughput**: > 1000 requests/segundo
- **Error Rate**: < 0.1%
- **Database Connections**: < 80% del máximo
- **Memory Usage**: < 80% de la disponible
- **CPU Usage**: < 70% promedio

#### Dashboards de Grafana
```json
{
  "dashboard": {
    "title": "Ferre-POS Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

## Troubleshooting

### 1. Problemas Comunes

#### Alto Uso de CPU
```bash
# Verificar procesos
top -p $(pgrep ferre-pos)

# Profile de CPU
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Verificar configuración GOMAXPROCS
cat /proc/$(pgrep ferre-pos)/environ | tr '\0' '\n' | grep GOMAXPROCS
```

#### Alto Uso de Memoria
```bash
# Profile de memoria
go tool pprof http://localhost:8080/debug/pprof/heap

# Verificar límites
systemctl show ferre-pos | grep Memory

# Monitorear GC
curl -s http://localhost:8080/debug/vars | jq '.memstats'
```

#### Problemas de Base de Datos
```sql
-- Verificar conexiones activas
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- Queries lentas
SELECT query, mean_exec_time, calls 
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 10;

-- Locks
SELECT * FROM pg_locks WHERE NOT granted;
```

### 2. Logs de Diagnóstico

#### Habilitar Debug Temporal
```bash
# Cambiar nivel de log temporalmente
sudo nano /opt/ferre-pos/configs/config.yaml
# Cambiar level: "debug"

# Recargar configuración (sin reiniciar)
sudo kill -HUP $(pgrep ferre-pos)

# Monitorear logs
sudo tail -f /var/log/ferre-pos/app.log | jq '.'
```

#### Análisis de Logs
```bash
# Errores más frecuentes
sudo grep -E '"level":"error"' /var/log/ferre-pos/app.log | \
    jq -r '.message' | sort | uniq -c | sort -nr

# Requests más lentas
sudo grep -E '"duration_ms":[0-9]{4,}' /var/log/ferre-pos/app.log | \
    jq -r '"\(.duration_ms)ms \(.method) \(.path)"' | sort -nr

# Análisis de tráfico por hora
sudo grep -E '"timestamp"' /var/log/ferre-pos/app.log | \
    jq -r '.timestamp' | cut -c1-13 | sort | uniq -c
```

## Checklist de Deployment

### Pre-Deployment
- [ ] Servidor configurado con requisitos mínimos
- [ ] PostgreSQL instalado y configurado
- [ ] Nginx instalado y configurado
- [ ] Certificados SSL configurados
- [ ] Firewall configurado
- [ ] Usuario del sistema creado
- [ ] Directorios y permisos configurados

### Deployment
- [ ] Código compilado para producción
- [ ] Archivos copiados a directorios correctos
- [ ] Configuración de producción aplicada
- [ ] Variables de entorno configuradas
- [ ] Servicio systemd creado y habilitado
- [ ] Base de datos migrada
- [ ] Servicio iniciado correctamente

### Post-Deployment
- [ ] Health checks funcionando
- [ ] Métricas siendo recolectadas
- [ ] Logs siendo generados correctamente
- [ ] Backups configurados y probados
- [ ] Monitoreo y alertas configurados
- [ ] Documentación actualizada
- [ ] Equipo notificado del deployment

---

**Autor**: Manus AI  
**Versión**: 1.0.0  
**Fecha**: 2024-01-XX

