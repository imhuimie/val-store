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

## API接口文档

### 认证方式

Val-Store API使用JWT (JSON Web Token) 进行认证。除了登录接口和公开接口外，所有受保护的接口都需要在请求头中包含有效的JWT令牌：

```
Authorization: Bearer <your_jwt_token>
```

JWT令牌通过登录接口获取，有效期默认为24小时。

### 响应格式

所有API响应都使用统一的JSON格式：

#### 成功响应格式

```json
{
  "status": 200,         // HTTP状态码
  "message": "成功消息",  // 操作成功的描述
  "data": {              // 响应数据（可选）
    // 具体数据字段
  }
}
```

#### 错误响应格式

```json
{
  "status": 400,         // HTTP错误状态码
  "message": "错误消息",  // 错误的简短描述
  "error": "详细错误信息" // 详细的错误原因（可选）
}
```

### 区域支持

Val-Store支持以下Valorant游戏区域：

- `ap`: 亚太地区
- `na`: 北美
- `eu`: 欧洲
- `kr`: 韩国
- `latam`: 拉丁美洲
- `br`: 巴西

### API接口详细说明

#### 1. 认证接口 (`/api/auth`)

##### 1.1 登录

- **URL**: `/api/auth/login`
- **方法**: `POST`
- **描述**: 使用Riot游戏账号用户名和密码登录
- **请求体**:
  ```json
  {
    "username": "your_riot_username",
    "password": "your_riot_password"
  }
  ```
- **响应**: 登录成功返回JWT令牌和用户信息
  ```json
  {
    "status": 200,
    "message": "登录成功",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIs...",
      "user": {
        "username": "your_username",
        "user_id": "your_user_id"
      }
    }
  }
  ```

##### 1.2 Cookie登录

- **URL**: `/api/auth/login/cookies`
- **方法**: `POST`
- **描述**: 使用Riot游戏Cookies登录
- **请求体**:
  ```json
  {
    "cookies": "ssid=xxx; csid=xxx; ...",
    "region": "ap"  // 可选，指定游戏区域
  }
  ```
- **响应**: 与常规登录相同

##### 1.3 健康检查

- **URL**: `/api/auth/ping`
- **方法**: `GET`
- **描述**: 检查服务是否正常运行
- **响应**:
  ```json
  {
    "status": 200,
    "message": "服务正常运行"
  }
  ```

#### 2. 商店接口 (`/api/shop`)

##### 2.1 获取每日商店

- **URL**: `/api/shop`
- **方法**: `GET`
- **描述**: 获取用户的每日商店皮肤信息
- **认证**: 需要JWT认证
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取商店数据",
    "data": {
      "featured_bundle": {
        // 精选捆绑包数据
      },
      "daily_offers": [
        {
          "offer_id": "skin_id",
          "name": "皮肤名称",
          "price": 1775,
          "image_url": "皮肤图片URL",
          "content_tier": "Deluxe", // 皮肤等级
          "expires_in": "23h 45m"   // 到期时间
        },
        // 更多每日皮肤...
      ]
    }
  }
  ```

#### 3. 皮肤接口 (`/api/skins`)

##### 3.1 获取所有皮肤列表

- **URL**: `/api/skins`
- **方法**: `GET`
- **描述**: 获取系统中所有皮肤的列表
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取皮肤列表",
    "data": [
      {
        "id": "skin_id",
        "name": "皮肤名称",
        "weapon": "武器类型",
        "price": 1775,
        "tier": "Deluxe",
        "image_url": "皮肤图片URL"
      },
      // 更多皮肤...
    ]
  }
  ```

##### 3.2 获取单个皮肤信息

- **URL**: `/api/skins/:id`
- **方法**: `GET`
- **描述**: 根据ID获取单个皮肤的详细信息
- **参数**: `id` - 皮肤的唯一标识符
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取皮肤信息",
    "data": {
      "id": "skin_id",
      "name": "皮肤名称",
      "weapon": "武器类型",
      "price": 1775,
      "tier": "Deluxe",
      "image_url": "皮肤图片URL",
      "description": "皮肤描述",
      "variants": [
        // 皮肤变体信息，如果有
      ],
      "videos": [
        // 皮肤相关视频，如果有
      ]
    }
  }
  ```

#### 4. 用户接口 (`/api/user`)

##### 4.1 获取用户信息

- **URL**: `/api/user/info`
- **方法**: `GET`
- **描述**: 获取当前登录用户的基本信息
- **认证**: 需要JWT认证
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取用户信息",
    "data": {
      "username": "your_username",
      "user_id": "your_user_id",
      "riot_username": "游戏用户名",
      "riot_tagline": "游戏标签",
      "region": "当前选择的区域"
    }
  }
  ```

##### 4.2 获取用户钱包信息

- **URL**: `/api/user/wallet`
- **方法**: `GET`
- **描述**: 获取用户在游戏中的虚拟货币余额
- **认证**: 需要JWT认证
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取钱包数据",
    "data": {
      "valorant_points": 1000,    // VP点数
      "radianite_points": 20,     // 辐能点数
      "kingdom_credits": 500      // 王国信用点
    }
  }
  ```

##### 4.3 设置用户区域

- **URL**: `/api/user/region`
- **方法**: `POST`
- **描述**: 设置用户的游戏区域
- **认证**: 需要JWT认证
- **请求体**:
  ```json
  {
    "region": "ap"  // 区域代码，可选值：ap, na, eu, kr, latam, br
  }
  ```
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功设置区域",
    "data": {
      "region": "ap"
    }
  }
  ```

##### 4.4 获取支持的区域列表

- **URL**: `/api/regions`
- **方法**: `GET`
- **描述**: 获取系统支持的所有游戏区域
- **响应**:
  ```json
  {
    "status": 200,
    "message": "成功获取支持的区域列表",
    "data": {
      "regions": ["ap", "na", "eu", "kr", "latam", "br"]
    }
  }
  ```

### API使用示例

以下是使用curl命令调用API接口的示例：

#### 登录并获取令牌

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}'
```

#### 使用令牌获取商店数据

```bash
curl -X GET http://localhost:8080/api/shop \
  -H "Authorization: Bearer your_jwt_token"
```

#### 设置游戏区域

```bash
curl -X POST http://localhost:8080/api/user/region \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_jwt_token" \
  -d '{"region":"eu"}'
```

#### 获取皮肤信息

```bash
curl -X GET http://localhost:8080/api/skins/skin_id_here
```

## 测试

使用`test_shop_api.sh`脚本测试商店API接口，该脚本提供了商店接口的基本测试功能：

```bash
./test_shop_api.sh
```

## 错误处理

API会返回标准的HTTP状态码和错误信息：

- `400` Bad Request - 请求参数有误
- `401` Unauthorized - 认证失败或令牌无效
- `404` Not Found - 资源不存在
- `500` Internal Server Error - 服务器内部错误 