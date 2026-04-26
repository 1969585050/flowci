-- 0004_project_pinned.sql
-- 项目置顶功能：pinned_at 非空表示已置顶，列表按 pinned_at DESC NULLS LAST 排序在前。
-- nullable，默认 NULL（未置顶）。

ALTER TABLE projects ADD COLUMN pinned_at DATETIME;
