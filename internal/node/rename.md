# 节点重命名表达式

重命名表达式使用 Go `text/template` 语法。预览接口和正式重命名都调用 `Rename`，因此解析、字段、函数和错误行为一致。

字段名区分大小写，只支持下表中的真实字段；不存在 `.index`、`.name` 或 `.Name`。

| 字段 | 含义 |
| --- | --- |
| `.Index` | 最终输出顺序中的序号 |
| `.Delay` | 延迟，单位毫秒 |
| `.DownloadSpeed` | 下载速度，单位 kb/s |
| `.TrafficMultiplier` | 流量扣费倍率，节点名称未标注时为 1 |
| `.Country.Alpha2` | ISO 3166-1 alpha-2 国家代码 |
| `.Country.NameEn` | 国家英文名 |
| `.Country.NameZh` | 国家中文名 |
| `.Country.Flag` | 国家或地区旗帜 emoji |

## 预览接口

`POST /api/v1/rename/preview` 需要登录，请求体只接收 `expression`：

```json
{
  "expression": "{{.Country.Flag}} {{.Country.NameZh}}-{{printf \"%03d\" .Index}}"
}
```

接口固定使用 `.Index=1`、`.Delay=123`、`.DownloadSpeed=10240`、`.TrafficMultiplier=0.5`、`.Country.Alpha2=CN` 预览。

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "result": "🇨🇳 中国-001"
  }
}
```

表达式为空、语法错误、字段不存在或函数参数错误时返回 HTTP 400，错误原因放在 `message` 中。

## 完整命名示例


表达式

```gotemplate
{{.Country.Flag}} {{printf "%03d" .Index}}{{if gt .DownloadSpeed 0}} | ⬇️ {{if ge .DownloadSpeed 1024}}{{div .DownloadSpeed 1024}}MB/s{{else}}{{.DownloadSpeed}}KB/s{{end}}{{end}}
```

结果

```text
🇨🇳 001 | ⬇️ 10MB/s
```

同一表达式在 `.DownloadSpeed=0` 时不会留下多余的分隔符。

表达式

```gotemplate
{{.Country.Flag}} {{printf "%03d" .Index}}{{if gt .DownloadSpeed 0}} | ⬇️ {{if ge .DownloadSpeed 1024}}{{div .DownloadSpeed 1024}}MB/s{{else}}{{.DownloadSpeed}}KB/s{{end}}{{end}}
```

结果

```text
🇨🇳 001
```

## 字段示例

以下示例均使用预览接口的固定数据。

### 序号

表达式

```gotemplate
{{.Index}}
```

结果

```text
1
```

### 延迟

表达式

```gotemplate
延迟 {{.Delay}} ms
```

结果

```text
延迟 123 ms
```

### 下载速度

表达式

```gotemplate
{{.DownloadSpeed}} kb/s
```

结果

```text
10240 kb/s
```

### 国家代码

表达式

```gotemplate
{{.Country.Alpha2}}-{{.Index}}
```

结果

```text
CN-1
```

### 国家英文名

表达式

```gotemplate
{{.Country.NameEn}}-{{.Index}}
```

结果

```text
China-1
```

### 国家中文名

表达式

```gotemplate
{{.Country.NameZh}}-{{.Index}}
```

结果

```text
中国-1
```

### 旗帜

表达式

```gotemplate
{{.Country.Flag}} {{.Country.NameZh}} {{.Index}}
```

结果

```text
🇨🇳 中国 1
```

## 格式化示例

### 序号补零

表达式

```gotemplate
{{printf "%03d" .Index}}
```

结果

```text
001
```

### 延迟补零并追加单位

表达式

```gotemplate
{{printf "%04dms" .Delay}}
```

结果

```text
0123ms
```

### 下载速度补零

表达式

```gotemplate
{{printf "%06d kb/s" .DownloadSpeed}}
```

结果

```text
010240 kb/s
```

### 一次格式化多个字段

表达式

```gotemplate
{{printf "%s-%s-%03d" .Country.Flag .Country.Alpha2 .Index}}
```

结果

```text
🇨🇳-CN-001
```

### 使用 `print` 拼接

表达式

```gotemplate
{{print .Country.NameEn "-" .Index}}
```

结果

```text
China-1
```

### 使用管道格式化

表达式

```gotemplate
{{.Country.NameEn | printf "%s Node"}}
```

结果

```text
China Node
```

### 算术结果通过管道补零

表达式

```gotemplate
{{add .Index 9 | printf "%03d"}}
```

结果

```text
010
```

### 去除左侧空白

表达式

```gotemplate
节点 {{- .Index}}
```

结果

```text
节点1
```

### 去除右侧空白

表达式

```gotemplate
{{.Country.NameZh -}} 节点
```

结果

```text
中国节点
```

### 模板变量

表达式

```gotemplate
{{$number := printf "%03d" .Index}}{{$number}}-{{.Country.Alpha2}}
```

结果

```text
001-CN
```

### 使用 `with` 缩短国家字段

表达式

```gotemplate
{{with .Country}}{{.Flag}} {{.NameZh}}{{end}}
```

结果

```text
🇨🇳 中国
```

## 条件判断示例

### 判断国家

表达式

```gotemplate
{{if eq .Country.Alpha2 "CN"}}国内{{else}}海外{{end}}-{{.Index}}
```

结果

```text
国内-1
```

### 判断延迟

表达式

```gotemplate
{{if lt .Delay 200}}低延迟{{else}}高延迟{{end}}
```

结果

```text
低延迟
```

### 仅在速度大于零时显示速度

表达式

```gotemplate
{{if gt .DownloadSpeed 0}}⬇️ {{if ge .DownloadSpeed 1024}}{{div .DownloadSpeed 1024}}MB/s{{else}}{{.DownloadSpeed}}KB/s{{end}}{{end}}
```

结果

```text
⬇️ 10MB/s
```

### 同时满足多个条件

表达式

```gotemplate
{{if and (lt .Delay 200) (ge .DownloadSpeed 10000)}}优选{{else}}普通{{end}}
```

结果

```text
优选
```

### 满足任一条件

表达式

```gotemplate
{{if or (eq .Country.Alpha2 "CN") (eq .Country.Alpha2 "HK")}}中文区{{else}}其他{{end}}
```

结果

```text
中文区
```

### 条件取反

表达式

```gotemplate
{{if not (eq .Country.Alpha2 "US")}}非美国{{else}}美国{{end}}
```

结果

```text
非美国
```

可用比较函数为 `eq`、`ne`、`lt`、`le`、`gt`、`ge`，逻辑函数为 `and`、`or`、`not`。

## 算术函数示例

所有算术函数均接收两个 `uint32`，只进行整数运算，不保留小数。

### 加法 `add`

表达式

```gotemplate
{{add .Index 1}}
```

结果

```text
2
```

### 减法 `sub`

表达式

```gotemplate
{{sub .Delay 23}}
```

结果

```text
100
```

### 减法小于零

表达式

```gotemplate
{{sub .Index 2}}
```

结果

```text
0
```

### 乘法 `mul`

表达式

```gotemplate
{{mul .Index 10}}
```

结果

```text
10
```

### 除法 `div`

表达式

```gotemplate
{{div .DownloadSpeed 1024}}
```

结果

```text
10
```

### 取模 `mod`

表达式

```gotemplate
{{mod .Delay 100}}
```

结果

```text
23
```

### 嵌套运算

表达式

```gotemplate
{{add (mul .Index 10) 5}}
```

结果

```text
15
```

### 除数为零

表达式

```gotemplate
{{div .DownloadSpeed 0}}
```

结果

```text
0
```

### 对零取模

表达式

```gotemplate
{{mod .DownloadSpeed 0}}
```

结果

```text
0
```

`sub` 小于零时返回 `0`；`div` 和 `mod` 的除数为 `0` 时也返回 `0`。

## 错误示例

### 字段不存在

表达式

```gotemplate
{{.Name}}
```

结果

```text
HTTP 400：执行模板失败，字段不存在
```

### 模板没有结束

表达式

```gotemplate
{{if .Delay}}低延迟
```

结果

```text
HTTP 400：解析模板失败，unexpected EOF
```

### 函数参数不足

表达式

```gotemplate
{{div .Delay}}
```

结果

```text
HTTP 400：执行模板失败，div 需要两个参数
```
