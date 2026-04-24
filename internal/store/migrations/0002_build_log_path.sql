-- 0002_build_log_path.sql
-- build log 从 DB 迁移到磁盘文件：
--   log_path 指向 <LogDir>/builds/<build-id>.log
--   log_size 记录字节数（列表查询展示，不需要读文件）
-- 原 log 列保留（不再写入新内容），阶段 3 结束后改迁移删除。
--
-- 注意：SQLite 的 ALTER TABLE ADD COLUMN 不支持 IF NOT EXISTS，
-- 但 schema_migrations 机制保证同一版本只跑一次。

ALTER TABLE build_records ADD COLUMN log_path TEXT NOT NULL DEFAULT '';
ALTER TABLE build_records ADD COLUMN log_size INTEGER NOT NULL DEFAULT 0;
