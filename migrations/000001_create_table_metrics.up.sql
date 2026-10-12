-- 000001_create_table_metrics.up.sql
-- Создание таблицы хранения метрик
CREATE TABLE metrics (
                         id    VARCHAR(255) NOT NULL,
                         mtype VARCHAR(16)  NOT NULL,
                         delta BIGINT,
                         value DOUBLE PRECISION,
                         PRIMARY KEY (id, mtype),
                         CHECK ((mtype = 'counter' AND delta IS NOT NULL AND value IS NULL)
                             OR (mtype = 'gauge'   AND value IS NOT NULL AND delta IS NULL))
);