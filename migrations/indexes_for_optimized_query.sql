-- Индексы для таблиц, используемых в оптимизированном запросе
-- Схема: rms

-- ============================================================================
-- rms.ch_common_results
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT ch_common_results_pk PRIMARY KEY (id)

-- Индексы
CREATE INDEX IF NOT EXISTS idx_ch_common_results_ci_id ON rms.ch_common_results USING btree (ci_id);
CREATE INDEX IF NOT EXISTS idx_ch_common_results_check_id ON rms.ch_common_results USING btree (check_id);
CREATE INDEX IF NOT EXISTS idx_ch_common_results_answer_id ON rms.ch_common_results USING btree (answer_id);

-- ============================================================================
-- rms.rr_checks
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_checks_pk PRIMARY KEY (id)

-- Индексы
CREATE INDEX IF NOT EXISTS add_idx_rr_checks_rr_id ON rms.rr_checks USING btree (rr_id);

-- ============================================================================
-- rms.datahub_sm_ci
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT datahub_sm_ci_pkey PRIMARY KEY (id)

-- Unique Index (автоматически создается при UNIQUE constraint)
-- CONSTRAINT datahub_sm_ci_ci_id_key UNIQUE (ci_id)

-- Индексы
CREATE INDEX IF NOT EXISTS datahub_sm_ci_providing_unit_index ON rms.datahub_sm_ci USING btree (providing_unit);
CREATE INDEX IF NOT EXISTS datahub_sm_ci_tribe_index ON rms.datahub_sm_ci USING btree (tribe);
CREATE INDEX IF NOT EXISTS idx_datahub_sm_ci_environment_id ON rms.datahub_sm_ci USING btree (environment_id);
CREATE INDEX IF NOT EXISTS idx_datahub_sm_ci_platform_id ON rms.datahub_sm_ci USING btree (platform_id);
CREATE INDEX IF NOT EXISTS idx_datahub_sm_ci_severity_id ON rms.datahub_sm_ci USING btree (severity_id);
CREATE INDEX IF NOT EXISTS idx_datahub_sm_ci_type_sm ON rms.datahub_sm_ci USING btree (it_service);
CREATE INDEX IF NOT EXISTS idx_dhci_composite_filter ON rms.datahub_sm_ci USING btree (id, severity_id, environment_id, platform_id, "name", ci_id, it_service_critical_level, status_detailed, platform_name, providing_unit, functional_block, tribe, it_manager_employee_name, reliability_group);

-- ============================================================================
-- rms.all_req_completing
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT all_req_completing_pkey PRIMARY KEY (id)

-- Индексы
CREATE INDEX IF NOT EXISTS all_req_completing_ci_blocking_id_index ON rms.all_req_completing USING btree (ci_blocking_id);
CREATE INDEX IF NOT EXISTS all_req_completing_req_id_index ON rms.all_req_completing USING btree (req_id);
CREATE INDEX IF NOT EXISTS idx_regres_ci_req_applicability ON rms.all_req_completing (ci_id, req_id, inclusion_in_radar, applicability_id, req_name, vnd_rr_code, ci_blocking_id, rr_radar_result_id);

-- ============================================================================
-- rms.c_answers_manual
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT c_answers_manual_pk PRIMARY KEY (id)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY

-- ============================================================================
-- rms.rr_registry
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_registry_pk PRIMARY KEY (id)

-- Индексы
CREATE INDEX IF NOT EXISTS idx_rr_registry_process_id ON rms.rr_registry USING btree (process_id);
CREATE INDEX IF NOT EXISTS idx_rr_registry_status_appl ON rms.rr_registry USING btree (status_id, id, applicability_id) INCLUDE (process_id);
CREATE INDEX IF NOT EXISTS idx_rr_registry_status_include ON rms.rr_registry USING btree (status_id, id) INCLUDE (applicability_id, process_id);

-- ============================================================================
-- rms.rr_problems
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_problems_pk PRIMARY KEY (id)

-- Unique Index (автоматически создается при UNIQUE constraint)
-- CONSTRAINT rr_problems_problem_id_key UNIQUE (problem_id)

-- Индексы
CREATE INDEX IF NOT EXISTS idx_rr_problems_ci_check_updated ON rms.rr_problems USING btree (ci_id, check_id, updated DESC) INCLUDE (problem_id, problem_status, end_date);

-- ============================================================================
-- rms.rr_checks_passed_status
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_checks_passed_status_pk PRIMARY KEY (id)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY

-- ============================================================================
-- rms.rr_applicability
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_applicability_pk PRIMARY KEY (id)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY

-- ============================================================================
-- rms.rr_process
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_process_pk PRIMARY KEY (id)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY

-- ============================================================================
-- rms.rr_radar_ci_blocking_status
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_radar_ci_blocking_status_pkey PRIMARY KEY (id)

-- Unique Index (автоматически создается при UNIQUE constraint)
-- CONSTRAINT rr_radar_ci_blocking_status_sysname_key UNIQUE (sysname)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY и UNIQUE

-- ============================================================================
-- rms.rr_radar_result
-- ============================================================================

-- Primary Key (автоматически создает индекс)
-- CONSTRAINT rr_radar_result_pkey PRIMARY KEY (id)

-- Примечание: Нет дополнительных индексов, кроме PRIMARY KEY

-- ============================================================================
-- Сводная информация
-- ============================================================================

-- Всего индексов (включая PRIMARY KEY и UNIQUE):
-- 1. ch_common_results: 4 индекса (1 PK + 3 обычных)
-- 2. rr_checks: 2 индекса (1 PK + 1 обычный)
-- 3. datahub_sm_ci: 9 индексов (1 PK + 1 UNIQUE + 7 обычных)
-- 4. all_req_completing: 4 индекса (1 PK + 3 обычных)
-- 5. c_answers_manual: 1 индекс (1 PK)
-- 6. rr_registry: 4 индекса (1 PK + 3 обычных)
-- 7. rr_problems: 3 индекса (1 PK + 1 UNIQUE + 1 обычный)
-- 8. rr_checks_passed_status: 1 индекс (1 PK)
-- 9. rr_applicability: 1 индекс (1 PK)
-- 10. rr_process: 1 индекс (1 PK)
-- 11. rr_radar_ci_blocking_status: 2 индекса (1 PK + 1 UNIQUE)
-- 12. rr_radar_result: 1 индекс (1 PK)

-- Итого: 33 индекса

-- ============================================================================
-- Критически важные индексы для оптимизированного запроса
-- ============================================================================

-- Эти индексы наиболее важны для производительности запроса:

-- 1. idx_ch_common_results_check_id - используется для JOIN с rr_checks
-- 2. idx_ch_common_results_answer_id - используется для JOIN с c_answers_manual
-- 3. idx_ch_common_results_ci_id - используется для JOIN с datahub_sm_ci
-- 4. add_idx_rr_checks_rr_id - используется для JOIN с rr_registry через all_req_completing
-- 5. idx_regres_ci_req_applicability - покрывающий индекс для all_req_completing (используется как Index Only Scan)
-- 6. idx_rr_problems_ci_check_updated - критически важен для фильтрации проблем (используется как Index Only Scan)
-- 7. idx_rr_registry_status_appl - используется для фильтрации по status_id и applicability_id
-- 8. idx_datahub_sm_ci_environment_id, idx_datahub_sm_ci_severity_id, idx_datahub_sm_ci_platform_id - для фильтрации CI
