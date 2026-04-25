-- 0003_project_git_repo.sql
-- 为 projects 加 Git 仓库元数据：
--   repo_url     - 仓库地址（HTTPS / SSH，空表示纯本地路径项目）
--   repo_branch  - 默认分支（空表示用 origin 默认）
--   last_pull_at - 最近一次成功 git pull 的时间
--
-- token 不入库，走 OS keyring（key="git:<projectID>"）。

ALTER TABLE projects ADD COLUMN repo_url TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN repo_branch TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN last_pull_at DATETIME;
