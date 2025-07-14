# Ferre-POS API

API REST para Sistema de Punto de Venta especializado en Ferreterías, desarrollado con Node.js y Fastify.

## 🚀 Características

- **Arquitectura Moderna**: Construido con Fastify para máximo rendimiento
- **Base de Datos**: PostgreSQL con pool de conexiones optimizado
- **Autenticación**: JWT con roles y permisos granulares
- **Validación**: Esquemas Joi para validación robusta de datos
- **Logging**: Sistema de logs estructurado con Winston
- **Documentación**: API documentada con Swagger/OpenAPI
- **Seguridad**: Rate limiting, CORS, Helmet y validaciones de entrada
- **Escalabilidad**: Diseño modular y preparado para microservicios

## 📋 Requisitos

- Node.js >= 18.0.0
- PostgreSQL >= 12
- npm >= 8.0.0

## 🛠️ Instalación

### 1. Clonar el repositorio

```bash
git clone https://github.com/ferre-pos/api.git
cd ferre-pos-api
```

### 2. Instalar dependencias

```bash
npm install
```

### 3. Configurar variables de entorno

```bash
cp .env.example .env
```

Editar el archivo `.env` con la configuración de tu entorno:

```env
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ferre_pos
DB_USER=ferre_pos_app
DB_PASSWORD=tu_password_seguro

# JWT
JWT_SECRET=tu_jwt_secret_muy_seguro

# Servidor
PORT=3000
NODE_ENV=development
```

### 4. Configurar base de datos

Ejecutar el script SQL proporcionado para crear la estructura de la base de datos:

```bash
psql -h localhost -U postgres -d ferre_pos -f ferre_pos_servidor_central_schema.sql
```

### 5. Iniciar el servidor

```bash
# Desarrollo
npm run dev

# Producción
npm start
```

## 📚 Documentación de la API

Una vez iniciado el servidor, la documentación interactiva estará disponible en:

- **Swagger UI**: http://localhost:3000/docs
- **Health Check**: http://localhost:3000/health
- **Info del Sistema**: http://localhost:3000/info

## 🏗️ Estructura del Proyecto

```
ferre-pos-api/
├── src/
│   ├── config/           # Configuración del sistema
│   │   ├── index.js      # Configuración principal
│   │   └── database.js   # Configuración de PostgreSQL
│   ├── controllers/      # Controladores de la API
│   │   ├── AuthController.js
│   │   ├── ProductoController.js
│   │   └── ...
│   ├── middleware/       # Middleware personalizado
│   │   └── validation.js # Validaciones con Joi
│   ├── models/           # Modelos de datos
│   │   ├── BaseModel.js  # Modelo base
│   │   ├── Usuario.js
│   │   ├── Producto.js
│   │   ├── Venta.js
│   │   └── ...
│   ├── routes/           # Definición de rutas
│   │   ├── auth.js
│   │   ├── productos.js
│   │   └── ...
│   ├── services/         # Servicios de negocio
│   ├── utils/            # Utilidades
│   │   └── logger.js     # Sistema de logging
│   └── server.js         # Servidor principal
├── tests/                # Tests automatizados
├── docs/                 # Documentación adicional
├── .env.example          # Variables de entorno de ejemplo
├── package.json
└── README.md
```

## 🔐 Autenticación

La API utiliza JWT (JSON Web Tokens) para autenticación. Para acceder a endpoints protegidos:

### 1. Obtener token

```bash
POST /api/auth/login
Content-Type: application/json

{
  "rut": "12345678-9",
  "password": "password123"
}
```

### 2. Usar token en requests

```bash
Authorization: Bearer <tu_jwt_token>
```

## 👥 Roles y Permisos

- **admin**: Acceso completo al sistema
- **gerente**: Gestión de productos, ventas y reportes
- **cajero**: Operaciones de venta y consultas
- **vendedor**: Consultas de productos y clientes

## 📊 Endpoints Principales

### Autenticación
- `POST /api/auth/login` - Iniciar sesión
- `POST /api/auth/logout` - Cerrar sesión
- `POST /api/auth/refresh` - Renovar token
- `GET /api/auth/profile` - Obtener perfil

### Productos
- `GET /api/productos` - Listar productos
- `POST /api/productos` - Crear producto
- `GET /api/productos/:id` - Obtener producto
- `PUT /api/productos/:id` - Actualizar producto
- `DELETE /api/productos/:id` - Eliminar producto
- `GET /api/productos/codigo/:codigo` - Buscar por código

### Ventas
- `POST /api/ventas` - Crear venta
- `GET /api/ventas` - Listar ventas
- `GET /api/ventas/:id` - Obtener venta
- `POST /api/ventas/:id/anular` - Anular venta

### Stock
- `GET /api/stock/producto/:id` - Stock de producto
- `POST /api/stock/movimiento` - Registrar movimiento
- `POST /api/stock/transferencia` - Transferir stock

## 🧪 Testing

```bash
# Ejecutar todos los tests
npm test

# Tests con coverage
npm run test:coverage

# Tests en modo watch
npm run test:watch
```

## 📝 Logging

El sistema utiliza Winston para logging estructurado:

- **Desarrollo**: Logs en consola con colores
- **Producción**: Logs en archivos con rotación automática

Niveles de log:
- `error`: Errores críticos
- `warn`: Advertencias y eventos de seguridad
- `info`: Información general y eventos de negocio
- `debug`: Información detallada para debugging

## 🔧 Scripts Disponibles

```bash
npm start          # Iniciar en producción
npm run dev        # Iniciar en desarrollo con nodemon
npm test           # Ejecutar tests
npm run lint       # Verificar código con ESLint
npm run lint:fix   # Corregir problemas de ESLint automáticamente
```

## 🚀 Despliegue

### Variables de Entorno de Producción

```env
NODE_ENV=production
PORT=3000
HOST=0.0.0.0

# Base de datos
DB_HOST=tu_host_db
DB_PORT=5432
DB_NAME=ferre_pos
DB_USER=ferre_pos_app
DB_PASSWORD=password_super_seguro
DB_SSL=true

# JWT
JWT_SECRET=jwt_secret_muy_seguro_y_largo

# Rate Limiting
RATE_LIMIT_MAX=1000
RATE_LIMIT_WINDOW=60000

# Logging
LOG_LEVEL=info
LOG_FILE=logs/ferre-pos-api.log
```

### Con PM2

```bash
# Instalar PM2
npm install -g pm2

# Iniciar aplicación
pm2 start src/server.js --name "ferre-pos-api"

# Ver logs
pm2 logs ferre-pos-api

# Reiniciar
pm2 restart ferre-pos-api
```

### Con Docker

```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY src/ ./src/

EXPOSE 3000

USER node

CMD ["npm", "start"]
```

## 🔒 Seguridad

- **Rate Limiting**: Protección contra ataques de fuerza bruta
- **CORS**: Configuración de orígenes permitidos
- **Helmet**: Headers de seguridad HTTP
- **Validación**: Sanitización y validación de todas las entradas
- **JWT**: Tokens seguros con expiración
- **Logging**: Auditoría completa de eventos de seguridad

## 📈 Monitoreo

### Health Check

```bash
GET /health
```

Respuesta:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "version": "1.0.0",
  "environment": "production",
  "database": {
    "status": "healthy",
    "connected": true
  },
  "uptime": 3600,
  "memory": {
    "rss": 52428800,
    "heapTotal": 29360128,
    "heapUsed": 20971520
  }
}
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🆘 Soporte

Para soporte técnico:
- Email: soporte@ferre-pos.cl
- Issues: [GitHub Issues](https://github.com/ferre-pos/api/issues)
- Documentación: [Wiki del Proyecto](https://github.com/ferre-pos/api/wiki)

## 🔄 Changelog

### v1.0.0 (2024-01-15)
- ✨ Implementación inicial de la API
- 🔐 Sistema de autenticación con JWT
- 📦 Gestión completa de productos
- 💰 Sistema de ventas y facturación
- 📊 Gestión de stock e inventario
- 🎯 Sistema de fidelización
- 📈 Reportes y estadísticas
- 📚 Documentación completa con Swagger

---

**Desarrollado con ❤️ para el ecosistema de ferreterías**

