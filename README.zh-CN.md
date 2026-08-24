# mc-saver

一个用 Go 编写的 Minecraft 存档备份工具。它根据 JSON 规则文件，将世界存档中*选定的部分*——区块区域文件和核心数据文件——打包成带日期的 ZIP 压缩包。

English version: [README.md](README.md)

## 功能

- **选择性备份** — 只备份你关心的维度和区块，而不是整个世界，压缩包更小、速度更快。
- **按维度配置** — 可以选择性配置和备份主世界（`overworld`）、下界（`the_nether`）、末地（`the_end`）以及任意自定义维度。
- **灵活的区域选择** — 既可以用矩形 `range` 规则，也可以用 `simple` 规则逐个指定区域坐标。

## 下载

在 [Releases](https://github.com/fovlin/mc-saver/releases)页面下载对应的二进制可执行文件。

## 示例方法

```bash
./mc-saver
# 启动一个向导，引导用户进行备份，与 repl 子命令等效

./mc-saver gencfg
# 生成一个默认的配置文件 - save-rule.json

./mc-saver run
# 以默认值开始一个备份，默认值包括：
# 配置文件 - save-rule.json
# 世界存档目录 - world
# 输出目录 - world-$time.zip

./mc-saver -c config.json run level level.zip
# 开始一个备份，指定配置文件 config.json，世界文件 level，输出文件 level.zip

./mc-saver -l run
# 以默认值开始一个备份，使用旧版模式，1.21.11 前的存档应使用此模式：
```

### 语法

```
mc-saver [-c <配置文件>] [-l] <命令> [参数...]
```

参数必须放在子命令之前；子命令之后的内容一律按位置参数处理。

### 命令

| 命令 | 默认值 | 说明 |
| --- | --- | --- |
| `run` | `world`、`.` | 备份存档。位置参数：`<存档目录>` 和 `<输出路径>`（见[输出路径](#输出路径)）。 |
| `gencfg` | `save-rule.json` | 生成默认配置文件后退出；若文件已存在则会覆盖。可选位置参数：`<配置文件>`。 |
| `help` | — | 显示内置帮助文本后退出。 |
| `repl` | — | 运行交互向导（与不写子命令时相同）。 |
| （无命令） | — | 运行交互向导：依次提示输入存档目录和输出路径；缺少配置文件时会询问是否生成。 |

### 参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-c` | `save-rule.json` | 备份规则文件（JSON）的路径。 |
| `-l` | 关闭 | 按旧版单文件夹布局（`DIM-1`/`DIM1`）备份，见[旧版布局模式](#旧版布局模式-l)。 |

## 配置说明

规则文件是 JSON 格式，包含两个顶层字段：`dimension` 和 `file`。

```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    },
    "minecraft:the_nether": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    },
    "minecraft:the_end": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    }
  },
  "file": [
    "level.dat",
    "data",
    "datapacks",
    "players"
  ]
}
```

### `dimension`

以维度的命名空间 ID（`<命名空间>:<ID>`）为键。每个维度的规则支持两种选择方式：


- `range` — 矩形区域数组。每一项是 `{ "from": [x, z], "to": [x, z] }`，坐标落在矩形内（含边界）的所有区域文件都会被备份。

```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    }
  }
}
```

- `simple` — 单个区域坐标数组，例如 `[x, z]`，用于精确指定个别区域。

```json
{
  "dimension": {
    "minecraft:overworld": {
      "simple": [
        [3, 3]
      ]
    }
  }
}
```

混用 `range` 和 `simple` 规则。
```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ],
      "simple": [
        [3, 3]
      ]
    }
  }
}
```


区域坐标与 Minecraft 的区域文件命名一致（`r.<x>.<z>.mca`，一个区域覆盖 512×512 方块）。

### `file`

世界根目录下需要包含的文件或文件夹列表。文件直接加入备份；文件夹会递归遍历其中的全部文件。

## 注意

- 输出文件会覆盖掉已存在的同名文件

- 使用旧版世界模式时，请手动修改配置文件里的 `file` 字段，删除 `data`，并将 `players` 改为 `playerdata`