# 运费明细表 (Fee Table)

一款简洁的 Android 运费记录管理应用，用于记录和导出运输费用明细表。

## 功能

- **表格管理**：按年月创建运费表，支持编辑和删除
- **记录管理**：添加、编辑、删除运费记录，包含日期、出发地、目的地、数量、单价、金额
- **地点管理**：下拉选择地点，按使用频率排序，支持快速添加新地点
- **标签功能**：为记录添加标签，支持按标签汇总导出
- **智能计算**：有单价时自动计算金额，无单价时手动输入固定费用
- **多格式导出**：
  - PNG 图片（可保存到相册或分享）
  - PDF 文档
  - Excel 表格

## 技术栈

- **语言**：Kotlin
- **UI 框架**：Jetpack Compose + Material 3
- **数据库**：Room
- **导航**：Navigation Compose
- **导出**：Android Canvas (PNG/PDF) + Apache POI (Excel)
- **架构**：MVVM (ViewModel + StateFlow)

## 构建

```bash
# Debug 构建
./gradlew assembleDebug

# Release 构建
./gradlew assembleRelease
```

## 最低要求

- Android SDK 26 (Android 8.0)
- Target SDK 34 (Android 14)

## 项目结构

```
app/src/main/java/com/example/feetable/
├── data/
│   ├── entity/          # Room 实体 (FeeTable, FeeRecord, Location, Tag)
│   ├── dao/             # 数据访问对象
│   ├── repository/      # 数据仓库
│   └── AppDatabase.kt   # Room 数据库定义
├── ui/
│   ├── home/            # 首页（表格列表）
│   ├── editor/          # 表格编辑页
│   ├── export/          # 导出预览页
│   ├── components/      # 通用组件（LocationDropdown, TagDropdown）
│   └── theme/           # 主题定义
├── export/              # 导出引擎 (PNG, PDF, Excel)
├── FeeTableApp.kt       # Application 类 & 导航图
└── MainActivity.kt      # 入口 Activity
```
