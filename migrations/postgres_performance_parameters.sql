-- SQL скрипт для получения параметров PostgreSQL, влияющих на производительность запросов
-- Выполнить: psql -U postgres -d metrics -f postgres_performance_parameters.sql

-- Основные параметры памяти
SELECT 
    '=== ПАРАМЕТРЫ ПАМЯТИ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'shared_buffers',
    'work_mem',
    'maintenance_work_mem',
    'effective_cache_size',
    'temp_buffers'
)
ORDER BY name;

-- Параметры планировщика запросов (стоимость операций)
SELECT 
    '=== ПАРАМЕТРЫ СТОИМОСТИ ОПЕРАЦИЙ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'random_page_cost',
    'seq_page_cost',
    'cpu_tuple_cost',
    'cpu_index_tuple_cost',
    'cpu_operator_cost',
    'effective_io_concurrency'
)
ORDER BY name;

-- Параметры параллельного выполнения
SELECT 
    '=== ПАРАМЕТРЫ ПАРАЛЛЕЛЬНОГО ВЫПОЛНЕНИЯ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'max_parallel_workers_per_gather',
    'max_parallel_workers',
    'max_worker_processes',
    'parallel_setup_cost',
    'parallel_tuple_cost'
)
ORDER BY name;

-- Параметры статистики и оптимизации
SELECT 
    '=== ПАРАМЕТРЫ СТАТИСТИКИ И ОПТИМИЗАЦИИ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'default_statistics_target',
    'join_collapse_limit',
    'from_collapse_limit',
    'geqo_threshold'
)
ORDER BY name;

-- Параметры включения/выключения методов выполнения
SELECT 
    '=== ПАРАМЕТРЫ МЕТОДОВ ВЫПОЛНЕНИЯ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'enable_hashjoin',
    'enable_mergejoin',
    'enable_nestloop',
    'enable_seqscan',
    'enable_indexscan',
    'enable_bitmapscan',
    'enable_indexonlyscan',
    'enable_material'
)
ORDER BY name;

-- Параметры WAL и checkpoint
SELECT 
    '=== ПАРАМЕТРЫ WAL И CHECKPOINT ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'wal_buffers',
    'min_wal_size',
    'max_wal_size',
    'checkpoint_completion_target',
    'checkpoint_timeout'
)
ORDER BY name;

-- Общие параметры
SELECT 
    '=== ОБЩИЕ ПАРАМЕТРЫ ===' as section;

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings
WHERE name IN (
    'max_connections',
    'autovacuum',
    'autovacuum_max_workers'
)
ORDER BY name;

-- Информация о версии PostgreSQL
SELECT 
    '=== ИНФОРМАЦИЯ О ВЕРСИИ ===' as section;

SELECT version();

-- Информация о размере базы данных
SELECT 
    '=== РАЗМЕР БАЗЫ ДАННЫХ ===' as section;

SELECT 
    pg_size_pretty(pg_database_size('metrics')) as database_size;
