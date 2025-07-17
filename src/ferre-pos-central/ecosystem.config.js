/**
 * Configuración PM2 para Sistema Ferre-POS
 * 
 * Configuración para deployment en producción con cluster mode,
 * monitoreo, logs y restart automático.
 */

module.exports = {
  apps: [
    {
      name: 'ferre-pos-api',
      script: 'src/server.js',
      instances: 'max', // Usar todos los cores disponibles
      exec_mode: 'cluster',
      env: {
        NODE_ENV: 'development',
        PORT: 3000
      },
      env_production: {
        NODE_ENV: 'production',
        PORT: 3000
      },
      env_staging: {
        NODE_ENV: 'staging',
        PORT: 3001
      },
      // Configuración de logs
      log_file: 'logs/combined.log',
      out_file: 'logs/out.log',
      error_file: 'logs/error.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      
      // Configuración de restart
      watch: false,
      ignore_watch: ['node_modules', 'logs', 'coverage', 'tests'],
      max_memory_restart: '1G',
      restart_delay: 4000,
      
      // Configuración de cluster
      kill_timeout: 5000,
      wait_ready: true,
      listen_timeout: 10000,
      
      // Configuración de monitoreo
      min_uptime: '10s',
      max_restarts: 10,
      
      // Variables de entorno específicas
      node_args: '--max-old-space-size=2048',
      
      // Configuración de autorestart
      autorestart: true,
      cron_restart: '0 2 * * *', // Restart diario a las 2 AM
      
      // Configuración de merge logs
      merge_logs: true,
      
      // Configuración de source map
      source_map_support: true,
      
      // Configuración de instancias
      increment_var: 'PORT',
      
      // Script de post-deploy
      post_update: ['npm install', 'npm run build']
    },
    {
      name: 'ferre-pos-workers',
      script: 'src/workers/workerManager.js',
      instances: 2,
      exec_mode: 'fork',
      env: {
        NODE_ENV: 'development',
        WORKER_TYPE: 'general'
      },
      env_production: {
        NODE_ENV: 'production',
        WORKER_TYPE: 'general'
      },
      log_file: 'logs/workers.log',
      out_file: 'logs/workers-out.log',
      error_file: 'logs/workers-error.log',
      max_memory_restart: '512M',
      restart_delay: 2000,
      autorestart: true,
      watch: false
    },
    {
      name: 'ferre-pos-scheduler',
      script: 'src/utils/scheduler.js',
      instances: 1,
      exec_mode: 'fork',
      env: {
        NODE_ENV: 'development'
      },
      env_production: {
        NODE_ENV: 'production'
      },
      log_file: 'logs/scheduler.log',
      out_file: 'logs/scheduler-out.log',
      error_file: 'logs/scheduler-error.log',
      max_memory_restart: '256M',
      restart_delay: 5000,
      autorestart: true,
      cron_restart: '0 0 * * *', // Restart diario a medianoche
      watch: false
    }
  ],

  deploy: {
    production: {
      user: 'deploy',
      host: ['server1.example.com', 'server2.example.com'],
      ref: 'origin/main',
      repo: 'git@github.com:ferre-pos/api.git',
      path: '/var/www/ferre-pos-api',
      'pre-deploy-local': '',
      'post-deploy': 'npm install && npm run build && pm2 reload ecosystem.config.js --env production',
      'pre-setup': '',
      'ssh_options': 'StrictHostKeyChecking=no'
    },
    staging: {
      user: 'deploy',
      host: 'staging.example.com',
      ref: 'origin/develop',
      repo: 'git@github.com:ferre-pos/api.git',
      path: '/var/www/ferre-pos-api-staging',
      'post-deploy': 'npm install && npm run build && pm2 reload ecosystem.config.js --env staging',
      'ssh_options': 'StrictHostKeyChecking=no'
    }
  }
}

