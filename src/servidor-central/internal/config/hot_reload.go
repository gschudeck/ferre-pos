package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// HotReloadManager gestiona la recarga en caliente de configuraciones
type HotReloadManager struct {
	configManager *ConfigManager
	watcher       *fsnotify.Watcher
	watchedFiles  map[string]string // path -> api name
	mutex         sync.RWMutex
	stopChan      chan bool
	reloadChan    chan ReloadEvent
	debounceTime  time.Duration
	lastReload    map[string]time.Time
}

// ReloadEvent representa un evento de recarga
type ReloadEvent struct {
	FilePath  string
	APIName   string
	EventType string
	Timestamp time.Time
}

// NewHotReloadManager crea un nuevo gestor de recarga en caliente
func NewHotReloadManager(configManager *ConfigManager) (*HotReloadManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creando watcher: %w", err)
	}

	return &HotReloadManager{
		configManager: configManager,
		watcher:       watcher,
		watchedFiles:  make(map[string]string),
		stopChan:      make(chan bool),
		reloadChan:    make(chan ReloadEvent, 100),
		debounceTime:  2 * time.Second, // Debounce de 2 segundos
		lastReload:    make(map[string]time.Time),
	}, nil
}

// Start inicia el sistema de recarga en caliente
func (hrm *HotReloadManager) Start() error {
	// Agregar archivos de configuración principales
	mainConfigPath := hrm.configManager.configPath
	if err := hrm.AddWatchFile(mainConfigPath, "main"); err != nil {
		return fmt.Errorf("error agregando archivo principal: %w", err)
	}

	// Agregar archivos de configuración por API
	configDir := filepath.Dir(mainConfigPath)
	apiConfigs := map[string]string{
		"pos":     filepath.Join(configDir, "pos", "pos-config.yaml"),
		"sync":    filepath.Join(configDir, "sync", "sync-config.yaml"),
		"labels":  filepath.Join(configDir, "labels", "labels-config.yaml"),
		"reports": filepath.Join(configDir, "reports", "reports-config.yaml"),
	}

	for apiName, configPath := range apiConfigs {
		if err := hrm.AddWatchFile(configPath, apiName); err != nil {
			log.Printf("Advertencia: No se pudo agregar archivo de configuración %s: %v", configPath, err)
		}
	}

	// Iniciar goroutines
	go hrm.watchFiles()
	go hrm.processReloadEvents()

	log.Println("Sistema de recarga en caliente iniciado")
	return nil
}

// Stop detiene el sistema de recarga en caliente
func (hrm *HotReloadManager) Stop() error {
	close(hrm.stopChan)

	if hrm.watcher != nil {
		if err := hrm.watcher.Close(); err != nil {
			return fmt.Errorf("error cerrando watcher: %w", err)
		}
	}

	log.Println("Sistema de recarga en caliente detenido")
	return nil
}

// AddWatchFile agrega un archivo para monitoreo
func (hrm *HotReloadManager) AddWatchFile(filePath, apiName string) error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	// Verificar si el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe: %s", filePath)
	}

	// Agregar al watcher
	if err := hrm.watcher.Add(filePath); err != nil {
		return fmt.Errorf("error agregando archivo al watcher: %w", err)
	}

	// Registrar archivo
	hrm.watchedFiles[filePath] = apiName
	log.Printf("Archivo agregado al monitoreo: %s (API: %s)", filePath, apiName)

	return nil
}

// RemoveWatchFile remueve un archivo del monitoreo
func (hrm *HotReloadManager) RemoveWatchFile(filePath string) error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	// Remover del watcher
	if err := hrm.watcher.Remove(filePath); err != nil {
		return fmt.Errorf("error removiendo archivo del watcher: %w", err)
	}

	// Desregistrar archivo
	delete(hrm.watchedFiles, filePath)
	delete(hrm.lastReload, filePath)
	log.Printf("Archivo removido del monitoreo: %s", filePath)

	return nil
}

// watchFiles monitorea cambios en archivos
func (hrm *HotReloadManager) watchFiles() {
	for {
		select {
		case event, ok := <-hrm.watcher.Events:
			if !ok {
				return
			}

			// Filtrar eventos relevantes
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				hrm.handleFileEvent(event)
			}

		case err, ok := <-hrm.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error en watcher: %v", err)

		case <-hrm.stopChan:
			return
		}
	}
}

// handleFileEvent maneja eventos de archivo
func (hrm *HotReloadManager) handleFileEvent(event fsnotify.Event) {
	hrm.mutex.RLock()
	apiName, exists := hrm.watchedFiles[event.Name]
	hrm.mutex.RUnlock()

	if !exists {
		return
	}

	// Verificar debounce
	now := time.Now()
	if lastReload, exists := hrm.lastReload[event.Name]; exists {
		if now.Sub(lastReload) < hrm.debounceTime {
			return // Ignorar evento por debounce
		}
	}

	// Crear evento de recarga
	reloadEvent := ReloadEvent{
		FilePath:  event.Name,
		APIName:   apiName,
		EventType: event.Op.String(),
		Timestamp: now,
	}

	// Enviar evento para procesamiento
	select {
	case hrm.reloadChan <- reloadEvent:
		hrm.lastReload[event.Name] = now
	default:
		log.Printf("Canal de recarga lleno, ignorando evento: %s", event.Name)
	}
}

// processReloadEvents procesa eventos de recarga
func (hrm *HotReloadManager) processReloadEvents() {
	for {
		select {
		case event := <-hrm.reloadChan:
			hrm.processReloadEvent(event)

		case <-hrm.stopChan:
			return
		}
	}
}

// processReloadEvent procesa un evento de recarga específico
func (hrm *HotReloadManager) processReloadEvent(event ReloadEvent) {
	log.Printf("Procesando recarga de configuración: %s (API: %s)", event.FilePath, event.APIName)

	// Validar archivo antes de recargar
	if err := hrm.validateConfigFile(event.FilePath, event.APIName); err != nil {
		log.Printf("Error validando configuración %s: %v", event.FilePath, err)
		return
	}

	// Recargar configuración según el tipo
	var err error
	switch event.APIName {
	case "main":
		err = hrm.reloadMainConfig(event.FilePath)
	default:
		err = hrm.reloadAPIConfig(event.FilePath, event.APIName)
	}

	if err != nil {
		log.Printf("Error recargando configuración %s: %v", event.FilePath, err)
		return
	}

	log.Printf("Configuración recargada exitosamente: %s (API: %s)", event.FilePath, event.APIName)
}

// validateConfigFile valida un archivo de configuración
func (hrm *HotReloadManager) validateConfigFile(filePath, apiName string) error {
	// Leer archivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %w", err)
	}

	// Validar sintaxis YAML
	var temp interface{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("error de sintaxis YAML: %w", err)
	}

	// Validaciones específicas por API
	switch apiName {
	case "main":
		return hrm.validateMainConfig(data)
	case "pos":
		return hrm.validatePOSConfig(data)
	case "sync":
		return hrm.validateSyncConfig(data)
	case "labels":
		return hrm.validateLabelsConfig(data)
	case "reports":
		return hrm.validateReportsConfig(data)
	}

	return nil
}

// validateMainConfig valida la configuración principal
func (hrm *HotReloadManager) validateMainConfig(data []byte) error {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración principal: %w", err)
	}

	// Validaciones básicas
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("puerto del servidor inválido: %d", config.Server.Port)
	}

	if config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret no puede estar vacío")
	}

	return nil
}

// validatePOSConfig valida la configuración del API POS
func (hrm *HotReloadManager) validatePOSConfig(data []byte) error {
	var config POSConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración POS: %w", err)
	}

	// Validaciones específicas
	if config.MaxConcurrentUsers <= 0 {
		return fmt.Errorf("max_concurrent_users debe ser mayor a 0")
	}

	if config.MaxVentaItems <= 0 {
		return fmt.Errorf("max_venta_items debe ser mayor a 0")
	}

	return nil
}

// validateSyncConfig valida la configuración del API Sync
func (hrm *HotReloadManager) validateSyncConfig(data []byte) error {
	var config SyncConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración Sync: %w", err)
	}

	// Validaciones específicas
	if config.MaxConcurrentSyncs <= 0 {
		return fmt.Errorf("max_concurrent_syncs debe ser mayor a 0")
	}

	if config.BatchSize <= 0 {
		return fmt.Errorf("batch_size debe ser mayor a 0")
	}

	validModes := []string{"manual", "auto_server", "auto_client"}
	validMode := false
	for _, mode := range validModes {
		if config.ConflictResolutionMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("conflict_resolution_mode inválido: %s", config.ConflictResolutionMode)
	}

	return nil
}

// validateLabelsConfig valida la configuración del API Labels
func (hrm *HotReloadManager) validateLabelsConfig(data []byte) error {
	var config LabelsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración Labels: %w", err)
	}

	// Validaciones específicas
	if config.MaxConcurrentJobs <= 0 {
		return fmt.Errorf("max_concurrent_jobs debe ser mayor a 0")
	}

	if config.MaxLabelsPerBatch <= 0 {
		return fmt.Errorf("max_labels_per_batch debe ser mayor a 0")
	}

	validFormats := []string{"pdf", "png", "jpg", "svg"}
	validFormat := false
	for _, format := range validFormats {
		if config.DefaultLabelFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("default_label_format inválido: %s", config.DefaultLabelFormat)
	}

	return nil
}

// validateReportsConfig valida la configuración del API Reports
func (hrm *HotReloadManager) validateReportsConfig(data []byte) error {
	var config ReportsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parseando configuración Reports: %w", err)
	}

	// Validaciones específicas
	if config.MaxConcurrentReports <= 0 {
		return fmt.Errorf("max_concurrent_reports debe ser mayor a 0")
	}

	if config.MaxReportSize <= 0 {
		return fmt.Errorf("max_report_size debe ser mayor a 0")
	}

	validFormats := []string{"pdf", "excel", "csv", "json", "html", "xml"}
	validFormat := false
	for _, format := range validFormats {
		if config.DefaultFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("default_format inválido: %s", config.DefaultFormat)
	}

	return nil
}

// reloadMainConfig recarga la configuración principal
func (hrm *HotReloadManager) reloadMainConfig(filePath string) error {
	// Crear nuevo ConfigManager temporal para validar
	tempManager := NewConfigManager(filePath)
	if err := tempManager.LoadConfig(); err != nil {
		return fmt.Errorf("error cargando configuración principal: %w", err)
	}

	// Validar configuración
	if err := tempManager.ValidateConfig(); err != nil {
		return fmt.Errorf("configuración principal inválida: %w", err)
	}

	// Aplicar configuración al manager principal
	newConfig := tempManager.GetConfig()
	if err := hrm.configManager.UpdateConfig(newConfig); err != nil {
		return fmt.Errorf("error actualizando configuración principal: %w", err)
	}

	return nil
}

// reloadAPIConfig recarga la configuración de una API específica
func (hrm *HotReloadManager) reloadAPIConfig(filePath, apiName string) error {
	// Leer archivo de configuración específica
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo configuración de API: %w", err)
	}

	// Parsear según el tipo de API
	var apiConfig interface{}
	switch apiName {
	case "pos":
		var config POSConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("error parseando configuración POS: %w", err)
		}
		apiConfig = config

	case "sync":
		var config SyncConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("error parseando configuración Sync: %w", err)
		}
		apiConfig = config

	case "labels":
		var config LabelsConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("error parseando configuración Labels: %w", err)
		}
		apiConfig = config

	case "reports":
		var config ReportsConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("error parseando configuración Reports: %w", err)
		}
		apiConfig = config

	default:
		return fmt.Errorf("API desconocida: %s", apiName)
	}

	// Actualizar configuración en el manager principal
	if err := hrm.configManager.UpdateAPIConfig(apiName, apiConfig); err != nil {
		return fmt.Errorf("error actualizando configuración de API %s: %w", apiName, err)
	}

	return nil
}

// GetWatchedFiles retorna la lista de archivos monitoreados
func (hrm *HotReloadManager) GetWatchedFiles() map[string]string {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()

	result := make(map[string]string)
	for path, apiName := range hrm.watchedFiles {
		result[path] = apiName
	}

	return result
}

// GetReloadStats retorna estadísticas de recarga
func (hrm *HotReloadManager) GetReloadStats() map[string]interface{} {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()

	stats := map[string]interface{}{
		"watched_files":     len(hrm.watchedFiles),
		"last_reload_times": make(map[string]time.Time),
		"debounce_time":     hrm.debounceTime,
		"queue_size":        len(hrm.reloadChan),
	}

	for path, lastTime := range hrm.lastReload {
		if apiName, exists := hrm.watchedFiles[path]; exists {
			stats["last_reload_times"].(map[string]time.Time)[apiName] = lastTime
		}
	}

	return stats
}

// SetDebounceTime configura el tiempo de debounce
func (hrm *HotReloadManager) SetDebounceTime(duration time.Duration) {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()
	hrm.debounceTime = duration
}

// ForceReload fuerza la recarga de un archivo específico
func (hrm *HotReloadManager) ForceReload(filePath string) error {
	hrm.mutex.RLock()
	apiName, exists := hrm.watchedFiles[filePath]
	hrm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("archivo no está siendo monitoreado: %s", filePath)
	}

	// Crear evento de recarga forzada
	reloadEvent := ReloadEvent{
		FilePath:  filePath,
		APIName:   apiName,
		EventType: "FORCE_RELOAD",
		Timestamp: time.Now(),
	}

	// Procesar inmediatamente
	hrm.processReloadEvent(reloadEvent)
	return nil
}

// IsHealthy verifica si el sistema de recarga está funcionando correctamente
func (hrm *HotReloadManager) IsHealthy() bool {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()

	// Verificar que el watcher esté funcionando
	if hrm.watcher == nil {
		return false
	}

	// Verificar que haya archivos monitoreados
	if len(hrm.watchedFiles) == 0 {
		return false
	}

	// Verificar que el canal no esté bloqueado
	if len(hrm.reloadChan) >= cap(hrm.reloadChan) {
		return false
	}

	return true
}
