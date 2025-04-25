# Val-Store 后端服务

这是Val-Store的后端服务器代码，使用Go语言开发，用于提供Valorant皮肤商店信息查询服务。

## 项目结构

```
server/
├── cmd/                # 应用程序入口点
│   └── server/         # 主服务器应用
├── internal/           # 私有应用程序和库代码
│   ├── api/            # API处理和路由
│   │   ├── handlers/   # HTTP处理器
│   │   ├── middleware/ # HTTP中间件
│   │   └── router.go   # 路由配置
│   ├── config/         # 配置管理
│   ├── models/         # 数据模型
│   ├── repositories/   # 数据存储和外部API交互
│   └── services/       # 业务逻辑
├── .env.example        # 环境变量示例
├── go.mod              # Go模块文件
└── go.sum              # Go模块校验和
```

## 使用方法

1. 复制`.env.example`到`.env`并填写相应配置
2. 运行`go build -o server ./cmd/server`编译服务器
3. 执行`./server`启动服务器

## API接口

- `/api/auth` - 用户认证相关接口
- `/api/shop` - 商店信息查询接口
- `/api/skins` - 皮肤信息查询接口
- `/api/user` - 用户信息管理接口

## 测试

使用`test_shop_api.sh`脚本测试商店API接口 