# ConnectionManager

基于命令行的连接管理工具。

## 功能特性

- 顶部模块栏：SSH、MySQL、PostgreSQL、Redis
- 主窗体：动态内容区域
- 状态栏：显示当前状态

## 操作说明

- `H/h` 或 `←`：切换到上一个模块
- `L/l` 或 `→`：切换到下一个模块  
- `ESC`：退出程序

## 运行程序

```bash
./connectionmanager
```

## 界面说明

- **Normal状态**：可使用HJKL或方向键进行导航
- **Edit状态**：进入编辑模式（暂未实现）
- 模块栏会根据终端宽度自动调整显示，超出时显示箭头指示
连接管理器，致力于管理和快速创建基于命令行工具的多种连接，目标包括但不限于：SSH、MySQL、PostgreSQL、Redis等等
