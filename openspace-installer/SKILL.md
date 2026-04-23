---
name: OpenSpace 安装器
version: 1.0.0
description: 自动化安装和配置 OpenSpace AI 代理进化引擎
author: OpenSpace
category: 系统
tags: [安装, 配置, OpenSpace, AI, 代理]
icon: 🚀

# 所需工具
requires:
  - run_shell
  - read_file
  - write_file

# 输入参数
input_schema:
  type: object
  properties:
    install_path:
      type: string
      description: OpenSpace 安装路径
      default: "~/OpenSpace"
    python_version:
      type: string
      description: 要求的 Python 版本 (3.12+)
      default: "3.12"
    install_dependencies:
      type: boolean
      description: 是否安装依赖
      default: true
    configure_api_key:
      type: boolean
      description: 是否配置 API Key
      default: false
    api_key:
      type: string
      description: DeepSeek API Key (当 configure_api_key 为 true 时必填)
      example: "sk-xxx"
    api_base_url:
      type: string
      description: API 基础 URL
      default: "https://api.deepseek.com"

# 输出参数
output_schema:
  type: object
  properties:
    success:
      type: boolean
      description: 安装是否成功
    version:
      type: string
      description: OpenSpace 版本
    install_path:
      type: string
      description: 安装路径
    python_version:
      type: string
      description: 检测到的 Python 版本
    steps:
      type: array
      description: 执行的步骤
      items:
        type: object
        properties:
          name:
            type: string
          status:
            type: boolean
          message:
            type: string
    error:
      type: string
      description: 错误信息

# 实现
implementation: |
  def install_openspace(install_path="~/OpenSpace", python_version="3.12", install_dependencies=True, configure_api_key=False, api_key="", api_base_url="https://api.deepseek.com"):
      """
      安装和配置 OpenSpace AI 代理进化引擎
      
      参数:
          install_path: 安装路径
          python_version: 要求的 Python 版本
          install_dependencies: 是否安装依赖
          configure_api_key: 是否配置 API Key
          api_key: DeepSeek API Key
          api_base_url: API 基础 URL
      
      返回:
          包含安装状态和详细信息的字典
      """
      import subprocess
      import os
      import re
      
      # 初始化结果
      result = {
          "success": False,
          "version": "",
          "install_path": install_path,
          "python_version": "",
          "steps": [],
          "error": ""
      }
      
      # 步骤 1: 检查 Python 版本
      try:
          python_check = subprocess.run(
              ["python", "--version"],
              capture_output=True,
              text=True
          )
          if python_check.returncode == 0:
              python_ver = python_check.stdout.strip()
              result["python_version"] = python_ver
              
              # 提取版本号
              ver_match = re.search(r"Python (\d+\.\d+)", python_ver)
              if ver_match:
                  major_minor = ver_match.group(1)
                  required = python_version
                  if float(major_minor) < float(required):
                      result["error"] = f"Python 版本过低，需要 {required}+，当前版本: {major_minor}"
                      return result
              
              result["steps"].append({
                  "name": "检查 Python 版本",
                  "status": True,
                  "message": f"Python 版本: {python_ver}"
              })
          else:
              result["error"] = "未找到 Python 安装"
              return result
      except Exception as e:
          result["error"] = f"检查 Python 版本失败: {str(e)}"
          return result
      
      # 步骤 2: 克隆 OpenSpace 代码
      try:
          # 展开路径
          install_path = os.path.expanduser(install_path)
          
          # 检查目录是否存在
          if os.path.exists(install_path):
              result["steps"].append({
                  "name": "克隆代码",
                  "status": True,
                  "message": f"目录已存在: {install_path}"
              })
          else:
              # 克隆代码
              clone_cmd = [
                  "git", "clone", "--filter=blob:none", "--sparse",
                  "https://github.com/HKUDS/OpenSpace.git",
                  install_path
              ]
              clone_result = subprocess.run(
                  clone_cmd,
                  capture_output=True,
                  text=True
              )
              
              if clone_result.returncode == 0:
                  # 配置 sparse checkout
                  sparse_cmd = [
                      "git", "sparse-checkout", "set", "*", "!assets"
                  ]
                  subprocess.run(
                      sparse_cmd,
                      cwd=install_path,
                      capture_output=True,
                      text=True
                  )
                  
                  result["steps"].append({
                      "name": "克隆代码",
                      "status": True,
                      "message": f"成功克隆到: {install_path}"
                  })
              else:
                  result["error"] = f"克隆代码失败: {clone_result.stderr}"
                  return result
      except Exception as e:
          result["error"] = f"克隆代码失败: {str(e)}"
          return result
      
      # 步骤 3: 安装依赖
      if install_dependencies:
          try:
              install_cmd = ["pip", "install", "-e", install_path]
              install_result = subprocess.run(
                  install_cmd,
                  capture_output=True,
                  text=True
              )
              
              if install_result.returncode == 0:
                  result["steps"].append({
                      "name": "安装依赖",
                      "status": True,
                      "message": "依赖安装成功"
                  })
              else:
                  result["error"] = f"依赖安装失败: {install_result.stderr}"
                  return result
          except Exception as e:
              result["error"] = f"依赖安装失败: {str(e)}"
              return result
      else:
          result["steps"].append({
              "name": "安装依赖",
              "status": True,
              "message": "跳过依赖安装"
          })
      
      # 步骤 4: 配置 API Key
      if configure_api_key and api_key:
          try:
              env_file = os.path.join(install_path, "openspace", ".env")
              env_content = f"# OpenSpace Environment Variables\n"
              env_content += f"OPENAI_API_KEY={api_key}\n"
              env_content += f"OPENAI_BASE_URL={api_base_url}\n"
              
              with open(env_file, "w", encoding="utf-8") as f:
                  f.write(env_content)
              
              result["steps"].append({
                  "name": "配置 API Key",
                  "status": True,
                  "message": "API Key 配置成功"
              })
          except Exception as e:
              result["error"] = f"配置 API Key 失败: {str(e)}"
              return result
      elif configure_api_key:
          result["error"] = "配置 API Key 时缺少 api_key 参数"
          return result
      else:
          result["steps"].append({
              "name": "配置 API Key",
              "status": True,
              "message": "跳过 API Key 配置"
          })
      
      # 步骤 5: 验证安装
      try:
          verify_cmd = ["python", "-m", "openspace", "--help"]
          verify_result = subprocess.run(
              verify_cmd,
              cwd=install_path,
              capture_output=True,
              text=True
          )
          
          if verify_result.returncode == 0:
              result["success"] = True
              result["version"] = "0.1.0"
              result["steps"].append({
                  "name": "验证安装",
                  "status": True,
                  "message": "OpenSpace 安装成功"
              })
          else:
              result["error"] = f"验证安装失败: {verify_result.stderr}"
              return result
      except Exception as e:
          result["error"] = f"验证安装失败: {str(e)}"
          return result
      
      return result

# 使用示例
usage:
  - name: 基本安装
    input:
      install_path: "~/OpenSpace"
      python_version: "3.12"
      install_dependencies: true
      configure_api_key: false
    expected_output:
      success: true
      version: "0.1.0"
      install_path: "~/OpenSpace"
      python_version: "Python 3.13.13"
      steps:
        - name: "检查 Python 版本"
          status: true
          message: "Python 版本: Python 3.13.13"
        - name: "克隆代码"
          status: true
          message: "成功克隆到: ~/OpenSpace"
        - name: "安装依赖"
          status: true
          message: "依赖安装成功"
        - name: "配置 API Key"
          status: true
          message: "跳过 API Key 配置"
        - name: "验证安装"
          status: true
          message: "OpenSpace 安装成功"
      error: ""
  
  - name: 带 API Key 配置
    input:
      install_path: "d:\\workspace\\OpenSpace"
      python_version: "3.12"
      install_dependencies: true
      configure_api_key: true
      api_key: "sk-8e5edc56b11a44c59940a50b4cc9a870"
      api_base_url: "https://api.deepseek.com"
    expected_output:
      success: true
      version: "0.1.0"
      install_path: "d:\\workspace\\OpenSpace"
      python_version: "Python 3.13.13"
      steps:
        - name: "检查 Python 版本"
          status: true
          message: "Python 版本: Python 3.13.13"
        - name: "克隆代码"
          status: true
          message: "成功克隆到: d:\\workspace\\OpenSpace"
        - name: "安装依赖"
          status: true
          message: "依赖安装成功"
        - name: "配置 API Key"
          status: true
          message: "API Key 配置成功"
        - name: "验证安装"
          status: true
          message: "OpenSpace 安装成功"
      error: ""
