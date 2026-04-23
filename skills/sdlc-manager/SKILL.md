---
name: 软件开发生命周期管理器
version: 3.0.0
description: 管理软件从产品原型到开发、测试、交付的全过程，每阶段必须通过质量评审（3轮挑刺打分≥95分）才能进入下一阶段，支持里程碑管理
author: OpenSpace
category: 项目管理
tags: [软件开发, 项目管理, 生命周期, 敏捷, 瀑布, 质量评审, 里程碑]
icon: 🔄

requires:
  - run_shell
  - read_file
  - write_file
  - list_dir

input_schema:
  type: object
  properties:
    action:
      type: string
      description: 执行的动作
      enum: ["create_prototype", "plan_development", "execute_development", "run_testing", "deploy_release", "track_progress", "full_lifecycle"]
      default: "full_lifecycle"
    project_name:
      type: string
      description: 项目名称
    project_path:
      type: string
      description: 项目路径
      default: "./project"
    methodology:
      type: string
      description: 开发方法论
      enum: ["敏捷", "瀑布", "DevOps", "迭代"]
      default: "敏捷"
    milestones:
      type: array
      description: 里程碑列表
    requirements:
      type: array
      description: 需求列表
    development_tasks:
      type: array
      description: 开发任务列表
    test_cases:
      type: array
      description: 测试用例列表
    deploy_target:
      type: string
      description: 部署目标环境
      enum: ["开发环境", "测试环境", "预生产环境", "生产环境"]
      default: "测试环境"
    config:
      type: object
      description: 额外配置
      properties:
        enable_auto_test:
          type: boolean
          default: false
        enable_ci_cd:
          type: boolean
          default: false
        enable_code_review:
          type: boolean
          default: true
        rollback_on_failure:
          type: boolean
          default: true
        quality_threshold:
          type: number
          description: 质量评审通过分数（默认95分）
          default: 95
        min_review_rounds:
          type: number
          description: 最少评审轮数（默认3轮）
          default: 3

output_schema:
  type: object
  properties:
    success:
      type: boolean
    action:
      type: string
    current_stage:
      type: string
    quality_review:
      type: object
      properties:
        rounds:
          type: array
        total_rounds:
          type: number
        passed:
          type: boolean
        final_score:
          type: number
        next_action:
          type: string
    milestone_report:
      type: object
    metrics:
      type: object
    next_steps:
      type: array
    error:
      type: string
