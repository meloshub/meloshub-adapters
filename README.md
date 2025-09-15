# meloshub-adapters

meloshub的适配器仓库

## 适配器规范

这是meloshub的适配器仓库，所有上传的适配器应遵守以下规范：

1. 请按照meloshub的适配器接口规范进行平台的具体实现

2. 适配器的网络通信部分应使用`meloshub/network`,日志模块应使用`meloshub/logging`。管理上下文会话统一使用adapter.Session类

3. 除官方适配器之外，请将适配器元数据中的适配器类型标注为：community（社区）类型

4. 如需编写测试，请在tests包下编写名为*_test.go的测试

## 适配器开发

> [!IMPORTANT]
> 注意：请不要手动修改适配器元数据文件adapter.yml与集合入口all.go这两个文件。

fork本仓库，在项目根目录下新建一个新适配器的package。

### 1. 定义适配器

定义一个新的适配器结构体，并继承adapter.Base：

```go
type ExampleAdapter struct {
	adapter.Base
}
```

### 2. 实现接口

需要实现adapter.Adapter中定义的所有方法：

```go
func (a *ExampleAdapter) Search(keyword string, options adapter.SearchOptions) ([]model.Song, error) {
	return []model.Song{}, nil
}

func (a *ExampleAdapter) PlayURL(id string) (string, error) {
	return "", nil
}

func (a *ExampleAdapter) Lyrics(id string) (string, error) {
	return "", nil
}

func (a *ExampleAdapter) AlbumDetail(id string) (model.Album, error) {
	return model.Album{}, nil
}
```

### 3. 注册适配器

在构造函数中使用Init方法初始化适配器，在init中使用adapter.Register接收构造函数：

```go
func init() {
	if err := adapter.Register(New()); err != nil {
		panic(fmt.Errorf("failed to register adapter: %w", err))
	}
}

func New() *ExampleAdapter {
	a := &ExampleAdapter{}
	a.Init("example")
	return a
}
```

### 4. 提交并触发CI

当你发起pull requests时，仓库的actions会执行一系列流程，主要包括：

1. 代码安全与格式化审查 

2. 适配器元数据(adapter.yml)生成 

3. 适配器新增，升级与移除变动通知

4. 项目集合入口(all.go)生成

如果没有其他的问题，当pr通过后，你的适配器就可以在主框架被正常导入和使用了。

另外，当适配器通过审核后，请手动执行git pull以同步适配器元数据文件adapter.yml与集合入口all.go