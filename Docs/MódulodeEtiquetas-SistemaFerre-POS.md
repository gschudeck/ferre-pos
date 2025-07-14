# Módulo de Etiquetas - Sistema Ferre-POS

## Descripción General

El Módulo de Etiquetas es un componente especializado del sistema Ferre-POS diseñado para confeccionar e imprimir etiquetas de precio profesionales con códigos de barras Code 39. Este módulo proporciona una solución integral para la gestión visual del inventario en ferreterías, facilitando la identificación de productos y mejorando la eficiencia operativa.

## Características Principales

### ✅ Funcionalidades Core
- **Búsqueda Avanzada de Productos**: Localización por código, descripción, marca o categoría
- **Generación de Códigos de Barras Code 39**: Cumplimiento con estándares industriales
- **Sistema de Plantillas Flexible**: Formatos predefinidos y personalizables
- **Vista Previa en Tiempo Real**: Verificación antes de impresión
- **Impresión Individual y Masiva**: Adaptable a diferentes volúmenes de trabajo
- **Gestión de Trabajos**: Historial y seguimiento de operaciones

### 🔧 Características Técnicas
- **Backend**: Flask con SQLAlchemy ORM
- **Frontend**: React con shadcn/ui y Tailwind CSS
- **Base de Datos**: PostgreSQL con esquema optimizado
- **APIs**: RESTful con documentación OpenAPI
- **Integración**: Seamless con sistema Ferre-POS principal
- **Seguridad**: Autenticación heredada y auditoría completa

### 📋 Información de Etiquetas
Cada etiqueta incluye:
- Código de producto
- Fabricante/Marca
- Modelo (cuando disponible)
- Descripción del producto
- Precio de venta
- Código de barras Code 39

## Estructura del Proyecto

```
modulo_etiquetas/
├── documentacion/
│   ├── modulo_etiquetas_especificacion.md    # Especificación técnica completa
│   └── README_Esquema_SQL.md                 # Documentación de base de datos
├── backend/
│   ├── src/
│   │   ├── models/                           # Modelos de datos
│   │   ├── routes/                           # APIs REST
│   │   └── main.py                           # Aplicación principal
│   └── requirements.txt                      # Dependencias Python
├── frontend/
│   ├── src/
│   │   ├── components/                       # Componentes React
│   │   └── App.jsx                           # Aplicación principal
│   └── package.json                          # Dependencias Node.js
└── database/
    ├── modulo_etiquetas_schema.sql           # Esquema de base de datos
    └── datos_iniciales.sql                   # Datos de ejemplo
```

## Instalación y Configuración

### Prerrequisitos
- Python 3.11+
- Node.js 20+
- PostgreSQL 14+
- Sistema Ferre-POS principal configurado

### 1. Configuración de Base de Datos

```sql
-- Ejecutar en PostgreSQL
\i modulo_etiquetas_schema.sql
```

### 2. Instalación del Backend

```bash
cd modulo_etiquetas_backend
python -m venv venv
source venv/bin/activate  # Linux/Mac
# o venv\Scripts\activate  # Windows
pip install -r requirements.txt
```

### 3. Configuración del Backend

Crear archivo `.env`:
```env
DATABASE_URL=postgresql://usuario:password@localhost/ferre_pos
SECRET_KEY=tu_clave_secreta_aqui
CORS_ORIGINS=http://localhost:3000
```

### 4. Instalación del Frontend

```bash
cd modulo_etiquetas_frontend
npm install
# o pnpm install
```

### 5. Ejecución en Desarrollo

**Backend:**
```bash
cd modulo_etiquetas_backend
source venv/bin/activate
python src/main.py
# Servidor en http://localhost:5000
```

**Frontend:**
```bash
cd modulo_etiquetas_frontend
npm run dev
# Aplicación en http://localhost:3000
```

## Uso del Sistema

### Generación de Etiquetas Individuales

1. **Buscar Producto**: Ingresa código o descripción en el campo de búsqueda
2. **Seleccionar Producto**: Haz clic en el producto deseado de los resultados
3. **Configurar Etiqueta**: Selecciona plantilla y cantidad de etiquetas
4. **Vista Previa**: Verifica el diseño antes de imprimir
5. **Imprimir**: Confirma la generación de etiquetas

### Generación Masiva de Etiquetas

1. **Filtrar Productos**: Utiliza criterios de categoría, marca o rango de precios
2. **Seleccionar Lote**: Marca los productos que requieren etiquetas
3. **Configurar Plantilla**: Selecciona formato apropiado para el lote
4. **Procesar**: Ejecuta la generación masiva con seguimiento de progreso

### Gestión de Plantillas

1. **Acceder a Plantillas**: Navega a la pestaña "Plantillas"
2. **Revisar Disponibles**: Visualiza plantillas existentes y sus características
3. **Configurar Predeterminadas**: Establece plantillas por defecto por categoría
4. **Monitorear Uso**: Revisa estadísticas de utilización

## APIs Disponibles

### Endpoints Principales

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/etiquetas/productos/buscar` | Buscar productos |
| POST | `/api/etiquetas/codigos-barras/generar` | Generar código de barras |
| GET | `/api/etiquetas/plantillas` | Listar plantillas |
| POST | `/api/etiquetas/vista-previa` | Generar vista previa |
| POST | `/api/etiquetas/imprimir` | Procesar impresión |
| GET | `/api/etiquetas/trabajos` | Historial de trabajos |

### Ejemplo de Uso de API

```javascript
// Buscar productos
const response = await fetch('/api/etiquetas/productos/buscar?q=martillo');
const data = await response.json();

// Generar etiquetas
const printJob = await fetch('/api/etiquetas/imprimir', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    productos_ids: ['1', '2', '3'],
    plantilla_id: 'plantilla-mediana',
    cantidad_por_producto: 2
  })
});
```

## Configuración de Impresoras

### Impresoras Térmicas Soportadas
- Zebra GK420d, GX420d, GX430t
- Citizen CL-S521, CL-S621, CL-S700
- TSC TTP-244CE, TTP-344M
- Honeywell PC42t, PC23d

### Configuración Manual
1. Acceder a configuraciones de impresora
2. Agregar nueva configuración con parámetros específicos
3. Configurar resolución, velocidad y tipo de papel
4. Probar conectividad y calidad de impresión

## Solución de Problemas

### Problemas Comunes

**Error: "Producto no encontrado"**
- Verificar que el producto existe en el sistema principal
- Comprobar sincronización de datos
- Revisar permisos de acceso

**Error: "Código de barras inválido"**
- Verificar que el código cumple con formato Code 39
- Eliminar caracteres especiales no soportados
- Verificar longitud máxima (43 caracteres)

**Error: "Impresora no responde"**
- Verificar conectividad física
- Comprobar configuración de driver
- Revisar estado de la impresora

### Logs y Diagnóstico

Los logs del sistema se encuentran en:
- Backend: `logs/etiquetas.log`
- Base de datos: Tabla `etiquetas_logs_operaciones`
- Frontend: Consola del navegador

## Mantenimiento

### Tareas Regulares
- **Limpieza de Cache**: Ejecutar limpieza de códigos de barras antiguos
- **Backup de Plantillas**: Respaldar configuraciones personalizadas
- **Monitoreo de Uso**: Revisar métricas de rendimiento
- **Actualización de Datos**: Sincronizar con sistema principal

### Comandos de Mantenimiento

```sql
-- Limpiar cache de códigos de barras (30 días)
SELECT limpiar_cache_codigos_barras(30);

-- Estadísticas de uso
SELECT * FROM vista_trabajos_impresion_resumen 
WHERE fecha_inicio >= NOW() - INTERVAL '7 days';
```

## Soporte y Contacto

### Documentación Adicional
- **Especificación Técnica Completa**: `modulo_etiquetas_especificacion.md`
- **Documentación de Base de Datos**: `README_Esquema_SQL.md`
- **Manual de Usuario**: Disponible en la aplicación

### Información de Versión
- **Versión**: 1.0.0
- **Fecha**: Julio 2025
- **Compatibilidad**: Sistema Ferre-POS v2.0+
- **Autor**: Manus AI

### Licencia
Este módulo es parte del sistema Ferre-POS y está sujeto a los términos de licencia del sistema principal.

---

**Nota**: Este módulo ha sido diseñado específicamente para ferreterías y puede requerir adaptaciones para otros tipos de comercio. Para consultas sobre personalización o soporte técnico, contactar al equipo de desarrollo del sistema Ferre-POS.

