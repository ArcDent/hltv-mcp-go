# AGENTS.md

## 项目身份
- 类型：HLTV MCP 服务完全重建
- 目标：基于 hltv-mcp 源码分析，重新构建一个更优的 HLTV MCP 服务
- 技术栈：TypeScript + MCP SDK + Python Flask 上游（待定）

## 项目静态结构
- `analysis-hltv-mcp.md` — 原始 hltv-mcp 项目源码分析文档

## 最近操作
- 2026-05-26：完成原始 hltv-mcp 项目的完整源码分析，产出 `analysis-hltv-mcp.md`

## 进行中
- 建立项目基础设施（CLAUDE.md、目录结构）

## 下一步
- 梳理项目级 CLAUDE.md
- 确定重建目标和改进方向

## 关键发现
- 原始项目采用 TypeScript MCP + Python Flask/Scrapy 双层架构
- 核心模块 HltvFacade 约1560行，承担过多职责
- 中文本地化体系完善（70+队伍+20+赛事映射）
- Python 上游依赖 Scrapy 爬虫，启动和管理复杂
