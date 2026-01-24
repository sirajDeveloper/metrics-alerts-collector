# Анализ плана выполнения оптимизированного запроса

## Основные метрики

- **Planning Time**: 6.990 ms
- **Execution Time**: 21.021 ms
- **Total Time**: ~28 ms
- **Количество строк**: 1398
- **Buffers (shared hit)**: 23,228 блоков

## Структура плана выполнения

### 1. CTE: filtered_base
- **Время выполнения**: 12.632 - 12.985 ms
- **Количество строк**: 1398
- **Buffers**: 10,643 shared hit
- **Операции**:
  - Unique (удаление дубликатов)
  - Sort (quicksort, Memory: 445kB)
  - Nested Loop с несколькими JOIN'ами

### 2. CTE: relevant_combinations
- Извлекает уникальные комбинации (ci_id_pk, ch_id) из filtered_base
- Используется для фильтрации проблем

### 3. CTE: problems_filtered
- **Window Functions**: ROW_NUMBER() и COUNT() OVER
- **Время выполнения**: ~0.289 ms для 698 уникальных комбинаций
- **Buffers**: 1,400 shared hit
- **Ключевая оптимизация**: Вычисляется один раз для релевантных данных, а не 1398 раз!

### 4. Основной SELECT
- **Nested Loop Left Join** с problems_filtered
- **Rows Removed by Join Filter**: 2,790 (нормально для LEFT JOIN)
- Все остальные JOIN'ы выполняются быстро через индексы

## Используемые индексы

1. **idx_ch_common_results_check_id** - Bitmap Index Scan
2. **idx_ch_common_results_answer_id** - Bitmap Index Scan
3. **idx_regres_ci_req_applicability** - Index Only Scan (покрывающий индекс!)
4. **datahub_sm_ci_pkey** - Index Scan
5. **idx_rr_problems_ci_check_updated** - Index Only Scan (критически важен!)
6. Все PRIMARY KEY индексы для справочных таблиц

## Ключевые оптимизации в плане

### ✅ Index Only Scan
- `idx_regres_ci_req_applicability` - покрывающий индекс, не требует обращения к таблице
- `idx_rr_problems_ci_check_updated` - покрывающий индекс с INCLUDE колонками

### ✅ Bitmap Index Scan
- Эффективное объединение нескольких индексов для ch_common_results

### ✅ CTE оптимизация
- problems_filtered вычисляется один раз вместо 1398 раз
- filtered_base материализуется один раз

### ✅ Hash Join
- Используется для небольших таблиц (rr_checks, rr_registry)

## Проблемные места (если есть)

1. **Join Filter**: `ch.rr_id = reqres.req_id` удаляет 128 строк
   - Это нормально, но можно рассмотреть создание индекса на (rr_id, id) для rr_checks

2. **Sort Method**: quicksort использует 445kB памяти
   - В пределах нормы для work_mem = 4MB

## Рекомендации по дальнейшей оптимизации

1. **Увеличить work_mem** до 64MB - для более эффективных сортировок
2. **Увеличить shared_buffers** до 2GB - больше данных в памяти
3. **Создать составной индекс** на `rr_checks (rr_id, id, active)` для оптимизации JOIN
4. **Рассмотреть материализованное представление** для часто используемых комбинаций

## Сравнение с исходным запросом

| Метрика | Исходный запрос | Оптимизированный | Улучшение |
|---------|----------------|------------------|-----------|
| Execution Time | 24,218 ms | 21 ms | **1154x быстрее** |
| Planning Time | 23.8 ms | 7 ms | 3.4x быстрее |
| WindowAgg выполнений | 1398 раз | 1 раз | **1398x меньше** |
| Обработка rr_problems | 12.6M строк | ~1,400 строк | **9000x меньше** |

## Выводы

План выполнения показывает эффективное использование:
- ✅ Покрывающих индексов (Index Only Scan)
- ✅ CTE для предварительной фильтрации
- ✅ Оптимальных методов JOIN
- ✅ Минимальное использование временных файлов

Запрос оптимизирован и готов к production использованию.
