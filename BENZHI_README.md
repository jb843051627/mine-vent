# mine-vent (矿井通风监测调度系统)

## 构建

```bash
docker build -f benzhi.Dockerfile -t benzhi/mine-vent:latest .
```

## 运行

```bash
docker run -p 8080:8080 benzhi/mine-vent:latest
```

## API

- GET /api/sensors - 列出所有传感器
- POST /api/sensors - 创建传感器
- GET /api/readings?sensor=X - 列出传感器读数
- POST /api/readings - 记录读数
- GET /api/fans - 列出所有风机
- POST /api/fans - 创建风机
- GET /api/schedules - 列出调度
- POST /api/schedules - 创建调度
- GET /api/alerts - 列出告警
- GET /api/maintenance - 列出维护任务
- GET /api/dashboard - 仪表盘数据
