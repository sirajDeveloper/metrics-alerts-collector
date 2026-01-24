-- Запрос для получения плана выполнения оптимизированного запроса
-- Форматы: TEXT (читаемый), JSON (для анализа), XML (альтернатива)

-- Вариант 1: Текстовый формат (читаемый)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
WITH filtered_base AS (
    SELECT DISTINCT
        ccr.id as ccr_id,
        ccr.ci_id,
        ccr.check_id,
        ccr.answer_id,
        ccr.description,
        ccr.answer_name,
        ch.id as ch_id,
        ch.name as ch_name,
        ch.rule_id,
        ch.rr_id,
        ci.id as ci_id_pk,
        ci.name as ci_name,
        ci.ci_id as ci_ci_id,
        ci.it_service_critical_level,
        ci.status_detailed,
        ci.platform_name,
        ci.providing_unit,
        ci.functional_block,
        ci.tribe,
        ci.it_manager_employee_name,
        ci.reliability_group,
        reqres.req_id,
        reqres.req_name,
        reqres.vnd_rr_code,
        reqres.ci_blocking_id,
        reqres.rr_radar_result_id,
        cam.status as cam_status,
        rr.process_id,
        rr.applicability_id
    FROM rms.ch_common_results ccr
    JOIN rms.rr_checks ch ON ccr.check_id = ch.id
    JOIN rms.datahub_sm_ci ci ON ccr.ci_id = ci.id
    JOIN rms.all_req_completing reqres ON ch.rr_id = reqres.req_id AND reqres.ci_id = ci.id
    JOIN rms.c_answers_manual cam ON cam.check_id = ch.id AND cam.id = ccr.answer_id
    JOIN rms.rr_registry rr ON rr.id = reqres.req_id
    WHERE reqres.inclusion_in_radar = true
      AND rr.status_id = 5
      AND ch.active = true
      AND ci.severity_id IN (1,2,3)
      AND ci.environment_id IN (1, 2)
      AND ci.platform_id IN (1, 2, 3, 4, 5, 6, 7, 8)
      AND rr.applicability_id IN (2)
),
relevant_combinations AS (
    SELECT DISTINCT ci_id_pk, ch_id
    FROM filtered_base
),
problems_filtered AS (
    SELECT 
        prob.*,
        ROW_NUMBER() OVER (PARTITION BY ci_id, check_id ORDER BY updated DESC) AS rn,
        COUNT(*) OVER (PARTITION BY ci_id, check_id) AS cnt
    FROM rms.rr_problems prob
    INNER JOIN relevant_combinations rc ON prob.ci_id = rc.ci_id_pk AND prob.check_id = rc.ch_id
)
SELECT
    fb.ci_name AS "ciName",
    fb.ci_ci_id AS "ciId",
    blockstat.name AS "blockStatusName",
    reqstat.name AS "reqRadarResult",
    fb.it_service_critical_level AS "severityName",
    fb.status_detailed AS "environmentName",
    fb.platform_name AS "platformName",
    fb.providing_unit AS "providingUnit",
    fb.functional_block AS "block",
    fb.tribe AS "tribe",
    fb.it_manager_employee_name AS "ciManagerName",
    fb.reliability_group AS "reliabilityGroup",
    fb.req_id AS "reqId",
    fb.req_name AS "reqName",
    fb.vnd_rr_code AS "vndRrCode",
    pr.process_manager AS "processManager",
    fb.ch_id AS "checkId",
    fb.ch_name AS "checkName",
    fb.rule_id AS "ruleId",
    prob.problem_id AS "problemNumber",
    prob.problem_status AS "problemStatus",
    prob.end_date AS "problemEndDate",
    fb.description AS "description",
    CASE
        WHEN rcps.status = true THEN cast('пройдена' as varchar(50))
        WHEN rcps.status = false THEN cast('не пройдена' as varchar(50))
        ELSE cast('' as varchar(50))
    END AS "checkPassed",
    fb.answer_name AS "answerName",
    appl.name AS "applicability"
FROM filtered_base fb
LEFT JOIN rms.rr_checks_passed_status rcps ON rcps.id = fb.cam_status
LEFT JOIN rms.rr_applicability appl ON appl.id = fb.applicability_id
LEFT JOIN rms.rr_process pr ON pr.id = fb.process_id
JOIN rms.rr_radar_ci_blocking_status blockstat ON blockstat.id = fb.ci_blocking_id
JOIN rms.rr_radar_result reqstat ON reqstat.id = fb.rr_radar_result_id
LEFT JOIN problems_filtered prob ON fb.ch_id = prob.check_id
    AND fb.ci_id_pk = prob.ci_id
    AND ((prob.cnt = 1) OR (prob.cnt > 1 AND prob.problem_status <> 'Закрыта' AND prob.rn = 1));

-- Вариант 2: JSON формат (для программного анализа)
-- EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT JSON) <запрос>

-- Вариант 3: XML формат
-- EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT XML) <запрос>

-- Вариант 4: YAML формат
-- EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT YAML) <запрос>
