# go-codec

从 `hkim-data-transfer/lib/data` 中提取出的可复用二进制编解码库。

## 范围

当前仓库只包含核心 codec 层：

- 基于反射的 marshal / unmarshal
- `encode` / `decode` 结构体标签
- 可配置的标量字节序
- 可配置的位域布局
- 定长字符串处理
- 可选的 `file` 标签钩子

它**不**包含协议组包、命令映射、MQ 封装或项目特定的存储行为。

## API

```go
import "github.com/weiwentao996/go-codec/codec"

b, err := codec.Marshal(v)
err := codec.Unmarshal(b, &v)
```

### 消费 buffer 的流式解码

```go
buf := bytes.NewBuffer(payload)
err := codec.Decode(buf, &header)
err = codec.Decode(buf, &body)
```

当调用方需要在同一个 `*bytes.Buffer` 上连续解码多段数据时，使用 `Decode`。
如果解码失败，buffer 可能已经被部分消费，目标对象也可能已经被部分填充。

## 配置

### 旧版兼容预设

```go
b, err := codec.Marshal(v, codec.WithLegacyHKimPreset())
err := codec.Unmarshal(b, &v, codec.WithLegacyHKimPreset())
```

对于需要从共享 buffer 分段解码的旧流程：

```go
buf := bytes.NewBuffer(payload)
err := codec.Decode(buf, &part1, codec.WithLegacyHKimPreset())
err = codec.Decode(buf, &part2, codec.WithLegacyHKimPreset())
```

这个预设会保留 `hkim-data-transfer` 的历史行为：

- `ByteOrder = BigEndian`
- `BitLayout = LegacySameDevice`

### 字节序

支持的全局字节序模式：

- `codec.BigEndian`
- `codec.LittleEndian`

示例：

```go
b, err := codec.Marshal(v, codec.WithByteOrder(codec.LittleEndian))
err := codec.Unmarshal(b, &v, codec.WithByteOrder(codec.LittleEndian))
```

说明：

- 字节序会影响 `uint16`、`uint32`、`uint64`、`float32`、`float64` 等多字节标量
- `uint8`、`int8`、`bool` 基本不受 endian 影响

### 位域布局

支持的全局位域布局：

- `codec.LegacySameDevice`
- `codec.LSBFirstLowToHigh`

示例：

```go
b, err := codec.Marshal(v, codec.WithBitLayout(codec.LSBFirstLowToHigh))
err := codec.Unmarshal(b, &v, codec.WithBitLayout(codec.LSBFirstLowToHigh))
```

#### `LegacySameDevice`

这是旧项目使用的兼容模式。

行为概述：

- 位域使用历史上的打包顺序
- 保持当前协议兼容性
- 默认配置就是这个模式

#### `LSBFirstLowToHigh`

这是一个面向测试和未来协议支持的备选布局。

行为概述：

- 低位优先输出
- 每个字节从低 bit 到高 bit 依次填充
- 与 `LegacySameDevice` 的编码结果不同，因此编码和解码必须使用同一布局

## Tags

当前支持的标签：

- `encode:"-"` / `decode:"-"`
- `bitCount:N`
- `subBitCount:N`
- `byteCount:N`
- `file`

示例：

```go
type Packet struct {
    Flags [8]uint8 `encode:"subBitCount:1" decode:"subBitCount:1"`
    Name  string   `encode:"byteCount:8" decode:"byteCount:8"`
    A     uint8    `encode:"bitCount:3" decode:"bitCount:3"`
}
```

## `file` 标签钩子

核心库本身不执行项目特定的文件 IO。
如果使用 `file` 标签，必须提供对应钩子。

### 编码侧文件读取器

```go
b, err := codec.Marshal(v, codec.WithFileReader(func(path string) ([]byte, error) {
    return os.ReadFile(path)
}))
```

### 解码侧文件写入器

```go
err := codec.Unmarshal(data, &v, codec.WithFileWriter(func(typeName string, data []byte) (string, error) {
    path := filepath.Join("out", typeName)
    return path, os.WriteFile(path, data, 0o644)
}))
```

如果没有配置钩子：

- encode 使用 `file` 标签时会返回 `ErrFileReaderNotConfigured`
- decode 使用 `file` 标签时会返回 `ErrFileWriterNotConfigured`

## 兼容性说明

默认配置会保持 `hkim-data-transfer` 的现有历史行为：

- 大端标量编码 / 解码
- 当前位域顺序（`LegacySameDevice`）
- 编码时定长字符串补零
- 解码时定长字符串按历史逻辑做 zero-to-space + trim
- 指针 / interface 解码兼容
- decode 侧兼容 `slice + byteCount` 的历史用法

## 当前状态

当前阶段重点：

- 行为兼容
- 显式的字节序和位域布局配置
- 可复用核心抽取

### 最近更新

- 修复了在 `bitCount` / `subBitCount` 字段与普通字节对齐字段切换时的字节对齐问题
- 增加了 `bitCount`、`subBitCount`、`byteCount`、`file` 的标签类型校验
- 引入结构体字段标签与分发决策的 metadata 缓存，减少重复反射开销
- 增加了流式 `Decode(*bytes.Buffer, ...)` 接口，支持共享 buffer 的顺序解码
- 恢复了 decode 侧 `slice + byteCount` 的兼容语义
- 增加了位域对齐、标签校验、重复调用、并发使用等回归测试
- 增加了小结构体、带标签结构体和嵌套结构体的 benchmark

### 暂未包含

- 字段级 endian 覆盖
- 字段级 bit layout 覆盖
- 协议 header/body/frame 抽象
- 命令注册表与传输适配层
