# Guía de Integración FERRE-POS APIs

**Sistema de Punto de Venta para Ferreterías**  
**Versión**: 1.0.0  
**Autor**: Manus AI  
**Fecha**: Enero 2025

---

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Configuración Inicial](#configuración-inicial)
3. [Flujos de Integración Comunes](#flujos-de-integración-comunes)
4. [Ejemplos de Código](#ejemplos-de-código)
5. [Manejo de Errores](#manejo-de-errores)
6. [Mejores Prácticas](#mejores-prácticas)
7. [Testing y Debugging](#testing-y-debugging)
8. [Casos de Uso Específicos](#casos-de-uso-específicos)

---

## Introducción

Esta guía proporciona instrucciones detalladas para integrar aplicaciones externas con el sistema FERRE-POS. El sistema está compuesto por 4 APIs REST especializadas que pueden utilizarse independientemente o en conjunto según las necesidades de integración.

### APIs Disponibles

- **API POS** (Puerto 8080): Operaciones de punto de venta, gestión de productos y usuarios
- **API Sync** (Puerto 8081): Sincronización entre terminales y servidor central
- **API Labels** (Puerto 8082): Generación de etiquetas y códigos de barras
- **API Reports** (Puerto 8083): Reportes, analytics y dashboards

### Requisitos Previos

- Conocimiento básico de APIs REST
- Capacidad para realizar requests HTTP
- Comprensión de autenticación JWT
- Acceso a credenciales del sistema FERRE-POS

## Configuración Inicial

### 1. Obtener Credenciales

Antes de comenzar la integración, necesitará obtener las credenciales apropiadas:

**Para integración como usuario:**
- Username y password de un usuario con permisos apropiados
- Identificación de sucursal y terminal (si aplica)

**Para integración como terminal:**
- Terminal ID único
- Terminal secret (clave secreta)
- MAC address de la terminal
- IP address de la terminal

### 2. Configurar Entorno

```bash
# Variables de entorno recomendadas
export FERRE_POS_BASE_URL="http://localhost"
export FERRE_POS_API_POS_PORT="8080"
export FERRE_POS_API_SYNC_PORT="8081"
export FERRE_POS_API_LABELS_PORT="8082"
export FERRE_POS_API_REPORTS_PORT="8083"

# Credenciales (usar variables de entorno seguras en producción)
export FERRE_POS_USERNAME="your_username"
export FERRE_POS_PASSWORD="your_password"
export FERRE_POS_TERMINAL_ID="your_terminal_id"
export FERRE_POS_TERMINAL_SECRET="your_terminal_secret"
```

### 3. Verificar Conectividad

```bash
# Verificar que todas las APIs estén disponibles
curl -f http://localhost:8080/health
curl -f http://localhost:8081/health
curl -f http://localhost:8082/health
curl -f http://localhost:8083/health
```

## Flujos de Integración Comunes

### Flujo 1: Autenticación y Gestión de Tokens

#### Paso 1: Login Inicial
```bash
# Login como usuario
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin_test",
    "password": "password123",
    "terminal_id": "test-terminal-1"
  }'
```

**Respuesta esperada:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": "test-user-1",
      "username": "admin_test",
      "rol": "admin"
    }
  }
}
```

#### Paso 2: Usar Token en Requests Subsecuentes
```bash
# Guardar token en variable
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Usar token en requests
curl -X GET "http://localhost:8080/api/v1/products" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

#### Paso 3: Renovar Token
```bash
# Cuando el token esté próximo a expirar
curl -X POST "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Flujo 2: Gestión de Productos

#### Listar Productos con Filtros
```bash
curl -X GET "http://localhost:8080/api/v1/products" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -G \
  -d "sucursal_id=test-sucursal-1" \
  -d "categoria_id=test-category-1" \
  -d "activo=true" \
  -d "page=1" \
  -d "per_page=20"
```

#### Buscar Producto por Código de Barras
```bash
curl -X GET "http://localhost:8080/api/v1/products/barcode/1234567890123" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -G \
  -d "sucursal_id=test-sucursal-1"
```

#### Crear Nuevo Producto
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "codigo": "NUEVO001",
    "codigo_barras": "1234567890999",
    "nombre": "Producto Nuevo",
    "descripcion": "Descripción del producto nuevo",
    "categoria_id": "test-category-1",
    "precio": 25000.00,
    "costo": 15000.00,
    "stock_minimo": 5,
    "activo": true
  }'
```

### Flujo 3: Procesamiento de Ventas

#### Crear Venta Completa
```bash
curl -X POST "http://localhost:8080/api/v1/sales" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sucursal_id": "test-sucursal-1",
    "terminal_id": "test-terminal-1",
    "cajero_id": "test-user-3",
    "vendedor_id": "test-user-2",
    "cliente": {
      "nombre": "Juan Pérez",
      "email": "juan.perez@email.com"
    },
    "tipo_documento": "boleta",
    "items": [
      {
        "producto_id": "test-product-1",
        "cantidad": 2,
        "precio_unitario": 15000.00
      },
      {
        "producto_id": "test-product-2",
        "cantidad": 1,
        "precio_unitario": 3500.00
      }
    ],
    "medios_pago": [
      {
        "medio_pago": "efectivo",
        "monto": 20000.00
      },
      {
        "medio_pago": "tarjeta_credito",
        "monto": 13500.00
      }
    ]
  }'
```

### Flujo 4: Sincronización de Terminal

#### Autenticación de Terminal
```bash
curl -X POST "http://localhost:8081/api/v1/sync/auth/terminal" \
  -H "Content-Type: application/json" \
  -d '{
    "terminal_id": "test-terminal-1",
    "terminal_secret": "secret_key_for_terminal_1",
    "mac_address": "00:11:22:33:44:55",
    "ip_address": "192.168.1.100",
    "software_version": "FERRE-POS Terminal v1.0.5",
    "location": {
      "sucursal_id": "test-sucursal-1",
      "zona": "Caja Principal"
    }
  }'
```

#### Descargar Cambios del Servidor
```bash
curl -X POST "http://localhost:8081/api/v1/sync/pull" \
  -H "Authorization: Bearer $TERMINAL_TOKEN" \
  -H "X-Terminal-ID: test-terminal-1" \
  -H "Content-Type: application/json" \
  -d '{
    "last_sync_timestamp": "2025-01-08T12:00:00Z",
    "sync_types": ["productos", "precios", "categorias"],
    "filters": {
      "sucursal_id": "test-sucursal-1",
      "only_active": true
    }
  }'
```

#### Enviar Cambios al Servidor
```bash
curl -X POST "http://localhost:8081/api/v1/sync/push" \
  -H "Authorization: Bearer $TERMINAL_TOKEN" \
  -H "X-Terminal-ID: test-terminal-1" \
  -H "Content-Type: application/json" \
  -d '{
    "terminal_id": "test-terminal-1",
    "sync_timestamp": "2025-01-08T13:20:00Z",
    "changes": {
      "ventas": {
        "created": [
          {
            "id": "local-sale-001",
            "numero_documento": "LOCAL-001",
            "fecha_venta": "2025-01-08T13:15:00Z",
            "total": 18500.00,
            "created_offline": true
          }
        ]
      }
    }
  }'
```

### Flujo 5: Generación de Etiquetas

#### Listar Plantillas Disponibles
```bash
curl -X GET "http://localhost:8082/api/v1/templates" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

#### Generar Etiquetas Múltiples
```bash
curl -X POST "http://localhost:8082/api/v1/labels/generate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plantilla_id": "template-basic",
    "productos": [
      {
        "producto_id": "test-product-1",
        "cantidad": 10
      },
      {
        "producto_id": "test-product-2",
        "cantidad": 25
      }
    ],
    "opciones": {
      "formato_salida": "pdf",
      "etiquetas_por_hoja": 12
    }
  }'
```

### Flujo 6: Obtener Reportes

#### Resumen Ejecutivo de Ventas
```bash
curl -X GET "http://localhost:8083/api/v1/reports/sales/summary" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -G \
  -d "fecha_inicio=2025-01-01" \
  -d "fecha_fin=2025-01-08" \
  -d "comparar_periodo_anterior=true"
```

#### Dashboard Ejecutivo
```bash
curl -X GET "http://localhost:8083/api/v1/reports/dashboard/executive" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -G \
  -d "periodo=mes"
```

## Ejemplos de Código

### JavaScript/Node.js

```javascript
const axios = require('axios');

class FerrePoSClient {
  constructor(baseUrl = 'http://localhost') {
    this.baseUrl = baseUrl;
    this.ports = {
      pos: 8080,
      sync: 8081,
      labels: 8082,
      reports: 8083
    };
    this.accessToken = null;
    this.refreshToken = null;
  }

  // Autenticación
  async login(username, password, terminalId = null) {
    try {
      const response = await axios.post(
        `${this.baseUrl}:${this.ports.pos}/api/v1/auth/login`,
        {
          username,
          password,
          terminal_id: terminalId
        }
      );

      if (response.data.success) {
        this.accessToken = response.data.data.access_token;
        this.refreshToken = response.data.data.refresh_token;
        return response.data.data;
      }
      throw new Error('Login failed');
    } catch (error) {
      console.error('Login error:', error.response?.data || error.message);
      throw error;
    }
  }

  // Configurar headers con autenticación
  getAuthHeaders() {
    return {
      'Authorization': `Bearer ${this.accessToken}`,
      'Content-Type': 'application/json'
    };
  }

  // Listar productos
  async getProducts(sucursalId, filters = {}) {
    try {
      const params = new URLSearchParams({
        sucursal_id: sucursalId,
        ...filters
      });

      const response = await axios.get(
        `${this.baseUrl}:${this.ports.pos}/api/v1/products?${params}`,
        { headers: this.getAuthHeaders() }
      );

      return response.data;
    } catch (error) {
      console.error('Get products error:', error.response?.data || error.message);
      throw error;
    }
  }

  // Crear venta
  async createSale(saleData) {
    try {
      const response = await axios.post(
        `${this.baseUrl}:${this.ports.pos}/api/v1/sales`,
        saleData,
        { headers: this.getAuthHeaders() }
      );

      return response.data;
    } catch (error) {
      console.error('Create sale error:', error.response?.data || error.message);
      throw error;
    }
  }

  // Obtener reporte de ventas
  async getSalesReport(fechaInicio, fechaFin, options = {}) {
    try {
      const params = new URLSearchParams({
        fecha_inicio: fechaInicio,
        fecha_fin: fechaFin,
        ...options
      });

      const response = await axios.get(
        `${this.baseUrl}:${this.ports.reports}/api/v1/reports/sales/summary?${params}`,
        { headers: this.getAuthHeaders() }
      );

      return response.data;
    } catch (error) {
      console.error('Get sales report error:', error.response?.data || error.message);
      throw error;
    }
  }

  // Generar etiquetas
  async generateLabels(plantillaId, productos, opciones = {}) {
    try {
      const response = await axios.post(
        `${this.baseUrl}:${this.ports.labels}/api/v1/labels/generate`,
        {
          plantilla_id: plantillaId,
          productos,
          opciones
        },
        { headers: this.getAuthHeaders() }
      );

      return response.data;
    } catch (error) {
      console.error('Generate labels error:', error.response?.data || error.message);
      throw error;
    }
  }
}

// Ejemplo de uso
async function example() {
  const client = new FerrePoSClient();

  try {
    // Login
    await client.login('admin_test', 'password123', 'test-terminal-1');
    console.log('Login exitoso');

    // Obtener productos
    const products = await client.getProducts('test-sucursal-1', {
      page: 1,
      per_page: 10
    });
    console.log('Productos:', products.data.length);

    // Crear venta
    const sale = await client.createSale({
      sucursal_id: 'test-sucursal-1',
      terminal_id: 'test-terminal-1',
      cajero_id: 'test-user-3',
      items: [
        {
          producto_id: 'test-product-1',
          cantidad: 1,
          precio_unitario: 15000.00
        }
      ],
      medios_pago: [
        {
          medio_pago: 'efectivo',
          monto: 15000.00
        }
      ]
    });
    console.log('Venta creada:', sale.data.id);

    // Obtener reporte
    const report = await client.getSalesReport('2025-01-01', '2025-01-08');
    console.log('Total ventas:', report.data.metricas_principales.total_ventas);

  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Ejecutar ejemplo
example();
```

### Python

```python
import requests
import json
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List

class FerrePoSClient:
    def __init__(self, base_url: str = "http://localhost"):
        self.base_url = base_url
        self.ports = {
            'pos': 8080,
            'sync': 8081,
            'labels': 8082,
            'reports': 8083
        }
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.session = requests.Session()

    def login(self, username: str, password: str, terminal_id: Optional[str] = None) -> Dict[str, Any]:
        """Autenticar usuario y obtener tokens"""
        url = f"{self.base_url}:{self.ports['pos']}/api/v1/auth/login"
        data = {
            "username": username,
            "password": password
        }
        if terminal_id:
            data["terminal_id"] = terminal_id

        response = self.session.post(url, json=data)
        response.raise_for_status()
        
        result = response.json()
        if result.get('success'):
            self.access_token = result['data']['access_token']
            self.refresh_token = result['data']['refresh_token']
            return result['data']
        else:
            raise Exception(f"Login failed: {result.get('error', {}).get('message')}")

    def _get_auth_headers(self) -> Dict[str, str]:
        """Obtener headers con autenticación"""
        if not self.access_token:
            raise Exception("No access token available. Please login first.")
        
        return {
            'Authorization': f'Bearer {self.access_token}',
            'Content-Type': 'application/json'
        }

    def get_products(self, sucursal_id: str, **filters) -> Dict[str, Any]:
        """Obtener lista de productos"""
        url = f"{self.base_url}:{self.ports['pos']}/api/v1/products"
        params = {'sucursal_id': sucursal_id, **filters}
        
        response = self.session.get(url, params=params, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def get_product_by_barcode(self, barcode: str, sucursal_id: str) -> Dict[str, Any]:
        """Buscar producto por código de barras"""
        url = f"{self.base_url}:{self.ports['pos']}/api/v1/products/barcode/{barcode}"
        params = {'sucursal_id': sucursal_id}
        
        response = self.session.get(url, params=params, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def create_product(self, product_data: Dict[str, Any]) -> Dict[str, Any]:
        """Crear nuevo producto"""
        url = f"{self.base_url}:{self.ports['pos']}/api/v1/products"
        
        response = self.session.post(url, json=product_data, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def create_sale(self, sale_data: Dict[str, Any]) -> Dict[str, Any]:
        """Crear nueva venta"""
        url = f"{self.base_url}:{self.ports['pos']}/api/v1/sales"
        
        response = self.session.post(url, json=sale_data, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def get_sales_report(self, fecha_inicio: str, fecha_fin: str, **options) -> Dict[str, Any]:
        """Obtener reporte de ventas"""
        url = f"{self.base_url}:{self.ports['reports']}/api/v1/reports/sales/summary"
        params = {
            'fecha_inicio': fecha_inicio,
            'fecha_fin': fecha_fin,
            **options
        }
        
        response = self.session.get(url, params=params, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def get_dashboard(self, periodo: str = 'mes') -> Dict[str, Any]:
        """Obtener dashboard ejecutivo"""
        url = f"{self.base_url}:{self.ports['reports']}/api/v1/reports/dashboard/executive"
        params = {'periodo': periodo}
        
        response = self.session.get(url, params=params, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

    def generate_labels(self, plantilla_id: str, productos: List[Dict], opciones: Dict = None) -> Dict[str, Any]:
        """Generar etiquetas"""
        url = f"{self.base_url}:{self.ports['labels']}/api/v1/labels/generate"
        data = {
            'plantilla_id': plantilla_id,
            'productos': productos,
            'opciones': opciones or {}
        }
        
        response = self.session.post(url, json=data, headers=self._get_auth_headers())
        response.raise_for_status()
        return response.json()

# Ejemplo de uso
def main():
    client = FerrePoSClient()
    
    try:
        # Login
        user_data = client.login('admin_test', 'password123', 'test-terminal-1')
        print(f"Login exitoso para usuario: {user_data['user']['username']}")
        
        # Obtener productos
        products = client.get_products('test-sucursal-1', page=1, per_page=5)
        print(f"Productos encontrados: {len(products['data'])}")
        
        # Buscar producto por código de barras
        try:
            product = client.get_product_by_barcode('1234567890123', 'test-sucursal-1')
            print(f"Producto encontrado: {product['data']['nombre']}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print("Producto no encontrado por código de barras")
        
        # Crear venta
        sale_data = {
            'sucursal_id': 'test-sucursal-1',
            'terminal_id': 'test-terminal-1',
            'cajero_id': 'test-user-3',
            'vendedor_id': 'test-user-2',
            'cliente': {
                'nombre': 'Cliente Python Test',
                'email': 'python@test.com'
            },
            'items': [
                {
                    'producto_id': 'test-product-1',
                    'cantidad': 1,
                    'precio_unitario': 15000.00
                }
            ],
            'medios_pago': [
                {
                    'medio_pago': 'efectivo',
                    'monto': 15000.00
                }
            ]
        }
        
        sale = client.create_sale(sale_data)
        print(f"Venta creada: {sale['data']['numero_documento']}")
        
        # Obtener reporte de ventas
        today = datetime.now().strftime('%Y-%m-%d')
        week_ago = (datetime.now() - timedelta(days=7)).strftime('%Y-%m-%d')
        
        report = client.get_sales_report(week_ago, today, comparar_periodo_anterior=True)
        print(f"Total ventas última semana: ${report['data']['metricas_principales']['total_ventas']:,.2f}")
        
        # Obtener dashboard
        dashboard = client.get_dashboard('mes')
        ventas_kpi = dashboard['data']['kpis_principales']['ventas_totales']
        print(f"Progreso ventas del mes: {ventas_kpi['progreso_porcentual']:.1f}%")
        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
```

### PHP

```php
<?php

class FerrePoSClient {
    private $baseUrl;
    private $ports;
    private $accessToken;
    private $refreshToken;
    
    public function __construct($baseUrl = 'http://localhost') {
        $this->baseUrl = $baseUrl;
        $this->ports = [
            'pos' => 8080,
            'sync' => 8081,
            'labels' => 8082,
            'reports' => 8083
        ];
        $this->accessToken = null;
        $this->refreshToken = null;
    }
    
    public function login($username, $password, $terminalId = null) {
        $url = $this->baseUrl . ':' . $this->ports['pos'] . '/api/v1/auth/login';
        $data = [
            'username' => $username,
            'password' => $password
        ];
        
        if ($terminalId) {
            $data['terminal_id'] = $terminalId;
        }
        
        $response = $this->makeRequest('POST', $url, $data);
        
        if ($response['success']) {
            $this->accessToken = $response['data']['access_token'];
            $this->refreshToken = $response['data']['refresh_token'];
            return $response['data'];
        }
        
        throw new Exception('Login failed: ' . ($response['error']['message'] ?? 'Unknown error'));
    }
    
    private function getAuthHeaders() {
        if (!$this->accessToken) {
            throw new Exception('No access token available. Please login first.');
        }
        
        return [
            'Authorization: Bearer ' . $this->accessToken,
            'Content-Type: application/json'
        ];
    }
    
    public function getProducts($sucursalId, $filters = []) {
        $url = $this->baseUrl . ':' . $this->ports['pos'] . '/api/v1/products';
        $params = array_merge(['sucursal_id' => $sucursalId], $filters);
        $url .= '?' . http_build_query($params);
        
        return $this->makeRequest('GET', $url, null, $this->getAuthHeaders());
    }
    
    public function createSale($saleData) {
        $url = $this->baseUrl . ':' . $this->ports['pos'] . '/api/v1/sales';
        return $this->makeRequest('POST', $url, $saleData, $this->getAuthHeaders());
    }
    
    public function getSalesReport($fechaInicio, $fechaFin, $options = []) {
        $url = $this->baseUrl . ':' . $this->ports['reports'] . '/api/v1/reports/sales/summary';
        $params = array_merge([
            'fecha_inicio' => $fechaInicio,
            'fecha_fin' => $fechaFin
        ], $options);
        $url .= '?' . http_build_query($params);
        
        return $this->makeRequest('GET', $url, null, $this->getAuthHeaders());
    }
    
    private function makeRequest($method, $url, $data = null, $headers = []) {
        $ch = curl_init();
        
        curl_setopt_array($ch, [
            CURLOPT_URL => $url,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_TIMEOUT => 30,
            CURLOPT_CONNECTTIMEOUT => 10
        ]);
        
        if ($data && in_array($method, ['POST', 'PUT', 'PATCH'])) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
        }
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);
        
        if ($error) {
            throw new Exception('cURL error: ' . $error);
        }
        
        $decodedResponse = json_decode($response, true);
        
        if ($httpCode >= 400) {
            $errorMessage = $decodedResponse['error']['message'] ?? 'HTTP error ' . $httpCode;
            throw new Exception($errorMessage);
        }
        
        return $decodedResponse;
    }
}

// Ejemplo de uso
try {
    $client = new FerrePoSClient();
    
    // Login
    $userData = $client->login('admin_test', 'password123', 'test-terminal-1');
    echo "Login exitoso para usuario: " . $userData['user']['username'] . "\n";
    
    // Obtener productos
    $products = $client->getProducts('test-sucursal-1', ['page' => 1, 'per_page' => 5]);
    echo "Productos encontrados: " . count($products['data']) . "\n";
    
    // Crear venta
    $saleData = [
        'sucursal_id' => 'test-sucursal-1',
        'terminal_id' => 'test-terminal-1',
        'cajero_id' => 'test-user-3',
        'items' => [
            [
                'producto_id' => 'test-product-1',
                'cantidad' => 1,
                'precio_unitario' => 15000.00
            ]
        ],
        'medios_pago' => [
            [
                'medio_pago' => 'efectivo',
                'monto' => 15000.00
            ]
        ]
    ];
    
    $sale = $client->createSale($saleData);
    echo "Venta creada: " . $sale['data']['numero_documento'] . "\n";
    
    // Obtener reporte
    $fechaInicio = date('Y-m-d', strtotime('-7 days'));
    $fechaFin = date('Y-m-d');
    
    $report = $client->getSalesReport($fechaInicio, $fechaFin, ['comparar_periodo_anterior' => true]);
    echo "Total ventas última semana: $" . number_format($report['data']['metricas_principales']['total_ventas'], 2) . "\n";
    
} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
?>
```

## Manejo de Errores

### Códigos de Error Comunes

El sistema FERRE-POS utiliza códigos de estado HTTP estándar y códigos de error específicos:

#### Errores de Cliente (4xx)

**400 Bad Request**
```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Los parámetros de la solicitud son inválidos",
    "details": [
      {
        "field": "sucursal_id",
        "message": "El campo sucursal_id es requerido"
      }
    ]
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**401 Unauthorized**
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Token de acceso requerido o inválido"
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**403 Forbidden**
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Permisos insuficientes para esta operación"
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**404 Not Found**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "El recurso solicitado no existe"
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**409 Conflict**
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "Ya existe un producto con este código",
    "details": [
      {
        "field": "codigo",
        "message": "El código TEST001 ya está en uso"
      }
    ]
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**422 Unprocessable Entity**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Los datos no cumplen con las reglas de validación",
    "details": [
      {
        "field": "precio",
        "message": "El precio debe ser mayor a 0"
      },
      {
        "field": "email",
        "message": "Formato de email inválido"
      }
    ]
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

**429 Too Many Requests**
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Límite de solicitudes excedido. Intente nuevamente en 60 segundos"
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

#### Errores de Servidor (5xx)

**500 Internal Server Error**
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Error interno del servidor. Por favor contacte al soporte"
  },
  "request_id": "req_123456789",
  "timestamp": "2025-01-08T16:00:00Z"
}
```

### Estrategias de Manejo de Errores

#### 1. Retry con Backoff Exponencial

```javascript
async function makeRequestWithRetry(requestFn, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await requestFn();
    } catch (error) {
      if (error.response?.status >= 500 && attempt < maxRetries) {
        const delay = Math.pow(2, attempt) * 1000; // 2s, 4s, 8s
        console.log(`Attempt ${attempt} failed, retrying in ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      throw error;
    }
  }
}

// Uso
try {
  const result = await makeRequestWithRetry(() => 
    client.getProducts('test-sucursal-1')
  );
} catch (error) {
  console.error('Request failed after retries:', error.message);
}
```

#### 2. Manejo de Rate Limiting

```javascript
class RateLimitHandler {
  constructor() {
    this.requestQueue = [];
    this.isProcessing = false;
  }

  async makeRequest(requestFn) {
    return new Promise((resolve, reject) => {
      this.requestQueue.push({ requestFn, resolve, reject });
      this.processQueue();
    });
  }

  async processQueue() {
    if (this.isProcessing || this.requestQueue.length === 0) {
      return;
    }

    this.isProcessing = true;

    while (this.requestQueue.length > 0) {
      const { requestFn, resolve, reject } = this.requestQueue.shift();

      try {
        const result = await requestFn();
        resolve(result);
      } catch (error) {
        if (error.response?.status === 429) {
          // Rate limit hit, wait and retry
          const retryAfter = error.response.headers['retry-after'] || 60;
          console.log(`Rate limit hit, waiting ${retryAfter} seconds...`);
          await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
          
          // Put request back in queue
          this.requestQueue.unshift({ requestFn, resolve, reject });
          continue;
        }
        reject(error);
      }

      // Small delay between requests
      await new Promise(resolve => setTimeout(resolve, 100));
    }

    this.isProcessing = false;
  }
}
```

#### 3. Validación de Datos Antes del Envío

```javascript
function validateSaleData(saleData) {
  const errors = [];

  // Validar campos requeridos
  if (!saleData.sucursal_id) {
    errors.push({ field: 'sucursal_id', message: 'Sucursal ID es requerido' });
  }

  if (!saleData.items || saleData.items.length === 0) {
    errors.push({ field: 'items', message: 'Al menos un item es requerido' });
  }

  // Validar items
  if (saleData.items) {
    saleData.items.forEach((item, index) => {
      if (!item.producto_id) {
        errors.push({ 
          field: `items[${index}].producto_id`, 
          message: 'Producto ID es requerido' 
        });
      }
      if (!item.cantidad || item.cantidad <= 0) {
        errors.push({ 
          field: `items[${index}].cantidad`, 
          message: 'Cantidad debe ser mayor a 0' 
        });
      }
      if (!item.precio_unitario || item.precio_unitario <= 0) {
        errors.push({ 
          field: `items[${index}].precio_unitario`, 
          message: 'Precio unitario debe ser mayor a 0' 
        });
      }
    });
  }

  // Validar medios de pago
  if (!saleData.medios_pago || saleData.medios_pago.length === 0) {
    errors.push({ field: 'medios_pago', message: 'Al menos un medio de pago es requerido' });
  }

  if (errors.length > 0) {
    throw new ValidationError('Datos de venta inválidos', errors);
  }
}

class ValidationError extends Error {
  constructor(message, errors) {
    super(message);
    this.name = 'ValidationError';
    this.errors = errors;
  }
}

// Uso
try {
  validateSaleData(saleData);
  const result = await client.createSale(saleData);
} catch (error) {
  if (error instanceof ValidationError) {
    console.error('Validation errors:', error.errors);
  } else {
    console.error('API error:', error.message);
  }
}
```

## Mejores Prácticas

### 1. Gestión de Tokens

```javascript
class TokenManager {
  constructor(client) {
    this.client = client;
    this.tokenRefreshPromise = null;
  }

  async ensureValidToken() {
    if (!this.client.accessToken) {
      throw new Error('No access token available');
    }

    // Verificar si el token está próximo a expirar
    const payload = JSON.parse(atob(this.client.accessToken.split('.')[1]));
    const exp = payload.exp * 1000;
    const now = Date.now();
    const timeUntilExpiry = exp - now;

    // Si expira en menos de 5 minutos, renovar
    if (timeUntilExpiry < 300000) {
      return this.refreshToken();
    }

    return this.client.accessToken;
  }

  async refreshToken() {
    // Evitar múltiples refreshes simultáneos
    if (this.tokenRefreshPromise) {
      return this.tokenRefreshPromise;
    }

    this.tokenRefreshPromise = this.client.refreshToken()
      .then(result => {
        this.tokenRefreshPromise = null;
        return result.access_token;
      })
      .catch(error => {
        this.tokenRefreshPromise = null;
        throw error;
      });

    return this.tokenRefreshPromise;
  }
}
```

### 2. Cacheo de Datos

```javascript
class DataCache {
  constructor(ttlSeconds = 300) { // 5 minutos por defecto
    this.cache = new Map();
    this.ttl = ttlSeconds * 1000;
  }

  set(key, value) {
    this.cache.set(key, {
      value,
      timestamp: Date.now()
    });
  }

  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;

    if (Date.now() - item.timestamp > this.ttl) {
      this.cache.delete(key);
      return null;
    }

    return item.value;
  }

  clear() {
    this.cache.clear();
  }
}

// Uso con productos
class ProductService {
  constructor(client) {
    this.client = client;
    this.cache = new DataCache(600); // 10 minutos para productos
  }

  async getProduct(productId, sucursalId) {
    const cacheKey = `product_${productId}_${sucursalId}`;
    let product = this.cache.get(cacheKey);

    if (!product) {
      const response = await this.client.getProduct(productId, sucursalId);
      product = response.data;
      this.cache.set(cacheKey, product);
    }

    return product;
  }
}
```

### 3. Logging y Monitoreo

```javascript
class APILogger {
  constructor(level = 'info') {
    this.level = level;
    this.levels = { error: 0, warn: 1, info: 2, debug: 3 };
  }

  log(level, message, data = {}) {
    if (this.levels[level] <= this.levels[this.level]) {
      const logEntry = {
        timestamp: new Date().toISOString(),
        level,
        message,
        ...data
      };
      console.log(JSON.stringify(logEntry));
    }
  }

  logRequest(method, url, data = null) {
    this.log('debug', 'API Request', {
      method,
      url,
      data: data ? JSON.stringify(data) : null
    });
  }

  logResponse(method, url, status, responseTime, data = null) {
    this.log('debug', 'API Response', {
      method,
      url,
      status,
      responseTime,
      dataSize: data ? JSON.stringify(data).length : 0
    });
  }

  logError(method, url, error) {
    this.log('error', 'API Error', {
      method,
      url,
      error: error.message,
      status: error.response?.status,
      details: error.response?.data
    });
  }
}
```

### 4. Configuración por Entorno

```javascript
class Config {
  constructor() {
    this.env = process.env.NODE_ENV || 'development';
    this.configs = {
      development: {
        baseUrl: 'http://localhost',
        timeout: 30000,
        retries: 3,
        logLevel: 'debug'
      },
      staging: {
        baseUrl: 'https://staging-api.ferrepos.com',
        timeout: 15000,
        retries: 2,
        logLevel: 'info'
      },
      production: {
        baseUrl: 'https://api.ferrepos.com',
        timeout: 10000,
        retries: 1,
        logLevel: 'warn'
      }
    };
  }

  get(key) {
    return this.configs[this.env][key];
  }

  getAll() {
    return this.configs[this.env];
  }
}

const config = new Config();
const client = new FerrePoSClient(config.get('baseUrl'));
```

## Testing y Debugging

### 1. Tests Unitarios

```javascript
const { expect } = require('chai');
const sinon = require('sinon');
const FerrePoSClient = require('./ferre-pos-client');

describe('FerrePoSClient', () => {
  let client;
  let mockAxios;

  beforeEach(() => {
    client = new FerrePoSClient();
    mockAxios = sinon.stub(client, 'session');
  });

  afterEach(() => {
    sinon.restore();
  });

  describe('login', () => {
    it('should login successfully with valid credentials', async () => {
      const mockResponse = {
        data: {
          success: true,
          data: {
            access_token: 'mock_token',
            refresh_token: 'mock_refresh',
            user: { id: '1', username: 'test' }
          }
        }
      };

      mockAxios.post.resolves(mockResponse);

      const result = await client.login('test', 'password');

      expect(result.access_token).to.equal('mock_token');
      expect(client.accessToken).to.equal('mock_token');
      expect(mockAxios.post.calledOnce).to.be.true;
    });

    it('should throw error with invalid credentials', async () => {
      mockAxios.post.rejects(new Error('Unauthorized'));

      try {
        await client.login('invalid', 'invalid');
        expect.fail('Should have thrown an error');
      } catch (error) {
        expect(error.message).to.include('Unauthorized');
      }
    });
  });

  describe('getProducts', () => {
    it('should fetch products with filters', async () => {
      const mockResponse = {
        data: {
          success: true,
          data: [{ id: '1', name: 'Product 1' }]
        }
      };

      client.accessToken = 'mock_token';
      mockAxios.get.resolves(mockResponse);

      const result = await client.getProducts('sucursal-1', { page: 1 });

      expect(result.data).to.have.length(1);
      expect(mockAxios.get.calledOnce).to.be.true;
    });
  });
});
```

### 2. Tests de Integración

```javascript
describe('Integration Tests', () => {
  let client;

  before(async () => {
    client = new FerrePoSClient('http://localhost');
    await client.login('admin_test', 'password123');
  });

  it('should create and retrieve a product', async () => {
    // Crear producto
    const productData = {
      codigo: 'TEST_INTEGRATION_001',
      nombre: 'Producto Test Integración',
      categoria_id: 'test-category-1',
      precio: 10000.00,
      costo: 6000.00
    };

    const createResponse = await client.createProduct(productData);
    expect(createResponse.success).to.be.true;

    const productId = createResponse.data.id;

    // Recuperar producto
    const getResponse = await client.getProduct(productId, 'test-sucursal-1');
    expect(getResponse.success).to.be.true;
    expect(getResponse.data.nombre).to.equal(productData.nombre);

    // Limpiar
    await client.deleteProduct(productId);
  });

  it('should process a complete sale flow', async () => {
    // Obtener productos disponibles
    const products = await client.getProducts('test-sucursal-1', { per_page: 1 });
    expect(products.data).to.have.length.greaterThan(0);

    const product = products.data[0];

    // Crear venta
    const saleData = {
      sucursal_id: 'test-sucursal-1',
      terminal_id: 'test-terminal-1',
      cajero_id: 'test-user-3',
      items: [{
        producto_id: product.id,
        cantidad: 1,
        precio_unitario: product.precio
      }],
      medios_pago: [{
        medio_pago: 'efectivo',
        monto: product.precio
      }]
    };

    const saleResponse = await client.createSale(saleData);
    expect(saleResponse.success).to.be.true;
    expect(saleResponse.data.total).to.equal(product.precio);
  });
});
```

### 3. Debugging

```javascript
class DebugClient extends FerrePoSClient {
  constructor(baseUrl, options = {}) {
    super(baseUrl);
    this.debug = options.debug || false;
    this.logger = new APILogger(options.logLevel || 'debug');
  }

  async makeRequest(method, url, data = null, headers = []) {
    const startTime = Date.now();
    
    if (this.debug) {
      this.logger.logRequest(method, url, data);
    }

    try {
      const response = await super.makeRequest(method, url, data, headers);
      
      if (this.debug) {
        const responseTime = Date.now() - startTime;
        this.logger.logResponse(method, url, 200, responseTime, response);
      }

      return response;
    } catch (error) {
      if (this.debug) {
        this.logger.logError(method, url, error);
      }
      throw error;
    }
  }
}

// Uso para debugging
const debugClient = new DebugClient('http://localhost', {
  debug: true,
  logLevel: 'debug'
});
```

## Casos de Uso Específicos

### 1. Integración con Sistema de Inventario Externo

```javascript
class InventorySync {
  constructor(ferrePoSClient, externalInventoryAPI) {
    this.ferrePos = ferrePoSClient;
    this.external = externalInventoryAPI;
    this.syncInterval = 300000; // 5 minutos
  }

  async startSync() {
    console.log('Iniciando sincronización de inventario...');
    
    // Sincronización inicial
    await this.syncProducts();
    await this.syncStock();

    // Programar sincronización periódica
    setInterval(async () => {
      try {
        await this.syncStock();
      } catch (error) {
        console.error('Error en sincronización periódica:', error.message);
      }
    }, this.syncInterval);
  }

  async syncProducts() {
    try {
      // Obtener productos del sistema externo
      const externalProducts = await this.external.getProducts();
      
      // Obtener productos actuales de FERRE-POS
      const ferrePoSProducts = await this.ferrePos.getProducts('test-sucursal-1');
      
      // Crear mapa de productos existentes
      const existingProducts = new Map();
      ferrePoSProducts.data.forEach(product => {
        existingProducts.set(product.codigo, product);
      });

      // Sincronizar cada producto
      for (const extProduct of externalProducts) {
        const existing = existingProducts.get(extProduct.code);
        
        if (existing) {
          // Actualizar si hay cambios
          if (this.hasProductChanges(existing, extProduct)) {
            await this.ferrePos.updateProduct(existing.id, {
              nombre: extProduct.name,
              precio: extProduct.price,
              costo: extProduct.cost
            });
            console.log(`Producto actualizado: ${extProduct.code}`);
          }
        } else {
          // Crear nuevo producto
          await this.ferrePos.createProduct({
            codigo: extProduct.code,
            codigo_barras: extProduct.barcode,
            nombre: extProduct.name,
            categoria_id: this.mapCategory(extProduct.category),
            precio: extProduct.price,
            costo: extProduct.cost
          });
          console.log(`Producto creado: ${extProduct.code}`);
        }
      }
    } catch (error) {
      console.error('Error sincronizando productos:', error.message);
      throw error;
    }
  }

  async syncStock() {
    try {
      // Obtener movimientos de stock del sistema externo
      const stockMovements = await this.external.getStockMovements();
      
      for (const movement of stockMovements) {
        // Buscar producto en FERRE-POS
        const product = await this.ferrePos.getProductByCode(movement.productCode);
        
        if (product.success) {
          // Crear movimiento de stock
          await this.ferrePos.createStockMovement({
            producto_id: product.data.id,
            sucursal_id: 'test-sucursal-1',
            tipo_movimiento: movement.type, // 'entrada' o 'salida'
            cantidad: movement.quantity,
            motivo: 'Sincronización automática',
            referencia_externa: movement.id
          });
        }
      }
      
      console.log(`Sincronizados ${stockMovements.length} movimientos de stock`);
    } catch (error) {
      console.error('Error sincronizando stock:', error.message);
      throw error;
    }
  }

  hasProductChanges(ferrePoSProduct, externalProduct) {
    return (
      ferrePoSProduct.nombre !== externalProduct.name ||
      ferrePoSProduct.precio !== externalProduct.price ||
      ferrePoSProduct.costo !== externalProduct.cost
    );
  }

  mapCategory(externalCategory) {
    // Mapear categorías del sistema externo a FERRE-POS
    const categoryMap = {
      'tools': 'test-category-1',
      'materials': 'test-category-2',
      'hardware': 'test-category-3',
      'electrical': 'test-category-4'
    };
    
    return categoryMap[externalCategory] || 'test-category-1';
  }
}
```

### 2. Generación Automática de Reportes

```javascript
class AutoReportGenerator {
  constructor(ferrePoSClient, emailService) {
    this.ferrePos = ferrePoSClient;
    this.email = emailService;
  }

  async generateDailyReport() {
    try {
      const today = new Date().toISOString().split('T')[0];
      const yesterday = new Date(Date.now() - 86400000).toISOString().split('T')[0];

      // Obtener datos del reporte
      const salesReport = await this.ferrePos.getSalesReport(yesterday, yesterday, {
        comparar_periodo_anterior: true
      });

      const inventoryReport = await this.ferrePos.getInventoryStatus({
        stock_critico: true
      });

      // Generar contenido del reporte
      const reportContent = this.generateReportHTML(salesReport.data, inventoryReport.data);

      // Enviar por email
      await this.email.send({
        to: ['gerente@ferreteria.com', 'supervisor@ferreteria.com'],
        subject: `Reporte Diario - ${yesterday}`,
        html: reportContent,
        attachments: await this.generateAttachments(yesterday)
      });

      console.log(`Reporte diario enviado para ${yesterday}`);
    } catch (error) {
      console.error('Error generando reporte diario:', error.message);
      throw error;
    }
  }

  generateReportHTML(salesData, inventoryData) {
    return `
      <html>
        <head>
          <style>
            body { font-family: Arial, sans-serif; }
            .header { background-color: #f0f0f0; padding: 20px; }
            .metric { margin: 10px 0; }
            .alert { color: red; font-weight: bold; }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>Reporte Diario de Ventas</h1>
            <p>Fecha: ${salesData.periodo.fecha_inicio}</p>
          </div>
          
          <h2>Métricas de Ventas</h2>
          <div class="metric">
            <strong>Total Ventas:</strong> $${salesData.metricas_principales.total_ventas.toLocaleString()}
          </div>
          <div class="metric">
            <strong>Transacciones:</strong> ${salesData.metricas_principales.cantidad_transacciones}
          </div>
          <div class="metric">
            <strong>Ticket Promedio:</strong> $${salesData.metricas_principales.ticket_promedio.toLocaleString()}
          </div>
          
          <h2>Alertas de Inventario</h2>
          ${inventoryData.productos_stock_critico.map(product => `
            <div class="alert">
              ⚠️ Stock crítico: ${product.nombre} (${product.stock_actual} unidades)
            </div>
          `).join('')}
          
          <h2>Top Productos</h2>
          <ul>
            ${salesData.top_productos.map(product => `
              <li>${product.nombre}: ${product.cantidad_vendida} unidades - $${product.monto_total.toLocaleString()}</li>
            `).join('')}
          </ul>
        </body>
      </html>
    `;
  }

  async generateAttachments(date) {
    // Generar reporte detallado en PDF
    const exportResponse = await this.ferrePos.exportReport({
      tipo_reporte: 'sales_detailed',
      parametros: {
        fecha_inicio: date,
        fecha_fin: date
      },
      formato_exportacion: 'pdf'
    });

    return [
      {
        filename: `reporte_detallado_${date}.pdf`,
        path: exportResponse.data.download_url
      }
    ];
  }

  // Programar ejecución diaria
  scheduleDaily() {
    const cron = require('node-cron');
    
    // Ejecutar todos los días a las 8:00 AM
    cron.schedule('0 8 * * *', async () => {
      try {
        await this.generateDailyReport();
      } catch (error) {
        console.error('Error en reporte programado:', error.message);
      }
    });
  }
}
```

### 3. Terminal POS Offline

```javascript
class OfflinePOSTerminal {
  constructor(ferrePoSClient, terminalId) {
    this.ferrePos = ferrePoSClient;
    this.terminalId = terminalId;
    this.isOnline = false;
    this.offlineQueue = [];
    this.localData = {
      products: new Map(),
      lastSync: null
    };
    
    this.startHeartbeat();
    this.startSyncProcess();
  }

  async initialize() {
    try {
      // Intentar conectar y sincronizar
      await this.ferrePos.authenticateTerminal(this.terminalId);
      await this.syncData();
      this.isOnline = true;
      console.log('Terminal inicializada en modo online');
    } catch (error) {
      console.log('Iniciando en modo offline:', error.message);
      this.isOnline = false;
      await this.loadLocalData();
    }
  }

  async processSale(saleData) {
    try {
      // Validar datos localmente
      this.validateSaleData(saleData);
      
      if (this.isOnline) {
        // Procesar online
        const result = await this.ferrePos.createSale(saleData);
        return result;
      } else {
        // Procesar offline
        const offlineSale = {
          ...saleData,
          id: this.generateOfflineId(),
          created_offline: true,
          timestamp: new Date().toISOString()
        };
        
        this.offlineQueue.push({
          type: 'sale',
          data: offlineSale
        });
        
        await this.saveLocalData();
        
        return {
          success: true,
          data: offlineSale,
          offline: true
        };
      }
    } catch (error) {
      console.error('Error procesando venta:', error.message);
      throw error;
    }
  }

  async syncData() {
    try {
      if (!this.isOnline) return;

      // Descargar cambios del servidor
      const changes = await this.ferrePos.pullChanges({
        last_sync_timestamp: this.localData.lastSync,
        sync_types: ['productos', 'precios']
      });

      // Actualizar datos locales
      if (changes.data.changes.productos) {
        changes.data.changes.productos.created?.forEach(product => {
          this.localData.products.set(product.id, product);
        });
        
        changes.data.changes.productos.updated?.forEach(product => {
          this.localData.products.set(product.id, product);
        });
        
        changes.data.changes.productos.deleted?.forEach(productId => {
          this.localData.products.delete(productId);
        });
      }

      // Enviar cambios pendientes
      if (this.offlineQueue.length > 0) {
        await this.pushOfflineChanges();
      }

      this.localData.lastSync = new Date().toISOString();
      await this.saveLocalData();
      
      console.log('Sincronización completada');
    } catch (error) {
      console.error('Error en sincronización:', error.message);
      this.isOnline = false;
    }
  }

  async pushOfflineChanges() {
    const changes = {
      ventas: {
        created: this.offlineQueue
          .filter(item => item.type === 'sale')
          .map(item => item.data)
      }
    };

    try {
      await this.ferrePos.pushChanges({
        terminal_id: this.terminalId,
        changes
      });
      
      // Limpiar cola después de envío exitoso
      this.offlineQueue = [];
      await this.saveLocalData();
      
      console.log('Cambios offline enviados exitosamente');
    } catch (error) {
      console.error('Error enviando cambios offline:', error.message);
      throw error;
    }
  }

  validateSaleData(saleData) {
    // Validar que los productos existan localmente
    for (const item of saleData.items) {
      const product = this.localData.products.get(item.producto_id);
      if (!product) {
        throw new Error(`Producto ${item.producto_id} no encontrado localmente`);
      }
      
      // Validar stock si está disponible
      if (product.stock_actual < item.cantidad) {
        throw new Error(`Stock insuficiente para ${product.nombre}`);
      }
    }
  }

  generateOfflineId() {
    return `offline_${this.terminalId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  async saveLocalData() {
    // Guardar en localStorage o archivo local
    const data = {
      products: Array.from(this.localData.products.entries()),
      lastSync: this.localData.lastSync,
      offlineQueue: this.offlineQueue
    };
    
    localStorage.setItem(`terminal_${this.terminalId}`, JSON.stringify(data));
  }

  async loadLocalData() {
    try {
      const saved = localStorage.getItem(`terminal_${this.terminalId}`);
      if (saved) {
        const data = JSON.parse(saved);
        this.localData.products = new Map(data.products);
        this.localData.lastSync = data.lastSync;
        this.offlineQueue = data.offlineQueue || [];
      }
    } catch (error) {
      console.error('Error cargando datos locales:', error.message);
    }
  }

  startHeartbeat() {
    setInterval(async () => {
      try {
        await this.ferrePos.sendHeartbeat({
          terminal_id: this.terminalId,
          status: 'online',
          pending_transactions: this.offlineQueue.length
        });
        
        if (!this.isOnline) {
          this.isOnline = true;
          console.log('Conexión restaurada');
          await this.syncData();
        }
      } catch (error) {
        if (this.isOnline) {
          this.isOnline = false;
          console.log('Conexión perdida, cambiando a modo offline');
        }
      }
    }, 30000); // Cada 30 segundos
  }

  startSyncProcess() {
    setInterval(async () => {
      if (this.isOnline) {
        await this.syncData();
      }
    }, 300000); // Cada 5 minutos
  }
}
```

---

Esta guía proporciona una base sólida para integrar aplicaciones con el sistema FERRE-POS. Para casos de uso específicos o preguntas adicionales, consulte la documentación completa o contacte al equipo de soporte técnico.

