# Módulo de Usuarios - Sistema Ferre-POS

## 📋 Descripción General

El módulo de Usuarios proporciona funcionalidades completas para la gestión de usuarios del sistema Ferre-POS. Incluye autenticación, autorización, gestión de perfiles, roles, permisos, recuperación de contraseñas, auditoría de accesos y administración de cuentas.

## 🎯 Funcionalidades Principales

### Gestión de Usuarios

1. **CRUD Completo**
   - Crear, leer, actualizar usuarios
   - Desactivación/reactivación (eliminación lógica)
   - Validaciones exhaustivas de datos
   - Auditoría completa de cambios

2. **Roles y Permisos**
   - **Admin**: Acceso completo al sistema
   - **Gerente**: Gestión de sucursal asignada
   - **Vendedor**: Operaciones de venta y consultas
   - **Cajero**: Operaciones de caja y pagos

3. **Autenticación Segura**
   - Login con RUT y contraseña
   - Bloqueo automático por intentos fallidos
   - Tokens JWT con expiración
   - Registro de intentos de acceso

### Gestión de Contraseñas

1. **Políticas de Seguridad**
   - Mínimo 8 caracteres
   - Al menos una mayúscula, minúscula, número y carácter especial
   - Hash con bcrypt y salt único
   - Forzar cambio en primer login

2. **Recuperación de Contraseñas**
   - Proceso seguro con tokens temporales
   - Tokens con expiración de 1 hora
   - Validación de integridad
   - Limpieza automática de tokens expirados

### Perfiles de Usuario

1. **Gestión de Perfil**
   - Actualización de datos personales
   - Cambio de contraseña personal
   - Visualización de información completa
   - Historial de accesos

2. **Información Completa**
   - Datos personales y contacto
   - Rol y permisos asignados
   - Sucursal de trabajo
   - Estadísticas de actividad

### Auditoría y Seguridad

1. **Registro de Accesos**
   - Intentos exitosos y fallidos
   - Direcciones IP y timestamps
   - Motivos de fallas
   - Filtros y búsquedas avanzadas

2. **Auditoría de Cambios**
   - Registro de todas las modificaciones
   - Trazabilidad completa
   - Identificación de responsables
   - Detalles de cambios realizados

## 🔧 API Endpoints

### Gestión de Usuarios

#### Obtener Lista de Usuarios

```http
GET /api/usuarios?page=1&limit=20&rol=vendedor&activo=true
Authorization: Bearer <token>
```

**Parámetros de consulta:**
- `sucursal_id`: Filtrar por sucursal
- `rol`: admin, gerente, vendedor, cajero
- `activo`: true/false
- `busqueda`: Búsqueda en nombre, email o RUT
- `page`: Número de página (default: 1)
- `limit`: Usuarios por página (default: 20, max: 100)
- `order_by`: nombre, email, rol, fecha_creacion, ultimo_acceso
- `order_direction`: ASC/DESC

**Respuesta:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "rut": "12345678-9",
      "nombre": "Juan Pérez",
      "email": "juan.perez@empresa.com",
      "telefono": "+56912345678",
      "rol": "vendedor",
      "sucursal_id": "uuid",
      "sucursal_nombre": "Sucursal Centro",
      "activo": true,
      "ultimo_acceso": "2024-01-15T10:30:00.000Z",
      "fecha_creacion": "2024-01-01T00:00:00.000Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "totalPages": 3
  }
}
```

#### Obtener Usuario Específico

```http
GET /api/usuarios/{id}
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rut": "12345678-9",
    "nombre": "Juan Pérez",
    "email": "juan.perez@empresa.com",
    "telefono": "+56912345678",
    "rol": "vendedor",
    "sucursal_id": "uuid",
    "sucursal_nombre": "Sucursal Centro",
    "activo": true,
    "ultimo_acceso": "2024-01-15T10:30:00.000Z",
    "debe_cambiar_password": false,
    "fecha_creacion": "2024-01-01T00:00:00.000Z",
    "fecha_modificacion": "2024-01-10T15:20:00.000Z",
    "creador_nombre": "Admin Sistema"
  }
}
```

#### Crear Nuevo Usuario

```http
POST /api/usuarios
Authorization: Bearer <token>
Content-Type: application/json

{
  "rut": "12345678-9",
  "nombre": "Juan Pérez",
  "email": "juan.perez@empresa.com",
  "telefono": "+56912345678",
  "rol": "vendedor",
  "sucursal_id": "uuid",
  "password": "TempPass123!",
  "debe_cambiar_password": true
}
```

#### Actualizar Usuario

```http
PUT /api/usuarios/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "nombre": "Juan Carlos Pérez",
  "telefono": "+56987654321",
  "rol": "gerente"
}
```

#### Desactivar Usuario

```http
POST /api/usuarios/{id}/desactivar
Authorization: Bearer <token>
Content-Type: application/json

{
  "motivo": "El empleado ya no trabaja en la empresa"
}
```

#### Reactivar Usuario

```http
POST /api/usuarios/{id}/reactivar
Authorization: Bearer <token>
```

### Gestión de Perfil

#### Obtener Perfil Propio

```http
GET /api/usuarios/perfil
Authorization: Bearer <token>
```

#### Actualizar Perfil Propio

```http
PUT /api/usuarios/perfil
Authorization: Bearer <token>
Content-Type: application/json

{
  "nombre": "Juan Carlos Pérez",
  "email": "juan.carlos@empresa.com",
  "telefono": "+56987654321"
}
```

### Gestión de Contraseñas

#### Cambiar Contraseña

```http
PUT /api/usuarios/{id}/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "password_actual": "CurrentPass123!",
  "password_nueva": "NewSecurePass123!"
}
```

**Nota:** Los administradores pueden cambiar contraseñas sin proporcionar la contraseña actual.

#### Iniciar Recuperación de Contraseña

```http
POST /api/usuarios/recuperar-password
Content-Type: application/json

{
  "email": "usuario@empresa.com"
}
```

#### Completar Recuperación de Contraseña

```http
POST /api/usuarios/completar-recuperacion
Content-Type: application/json

{
  "token": "token_de_recuperacion_64_caracteres",
  "password_nueva": "NewRecoveredPass123!"
}
```

#### Forzar Cambio de Contraseña

```http
POST /api/usuarios/{id}/forzar-cambio-password
Authorization: Bearer <token>
```

### Administración Avanzada

#### Desbloquear Usuario

```http
POST /api/usuarios/{id}/desbloquear
Authorization: Bearer <token>
```

#### Obtener Historial de Accesos

```http
GET /api/usuarios/{id}/historial-accesos?fecha_inicio=2024-01-01&exitosos=true
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": [
    {
      "fecha": "2024-01-15T10:30:00.000Z",
      "exitoso": true,
      "motivo": "LOGIN_EXITOSO",
      "ip_address": "192.168.1.100"
    },
    {
      "fecha": "2024-01-15T09:45:00.000Z",
      "exitoso": false,
      "motivo": "PASSWORD_INCORRECTA",
      "ip_address": "192.168.1.100"
    }
  ]
}
```

#### Obtener Estadísticas de Usuarios

```http
GET /api/usuarios/estadisticas?sucursal_id=uuid
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "total_usuarios": 45,
    "usuarios_activos": 42,
    "usuarios_inactivos": 3,
    "administradores": 2,
    "gerentes": 5,
    "vendedores": 25,
    "cajeros": 13,
    "activos_hoy": 38,
    "activos_semana": 44
  }
}
```

#### Ejecutar Mantenimiento

```http
POST /api/usuarios/mantenimiento
Authorization: Bearer <token>
```

**Respuesta:**
```json
{
  "success": true,
  "message": "Mantenimiento ejecutado exitosamente",
  "data": {
    "timestamp": "2024-01-15T15:30:00.000Z",
    "tareas_ejecutadas": [
      {
        "tarea": "limpiar_tokens_expirados",
        "resultado": { "tokensLimpiados": 5 }
      },
      {
        "tarea": "desbloquear_usuarios_expirados",
        "resultado": { "usuariosDesbloqueados": 2 }
      }
    ]
  }
}
```

## 🔒 Permisos y Seguridad

### Matriz de Permisos

| Acción | Admin | Gerente | Vendedor | Cajero |
|--------|-------|---------|----------|--------|
| Ver lista usuarios | ✅ | ✅ (sucursal) | ❌ | ❌ |
| Ver usuario específico | ✅ | ✅ (sucursal) | ✅ (propio) | ✅ (propio) |
| Crear usuario | ✅ | ❌ | ❌ | ❌ |
| Actualizar usuario | ✅ | ✅ (sucursal) | ✅ (propio limitado) | ✅ (propio limitado) |
| Desactivar usuario | ✅ | ❌ | ❌ | ❌ |
| Reactivar usuario | ✅ | ❌ | ❌ | ❌ |
| Cambiar contraseña | ✅ | ✅ (propia) | ✅ (propia) | ✅ (propia) |
| Forzar cambio contraseña | ✅ | ❌ | ❌ | ❌ |
| Desbloquear usuario | ✅ | ❌ | ❌ | ❌ |
| Ver historial accesos | ✅ | ✅ (sucursal) | ✅ (propio) | ✅ (propio) |
| Ver estadísticas | ✅ | ✅ (sucursal) | ❌ | ❌ |
| Ejecutar mantenimiento | ✅ | ❌ | ❌ | ❌ |

### Políticas de Seguridad

1. **Autenticación**
   - JWT obligatorio en todas las rutas protegidas
   - Tokens con expiración configurable
   - Refresh tokens para renovación automática

2. **Autorización**
   - Permisos granulares por rol
   - Restricciones por sucursal para gerentes
   - Auto-gestión limitada para usuarios normales

3. **Contraseñas**
   - Política de complejidad estricta
   - Hash con bcrypt y salt único
   - Forzar cambio en primer login
   - Recuperación segura con tokens temporales

4. **Bloqueos de Seguridad**
   - Bloqueo automático tras 5 intentos fallidos
   - Bloqueo temporal de 30 minutos
   - Desbloqueo manual por administradores
   - Limpieza automática de bloqueos expirados

## 📊 Validaciones y Reglas de Negocio

### Validaciones de Datos

#### RUT
- Formato: 12345678-9 o 12345678-K
- Validación de dígito verificador
- Único en el sistema

#### Email
- Formato válido de email
- Único en el sistema
- Máximo 100 caracteres

#### Contraseña
- Mínimo 8 caracteres
- Al menos una letra mayúscula
- Al menos una letra minúscula
- Al menos un número
- Al menos un carácter especial (!@#$%^&*(),.?":{}|<>)

#### Teléfono
- Formato internacional recomendado (+56912345678)
- Máximo 20 caracteres
- Opcional

#### Nombre
- Mínimo 2 caracteres
- Máximo 100 caracteres
- Solo letras, espacios y acentos

### Reglas de Negocio

1. **Creación de Usuarios**
   - Solo administradores pueden crear usuarios
   - Vendedores y cajeros requieren sucursal asignada
   - Contraseña temporal obligatoria
   - Debe cambiar contraseña en primer login

2. **Actualización de Usuarios**
   - Usuarios pueden actualizar su propio perfil (limitado)
   - Gerentes pueden actualizar usuarios de su sucursal
   - Administradores pueden actualizar cualquier usuario
   - No se puede cambiar el RUT

3. **Desactivación**
   - Solo administradores pueden desactivar usuarios
   - No se puede auto-desactivar
   - Requiere motivo obligatorio
   - Eliminación lógica (no física)

4. **Roles y Sucursales**
   - Vendedores y cajeros deben tener sucursal asignada
   - Gerentes pueden gestionar solo su sucursal
   - Administradores tienen acceso global
   - Un usuario solo puede tener un rol

## 🔄 Flujos de Trabajo

### Flujo de Creación de Usuario

1. **Validación**: Admin valida datos del nuevo usuario
2. **Creación**: Sistema crea usuario con contraseña temporal
3. **Notificación**: Usuario recibe credenciales (email/manual)
4. **Primer Login**: Usuario debe cambiar contraseña
5. **Activación**: Usuario queda activo en el sistema

### Flujo de Recuperación de Contraseña

1. **Solicitud**: Usuario solicita recuperación con email
2. **Validación**: Sistema valida email y usuario activo
3. **Token**: Sistema genera token temporal (1 hora)
4. **Notificación**: Usuario recibe link de recuperación
5. **Cambio**: Usuario establece nueva contraseña
6. **Confirmación**: Sistema confirma cambio exitoso

### Flujo de Bloqueo por Intentos Fallidos

1. **Intento Fallido**: Usuario ingresa credenciales incorrectas
2. **Contador**: Sistema incrementa contador de intentos
3. **Bloqueo**: Tras 5 intentos, usuario se bloquea 30 minutos
4. **Notificación**: Sistema registra evento de bloqueo
5. **Desbloqueo**: Automático tras tiempo o manual por admin

### Flujo de Auditoría

1. **Evento**: Ocurre acción significativa en usuario
2. **Registro**: Sistema registra evento con detalles
3. **Metadatos**: Incluye timestamp, usuario responsable, cambios
4. **Almacenamiento**: Evento se guarda en tabla de auditoría
5. **Consulta**: Disponible para reportes y análisis

## 📈 Monitoreo y Métricas

### Métricas Clave

1. **Actividad de Usuarios**
   - Usuarios activos por día/semana/mes
   - Frecuencia de login por usuario
   - Tiempo promedio de sesión
   - Distribución por roles

2. **Seguridad**
   - Intentos de login fallidos
   - Usuarios bloqueados
   - Recuperaciones de contraseña
   - Cambios de contraseña forzados

3. **Administración**
   - Usuarios creados/desactivados por período
   - Cambios de rol y permisos
   - Actividad por sucursal
   - Usuarios inactivos prolongados

### Alertas Automáticas

1. **Seguridad**
   - Múltiples intentos fallidos desde misma IP
   - Recuperaciones de contraseña masivas
   - Logins desde ubicaciones inusuales
   - Cambios de rol no autorizados

2. **Operacionales**
   - Usuarios sin actividad por 30+ días
   - Contraseñas no cambiadas en 90+ días
   - Usuarios sin sucursal asignada
   - Tokens de recuperación no utilizados

## 🛠️ Configuración

### Variables de Entorno

```env
# Configuración de autenticación
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=7d

# Configuración de contraseñas
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBERS=true
PASSWORD_REQUIRE_SPECIAL=true

# Configuración de bloqueos
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION_MINUTES=30
TOKEN_RECOVERY_EXPIRATION_HOURS=1

# Configuración de limpieza
CLEANUP_EXPIRED_TOKENS_HOURS=24
CLEANUP_OLD_ACCESS_LOGS_DAYS=90
```

### Configuración de Base de Datos

El módulo utiliza las siguientes tablas:

- `usuarios`: Información principal de usuarios
- `intentos_acceso`: Registro de intentos de login
- `auditoria_usuarios`: Auditoría de cambios
- `sucursales`: Información de sucursales (referencia)

### Scripts de Inicialización

```sql
-- Crear usuario administrador inicial
INSERT INTO usuarios (rut, nombre, email, rol, password_hash, salt, activo, debe_cambiar_password)
VALUES ('11111111-1', 'Administrador', 'admin@empresa.com', 'admin', 'hash', 'salt', true, true);

-- Crear índices para optimización
CREATE INDEX idx_usuarios_rut ON usuarios(rut);
CREATE INDEX idx_usuarios_email ON usuarios(email);
CREATE INDEX idx_usuarios_activo ON usuarios(activo);
CREATE INDEX idx_intentos_acceso_fecha ON intentos_acceso(fecha);
CREATE INDEX idx_auditoria_usuarios_fecha ON auditoria_usuarios(fecha);
```

## 🧪 Testing

### Ejecutar Tests

```bash
# Tests específicos del módulo
npm test tests/usuarios.test.js

# Tests de integración
npm run test:integration -- --grep "usuarios"

# Coverage del módulo
npm run test:coverage -- tests/usuarios.test.js
```

### Casos de Prueba Incluidos

1. **CRUD de Usuarios**
   - Crear, leer, actualizar usuarios
   - Validaciones de datos
   - Permisos por rol
   - Manejo de errores

2. **Autenticación y Autorización**
   - Login exitoso y fallido
   - Bloqueos por intentos
   - Permisos granulares
   - Tokens JWT

3. **Gestión de Contraseñas**
   - Cambio de contraseña
   - Recuperación segura
   - Validaciones de complejidad
   - Forzar cambios

4. **Perfiles y Configuración**
   - Actualización de perfil
   - Gestión de datos personales
   - Historial de accesos
   - Estadísticas

5. **Administración**
   - Desactivación/reactivación
   - Desbloqueos manuales
   - Mantenimiento automático
   - Auditoría completa

## 🔧 Mantenimiento

### Tareas Programadas Recomendadas

1. **Limpieza Automática**
   - Tokens de recuperación expirados: Cada hora
   - Desbloqueo de usuarios: Cada 5 minutos
   - Logs de acceso antiguos: Semanal
   - Auditoría antigua: Mensual

2. **Monitoreo Continuo**
   - Usuarios inactivos: Diario
   - Intentos de acceso sospechosos: Tiempo real
   - Cambios de permisos: Inmediato
   - Estadísticas de uso: Diario

### Scripts de Utilidad

```bash
# Limpiar tokens expirados
node scripts/cleanup-expired-tokens.js

# Desbloquear usuarios
node scripts/unlock-expired-users.js

# Generar reporte de usuarios inactivos
node scripts/inactive-users-report.js --days 30

# Auditoría de cambios recientes
node scripts/audit-report.js --days 7
```

### Backup y Recuperación

- **Datos críticos**: usuarios, intentos_acceso, auditoria_usuarios
- **Frecuencia**: Backup diario con retención de 30 días
- **Recuperación**: Scripts automatizados para restauración
- **Validación**: Verificación de integridad post-backup

## 📞 Soporte

Para soporte específico del módulo de usuarios:

- **Documentación API**: `/docs/api` en el servidor
- **Logs de autenticación**: `logs/auth.log`
- **Logs de auditoría**: `logs/audit.log`
- **Métricas en vivo**: Dashboard de administración
- **Contacto**: soporte@ferre-pos.cl

---

**Módulo desarrollado para gestión completa de usuarios del sistema Ferre-POS** 👥🔐

