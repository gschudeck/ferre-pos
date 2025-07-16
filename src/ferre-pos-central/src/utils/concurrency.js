/**
 * Utilidades de Concurrencia - Sistema Ferre-POS
 * 
 * Proporciona herramientas para optimizar operaciones concurrentes,
 * incluyendo workers, pools de tareas y gestión de recursos.
 */

const { Worker, isMainThread, parentPort, workerData } = require('worker_threads')
const { EventEmitter } = require('events')
const path = require('path')
const logger = require('./logger')

/**
 * Pool de Workers para operaciones CPU-intensivas
 */
class WorkerPool extends EventEmitter {
  constructor(workerScript, poolSize = require('os').cpus().length) {
    super()
    this.workerScript = workerScript
    this.poolSize = poolSize
    this.workers = []
    this.queue = []
    this.activeJobs = new Map()
    this.jobIdCounter = 0
    
    this.initializeWorkers()
  }

  /**
   * Inicializa el pool de workers
   */
  initializeWorkers() {
    for (let i = 0; i < this.poolSize; i++) {
      this.createWorker()
    }
    
    logger.info(`Worker pool inicializado con ${this.poolSize} workers`, {
      script: this.workerScript,
      poolSize: this.poolSize
    })
  }

  /**
   * Crea un nuevo worker
   */
  createWorker() {
    const worker = new Worker(this.workerScript)
    
    worker.on('message', (result) => {
      const { jobId, success, data, error } = result
      const job = this.activeJobs.get(jobId)
      
      if (job) {
        this.activeJobs.delete(jobId)
        
        if (success) {
          job.resolve(data)
        } else {
          job.reject(new Error(error))
        }
        
        // Procesar siguiente trabajo en cola
        this.processNextJob(worker)
      }
    })

    worker.on('error', (error) => {
      logger.error('Error en worker:', error)
      this.emit('workerError', error)
      
      // Reemplazar worker defectuoso
      this.replaceWorker(worker)
    })

    worker.on('exit', (code) => {
      if (code !== 0) {
        logger.warn(`Worker terminó con código ${code}`)
        this.replaceWorker(worker)
      }
    })

    worker.isAvailable = true
    this.workers.push(worker)
    
    return worker
  }

  /**
   * Reemplaza un worker defectuoso
   */
  replaceWorker(oldWorker) {
    const index = this.workers.indexOf(oldWorker)
    if (index !== -1) {
      this.workers.splice(index, 1)
      oldWorker.terminate()
      this.createWorker()
    }
  }

  /**
   * Ejecuta una tarea en el pool
   */
  async execute(data, timeout = 30000) {
    return new Promise((resolve, reject) => {
      const jobId = ++this.jobIdCounter
      const job = {
        id: jobId,
        data,
        resolve,
        reject,
        timestamp: Date.now()
      }

      // Configurar timeout
      const timeoutId = setTimeout(() => {
        this.activeJobs.delete(jobId)
        reject(new Error(`Worker timeout después de ${timeout}ms`))
      }, timeout)

      job.timeoutId = timeoutId
      this.activeJobs.set(jobId, job)

      // Buscar worker disponible
      const availableWorker = this.workers.find(w => w.isAvailable)
      
      if (availableWorker) {
        this.assignJob(availableWorker, job)
      } else {
        // Agregar a cola
        this.queue.push(job)
      }
    })
  }

  /**
   * Asigna un trabajo a un worker
   */
  assignJob(worker, job) {
    worker.isAvailable = false
    worker.postMessage({
      jobId: job.id,
      data: job.data
    })
  }

  /**
   * Procesa el siguiente trabajo en cola
   */
  processNextJob(worker) {
    worker.isAvailable = true
    
    if (this.queue.length > 0) {
      const nextJob = this.queue.shift()
      this.assignJob(worker, nextJob)
    }
  }

  /**
   * Cierra el pool de workers
   */
  async close() {
    // Esperar trabajos activos
    while (this.activeJobs.size > 0) {
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    // Terminar todos los workers
    await Promise.all(
      this.workers.map(worker => worker.terminate())
    )

    this.workers = []
    logger.info('Worker pool cerrado')
  }

  /**
   * Obtiene estadísticas del pool
   */
  getStats() {
    return {
      poolSize: this.poolSize,
      activeJobs: this.activeJobs.size,
      queuedJobs: this.queue.length,
      availableWorkers: this.workers.filter(w => w.isAvailable).length
    }
  }
}

/**
 * Gestor de tareas concurrentes con límite
 */
class ConcurrencyManager {
  constructor(maxConcurrency = 10) {
    this.maxConcurrency = maxConcurrency
    this.running = 0
    this.queue = []
  }

  /**
   * Ejecuta una función con límite de concurrencia
   */
  async execute(fn) {
    return new Promise((resolve, reject) => {
      const task = {
        fn,
        resolve,
        reject
      }

      if (this.running < this.maxConcurrency) {
        this.runTask(task)
      } else {
        this.queue.push(task)
      }
    })
  }

  /**
   * Ejecuta una tarea
   */
  async runTask(task) {
    this.running++
    
    try {
      const result = await task.fn()
      task.resolve(result)
    } catch (error) {
      task.reject(error)
    } finally {
      this.running--
      this.processQueue()
    }
  }

  /**
   * Procesa la cola de tareas
   */
  processQueue() {
    if (this.queue.length > 0 && this.running < this.maxConcurrency) {
      const nextTask = this.queue.shift()
      this.runTask(nextTask)
    }
  }

  /**
   * Ejecuta múltiples tareas en lotes
   */
  async executeBatch(tasks, batchSize = this.maxConcurrency) {
    const results = []
    
    for (let i = 0; i < tasks.length; i += batchSize) {
      const batch = tasks.slice(i, i + batchSize)
      const batchPromises = batch.map(task => this.execute(task))
      const batchResults = await Promise.allSettled(batchPromises)
      results.push(...batchResults)
    }
    
    return results
  }
}

/**
 * Utilidades para operaciones paralelas comunes
 */
class ParallelUtils {
  /**
   * Ejecuta consultas de base de datos en paralelo
   */
  static async parallelQueries(database, queries) {
    const startTime = Date.now()
    
    try {
      const results = await Promise.all(
        queries.map(async ({ query, params, name }) => {
          const queryStart = Date.now()
          const result = await database.query(query, params)
          const queryTime = Date.now() - queryStart
          
          logger.debug(`Query paralela completada: ${name}`, {
            duration: queryTime,
            rowCount: result.rowCount
          })
          
          return { name, result, duration: queryTime }
        })
      )
      
      const totalTime = Date.now() - startTime
      logger.info('Consultas paralelas completadas', {
        totalQueries: queries.length,
        totalDuration: totalTime,
        averageDuration: totalTime / queries.length
      })
      
      return results
    } catch (error) {
      logger.error('Error en consultas paralelas:', error)
      throw error
    }
  }

  /**
   * Procesa arrays grandes en chunks paralelos
   */
  static async processInChunks(array, processor, chunkSize = 100, maxConcurrency = 5) {
    const chunks = []
    
    // Dividir array en chunks
    for (let i = 0; i < array.length; i += chunkSize) {
      chunks.push(array.slice(i, i + chunkSize))
    }
    
    const concurrencyManager = new ConcurrencyManager(maxConcurrency)
    
    // Procesar chunks en paralelo
    const results = await Promise.all(
      chunks.map((chunk, index) => 
        concurrencyManager.execute(async () => {
          logger.debug(`Procesando chunk ${index + 1}/${chunks.length}`, {
            chunkSize: chunk.length
          })
          
          return await processor(chunk, index)
        })
      )
    )
    
    // Combinar resultados
    return results.flat()
  }

  /**
   * Ejecuta validaciones en paralelo con early exit
   */
  static async parallelValidations(validations) {
    const results = await Promise.allSettled(
      validations.map(async (validation, index) => {
        try {
          const result = await validation()
          return { index, success: true, result }
        } catch (error) {
          return { index, success: false, error: error.message }
        }
      })
    )
    
    const failures = results
      .filter(r => r.status === 'fulfilled' && !r.value.success)
      .map(r => r.value)
    
    if (failures.length > 0) {
      throw new Error(`Validaciones fallidas: ${failures.map(f => f.error).join(', ')}`)
    }
    
    return results
      .filter(r => r.status === 'fulfilled' && r.value.success)
      .map(r => r.value.result)
  }

  /**
   * Retry con backoff exponencial
   */
  static async retryWithBackoff(fn, maxRetries = 3, baseDelay = 1000) {
    let lastError
    
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        return await fn()
      } catch (error) {
        lastError = error
        
        if (attempt === maxRetries) {
          break
        }
        
        const delay = baseDelay * Math.pow(2, attempt)
        logger.warn(`Intento ${attempt + 1} falló, reintentando en ${delay}ms`, {
          error: error.message
        })
        
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }
    
    throw lastError
  }
}

/**
 * Cache con TTL para operaciones costosas
 */
class TTLCache {
  constructor(defaultTTL = 300000) { // 5 minutos por defecto
    this.cache = new Map()
    this.timers = new Map()
    this.defaultTTL = defaultTTL
  }

  /**
   * Obtiene un valor del cache o lo calcula
   */
  async get(key, calculator, ttl = this.defaultTTL) {
    if (this.cache.has(key)) {
      return this.cache.get(key)
    }

    const value = await calculator()
    this.set(key, value, ttl)
    return value
  }

  /**
   * Establece un valor en el cache
   */
  set(key, value, ttl = this.defaultTTL) {
    // Limpiar timer existente
    if (this.timers.has(key)) {
      clearTimeout(this.timers.get(key))
    }

    this.cache.set(key, value)
    
    // Configurar expiración
    const timer = setTimeout(() => {
      this.cache.delete(key)
      this.timers.delete(key)
    }, ttl)
    
    this.timers.set(key, timer)
  }

  /**
   * Elimina un valor del cache
   */
  delete(key) {
    if (this.timers.has(key)) {
      clearTimeout(this.timers.get(key))
      this.timers.delete(key)
    }
    
    return this.cache.delete(key)
  }

  /**
   * Limpia todo el cache
   */
  clear() {
    for (const timer of this.timers.values()) {
      clearTimeout(timer)
    }
    
    this.cache.clear()
    this.timers.clear()
  }

  /**
   * Obtiene estadísticas del cache
   */
  getStats() {
    return {
      size: this.cache.size,
      keys: Array.from(this.cache.keys())
    }
  }
}

module.exports = {
  WorkerPool,
  ConcurrencyManager,
  ParallelUtils,
  TTLCache
}

