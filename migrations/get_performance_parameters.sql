-- Запрос для получения параметров PostgreSQL, влияющих на производительность запросов
-- Выполнить: psql -U postgres -d metrics -f get_performance_parameters.sql

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
    'random_page_cost',
    'seq_page_cost',
    'cpu_tuple_cost',
    'cpu_index_tuple_cost',
    'cpu_operator_cost',
    'max_parallel_workers_per_gather',
    'max_parallel_workers',
    'max_worker_processes',
    'effective_io_concurrency',
    'default_statistics_target',
    'checkpoint_completion_target',
    'wal_buffers',
    'min_wal_size',
    'max_wal_size',
    'checkpoint_timeout',
    'max_connections',
    'temp_buffers',
    'join_collapse_limit',
    'from_collapse_limit',
    'geqo_threshold',
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
