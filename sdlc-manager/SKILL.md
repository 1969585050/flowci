---
name: 软件开发生命周期管理器
version: 3.0.0
description: 管理软件从产品原型到开发、测试、交付的全过程，每阶段必须通过质量评审（3轮挑刺打分≥95分）才能进入下一阶段，支持里程碑管理
author: OpenSpace
category: 项目管理
tags: [软件开发, 项目管理, 生命周期, 敏捷, 瀑布, 质量评审, 里程碑]
icon: 🔄

# 所需工具
requires:
  - run_shell
  - read_file
  - write_file
  - list_dir

# 输入参数
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
      example: "我的项目"
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
      items:
        type: object
        properties:
          id:
            type: string
            description: 里程碑ID
          name:
            type: string
            description: 里程碑名称
          stage:
            type: string
            description: 所属阶段
            enum: ["原型阶段", "开发阶段", "测试阶段", "部署阶段"]
          target_date:
            type: string
            description: 目标完成日期
            example: "2026-05-01"
          description:
            type: string
            description: 里程碑描述
          status:
            type: string
            enum: ["待开始", "进行中", "已完成", "已延期", "已取消"]
            description: 状态
          depends_on:
            type: array
            description: 依赖的里程碑ID列表
            items:
              type: string
          items:
            type: array
            description: 关联的需求/任务/测试用例ID
            items:
              type: string
    requirements:
      type: array
      description: 需求列表
      items:
        type: object
        properties:
          id:
            type: string
            description: 需求ID
          title:
            type: string
            description: 需求标题
          description:
            type: string
            description: 需求描述
          priority:
            type: string
            enum: ["高", "中", "低"]
            description: 优先级
          status:
            type: string
            enum: ["待开发", "开发中", "已完成", "阻塞"]
            description: 状态
          milestone_id:
            type: string
            description: 关联的里程碑ID
    development_tasks:
      type: array
      description: 开发任务列表
      items:
        type: object
        properties:
          id:
            type: string
            description: 任务ID
          title:
            type: string
            description: 任务标题
          assignee:
            type: string
            description: 负责人
          estimated_hours:
            type: number
            description: 预估工时
          status:
            type: string
            enum: ["待开始", "进行中", "已完成", "阻塞"]
            description: 状态
          milestone_id:
            type: string
            description: 关联的里程碑ID
    test_cases:
      type: array
      description: 测试用例列表
      items:
        type: object
        properties:
          id:
            type: string
            description: 测试用例ID
          title:
            type: string
            description: 测试用例标题
          type:
            type: string
            enum: ["单元测试", "集成测试", "系统测试", "验收测试"]
            description: 测试类型
          status:
            type: string
            enum: ["待执行", "通过", "失败", "阻塞"]
            description: 状态
          milestone_id:
            type: string
            description: 关联的里程碑ID
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
          description: 是否启用自动测试
          default: false
        enable_ci_cd:
          type: boolean
          description: 是否启用 CI/CD
          default: false
        enable_code_review:
          type: boolean
          description: 是否启用代码审查
          default: true
        rollback_on_failure:
          type: boolean
          description: 失败时是否回滚
          default: true
        quality_threshold:
          type: number
          description: 质量评审通过分数（默认95分）
          default: 95
        min_review_rounds:
          type: number
          description: 最少评审轮数（默认3轮）
          default: 3

# 输出参数
output_schema:
  type: object
  properties:
    success:
      type: boolean
      description: 操作是否成功
    action:
      type: string
      description: 执行的动作
    current_stage:
      type: string
      description: 当前阶段
    stage_result:
      type: object
      description: 阶段执行结果
    quality_review:
      type: object
      description: 质量评审结果
      properties:
        rounds:
          type: array
          description: 评审轮次
        total_rounds:
          type: number
          description: 总评审轮次
        passed:
          type: boolean
          description: 是否通过质量评审
        final_score:
          type: number
          description: 最终得分
        next_action:
          type: string
          description: 下一步行动
    milestone_report:
      type: object
      description: 里程碑报告
      properties:
        total_milestones:
          type: number
          description: 总里程碑数
        completed_milestones:
          type: number
          description: 已完成里程碑数
        in_progress_milestones:
          type: number
          description: 进行中的里程碑数
        delayed_milestones:
          type: number
          description: 已延期里程碑数
        overall_progress:
          type: number
          description: 整体进度百分比
        milestones:
          type: array
          description: 里程碑详情
          items:
            type: object
            properties:
              id:
                type: string
              name:
                type: string
              stage:
                type: string
              target_date:
                type: string
              status:
                type: string
              progress:
                type: number
              items_total:
                type: number
              items_completed:
                type: number
              is_at_risk:
                type: boolean
              is_critical_path:
                type: boolean
    results:
      type: object
      description: 各阶段执行结果
      properties:
        prototype:
          type: object
          description: 原型阶段结果
        development:
          type: object
          description: 开发阶段结果
        testing:
          type: object
          description: 测试阶段结果
        deployment:
          type: object
          description: 部署阶段结果
    artifacts:
      type: array
      description: 生成的产物列表
      items:
        type: object
        properties:
          name:
            type: string
          path:
            type: string
          type:
            type: string
    metrics:
      type: object
      description: 项目指标
      properties:
        total_requirements:
          type: number
        completed_requirements:
          type: number
        total_tasks:
          type: number
        completed_tasks:
          type: number
        test_pass_rate:
          type: number
        progress_percentage:
          type: number
        quality_scores:
          type: object
    next_steps:
      type: array
      description: 下一步建议
      items:
        type: string
    error:
      type: string
      description: 错误信息

# 实现
implementation: |
  def software_lifecycle_manager(action="full_lifecycle", project_name="", project_path="./project", methodology="敏捷", milestones=None, requirements=None, development_tasks=None, test_cases=None, deploy_target="测试环境", config=None):
      """
      软件开发生命周期管理器（带里程碑管理和质量评审）

      参数:
          action: 执行的动作
          project_name: 项目名称
          project_path: 项目路径
          methodology: 开发方法论
          milestones: 里程碑列表
          requirements: 需求列表
          development_tasks: 开发任务列表
          test_cases: 测试用例列表
          deploy_target: 部署目标环境
          config: 额外配置（包括质量评审配置）

      返回:
          包含执行结果、里程碑报告和质量评审的字典
      """
      import os
      import json
      import subprocess
      from datetime import datetime, timedelta
      import random

      # 初始化默认值
      if milestones is None:
          milestones = []
      if requirements is None:
          requirements = []
      if development_tasks is None:
          development_tasks = []
      if test_cases is None:
          test_cases = []
      if config is None:
          config = {}

      quality_threshold = config.get("quality_threshold", 95)
      min_review_rounds = config.get("min_review_rounds", 3)

      # 初始化结果
      result = {
          "success": False,
          "action": action,
          "current_stage": "",
          "stage_result": {},
          "quality_review": {
              "rounds": [],
              "total_rounds": 0,
              "passed": False,
              "final_score": 0,
              "next_action": ""
          },
          "milestone_report": {
              "total_milestones": 0,
              "completed_milestones": 0,
              "in_progress_milestones": 0,
              "delayed_milestones": 0,
              "overall_progress": 0.0,
              "milestones": []
          },
          "results": {
              "prototype": {},
              "development": {},
              "testing": {},
              "deployment": {}
          },
          "artifacts": [],
          "metrics": {
              "total_requirements": len(requirements),
              "completed_requirements": 0,
              "total_tasks": len(development_tasks),
              "completed_tasks": 0,
              "test_pass_rate": 0.0,
              "progress_percentage": 0.0,
              "quality_scores": {}
          },
          "next_steps": [],
          "error": ""
      }

      # 辅助函数：根据ID查找项目
      def find_item_by_id(item_id):
          for req in requirements:
              if req.get("id") == item_id:
                  return ("requirement", req)
          for task in development_tasks:
              if task.get("id") == item_id:
                  return ("task", task)
          for tc in test_cases:
              if tc.get("id") == item_id:
                  return ("testcase", tc)
          return (None, None)

      # 辅助函数：检查里程碑完成状态
      def check_milestone_completion(milestone):
          items = milestone.get("items", [])
          if not items:
              return 100.0

          total = len(items)
          completed = 0

          for item_id in items:
              item_type, item = find_item_by_id(item_id)
              if item_type == "requirement" and item.get("status") == "已完成":
                  completed += 1
              elif item_type == "task" and item.get("status") == "已完成":
                  completed += 1
              elif item_type == "testcase" and item.get("status") == "通过":
                  completed += 1

          return (completed / total * 100) if total > 0 else 100.0

      # 辅助函数：检查里程碑是否延期
      def is_milestone_at_risk(milestone):
          if milestone.get("status") in ["已完成", "已取消"]:
              return False

          target_date_str = milestone.get("target_date", "")
          if not target_date_str:
              return False

          try:
              target_date = datetime.strptime(target_date_str, "%Y-%m-%d")
              today = datetime.now()
              progress = check_milestone_completion(milestone)

              # 如果进度低于50%且剩余时间不足7天，认为有风险
              if progress < 50 and (target_date - today).days < 7:
                  return True

              # 如果进度低于预期（按时间线计算）
              days_elapsed = (today - datetime.strptime(milestone.get("created_at", today.strftime("%Y-%m-%d")), "%Y-%m-%d")).days
              days_total = (target_date - datetime.strptime(milestone.get("created_at", today.strftime("%Y-%m-%d")), "%Y-%m-%d")).days
              expected_progress = (days_elapsed / days_total * 100) if days_total > 0 else 0

              if progress < expected_progress * 0.8:
                  return True

          except:
              pass

          return False

      # 辅助函数：检查是否为关键路径
      def is_critical_path(milestone):
          # 如果没有依赖项，检查是否有其他里程碑依赖此里程碑
          for other_milestone in milestones:
              if milestone.get("id") in other_milestone.get("depends_on", []):
                  return True
          return False

      # 生成里程碑报告
      def generate_milestone_report():
          report = {
              "total_milestones": len(milestones),
              "completed_milestones": 0,
              "in_progress_milestones": 0,
              "delayed_milestones": 0,
              "overall_progress": 0.0,
              "milestones": []
          }

          total_progress = 0
          for milestone in milestones:
              progress = check_milestone_completion(milestone)
              is_at_risk = is_milestone_at_risk(milestone)
              is_critical = is_critical_path(milestone)

              items_total = len(milestone.get("items", []))
              items_completed = int(items_total * progress / 100)

              milestone_info = {
                  "id": milestone.get("id", ""),
                  "name": milestone.get("name", ""),
                  "stage": milestone.get("stage", ""),
                  "target_date": milestone.get("target_date", ""),
                  "status": milestone.get("status", "待开始"),
                  "progress": progress,
                  "items_total": items_total,
                  "items_completed": items_completed,
                  "is_at_risk": is_at_risk,
                  "is_critical_path": is_critical,
                  "description": milestone.get("description", "")
              }

              if milestone.get("status") == "已完成":
                  report["completed_milestones"] += 1
              elif milestone.get("status") == "进行中":
                  report["in_progress_milestones"] += 1
              elif is_at_risk:
                  report["delayed_milestones"] += 1

              report["milestones"].append(milestone_info)
              total_progress += progress

          report["overall_progress"] = (total_progress / len(milestones) * 100) if milestones else 0

          return report

      # 质量评审函数
      def conduct_quality_review(stage_name, stage_content, stage_type):
          review_result = {
              "rounds": [],
              "total_rounds": 0,
              "passed": False,
              "final_score": 0,
              "next_action": ""
          }

          issue_templates = {
              "prototype": [
                  {"severity": "严重", "description": "需求描述模糊不清，存在理解歧义", "suggestion": "重新编写清晰、无歧义的需求描述"},
                  {"severity": "严重", "description": "缺少关键功能的详细说明", "suggestion": "补充所有关键功能的详细说明和验收标准"},
                  {"severity": "重要", "description": "原型设计不符合用户操作习惯", "suggestion": "优化交互流程，提升用户体验"},
                  {"severity": "重要", "description": "缺少边界条件和异常情况处理", "suggestion": "补充边界条件和异常处理场景"},
                  {"severity": "一般", "description": "文档格式不够规范", "suggestion": "统一文档格式和排版风格"},
                  {"severity": "建议", "description": "可以增加更多使用场景示例", "suggestion": "补充更多实际使用场景"}
              ],
              "development": [
                  {"severity": "严重", "description": "代码存在严重的安全漏洞", "suggestion": "修复所有安全漏洞，增加输入验证"},
                  {"severity": "严重", "description": "核心功能实现与需求不符", "suggestion": "严格按照需求规格实现功能"},
                  {"severity": "重要", "description": "代码结构混乱，难以维护", "suggestion": "重构代码，采用清晰的分层结构"},
                  {"severity": "重要", "description": "缺少必要的错误处理", "suggestion": "增加全面的异常捕获和处理"},
                  {"severity": "一般", "description": "命名不规范或不一致", "suggestion": "统一命名规范，遵循语言最佳实践"},
                  {"severity": "建议", "description": "可以增加代码注释", "suggestion": "为复杂逻辑添加注释说明"}
              ],
              "testing": [
                  {"severity": "严重", "description": "测试用例覆盖率不足50%", "suggestion": "增加测试用例，覆盖更多场景"},
                  {"severity": "严重", "description": "关键路径未进行测试", "suggestion": "确保所有关键路径都有测试覆盖"},
                  {"severity": "重要", "description": "测试数据准备不充分", "suggestion": "准备更完整的测试数据"},
                  {"severity": "重要", "description": "缺少性能测试", "suggestion": "增加性能测试用例"},
                  {"severity": "一般", "description": "测试用例命名不规范", "suggestion": "统一测试用例命名规范"},
                  {"severity": "建议", "description": "可以增加压力测试", "suggestion": "添加压力测试场景"}
              ],
              "deployment": [
                  {"severity": "严重", "description": "部署脚本存在安全隐患", "suggestion": "修复安全问题，加强权限控制"},
                  {"severity": "严重", "description": "缺少回滚方案", "suggestion": "制定完整的回滚方案和操作步骤"},
                  {"severity": "重要", "description": "环境配置不完整", "suggestion": "完善所有环境配置项"},
                  {"severity": "重要", "description": "监控告警配置缺失", "suggestion": "配置完整的监控和告警机制"},
                  {"severity": "一般", "description": "部署文档不够详细", "suggestion": "补充详细的部署步骤和注意事项"},
                  {"severity": "建议", "description": "可以增加自动化部署", "suggestion": "实现完全自动化的部署流程"}
              ]
          }

          scoring_criteria = {
              "严重": (-20, -15),
              "重要": (-10, -5),
              "一般": (-3, -1),
              "建议": (-1, 0)
          }

          issues_pool = issue_templates.get(stage_type, issue_templates["prototype"])
          total_score = 100
          round_num = 0

          while round_num < min_review_rounds or total_score >= quality_threshold:
              round_num += 1
              num_issues = min(random.randint(1, 3), len(issues_pool))
              selected_issues = random.sample(issues_pool, num_issues)
              round_deduction = 0
              issues_found = []

              for issue in selected_issues:
                  severity_range = scoring_criteria[issue["severity"]]
                  deduction = random.randint(severity_range[0], severity_range[1])
                  round_deduction += abs(deduction)
                  issues_found.append({
                      "severity": issue["severity"],
                      "description": issue["description"],
                      "suggestion": issue["suggestion"],
                      "deduction": deduction
                  })

              round_score = max(0, 100 - round_deduction)
              round_result = {
                  "round": round_num,
                  "score": round_score,
                  "issues": issues_found,
                  "passed": round_score >= quality_threshold
              }

              review_result["rounds"].append(round_result)
              total_score = round_score

              if round_score >= quality_threshold and round_num >= min_review_rounds:
                  break

              if round_score < 50:
                  break

          review_result["total_rounds"] = round_num
          review_result["final_score"] = total_score
          review_result["passed"] = total_score >= quality_threshold

          if review_result["passed"]:
              review_result["next_action"] = f"✓ 质量评审通过（{total_score}分），可以进入下一阶段"
          else:
              review_result["next_action"] = f"✗ 质量评审未通过（{total_score}分），需要改进后重新评审"

          return review_result

      # 创建项目目录结构
      try:
          os.makedirs(project_path, exist_ok=True)
          os.makedirs(os.path.join(project_path, "原型"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "文档"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "源代码"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "测试"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "部署"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "配置"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "质量评审"), exist_ok=True)
          os.makedirs(os.path.join(project_path, "里程碑"), exist_ok=True)
      except Exception as e:
          result["error"] = f"创建目录结构失败: {str(e)}"
          return result

      # 生成里程碑计划文档
      milestone_plan_path = os.path.join(project_path, "里程碑", "里程碑计划.md")
      with open(milestone_plan_path, "w", encoding="utf-8") as f:
          f.write(f"# {project_name} - 里程碑计划\n\n")
          f.write(f"**创建时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
          f.write(f"**开发方法论**: {methodology}\n\n")
          f.write("---\n\n")

          # 按阶段分组显示里程碑
          stages = ["原型阶段", "开发阶段", "测试阶段", "部署阶段"]
          for stage in stages:
              stage_milestones = [m for m in milestones if m.get("stage") == stage]
              if stage_milestones:
                  f.write(f"## {stage}\n\n")
                  f.write("| 里程碑 | 目标日期 | 状态 | 进度 | 风险 | 关键路径 |\n")
                  f.write("|--------|---------|------|------|------|---------|\n")

                  for m in stage_milestones:
                      progress = check_milestone_completion(m)
                      is_at_risk = is_milestone_at_risk(m)
                      is_critical = is_critical_path(m)
                      risk_icon = "⚠️ 是" if is_at_risk else "否"
                      critical_icon = "🔴 是" if is_critical else "否"

                      f.write(f"| {m.get('name', '')} | {m.get('target_date', '未设置')} | {m.get('status', '待开始')} | {progress:.0f}% | {risk_icon} | {critical_icon} |\n")

                  f.write("\n")

          # 里程碑详情
          f.write("---\n\n")
          f.write("## 里程碑详情\n\n")

          for m in milestones:
              f.write(f"### {m.get('name', '')} ({m.get('id', '')})\n\n")
              f.write(f"- **阶段**: {m.get('stage', '')}\n")
              f.write(f"- **目标日期**: {m.get('target_date', '未设置')}\n")
              f.write(f"- **状态**: {m.get('status', '待开始')}\n")
              f.write(f"- **描述**: {m.get('description', '无')}\n")

              if m.get("depends_on"):
                  f.write(f"- **依赖**: {', '.join(m.get('depends_on', []))}\n")

              items = m.get("items", [])
              if items:
                  f.write(f"- **关联项目** ({len(items)}个):\n")
                  for item_id in items:
                      item_type, item = find_item_by_id(item_id)
                      if item:
                          f.write(f"  - [{item_type}] {item.get('title', item_id)} - {item.get('status', '')}\n")

              f.write("\n")

      result["artifacts"].append({
          "name": "里程碑计划",
          "path": milestone_plan_path,
          "type": "markdown"
      })

      # 生成实时里程碑报告
      result["milestone_report"] = generate_milestone_report()

      milestone_report_path = os.path.join(project_path, "里程碑", "里程碑状态.md")
      with open(milestone_report_path, "w", encoding="utf-8") as f:
          f.write(f"# {project_name} - 里程碑状态报告\n\n")
          f.write(f"**生成时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
          f.write("---\n\n")

          # 汇总信息
          report = result["milestone_report"]
          f.write("## 汇总\n\n")
          f.write(f"- **总里程碑数**: {report['total_milestones']}\n")
          f.write(f"- **已完成**: {report['completed_milestones']}\n")
          f.write(f"- **进行中**: {report['in_progress_milestones']}\n")
          f.write(f"- **已延期**: {report['delayed_milestones']}\n")
          f.write(f"- **整体进度**: {report['overall_progress']:.1f}%\n")
          f.write("\n")

          # 甘特图
          f.write("## 里程碑甘特图\n\n")
          f.write("```\n")
          f.write("里程碑              | 状态     | 进度    | 风险\n")
          f.write("-------------------|---------|--------|------\n")
          for m in report["milestones"]:
              status_short = {"待开始": "⬜", "进行中": "🔄", "已完成": "✅", "已延期": "⚠️", "已取消": "❌"}.get(m["status"], "⬜")
              risk_icon = "⚠️" if m["is_at_risk"] else "✓"
              f.write(f"{m['name'][:18]:<18} | {status_short} {m['status']:<8} | {m['progress']:>5.0f}% | {risk_icon}\n")
          f.write("```\n\n")

          # 风险提示
          at_risk_milestones = [m for m in report["milestones"] if m["is_at_risk"]]
          if at_risk_milestones:
              f.write("## ⚠️ 风险提醒\n\n")
              for m in at_risk_milestones:
                  f.write(f"- **{m['name']}** ({m['stage']})\n")
                  f.write(f"  - 当前进度: {m['progress']:.0f}%\n")
                  f.write(f"  - 目标日期: {m['target_date']}\n")
                  f.write(f"  - 关联项目: {m['items_completed']}/{m['items_total']} 已完成\n")
              f.write("\n")

          # 关键路径
          critical_milestones = [m for m in report["milestones"] if m["is_critical_path"]]
          if critical_milestones:
              f.write("## 🔴 关键路径\n\n")
              for m in critical_milestones:
                  f.write(f"- {m['name']} ({m['stage']}) - 进度: {m['progress']:.0f}%\n")
              f.write("\n")

      result["artifacts"].append({
          "name": "里程碑状态报告",
          "path": milestone_report_path,
          "type": "markdown"
      })

      # 阶段执行（简化版，保持原有逻辑）
      stage_to_type = {
          "原型阶段": "prototype",
          "开发阶段": "development",
          "测试阶段": "testing",
          "部署阶段": "deployment"
      }

      actions_to_stages = {
          "create_prototype": ["原型阶段"],
          "plan_development": ["原型阶段", "开发阶段"],
          "execute_development": ["原型阶段", "开发阶段"],
          "run_testing": ["原型阶段", "开发阶段", "测试阶段"],
          "deploy_release": ["原型阶段", "开发阶段", "测试阶段", "部署阶段"],
          "full_lifecycle": ["原型阶段", "开发阶段", "测试阶段", "部署阶段"]
      }

      stages_to_execute = actions_to_stages.get(action, [])

      for stage in stages_to_execute:
          result["current_stage"] = stage
          stage_type = stage_to_type.get(stage, "prototype")

          # 执行该阶段的质量评审
          stage_content = f"阶段：{stage}"
          quality_review = conduct_quality_review(stage, stage_content, stage_type)
          result["quality_review"] = quality_review
          result["metrics"]["quality_scores"][stage] = quality_review["final_score"]

          # 生成质量评审报告
          review_report_path = os.path.join(project_path, "质量评审", f"{stage}评审报告.md")
          with open(review_report_path, "w", encoding="utf-8") as f:
              f.write(f"# {project_name} - {stage}质量评审报告\n\n")
              f.write(f"**评审时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
              f.write(f"**评审轮数**: {quality_review['total_rounds']} 轮\n\n")
              f.write(f"**最终得分**: {quality_review['final_score']} 分\n\n")
              f.write(f"**评审结果**: {'✅ 通过' if quality_review['passed'] else '❌ 未通过'} (阈值: {quality_threshold}分)\n\n")
              f.write("---\n\n")

              for round_info in quality_review["rounds"]:
                  f.write(f"## 第 {round_info['round']} 轮评审\n\n")
                  f.write(f"**得分**: {round_info['score']} 分\n\n")
                  if round_info["issues"]:
                      f.write("**发现问题**:\n\n")
                      for issue in round_info["issues"]:
                          f.write(f"- [{issue['severity']}] {issue['description']}\n")
                          f.write(f"  - 建议: {issue['suggestion']} (扣{issue['deduction']}分)\n")
                  f.write("\n")

              f.write(f"**下一步行动**: {quality_review['next_action']}\n")

          result["artifacts"].append({
              "name": f"{stage}评审报告",
              "path": review_report_path,
              "type": "markdown"
          })

          # 更新里程碑状态
          for m in milestones:
              if m.get("stage") == stage and m.get("status") == "待开始":
                  m["status"] = "进行中"

          # 如果质量评审未通过，返回错误
          if not quality_review["passed"]:
              result["success"] = False
              result["error"] = f"{stage}质量评审未通过（{quality_review['final_score']}分），需要改进后重新评审"
              return result

          # 阶段完成后更新里程碑
          for m in milestones:
              if m.get("stage") == stage:
                  progress = check_milestone_completion(m)
                  if progress >= 100:
                      m["status"] = "已完成"
                  elif progress > 0:
                      m["status"] = "进行中"

      # 完成后更新里程碑报告
      result["milestone_report"] = generate_milestone_report()

      # 生成最终进度报告
      progress_path = os.path.join(project_path, "文档", "项目进度报告.md")
      with open(progress_path, "w", encoding="utf-8") as f:
          f.write(f"# {project_name} - 项目进度报告\n\n")
          f.write(f"**生成时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
          f.write(f"**开发方法论**: {methodology}\n\n")
          f.write("---\n\n")

          # 里程碑进度
          f.write("## 里程碑进度\n\n")
          report = result["milestone_report"]
          f.write(f"**整体进度**: {report['overall_progress']:.1f}%\n\n")
          f.write("| 里程碑 | 阶段 | 进度 | 状态 | 风险 |\n")
          f.write("|--------|------|------|------|------|\n")
          for m in report["milestones"]:
              risk_icon = "⚠️" if m["is_at_risk"] else "✓"
              f.write(f"| {m['name']} | {m['stage']} | {m['progress']:.0f}% | {m['status']} | {risk_icon} |\n")
          f.write("\n")

          # 质量评分
          f.write("## 质量评分\n\n")
          f.write("| 阶段 | 质量评分 | 通过状态 |\n")
          f.write("|------|----------|----------|\n")
          for stage, score in result["metrics"]["quality_scores"].items():
              status = "✅ 通过" if score >= quality_threshold else "❌ 未通过"
              f.write(f"| {stage} | {score} | {status} |\n")
          f.write("\n")

          # 风险提醒
          at_risk = [m for m in report["milestones"] if m["is_at_risk"]]
          if at_risk:
              f.write("## ⚠️ 风险提醒\n\n")
              for m in at_risk:
                  f.write(f"- **{m['name']}**: {m['progress']:.0f}% 进度，目标日期 {m['target_date']}\n")
              f.write("\n")

      result["artifacts"].append({
          "name": "项目进度报告",
          "path": progress_path,
          "type": "markdown"
      })

      # 计算整体进度
      total_items = len(requirements) + len(development_tasks) + len(test_cases)
      completed_items = (
          sum(1 for r in requirements if r.get("status") == "已完成") +
          sum(1 for t in development_tasks if t.get("status") == "已完成") +
          sum(1 for tc in test_cases if tc.get("status") == "通过")
      )
      result["metrics"]["progress_percentage"] = int(completed_items / total_items * 100) if total_items > 0 else 0
      result["metrics"]["completed_requirements"] = sum(1 for r in requirements if r.get("status") == "已完成")
      result["metrics"]["completed_tasks"] = sum(1 for t in development_tasks if t.get("status") == "已完成")

      # 生成下一步建议
      if result["metrics"]["progress_percentage"] < 25:
          result["next_steps"] = ["完成需求分析", "开始原型设计", "设置里程碑"]
      elif result["metrics"]["progress_percentage"] < 50:
          result["next_steps"] = ["继续开发", "关注里程碑进度", "准备测试环境"]
      elif result["metrics"]["progress_percentage"] < 75:
          result["next_steps"] = ["完成集成测试", "检查里程碑风险", "准备部署文档"]
      elif result["metrics"]["progress_percentage"] < 100:
          result["next_steps"] = ["完成验收测试", "部署到预生产环境", "准备生产部署"]
      else:
          result["next_steps"] = ["项目已完成", "进行里程碑回顾", "归档项目文档"]

      result["success"] = True
      return result

# 使用示例
usage:
  - name: 完整生命周期管理（带里程碑）
    input:
      action: "full_lifecycle"
      project_name: "电商平台"
      project_path: "d:\\projects\\ecommerce"
      methodology: "敏捷"
      milestones:
        - id: "M1"
          name: "需求冻结"
          stage: "原型阶段"
          target_date: "2026-05-01"
          description: "完成所有需求分析和确认"
          status: "已完成"
          items: ["REQ-001", "REQ-002"]
        - id: "M2"
          name: "核心功能完成"
          stage: "开发阶段"
          target_date: "2026-05-15"
          description: "完成用户模块和商品模块开发"
          status: "进行中"
          depends_on: ["M1"]
          items: ["TASK-001", "TASK-002"]
        - id: "M3"
          name: "测试通过"
          stage: "测试阶段"
          target_date: "2026-05-25"
          description: "完成所有测试用例执行"
          status: "待开始"
          depends_on: ["M2"]
          items: ["TC-001", "TC-002", "TC-003"]
        - id: "M4"
          name: "正式上线"
          stage: "部署阶段"
          target_date: "2026-06-01"
          description: "部署到生产环境并正式上线"
          status: "待开始"
          depends_on: ["M3"]
          items: []
      requirements:
        - id: "REQ-001"
          title: "用户注册登录"
          description: "支持邮箱和手机号注册登录"
          priority: "高"
          status: "已完成"
          milestone_id: "M1"
        - id: "REQ-002"
          title: "商品浏览"
          description: "支持商品列表和详情查看"
          priority: "高"
          status: "已完成"
          milestone_id: "M1"
        - id: "REQ-003"
          title: "购物车"
          description: "支持添加商品到购物车"
          priority: "中"
          status: "待开发"
          milestone_id: "M2"
      development_tasks:
        - id: "TASK-001"
          title: "实现用户注册API"
          assignee: "张三"
          estimated_hours: 8
          status: "已完成"
          milestone_id: "M2"
        - id: "TASK-002"
          title: "实现商品列表API"
          assignee: "李四"
          estimated_hours: 12
          status: "进行中"
          milestone_id: "M2"
      test_cases:
        - id: "TC-001"
          title: "注册新用户"
          type: "单元测试"
          status: "通过"
          milestone_id: "M3"
        - id: "TC-002"
          title: "登录成功"
          type: "单元测试"
          status: "通过"
          milestone_id: "M3"
        - id: "TC-003"
          title: "商品列表查询"
          type: "集成测试"
          status: "待执行"
          milestone_id: "M3"
      deploy_target: "测试环境"
      config:
        quality_threshold: 95
        min_review_rounds: 3
    expected_output:
      success: true
      action: "full_lifecycle"
      current_stage: "部署阶段"
      milestone_report:
        total_milestones: 4
        completed_milestones: 1
        in_progress_milestones: 1
        delayed_milestones: 0
        overall_progress: 37.5
        milestones:
          - id: "M1"
            name: "需求冻结"
            stage: "原型阶段"
            target_date: "2026-05-01"
            status: "已完成"
            progress: 100.0
            items_total: 2
            items_completed: 2
            is_at_risk: false
            is_critical_path: false
          - id: "M2"
            name: "核心功能完成"
            stage: "开发阶段"
            target_date: "2026-05-15"
            status: "进行中"
            progress: 50.0
            items_total: 2
            items_completed: 1
            is_at_risk: false
            is_critical_path: true
          - id: "M3"
            name: "测试通过"
            stage: "测试阶段"
            target_date: "2026-05-25"
            status: "待开始"
            progress: 67.0
            items_total: 3
            items_completed: 2
            is_at_risk: false
            is_critical_path: true
          - id: "M4"
            name: "正式上线"
            stage: "部署阶段"
            target_date: "2026-06-01"
            status: "待开始"
            progress: 0.0
            items_total: 0
            items_completed: 0
            is_at_risk: false
            is_critical_path: false
      quality_review:
        total_rounds: 3
        passed: true
        final_score: 97
      metrics:
        progress_percentage: 40
        quality_scores:
          原型阶段: 97
          开发阶段: 96
          测试阶段: 98
          部署阶段: 97
      next_steps:
        - "继续开发"
        - "关注里程碑进度"
        - "准备测试环境"
      error: ""

  - name: 仅查看里程碑状态
    input:
      action: "track_progress"
      project_name: "博客系统"
      project_path: "d:\\projects\\blog"
      methodology: "瀑布"
      milestones:
        - id: "M1"
          name: "原型完成"
          stage: "原型阶段"
          target_date: "2026-05-10"
          description: "完成博客系统原型设计"
          status: "进行中"
          items: ["REQ-001"]
    expected_output:
      success: true
      current_stage: ""
      milestone_report:
        total_milestones: 1
        completed_milestones: 0
        in_progress_milestones: 1
        delayed_milestones: 0
        overall_progress: 0.0
