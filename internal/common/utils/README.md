# HTTP Client 通用HTTP客户端

## 概述

`http_client.go` 提供了一个通用的HTTP客户端封装，简化了HTTP请求的发送和响应处理。

## 主要特性

- 支持GET、POST、PUT、DELETE等HTTP方法
- 自动JSON序列化/反序列化
- 内置错误处理和状态码检查
- 支持OAuth token认证
- 可配置的请求头和超时时间
- 支持查询参数

## 使用示例

### 1. 基本HTTP客户端

```go
ctx := context.Background()

// 创建HTTP客户端
client := utils.NewHTTPClient(utils.HTTPClientConfig{
    BaseURL:   "https://api.example.com",
    Timeout:   10 * time.Second,
    UserAgent: "my-app/1.0",
})

// GET请求
resp, err := client.Get(ctx, "/users", map[string]string{
    "page": "1",
    "limit": "10",
})
if err != nil {
    log.Fatal(err)
}

// 获取JSON响应
var users []User
if err := client.GetJSON(ctx, "/users", nil, &users); err != nil {
    log.Fatal(err)
}

// POST请求
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
var createdUser User
if err := client.PostJSON(ctx, "/users", nil, CreateUserRequest{
    Name:  "张三",
    Email: "zhangsan@example.com",
}, &createdUser); err != nil {
    log.Fatal(err)
}
```

### 2. OAuth客户端

```go
// 创建OAuth客户端（自动添加Authorization header）
oauthClient := utils.NewOAuthClient(
    utils.HTTPClientConfig{
        BaseURL:   "https://api.github.com",
        Timeout:   10 * time.Second,
        UserAgent: "my-app/1.0",
    },
    "your-access-token",
)

// 使用token访问受保护的API
var user User
if err := oauthClient.GetJSONWithToken(ctx, "/user", nil, &user); err != nil {
    log.Fatal(err)
}
```

### 3. 自定义请求

```go
// 自定义请求头
headers := map[string]string{
    "X-Custom-Header": "value",
}

// 发送自定义请求
resp, err := client.Request(
    ctx,
    http.MethodPost,
    "/api/endpoint",
    map[string]string{"param": "value"},
    payload,
    headers,
)
if err != nil {
    log.Fatal(err)
}
```

## API参考

### HTTPClient

- `NewHTTPClient(config HTTPClientConfig) *HTTPClient`: 创建新的HTTP客户端
- `Get(ctx, endpoint, params) (*HTTPResponse, error)`: 发送GET请求
- `Post(ctx, endpoint, params, body) (*HTTPResponse, error)`: 发送POST请求
- `Put(ctx, endpoint, params, body) (*HTTPResponse, error)`: 发送PUT请求
- `Delete(ctx, endpoint, params) (*HTTPResponse, error)`: 发送DELETE请求
- `GetJSON(ctx, endpoint, params, result) error`: GET请求并解析JSON响应
- `PostJSON(ctx, endpoint, params, body, result) error`: POST请求并解析JSON响应
- `Request(ctx, method, endpoint, params, body, headers) (*HTTPResponse, error)`: 发送自定义请求

### OAuthClient

- `NewOAuthClient(config HTTPClientConfig, accessToken string) *OAuthClient`: 创建OAuth客户端
- `OAuthGet(ctx, endpoint, params) (*HTTPResponse, error)`: 使用token发送GET请求
- `OAuthPost(ctx, endpoint, params, body) (*HTTPResponse, error)`: 使用token发送POST请求
- `GetJSONWithToken(ctx, endpoint, params, result) error`: 使用token获取JSON响应

### HTTPResponse

- `StatusCode`: HTTP状态码
- `Headers`: 响应头
- `Body`: 响应体（字节数组）
- `RequestID`: 请求ID（从响应头获取）

## 最佳实践

1. **重用客户端**: 为同一API端点创建客户端实例并重用，而不是为每个请求创建新实例
2. **设置超时**: 始终设置合理的超时时间，避免请求挂起
3. **错误处理**: 检查并处理所有HTTP请求错误
4. **上下文**: 使用context来控制请求的生命周期，特别是用于取消和超时
5. **日志记录**: 记录关键信息如请求URL、状态码和响应时间

## 在OAuth中的使用

在`oauth_repo.go`中，我们使用HTTP客户端来简化第三方OAuth API的调用：

```go
// GitHub OAuth
oauthClient := utils.NewOAuthClient(
    utils.HTTPClientConfig{
        BaseURL:   "https://api.github.com",
        Timeout:   10 * time.Second,
        UserAgent: "auth-perm/1.0",
    },
    token.AccessToken,
)

var githubUser GitHubUser
if err := oauthClient.GetJSONWithToken(ctx, "/user", nil, &githubUser); err != nil {
    return nil, errors.WrapBizError(err, "Failed to fetch GitHub user info")
}
```

这种方式比手动创建HTTP请求更简洁、更安全，并且提供了统一的错误处理。
