# OpenSpace Installer Skill

## 描述

自动安装和配置 OpenSpace (HKUDS/OpenSpace) 的 Skill。

## 功能

- 克隆 OpenSpace 仓库
- 配置 API 密钥
- 验证安装
- 启动 OpenSpace 服务

## 使用方法

```
/install-openspace --api-key <your-api-key> --model <model-name>
```

## 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| --api-key | string | 是 | DeepSeek API 密钥 |
| --model | string | 否 | 模型名称，默认 deepseek-chat |
| --port | number | 否 | 服务端口，默认 8000 |

## 示例

```bash
/install-openspace --api-key sk-xxxx --model deepseek-chat
```

## 依赖

- Git
- Python 3.10+
- DeepSeek API 访问权限

## 版本

v1.0.0
