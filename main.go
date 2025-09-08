package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

// 应用程序状态枚举
type AppState int

const (
	Normal AppState = iota // 正常状态，可以进行导航操作
	Edit                   // 编辑状态，用于编辑连接信息等
)

// 应用程序主结构体，包含所有UI组件和状态信息
type App struct {
	// tview相关组件
	app         *tview.Application // 主应用程序实例
	grid        *tview.Grid        // 主Grid布局容器
	moduleBar   *tview.TextView    // 顶部模块栏，显示模块选择
	mainPanel   *tview.TextView    // 中间主面板，显示主要内容
	statusBar   *tview.TextView    // 底部状态栏，显示当前状态信息
	confirmBox  *tview.TextView    // 确认退出的文本框
	confirmGrid *tview.Grid        // 确认对话框的网格布局

	// 应用程序状态
	state          AppState // 当前应用状态（Normal或Edit）
	modules        []string // 可用的模块列表
	currentModule  int      // 当前选中的模块索引
	showingConfirm bool     // 是否正在显示确认对话框
}

// 创建新的应用程序实例，初始化所有默认值
func NewApp() *App {
	return &App{
		app:            tview.NewApplication(),                          // 创建tview应用实例
		state:          Normal,                                          // 初始状态为Normal
		modules:        []string{"SSH", "MySQL", "PostgreSQL", "Redis"}, // 定义可用模块列表
		currentModule:  0,                                               // 默认选中第一个模块（SSH）
		showingConfirm: false,                                           // 初始不显示确认对话框
	}
}

// 初始化用户界面，设置所有UI组件和布局
func (a *App) initUI() {
	// 设置全局边框样式为双线，创建统一的视觉效果
	tview.Borders.Horizontal = '═'  // 水平边框字符
	tview.Borders.Vertical = '║'    // 垂直边框字符
	tview.Borders.TopLeft = '╔'     // 左上角边框字符
	tview.Borders.TopRight = '╗'    // 右上角边框字符
	tview.Borders.BottomLeft = '╚'  // 左下角边框字符
	tview.Borders.BottomRight = '╝' // 右下角边框字符
	tview.Borders.BottomT = '╩'     // 底部T形连接
	tview.Borders.LeftT = '╠'       // 左侧T形连接
	tview.Borders.RightT = '╣'      // 右侧T形连接
	tview.Borders.TopT = '╦'        // 顶部T形连接
	tview.Borders.Cross = '╬'       // 十字交叉连接

	// 创建顶部模块栏 - 水平显示可用模块
	a.moduleBar = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetScrollable(false)
	a.moduleBar.SetBorder(true).SetTitle("模块选择").SetTitleAlign(tview.AlignLeft)

	// 创建中间主面板 - 显示主要内容
	a.mainPanel = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetScrollable(true)
	a.mainPanel.SetBorder(true).SetTitle("主要内容").SetTitleAlign(tview.AlignLeft)

	// 创建底部状态栏组件，用于显示应用程序状态信息
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("准备就绪...")
	a.statusBar.SetBorder(true).SetTitle("状态").SetTitleAlign(tview.AlignLeft)

	// 创建确认退出对话框的Grid布局 - 居中显示小框
	a.confirmGrid = tview.NewGrid().
		SetRows(0, 7, 0).     // 上下留空，中间7行给确认框
		SetColumns(0, 40, 0). // 左右留空，中间40列给确认框
		SetBorders(false)

	// 创建确认退出对话框组件 - 小巧的居中框
	a.confirmBox = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(false)
	a.confirmBox.SetBorder(true).
		SetTitle("确认退出").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorYellow)

	// 将确认框添加到Grid中央
	a.confirmGrid.AddItem(a.confirmBox, 1, 1, 1, 1, 0, 0, true)

	// 使用Grid布局创建垂直三行布局
	a.grid = tview.NewGrid().
		SetRows(3, 0, 3). // 3行：模块栏(3行含边框), 主面板(占据剩余空间), 状态栏(3行含边框)
		SetColumns(0).    // 1列：占据全部宽度
		SetBorders(false) // 关闭Grid边框，使用各组件自己的边框

	// 设置Grid的标题和对齐方式
	a.grid.SetTitle("ConnectionManager")
	a.grid.SetTitleAlign(tview.AlignCenter)

	// 添加组件到Grid（垂直排列）
	a.grid.AddItem(a.moduleBar, 0, 0, 1, 1, 0, 0, true). // 模块栏：第0行，可聚焦
								AddItem(a.mainPanel, 1, 0, 1, 1, 0, 0, false). // 主面板：第1行
								AddItem(a.statusBar, 2, 0, 1, 1, 0, 0, false)  // 状态栏：第2行

	// 初始化更新界面内容
	a.updateModuleBar()
	a.updateMainPanel()
	a.updateStatusBar()

	// 设置初始焦点
	a.setInitialFocus()

	// 设置全局键盘事件处理器，捕获用户的键盘输入
	a.app.SetInputCapture(a.handleKeyEvent)

	// 设置根界面组件并启用全屏模式
	a.app.SetRoot(a.grid, true)
}

// 更新整个界面内容
func (a *App) updateUI() {
	a.updateModuleBar() // 更新模块栏
	a.updateMainPanel() // 更新主面板
	a.updateStatusBar() // 更新状态栏
}

// 设置初始焦点
func (a *App) setInitialFocus() {
	a.moduleBar.SetBorderColor(tcell.ColorYellow)
	a.mainPanel.SetBorderColor(tcell.ColorWhite)
	a.app.SetFocus(a.moduleBar)
}

// 更新模块栏显示（顶部水平模块选择栏）
func (a *App) updateModuleBar() {
	content := "  " // 左侧间距

	for i, module := range a.modules {
		if i > 0 {
			content += "  " // 模块间距
		}

		if i == a.currentModule {
			// 选中状态：高亮显示
			content += fmt.Sprintf("[white:blue:b] %s [-:-:-]", module)
		} else {
			// 未选中状态：正常显示
			content += fmt.Sprintf(" %s ", module)
		}
	}

	a.moduleBar.SetText(content)
}

// 更新主面板显示（中间主要内容）
func (a *App) updateMainPanel() {
	currentModule := a.modules[a.currentModule]
	content := fmt.Sprintf("[yellow]%s 连接管理[-]\n\n", currentModule)

	switch currentModule {
	case "SSH":
		content += "SSH 连接配置:\n\n"
		content += "主机: example.com\n"
		content += "端口: 22\n"
		content += "用户: user\n"
		content += "认证: 密钥认证\n\n"
		content += "[green]连接状态: 就绪[-]\n\n"
		content += "可用连接:\n"
		content += "• SSH-Server-01 (192.168.1.10)\n"
		content += "• SSH-Server-02 (192.168.1.11)\n"
		content += "• Production-Server (prod.example.com)\n"
	case "MySQL":
		content += "MySQL 数据库配置:\n\n"
		content += "主机: localhost\n"
		content += "端口: 3306\n"
		content += "数据库: myapp\n"
		content += "用户: root\n\n"
		content += "[red]连接状态: 未连接[-]\n\n"
		content += "可用数据库:\n"
		content += "• MySQL-DB-01 (localhost:3306)\n"
		content += "• MySQL-DB-02 (db.example.com:3306)\n"
	case "PostgreSQL":
		content += "PostgreSQL 数据库配置:\n\n"
		content += "主机: localhost\n"
		content += "端口: 5432\n"
		content += "数据库: postgres\n"
		content += "用户: postgres\n\n"
		content += "[yellow]连接状态: 连接中...[-]\n\n"
		content += "可用数据库:\n"
		content += "• PostgreSQL-Main (localhost:5432)\n"
		content += "• PostgreSQL-Analytics (analytics.example.com:5432)\n"
	case "Redis":
		content += "Redis 缓存配置:\n\n"
		content += "主机: localhost\n"
		content += "端口: 6379\n"
		content += "数据库: 0\n"
		content += "认证: 无\n\n"
		content += "[green]连接状态: 已连接[-]\n\n"
		content += "可用缓存实例:\n"
		content += "• Redis-Cache-01 (localhost:6379)\n"
		content += "• Redis-Session (session.example.com:6379)\n"
	}

	a.mainPanel.SetText(content)
}

// 更新确认对话框显示
func (a *App) updateConfirmBox() {
	content := "\n[yellow]确定要退出程序吗？[-]\n\n"
	content += "[green]Yes (Y)[-]    [red]No (N)[-]\n"

	a.confirmBox.SetText(content)
}
func (a *App) updateStatusBar() {
	stateText := ""
	switch a.state {
	case Normal:
		stateText = "Normal"
	case Edit:
		stateText = "Edit"
	}

	statusText := fmt.Sprintf("[yellow]状态: %s[-] | [blue]当前模块: %s[-] | [gray]Q: 退出, ←→/H/L: 切换模块[-]",
		stateText, a.modules[a.currentModule])

	a.statusBar.SetText(statusText)
}

// 处理键盘事件
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	// 如果正在显示确认对话框，只处理Y/N键
	if a.showingConfirm {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'y', 'Y':
				a.app.Stop() // 选择Yes，退出程序
				return nil
			case 'n', 'N':
				a.hideExitConfirmation() // 选择No，返回主界面
				return nil
			}
		}
		return event
	}

	// 正常模式下的按键处理
	if a.state != Normal {
		return event
	}

	switch event.Key() {
	case tcell.KeyLeft:
		a.moveToPreviousModule()
		return nil
	case tcell.KeyRight:
		a.moveToNextModule()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'h', 'H':
			a.moveToPreviousModule()
			return nil
		case 'l', 'L':
			a.moveToNextModule()
			return nil
		case 'q', 'Q':
			a.showExitConfirmation()
			return nil
		}
	}

	return event
}

// 显示退出确认对话框
func (a *App) showExitConfirmation() {
	a.showingConfirm = true
	a.updateConfirmBox()
	a.app.SetRoot(a.confirmGrid, true)
}

// 隐藏退出确认对话框
func (a *App) hideExitConfirmation() {
	a.showingConfirm = false
	a.app.SetRoot(a.grid, true)
}

// 移动到上一个模块
func (a *App) moveToPreviousModule() {
	if a.currentModule > 0 {
		a.currentModule--
		a.updateUI()
	}
}

// 移动到下一个模块
func (a *App) moveToNextModule() {
	if a.currentModule < len(a.modules)-1 {
		a.currentModule++
		a.updateUI()
	}
}

// 运行应用程序
func (a *App) Run() error {
	return a.app.Run()
}

// 主函数
func main() {
	// 初始化配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.connectionmanager")
	viper.AutomaticEnv()

	// 读取配置文件（如果存在）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("读取配置文件错误: %v\n", err)
			os.Exit(1)
		}
	}

	// 创建应用程序
	app := NewApp()

	// 初始化界面
	app.initUI()

	// 运行应用程序
	if err := app.Run(); err != nil {
		fmt.Printf("运行应用程序错误: %v\n", err)
		os.Exit(1)
	}
}
