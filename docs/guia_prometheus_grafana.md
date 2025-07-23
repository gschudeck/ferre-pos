# Guía de Monitoreo con Prometheus y Grafana

## 1. Instalación de Prometheus en Ubuntu 24.04

```bash
sudo useradd --no-create-home --shell /bin/false prometheus
sudo mkdir /etc/prometheus /var/lib/prometheus
cd /tmp
curl -LO https://github.com/prometheus/prometheus/releases/latest/download/prometheus-2.52.0.linux-amd64.tar.gz
tar xvf prometheus-*.tar.gz
cd prometheus-*.linux-amd64
sudo cp prometheus promtool /usr/local/bin/
sudo cp -r consoles/ console_libraries/ /etc/prometheus/
sudo cp prometheus.yml /etc/prometheus/
sudo chown -R prometheus:prometheus /etc/prometheus /var/lib/prometheus
sudo chown prometheus:prometheus /usr/local/bin/prometheus /usr/local/bin/promtool
```

### Servicio systemd

```ini
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
ExecStart=/usr/local/bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/var/lib/prometheus

[Install]
WantedBy=default.target
```

```bash
sudo nano /etc/systemd/system/prometheus.service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus
```

---

## 2. Instalación de Grafana en Ubuntu 24.04

```bash
sudo apt update
sudo apt install -y software-properties-common
sudo add-apt-repository "deb [arch=amd64] https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt update
sudo apt install grafana
sudo systemctl enable grafana-server
sudo systemctl start grafana-server
```

---

## 3. Configurar Prometheus como fuente de datos

Archivo: `/etc/grafana/provisioning/datasources/prometheus.yaml`

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://localhost:9090
    isDefault: true
    editable: true
```

```bash
sudo systemctl restart grafana-server
```

---

## 4. Cargar Dashboard completo

Archivo: `/etc/grafana/provisioning/dashboards/prometheus_dashboards.yaml`

```yaml
apiVersion: 1

providers:
  - name: 'Prometheus Default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /var/lib/grafana/dashboards
```

Dashboard JSON: `/var/lib/grafana/dashboards/prometheus-completo.json`  
(Puedes usar el archivo que te entregué en el ZIP)

---

## 5. Monitorear otros servidores

### En cada servidor remoto:

```bash
sudo useradd --no-create-home --shell /bin/false node_exporter
cd /tmp
curl -LO https://github.com/prometheus/node_exporter/releases/latest/download/node_exporter-1.8.1.linux-amd64.tar.gz
tar xvf node_exporter-*.tar.gz
sudo cp node_exporter-*/node_exporter /usr/local/bin/
```

Servicio systemd `/etc/systemd/system/node_exporter.service`:

```ini
[Unit]
Description=Node Exporter
After=network.target

[Service]
User=node_exporter
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=default.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable node_exporter
sudo systemctl start node_exporter
```

### Agregar en Prometheus principal:

```yaml
  - job_name: 'servidores_remotos'
    static_configs:
      - targets:
        - '192.168.1.100:9100'
        - '192.168.1.101:9100'
```

```bash
sudo systemctl restart prometheus
```

---

## 6. Monitorear procesos específicos (ejemplo: nginx)

### Script: `/usr/local/bin/check_nginx.sh`

```bash
#!/bin/bash
pgrep nginx >/dev/null
if [ $? -eq 0 ]; then
    echo "proceso_nginx_up 1" > /var/lib/node_exporter/nginx.prom
else
    echo "proceso_nginx_up 0" > /var/lib/node_exporter/nginx.prom
fi
```

```bash
chmod +x /usr/local/bin/check_nginx.sh
mkdir -p /var/lib/node_exporter
crontab -e
# Agrega:
* * * * * /usr/local/bin/check_nginx.sh
```

Ejecutar Node Exporter con soporte textfile:

```bash
/usr/local/bin/node_exporter --collector.textfile.directory=/var/lib/node_exporter
```

### Consulta en Grafana:

```promql
proceso_nginx_up{instance="192.168.1.100:9100"}
```

---
