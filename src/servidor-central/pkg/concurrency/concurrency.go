// Package concurrency proporciona utilidades avanzadas de concurrencia
// con notación húngara y prevención de race conditions
package concurrency

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"ferre-pos-servidor-central/pkg/errors"
	"ferre-pos-servidor-central/pkg/logger"
)

// structWorkerPool pool de workers para procesamiento concurrente
type structWorkerPool struct {
	IntWorkers       int
	ChanJobs         chan structJob
	ChanResults      chan structJobResult
	ChanStop         chan struct{}
	WaitGroupWorkers sync.WaitGroup
	PtrLogger        logger.InterfaceLogger
	BoolStarted      bool
	MutexState       sync.RWMutex
	Int64JobsTotal   int64
	Int64JobsSuccess int64
	Int64JobsError   int64
}

// structJob representa un trabajo a procesar
type structJob struct {
	UuidID      string
	FuncTask    func(context.Context) (interface{}, error)
	CtxContext  context.Context
	TimeCreated time.Time
	IntPriority int
	MapMetadata map[string]interface{}
}

// structJobResult resultado de un trabajo
type structJobResult struct {
	UuidJobID     string
	ObjResult     interface{}
	ErrError      error
	TimeDuration  time.Duration
	TimeCompleted time.Time
}

// interfaceWorkerPool interfaz del pool de workers
type interfaceWorkerPool interface {
	Start() error
	Stop() error
	SubmitJob(structJob) error
	GetResults() <-chan structJobResult
	GetStats() structWorkerPoolStats
	IsRunning() bool
}

// structWorkerPoolStats estadísticas del pool
type structWorkerPoolStats struct {
	IntWorkers       int     `json:"workers"`
	Int64JobsTotal   int64   `json:"jobs_total"`
	Int64JobsSuccess int64   `json:"jobs_success"`
	Int64JobsError   int64   `json:"jobs_error"`
	Int64JobsPending int64   `json:"jobs_pending"`
	FltSuccessRate   float64 `json:"success_rate"`
}

// structSafeMap mapa thread-safe
type structSafeMap struct {
	MapData map[string]interface{}
	MutexRW sync.RWMutex
}

// interfaceSafeMap interfaz del mapa thread-safe
type interfaceSafeMap interface {
	Set(string, interface{})
	Get(string) (interface{}, bool)
	Delete(string)
	Keys() []string
	Size() int
	Clear()
	ForEach(func(string, interface{}))
}

// structSafeCounter contador thread-safe
type structSafeCounter struct {
	Int64Value int64
}

// interfaceSafeCounter interfaz del contador thread-safe
type interfaceSafeCounter interface {
	Increment() int64
	Decrement() int64
	Add(int64) int64
	Get() int64
	Set(int64)
	Reset()
}

// structRateLimiter limitador de velocidad
type structRateLimiter struct {
	IntLimit       int
	DurationWindow time.Duration
	MapRequests    map[string][]time.Time
	MutexRequests  sync.RWMutex
	ChanCleanup    chan struct{}
	BoolRunning    bool
}

// interfaceRateLimiter interfaz del limitador de velocidad
type interfaceRateLimiter interface {
	Allow(string) bool
	GetRemaining(string) int
	Reset(string)
	Stop()
}

// structCircuitBreaker circuit breaker para prevenir cascadas de fallos
type structCircuitBreaker struct {
	StrName              string
	IntMaxFailures       int
	DurationTimeout      time.Duration
	DurationResetTimeout time.Duration
	EnumState            enumCircuitBreakerState
	Int64FailureCount    int64
	TimeLastFailure      time.Time
	MutexState           sync.RWMutex
	PtrLogger            logger.InterfaceLogger
}

// enumCircuitBreakerState estados del circuit breaker
type enumCircuitBreakerState int

const (
	EnumCircuitBreakerStateClosed enumCircuitBreakerState = iota
	EnumCircuitBreakerStateOpen
	EnumCircuitBreakerStateHalfOpen
)

// interfaceCircuitBreaker interfaz del circuit breaker
type interfaceCircuitBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
	GetState() enumCircuitBreakerState
	GetFailureCount() int64
	Reset()
}

// NewWorkerPool crea un nuevo pool de workers
func NewWorkerPool(intWorkers int, intBufferSize int, ptrLogger logger.InterfaceLogger) interfaceWorkerPool {
	if intWorkers <= 0 {
		intWorkers = runtime.NumCPU()
	}

	if intBufferSize <= 0 {
		intBufferSize = intWorkers * 2
	}

	return &structWorkerPool{
		IntWorkers:  intWorkers,
		ChanJobs:    make(chan structJob, intBufferSize),
		ChanResults: make(chan structJobResult, intBufferSize),
		ChanStop:    make(chan struct{}),
		PtrLogger:   ptrLogger,
	}
}

// Start inicia el pool de workers
func (ptrPool *structWorkerPool) Start() error {
	ptrPool.MutexState.Lock()
	defer ptrPool.MutexState.Unlock()

	if ptrPool.BoolStarted {
		return errors.New("Worker pool ya está iniciado")
	}

	ptrPool.BoolStarted = true

	// Iniciar workers
	for intI := 0; intI < ptrPool.IntWorkers; intI++ {
		ptrPool.WaitGroupWorkers.Add(1)
		go ptrPool.worker(intI)
	}

	ptrPool.PtrLogger.Info("Worker pool iniciado",
		zap.Int("workers", ptrPool.IntWorkers),
		zap.Int("buffer_size", cap(ptrPool.ChanJobs)),
	)

	return nil
}

// Stop detiene el pool de workers
func (ptrPool *structWorkerPool) Stop() error {
	ptrPool.MutexState.Lock()
	defer ptrPool.MutexState.Unlock()

	if !ptrPool.BoolStarted {
		return errors.New("Worker pool no está iniciado")
	}

	close(ptrPool.ChanStop)
	ptrPool.WaitGroupWorkers.Wait()
	close(ptrPool.ChanJobs)
	close(ptrPool.ChanResults)

	ptrPool.BoolStarted = false

	ptrPool.PtrLogger.Info("Worker pool detenido",
		zap.Int64("jobs_processed", atomic.LoadInt64(&ptrPool.Int64JobsTotal)),
	)

	return nil
}

// SubmitJob envía un trabajo al pool
func (ptrPool *structWorkerPool) SubmitJob(structJobData structJob) error {
	ptrPool.MutexState.RLock()
	defer ptrPool.MutexState.RUnlock()

	if !ptrPool.BoolStarted {
		return errors.New("Worker pool no está iniciado")
	}

	select {
	case ptrPool.ChanJobs <- structJobData:
		atomic.AddInt64(&ptrPool.Int64JobsTotal, 1)
		return nil
	default:
		return errors.New("Worker pool está lleno")
	}
}

// GetResults obtiene el canal de resultados
func (ptrPool *structWorkerPool) GetResults() <-chan structJobResult {
	return ptrPool.ChanResults
}

// GetStats obtiene estadísticas del pool
func (ptrPool *structWorkerPool) GetStats() structWorkerPoolStats {
	int64Total := atomic.LoadInt64(&ptrPool.Int64JobsTotal)
	int64Success := atomic.LoadInt64(&ptrPool.Int64JobsSuccess)
	int64Error := atomic.LoadInt64(&ptrPool.Int64JobsError)
	int64Pending := int64(len(ptrPool.ChanJobs))

	var fltSuccessRate float64
	if int64Total > 0 {
		fltSuccessRate = float64(int64Success) / float64(int64Total) * 100
	}

	return structWorkerPoolStats{
		IntWorkers:       ptrPool.IntWorkers,
		Int64JobsTotal:   int64Total,
		Int64JobsSuccess: int64Success,
		Int64JobsError:   int64Error,
		Int64JobsPending: int64Pending,
		FltSuccessRate:   fltSuccessRate,
	}
}

// IsRunning verifica si el pool está ejecutándose
func (ptrPool *structWorkerPool) IsRunning() bool {
	ptrPool.MutexState.RLock()
	defer ptrPool.MutexState.RUnlock()
	return ptrPool.BoolStarted
}

// worker función del worker
func (ptrPool *structWorkerPool) worker(intWorkerID int) {
	defer ptrPool.WaitGroupWorkers.Done()

	ptrWorkerLogger := ptrPool.PtrLogger.With(zap.Int("worker_id", intWorkerID))
	ptrWorkerLogger.Debug("Worker iniciado")

	for {
		select {
		case structJobData := <-ptrPool.ChanJobs:
			ptrPool.processJob(structJobData, ptrWorkerLogger)
		case <-ptrPool.ChanStop:
			ptrWorkerLogger.Debug("Worker detenido")
			return
		}
	}
}

// processJob procesa un trabajo
func (ptrPool *structWorkerPool) processJob(structJobData structJob, ptrLogger logger.InterfaceLogger) {
	timeStart := time.Now()

	ptrLogger.Debug("Procesando trabajo",
		zap.String("job_id", structJobData.UuidID),
		zap.Int("priority", structJobData.IntPriority),
	)

	objResult, err := structJobData.FuncTask(structJobData.CtxContext)
	timeDuration := time.Since(timeStart)

	structResult := structJobResult{
		UuidJobID:     structJobData.UuidID,
		ObjResult:     objResult,
		ErrError:      err,
		TimeDuration:  timeDuration,
		TimeCompleted: time.Now(),
	}

	if err != nil {
		atomic.AddInt64(&ptrPool.Int64JobsError, 1)
		ptrLogger.Error("Trabajo falló",
			zap.String("job_id", structJobData.UuidID),
			zap.Error(err),
			zap.Duration("duration", timeDuration),
		)
	} else {
		atomic.AddInt64(&ptrPool.Int64JobsSuccess, 1)
		ptrLogger.Debug("Trabajo completado",
			zap.String("job_id", structJobData.UuidID),
			zap.Duration("duration", timeDuration),
		)
	}

	select {
	case ptrPool.ChanResults <- structResult:
	default:
		ptrLogger.Warn("Canal de resultados lleno, descartando resultado",
			zap.String("job_id", structJobData.UuidID),
		)
	}
}

// NewSafeMap crea un nuevo mapa thread-safe
func NewSafeMap() interfaceSafeMap {
	return &structSafeMap{
		MapData: make(map[string]interface{}),
	}
}

// Set establece un valor en el mapa
func (ptrMap *structSafeMap) Set(strKey string, objValue interface{}) {
	ptrMap.MutexRW.Lock()
	defer ptrMap.MutexRW.Unlock()
	ptrMap.MapData[strKey] = objValue
}

// Get obtiene un valor del mapa
func (ptrMap *structSafeMap) Get(strKey string) (interface{}, bool) {
	ptrMap.MutexRW.RLock()
	defer ptrMap.MutexRW.RUnlock()
	objValue, boolExists := ptrMap.MapData[strKey]
	return objValue, boolExists
}

// Delete elimina un valor del mapa
func (ptrMap *structSafeMap) Delete(strKey string) {
	ptrMap.MutexRW.Lock()
	defer ptrMap.MutexRW.Unlock()
	delete(ptrMap.MapData, strKey)
}

// Keys obtiene todas las claves del mapa
func (ptrMap *structSafeMap) Keys() []string {
	ptrMap.MutexRW.RLock()
	defer ptrMap.MutexRW.RUnlock()

	arrKeys := make([]string, 0, len(ptrMap.MapData))
	for strKey := range ptrMap.MapData {
		arrKeys = append(arrKeys, strKey)
	}
	return arrKeys
}

// Size obtiene el tamaño del mapa
func (ptrMap *structSafeMap) Size() int {
	ptrMap.MutexRW.RLock()
	defer ptrMap.MutexRW.RUnlock()
	return len(ptrMap.MapData)
}

// Clear limpia el mapa
func (ptrMap *structSafeMap) Clear() {
	ptrMap.MutexRW.Lock()
	defer ptrMap.MutexRW.Unlock()
	ptrMap.MapData = make(map[string]interface{})
}

// ForEach itera sobre el mapa
func (ptrMap *structSafeMap) ForEach(funcCallback func(string, interface{})) {
	ptrMap.MutexRW.RLock()
	defer ptrMap.MutexRW.RUnlock()

	for strKey, objValue := range ptrMap.MapData {
		funcCallback(strKey, objValue)
	}
}

// NewSafeCounter crea un nuevo contador thread-safe
func NewSafeCounter() interfaceSafeCounter {
	return &structSafeCounter{}
}

// Increment incrementa el contador
func (ptrCounter *structSafeCounter) Increment() int64 {
	return atomic.AddInt64(&ptrCounter.Int64Value, 1)
}

// Decrement decrementa el contador
func (ptrCounter *structSafeCounter) Decrement() int64 {
	return atomic.AddInt64(&ptrCounter.Int64Value, -1)
}

// Add agrega un valor al contador
func (ptrCounter *structSafeCounter) Add(int64Delta int64) int64 {
	return atomic.AddInt64(&ptrCounter.Int64Value, int64Delta)
}

// Get obtiene el valor del contador
func (ptrCounter *structSafeCounter) Get() int64 {
	return atomic.LoadInt64(&ptrCounter.Int64Value)
}

// Set establece el valor del contador
func (ptrCounter *structSafeCounter) Set(int64Value int64) {
	atomic.StoreInt64(&ptrCounter.Int64Value, int64Value)
}

// Reset resetea el contador a cero
func (ptrCounter *structSafeCounter) Reset() {
	atomic.StoreInt64(&ptrCounter.Int64Value, 0)
}

// NewRateLimiter crea un nuevo limitador de velocidad
func NewRateLimiter(intLimit int, durationWindow time.Duration) interfaceRateLimiter {
	ptrLimiter := &structRateLimiter{
		IntLimit:       intLimit,
		DurationWindow: durationWindow,
		MapRequests:    make(map[string][]time.Time),
		ChanCleanup:    make(chan struct{}),
		BoolRunning:    true,
	}

	// Iniciar limpieza periódica
	go ptrLimiter.cleanup()

	return ptrLimiter
}

// Allow verifica si una request está permitida
func (ptrLimiter *structRateLimiter) Allow(strKey string) bool {
	ptrLimiter.MutexRequests.Lock()
	defer ptrLimiter.MutexRequests.Unlock()

	timeNow := time.Now()
	timeWindowStart := timeNow.Add(-ptrLimiter.DurationWindow)

	// Obtener requests existentes
	arrRequests, boolExists := ptrLimiter.MapRequests[strKey]
	if !boolExists {
		arrRequests = make([]time.Time, 0)
	}

	// Filtrar requests dentro de la ventana
	arrValidRequests := make([]time.Time, 0)
	for _, timeRequest := range arrRequests {
		if timeRequest.After(timeWindowStart) {
			arrValidRequests = append(arrValidRequests, timeRequest)
		}
	}

	// Verificar límite
	if len(arrValidRequests) >= ptrLimiter.IntLimit {
		ptrLimiter.MapRequests[strKey] = arrValidRequests
		return false
	}

	// Agregar nueva request
	arrValidRequests = append(arrValidRequests, timeNow)
	ptrLimiter.MapRequests[strKey] = arrValidRequests

	return true
}

// GetRemaining obtiene el número de requests restantes
func (ptrLimiter *structRateLimiter) GetRemaining(strKey string) int {
	ptrLimiter.MutexRequests.RLock()
	defer ptrLimiter.MutexRequests.RUnlock()

	arrRequests, boolExists := ptrLimiter.MapRequests[strKey]
	if !boolExists {
		return ptrLimiter.IntLimit
	}

	timeNow := time.Now()
	timeWindowStart := timeNow.Add(-ptrLimiter.DurationWindow)

	intValidRequests := 0
	for _, timeRequest := range arrRequests {
		if timeRequest.After(timeWindowStart) {
			intValidRequests++
		}
	}

	intRemaining := ptrLimiter.IntLimit - intValidRequests
	if intRemaining < 0 {
		intRemaining = 0
	}

	return intRemaining
}

// Reset resetea el limitador para una clave
func (ptrLimiter *structRateLimiter) Reset(strKey string) {
	ptrLimiter.MutexRequests.Lock()
	defer ptrLimiter.MutexRequests.Unlock()
	delete(ptrLimiter.MapRequests, strKey)
}

// Stop detiene el limitador
func (ptrLimiter *structRateLimiter) Stop() {
	ptrLimiter.BoolRunning = false
	close(ptrLimiter.ChanCleanup)
}

// cleanup limpia requests expiradas
func (ptrLimiter *structRateLimiter) cleanup() {
	ptrTicker := time.NewTicker(ptrLimiter.DurationWindow)
	defer ptrTicker.Stop()

	for {
		select {
		case <-ptrTicker.C:
			ptrLimiter.cleanupExpiredRequests()
		case <-ptrLimiter.ChanCleanup:
			return
		}
	}
}

// cleanupExpiredRequests limpia requests expiradas
func (ptrLimiter *structRateLimiter) cleanupExpiredRequests() {
	ptrLimiter.MutexRequests.Lock()
	defer ptrLimiter.MutexRequests.Unlock()

	timeNow := time.Now()
	timeWindowStart := timeNow.Add(-ptrLimiter.DurationWindow)

	for strKey, arrRequests := range ptrLimiter.MapRequests {
		arrValidRequests := make([]time.Time, 0)
		for _, timeRequest := range arrRequests {
			if timeRequest.After(timeWindowStart) {
				arrValidRequests = append(arrValidRequests, timeRequest)
			}
		}

		if len(arrValidRequests) == 0 {
			delete(ptrLimiter.MapRequests, strKey)
		} else {
			ptrLimiter.MapRequests[strKey] = arrValidRequests
		}
	}
}

// NewCircuitBreaker crea un nuevo circuit breaker
func NewCircuitBreaker(strName string, intMaxFailures int, durationTimeout, durationResetTimeout time.Duration, ptrLogger logger.InterfaceLogger) interfaceCircuitBreaker {
	return &structCircuitBreaker{
		StrName:              strName,
		IntMaxFailures:       intMaxFailures,
		DurationTimeout:      durationTimeout,
		DurationResetTimeout: durationResetTimeout,
		EnumState:            EnumCircuitBreakerStateClosed,
		PtrLogger:            ptrLogger,
	}
}

// Execute ejecuta una función con circuit breaker
func (ptrBreaker *structCircuitBreaker) Execute(funcOperation func() (interface{}, error)) (interface{}, error) {
	ptrBreaker.MutexState.Lock()
	defer ptrBreaker.MutexState.Unlock()

	switch ptrBreaker.EnumState {
	case EnumCircuitBreakerStateOpen:
		if time.Since(ptrBreaker.TimeLastFailure) > ptrBreaker.DurationResetTimeout {
			ptrBreaker.EnumState = EnumCircuitBreakerStateHalfOpen
			ptrBreaker.PtrLogger.Info("Circuit breaker cambiando a half-open",
				zap.String("name", ptrBreaker.StrName),
			)
		} else {
			return nil, errors.NewTimeout("Circuit breaker está abierto", "CIRCUIT_BREAKER_OPEN")
		}

	case EnumCircuitBreakerStateHalfOpen:
		// Permitir una request de prueba
	}

	// Ejecutar operación
	objResult, err := funcOperation()

	if err != nil {
		ptrBreaker.recordFailure()
		return nil, err
	}

	ptrBreaker.recordSuccess()
	return objResult, nil
}

// GetState obtiene el estado actual
func (ptrBreaker *structCircuitBreaker) GetState() enumCircuitBreakerState {
	ptrBreaker.MutexState.RLock()
	defer ptrBreaker.MutexState.RUnlock()
	return ptrBreaker.EnumState
}

// GetFailureCount obtiene el número de fallos
func (ptrBreaker *structCircuitBreaker) GetFailureCount() int64 {
	return atomic.LoadInt64(&ptrBreaker.Int64FailureCount)
}

// Reset resetea el circuit breaker
func (ptrBreaker *structCircuitBreaker) Reset() {
	ptrBreaker.MutexState.Lock()
	defer ptrBreaker.MutexState.Unlock()

	ptrBreaker.EnumState = EnumCircuitBreakerStateClosed
	atomic.StoreInt64(&ptrBreaker.Int64FailureCount, 0)

	ptrBreaker.PtrLogger.Info("Circuit breaker reseteado",
		zap.String("name", ptrBreaker.StrName),
	)
}

// recordFailure registra un fallo
func (ptrBreaker *structCircuitBreaker) recordFailure() {
	int64Failures := atomic.AddInt64(&ptrBreaker.Int64FailureCount, 1)
	ptrBreaker.TimeLastFailure = time.Now()

	if int64Failures >= int64(ptrBreaker.IntMaxFailures) {
		ptrBreaker.EnumState = EnumCircuitBreakerStateOpen
		ptrBreaker.PtrLogger.Warn("Circuit breaker abierto por exceso de fallos",
			zap.String("name", ptrBreaker.StrName),
			zap.Int64("failures", int64Failures),
		)
	}
}

// recordSuccess registra un éxito
func (ptrBreaker *structCircuitBreaker) recordSuccess() {
	if ptrBreaker.EnumState == EnumCircuitBreakerStateHalfOpen {
		ptrBreaker.EnumState = EnumCircuitBreakerStateClosed
		atomic.StoreInt64(&ptrBreaker.Int64FailureCount, 0)
		ptrBreaker.PtrLogger.Info("Circuit breaker cerrado después de éxito",
			zap.String("name", ptrBreaker.StrName),
		)
	}
}

// Funciones de utilidad

// RunWithTimeout ejecuta una función con timeout
func RunWithTimeout(funcOperation func() (interface{}, error), durationTimeout time.Duration) (interface{}, error) {
	chanResult := make(chan interface{}, 1)
	chanError := make(chan error, 1)

	go func() {
		objResult, err := funcOperation()
		if err != nil {
			chanError <- err
		} else {
			chanResult <- objResult
		}
	}()

	select {
	case objResult := <-chanResult:
		return objResult, nil
	case err := <-chanError:
		return nil, err
	case <-time.After(durationTimeout):
		return nil, errors.NewTimeout("Operación excedió el timeout", "OPERATION_TIMEOUT")
	}
}

// Retry ejecuta una función con reintentos
func Retry(funcOperation func() error, intMaxRetries int, durationDelay time.Duration) error {
	var errLast error

	for intI := 0; intI <= intMaxRetries; intI++ {
		if intI > 0 {
			time.Sleep(durationDelay)
		}

		if err := funcOperation(); err != nil {
			errLast = err
			continue
		}

		return nil
	}

	return errors.Wrapf(errLast, "operación falló después de %d reintentos", intMaxRetries)
}

// RetryWithBackoff ejecuta una función con reintentos y backoff exponencial
func RetryWithBackoff(funcOperation func() error, intMaxRetries int, durationInitialDelay time.Duration) error {
	var errLast error
	durationDelay := durationInitialDelay

	for intI := 0; intI <= intMaxRetries; intI++ {
		if intI > 0 {
			time.Sleep(durationDelay)
			durationDelay *= 2 // Backoff exponencial
		}

		if err := funcOperation(); err != nil {
			errLast = err
			continue
		}

		return nil
	}

	return errors.Wrapf(errLast, "operación falló después de %d reintentos con backoff", intMaxRetries)
}
