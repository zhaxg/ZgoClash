# ZgoClash

一个紧凑的 Windows mihomo 客户端：系统代理、TUN、订阅管理、节点选择与实时连接都在一个原生界面里。

界面基于 [MyGo](https://github.com/egoist/mygo) 的原生 UI（不是网页），内核是 [mihomo](https://github.com/MetaCubeX/mihomo)，以库的方式跑在同一个进程里。

## 功能

- **概览**：出站模式、系统代理与 TUN 开关、上下行流量与内存占用、服务连通性、实时连接。
- **代理**：按分组浏览节点，切换节点、测延迟。
- **订阅**：订阅链接 / 本地文件 / 手动配置，可编辑名称与链接、一键同步全部。
- **连接**：实时连接列表，可查看详情与断开全部。
- **日志**：mihomo 内核日志，可过滤与自动滚动。
- **设置**：外观、启动行为、端口、TUN 栈、延迟测试地址、订阅自动更新等。
- 托盘常驻：窗口最小化或收进托盘后暂停轮询并释放内存，核心继续在后台工作。

## 构建

需要 Go 1.27 与 [MyGo](https://github.com/egoist/mygo) 的工具链：

```bash
go tool mygo build      # 产物在 bin/windows-amd64/
go tool mygo dev        # 开发模式，改动自动热重载
```

## 目录

| 路径 | 内容 |
| --- | --- |
| `main.go` | 窗口、托盘与单实例 |
| `internal/app` | 设置、订阅与 mihomo 核心的生命周期 |
| `internal/engine` | mihomo 的封装与 REST 轮询 |
| `internal/ui` | 界面（概览、代理、订阅、连接、日志、设置） |
| `resources` | 应用图标与托盘图标 |
