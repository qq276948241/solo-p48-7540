# 地牢探索游戏 - 架构文档

## 项目概览

这是一个用 Go + termbox-go 实现的终端 Roguelike 地牢探索小游戏。玩家操控 `@` 在 40×20 的随机地图上探索，收集金币、装备、击败怪物、逐层深入。

---

## 一、文件职责总览

```
project48/
├── types.go        # 所有类型定义 + 常量 + 工具函数（所有其他文件的基础）
├── equipment.go    # 装备系统 + 背包管理（Inventory 方法 + 随机掉落）
├── mapgen.go       # 地图生成（房间放置 + 走廊连接 + 物品/怪物摆放）
├── combat.go       # 战斗逻辑（怪物查找 + 伤害计算 + 胜负判定）
├── game.go         # 游戏核心逻辑（移动 + 拾取 + 状态更新 + 消息系统）
├── render.go       # UI 渲染（termbox 绘制 + 背包栏 + 状态栏）
└── main.go         # 程序入口（termbox 初始化 + 事件主循环）
```

### 各文件详细说明

| 文件 | 依赖 | 对外提供 | 不包含 |
|---|---|---|---|
| **[types.go](types.go)** | 无（除标准库 `time`） | `Tile`/`ItemType`/`EquipType` 枚举<br>`Equipment`/`Monster`/`Player`/`Room`/`Game` 结构体<br>`MapW`/`MapH`/`InvMax` 常量<br>`maxInt()` 工具函数 | 任何业务逻辑 |
| **[equipment.go](equipment.go)** | types.go | `Inventory.TotalAtk()`/`TotalDef()`<br>`Player.ATK()`/`DEF()`/`AddEquip()`<br>`randomEquipment(level)` | 不直接读写 Game 状态 |
| **[mapgen.go](mapgen.go)** | types.go, equipment.go | `Game.GenerateLevel()`<br>`placeRooms()`/`connectRooms()`<br>`findStairsPos()` | 不处理玩家输入 |
| **[combat.go](combat.go)** | types.go, equipment.go | `Game.combat(monster)`<br>`monsterAt(x,y)`/`allMonstersDead()` | 不处理渲染 |
| **[game.go](game.go)** | types.go, equipment.go, combat.go, mapgen.go | `NewGame()`<br>`Game.Move(dx,dy)`<br>`Game.SetMsg()` | 不包含 termbox 渲染代码 |
| **[render.go](render.go)** | types.go, equipment.go, game.go | `Game.Render()`<br>`drawText()`<br>`clearScreen()` | 不修改 Game 状态（只读） |
| **[main.go](main.go)** | 所有文件 | `main()` 入口函数 | 不包含业务逻辑 |

---

## 二、核心结构体关系

```
          Game (全局游戏状态)
          ├── Level: int            (当前层数)
          ├── Grid: [][]Tile        (40x20 地图, TWall/TFloor)
          ├── Items: map[[2]int]ItemType
          │         (坐标 → 物品类型: IGold/IPotion/IChest/IStairs)
          ├── Monsters: []*Monster  (所有怪物数组)
          ├── Player: Player        (玩家角色)
          ├── Over: bool            (游戏结束标记)
          ├── Message: string       (底部消息栏)
          └── MsgTime: time.Time    (消息过期时间)

  Player
  ├── X, Y: int       (坐标)
  ├── HP, MaxHP: int  (当前/最大生命)
  ├── BaseATK: int    (基础攻击力)
  ├── BaseDEF: int    (基础防御力)
  ├── Gold: int       (金币数 = 分数)
  └── Inv: Inventory  (装备背包, 最多3件)
         │
         └── []*Equipment
             ├── Name: string       ("长剑", "锁子甲"...)
             ├── Type: EquipType    (EWeapon/EArmor)
             ├── Atk: int           (攻击力加成)
             └── Def: int           (防御力加成)

  Monster
  ├── X, Y: int
  ├── HP, MaxHP: int
  ├── ATK: int
  └── DEF: int
```

**关键设计点：**
- `Player.ATK()` 和 `Player.DEF()` 是**方法**，不是字段——自动汇总 `BaseATK/BaseDEF` + 背包所有装备加成，调用方无需关心装备是否存在
- `Inventory` 是切片别名，用 `append(Inv[1:], new)` 实现 FIFO 替换策略（满 3 件踢掉最早的）
- `Game.Monsters` 存指针，战斗时直接修改 HP，死亡后保留在数组中但 `HP <= 0` 被忽略

---

## 三、完整数据流：从按键到画面刷新

这是游戏最核心的执行路径，每个环节对应一个文件的一个函数。

### 第 1 步：用户按键 → 主循环捕获（main.go）

```
  用户按 'W'
     │
     ▼
  termbox.PollEvent() 阻塞等待
     │
     ▼
  ev.Ch == 'w' → 调用 game.Move(0, -1)
     │
     ▼
  调用 game.Render() 重绘整个屏幕
```

### 第 2 步：移动逻辑（game.go → Move）

`Game.Move(dx, dy)` 按以下顺序处理：

```
  检查游戏是否已结束? ──是──> 直接返回
     │否
     ▼
  计算目标坐标 nx, ny = (x+dx, y+dy)
     │
     ▼
  边界检查 + 墙壁检查 ──是墙壁──> 返回
     │否
     ▼
  目标格有怪物? ──是──> 调用 combat() 战斗（见第3步）
     │否
     ▼
  更新玩家坐标: Player.X/Y = nx/ny
     │
     ▼
  目标格有物品?
     ├─ IGold   → BaseATK++, Gold++, SetMsg("拾取金币!")
     ├─ IPotion → HP += 5（不超上限）, SetMsg("恢复 X HP")
     ├─ IChest  → randomEquipment() 出装备 → AddEquip() 入背包
     │            满了? 是→踢掉第0件, SetMsg("背包已满! 丢弃[旧]装入[新]!")
     │            否→直接装入, SetMsg("★宝箱开出...!")
     └─ IStairs → Level++, GenerateLevel() 生成新地图（见第4步）
```

### 第 3 步：战斗逻辑（combat.go → combat）

```
  玩家伤害 pDmg = max(1, Player.ATK() - Monster.DEF)
  怪物伤害 mDmg = max(1, Monster.ATK - Player.DEF())
     │
     ▼
  循环扣血直到一方 HP <= 0:
    怪物 HP -= pDmg ──死──> 检查是否全层清怪 ──是──> 刷出楼梯
       │活
       ▼
    玩家 HP -= mDmg ──死──> Over = true, SetMsg("你被怪物杀死了...")
```

### 第 4 步：地图生成（mapgen.go → GenerateLevel）

```
  先备份 Player 的 HP/MaxHP/BaseATK/BaseDEF/Gold/Inv 到局部变量
     │
     ▼
  Grid 初始化为全墙
     │
     ▼
  placeRooms() 随机放最多 6 个不重叠矩形房间
     │
     ▼
  connectRooms() 用 L 形走廊连通相邻房间
     │
     ▼
  设置出生点: 第一个房间中心
     │
     ▼
  恢复 Player 的所有备份字段（**确保装备不丢**）
     │
     ▼
  随机放置: 金币(3~5个) / 血瓶(1~2个) / 宝箱(1~2个) / 怪物(2+Level~4+Level个)
```

### 第 5 步：渲染（render.go → Render）

每次事件处理完毕后无条件调用一次 `Render()`：

```
  termbox.Clear() 清屏
     │
     ▼
  顶栏状态栏: " 第N层 | HP:x/20 | ATK:x | DEF:x | 金币:x "（蓝底白字）
     │
     ▼
  地图区 (40×20):
    遍历每个格子:
      TWall → '#' (白)
      TFloor → '.' (暗灰)
      Items 覆盖: '$' '+' 'C' '>'
      Monsters 覆盖: 'E' (红)
      Player 覆盖: '@' (黄)
     │
     ▼
  右侧背包栏 (x=42 开始):
    "══════════════"
    " 背  包 "
    "══════════════"
    "[1] 长剑 攻+2"   (武器黄色)
    "[2] 锁子甲 防+1" (护甲绿色)
    "[3] (空)"       (灰色)
     │
     ▼
  底部消息栏: 3秒自动过期, 游戏结束时固定显示得分
     │
     ▼
  操作提示栏: "WASD移动 | $金币 | +血瓶 | C宝箱 | E怪物 | >楼梯"
     │
     ▼
  termbox.Flush() 输出到终端
```

---

## 四、主循环流程图（ASCII）

```
┌───────────────────────────────────────────────────────────┐
│                    程序启动 (main.go)                     │
│  rand.Seed() ──> termbox.Init() ──> NewGame()             │
└─────────────────────────────┬─────────────────────────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │  首次 Render()    │ 画出初始地图
                    └─────────┬─────────┘
                              │
┌─────────────────────────────▼─────────────────────────────┐
│                    主循环 (main loop)                     │
│                                                           │
│   ┌───────────────────────────────────────────────────┐   │
│   │           termbox.PollEvent() 阻塞等待            │   │
│   └─────────────────────┬─────────────────────────────┘   │
│                         │                                 │
│           ┌─────────────┴─────────────┐                   │
│           │  是什么事件?               │                   │
│           ├─ Esc/Error ──> 退出循环 ───┼──> break mainloop│
│           ├─ 'q' + 游戏结束 ──> 退出 ──┘                   │
│           │                                               │
│           └─ 'w'/'a'/'s'/'d' ──> game.Move(dx, dy)       │
│                                  │                        │
│                                  ▼                        │
│                        ┌───────────────────┐              │
│                        │  游戏逻辑更新     │              │
│                        │  (game.go)        │              │
│                        │  - 移动 / 战斗    │              │
│                        │  - 拾取 / 下楼    │              │
│                        │  - 状态变更       │              │
│                        └─────────┬─────────┘              │
│                                  │                        │
│                                  ▼                        │
│                        ┌───────────────────┐              │
│                        │   game.Render()   │              │
│                        │   (render.go)     │              │
│                        │  重绘整个屏幕     │              │
│                        └─────────┬─────────┘              │
│                                  │                        │
│                                  └───────────┐            │
│                                              │            │
└──────────────────────────────────────────────┘            │
                                                           │
                                                           ▼
                                               termbox.Close()
                                               程序退出
```

**数据流单向性原则：**
- `main.go` → 调用 `game.Move()` / `game.Render()` → 不直接修改 Game 字段
- `game.go` → 修改 Game 状态 → 不调用 termbox
- `render.go` → 只读 Game 状态 → 不做任何修改
- `mapgen.go` / `combat.go` / `equipment.go` → 纯逻辑层

这个设计确保你可以安全地修改任何一个文件而不会意外破坏其他模块。

---

## 五、常见修改指引

### 想加新物品？
1. 在 [types.go](types.go#L19-L36) 的 `ItemType` 加新常量
2. 在 [mapgen.go](mapgen.go) 的 `GenerateLevel()` 里增加随机放置
3. 在 [game.go](game.go#L50-L86) 的 `Move()` switch 里加拾取逻辑
4. 在 [render.go](render.go#L40-L48) 的 `Render()` 里加地图字符

### 想改装备数值？
直接改 [equipment.go](equipment.go#L47-L62) 的 `randomEquipment()` 里的 `bonus` 计算公式。

### 想改战斗公式？
改 [combat.go](combat.go#L18-L41) 的 `combat()` 里的 `pDmg` / `mDmg` 计算。

### 想改背包容量？
改 [types.go](types.go#L5-L10) 的 `InvMax` 常量，同时改 [render.go](render.go#L83-L111) 里的背包栏行数。

---

## 六、编译与运行

```bash
cd project48
go build -o dungeon.exe .
./dungeon.exe
```

**操作说明：**
- `W/A/S/D` - 上下左右移动
- `ESC` - 随时退出
- `Q` - 游戏结束后退出
