# mc-saver

一个用 Go 编写的 Minecraft 存档备份工具。它根据 JSON 规则文件，将世界存档中*选定的部分*——区块区域文件和核心数据文件——打包成带日期的 ZIP 压缩包。

English version: [README.md](README.md)

## 功能

- **选择性备份** — 只备份你关心的区块，而不是整个世界，压缩包更小、速度更快。
- **按维度配置** — 可以分别配置主世界（`overworld`）、下界（`the_nether`）、末地（`the_end`）以及任意自定义维度。
- **灵活的区域选择** — 既可以用矩形 `range` 规则，也可以用 `simple` 规则逐个指定区域坐标。
- **核心数据支持** — 可包含世界根目录下的文件和文件夹，如 `level.dat`、`data`、`datapacks`、`players`。
- **自动命名压缩包** — 输出文件自动命名为 `<存档名>-YYYY-MM-DD.zip`。
- **跨平台** — 一个脚本即可交叉编译 Linux、Windows、macOS（amd64/arm64）。
- **命令行简单** — 只需要一个配置文件、一个存档目录，然后运行即可。

## 环境要求

- Go 1.26.5 及以上版本，用于从源码构建（也可以直接使用 Release 发布里的程序）。
- 一个 Minecraft Java 版世界存档目录。程序会校验目录是否存在且包含 `level.dat`。

## 编译

为当前平台编译：

```bash
go build -o mc-saver .
```

或者使用项目自带的脚本交叉编译所有支持的平台（产物输出到 `build/`）：

```bash
./build.sh
```

## 使用方法

```
# 语法说明：-c 指定配置文件，-t 指定存档目录，-o 指定输出目录
mc-saver [-c <配置文件>] [-t <存档目录>] [-o <输出目录>] [old]

# gencfg：生成默认配置文件后退出（将 -c 放在 gencfg 之前可指定输出路径）
mc-saver [-c <配置文件>] gencfg
```

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-c` | `save-rule.json` | 备份规则文件（JSON）的路径。 |
| `-t` | `world` | 要备份的世界存档目录。 |
| `-o` | `.` | ZIP 压缩包的输出目录。 |

示例：

```bash
# 使用当前目录下的 save-rule.json 备份 ./world
./mc-saver

# 指定要备份的存档
./mc-saver -t /path/to/world

# 使用自定义规则文件
./mc-saver -c rules.json -t /path/to/world

# 指定输出目录
./mc-saver -t /path/to/world -o /path/to/backups

# 生成默认的 save-rule.json 后退出
./mc-saver gencfg

# 按旧版存档布局备份（old 模式，见下文说明）
./mc-saver -t /path/to/old_world old
```

备份完成后会在输出目录（默认当前目录）生成以存档名和日期命名的 ZIP 文件，例如 `world-2026-08-13.zip`。

## old 模式（旧版存档布局）

旧版 Minecraft 使用单文件夹布局：主世界内容直接位于存档根目录，下界和末地分别是 `DIM-1` 和 `DIM1` 文件夹，而不是现代的 `dimensions/<命名空间>/<ID>/`。在命令末尾加一个 `old` 参数即可按旧版布局备份：

```bash
./mc-saver -c rules.json -t /path/to/old_world -o /path/to/backups old
```

`old` 是位置参数，必须放在所有选项（`-c`、`-t`、`-o`）之后，否则后面的选项不会被解析。

旧版布局中，主世界的数据目录就是存档根目录下的 `data/`，old 模式的维度备份已经包含它，所以 `file` 列表里不要重复写 `"data"`，否则同一批文件会重复进包。使用 `gencfg` 生成的默认配置运行 old 模式前，请先从 `file` 列表中删除 `"data"`。

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
- `simple` — 单个区域坐标数组，例如 `[x, z]`，用于精确指定个别区域。

区域坐标与 Minecraft 的区域文件命名一致（`r.<x>.<z>.mca`，一个区域覆盖 512×512 方块）。如果某个维度同时配置了 `range` 和 `simple`，两种规则都会生效。

对于选中的每个区域，程序会从 `dimensions/<命名空间>/<ID>/` 下的 `region/`、`entities/`、`poi/` 目录中收集对应的 `r.<x>.<z>.mca` 文件，同时也会包含该维度 `data/` 目录下的全部文件。

配置中出现的每个维度，其 `dimensions/<命名空间>/<ID>/` 目录必须真实存在，否则备份会直接失败。下界、末地等维度的目录通常要等玩家第一次进入才会生成，所以新存档或从未去过的维度需要从配置中移除对应规则。

### `file`

世界根目录下需要包含的文件或文件夹列表。文件直接加入备份；文件夹会递归遍历其中的全部文件。

缺失的区域文件会输出警告并跳过，不会中断备份；但 `file` 列表的规则更严格：列表中的文件或目录不存在时，备份会直接失败。请确保列表里的路径真实存在（例如旧版本或新生成的世界可能没有 `datapacks`、`players` 目录）。

`file` 列表只应填写维度规则未覆盖的根级条目。不要填写 `dimensions`——它会绕过 `range`/`simple` 的区域选择，把整个维度的全部区域文件都打进压缩包，同时与维度规则产生重复条目。

## 项目结构

```
.
├── main.go          # 程序入口：命令行参数、存档校验、ZIP 打包
├── parse/           # 配置解析与文件收集（通过回调直接写入 ZIP）
├── record/          # 彩色控制台日志（INFO/WARN/ERROR）
├── build.sh         # 交叉编译脚本（Linux/Windows/macOS，amd64/arm64）
├── save-rule.json   # 示例备份规则文件
└── build/           # build.sh 编译产物的输出目录
```

## 补充说明

- 压缩包内保留了存档原有的目录结构（包括顶层存档文件夹名），可以直接放回存档目录使用。
- 默认配置会在三个原版维度中各备份 `(-1, -1)` 到 `(1, 1)` 的 3×3 区域，请根据你实际游玩的区域调整。
