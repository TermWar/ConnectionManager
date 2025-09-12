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
	hoveredModule  int      // 当前悬停的模块索引（键盘导航）
	showingConfirm bool     // 是否正在显示确认对话框

	// 树状结构导航状态
	inTreeView      bool            // 是否进入了树状视图导航模式
	selectedProject int             // 当前选中的项目索引
	selectedEnv     int             // 当前选中的环境索引
	selectedConn    int             // 当前选中的连接索引
	treeLevel       int             // 当前所在的树级别 (0=项目, 1=环境, 2=连接)
	expandedNodes   map[string]bool // 展开状态记录
}

// 创建新的应用程序实例，初始化所有默认值
func NewApp() *App {
	return &App{
		app:            tview.NewApplication(),                          // 创建tview应用实例
		state:          Normal,                                          // 初始状态为Normal
		modules:        []string{"SSH", "MySQL", "PostgreSQL", "Redis"}, // 定义可用模块列表
		currentModule:  0,                                               // 默认选中第一个模块（SSH）
		hoveredModule:  0,                                               // 默认悬停模块与选中模块一致
		showingConfirm: false,                                           // 初始不显示确认对话框

		// 树状结构导航初始状态
		inTreeView:      false,                 // 初始不在树状视图中
		selectedProject: 0,                     // 默认选中第一个项目
		selectedEnv:     0,                     // 默认选中第一个环境
		selectedConn:    0,                     // 默认选中第一个连接
		treeLevel:       0,                     // 初始在项目级别
		expandedNodes:   make(map[string]bool), // 初始化展开状态映射
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
			// 已选中状态：蓝色背景 + 方括号
			content += fmt.Sprintf("[white:blue:b][ %s ][-:-:-]", module)
		} else if i == a.hoveredModule && i != a.currentModule {
			// 悬停状态：黄色边框 + 方括号
			content += fmt.Sprintf("[yellow][ %s ][-]", module)
		} else {
			// 普通状态：无边框
			content += fmt.Sprintf(" %s ", module)
		}
	}

	a.moduleBar.SetText(content)
}

// 更新主面板显示（中间主要内容）
func (a *App) updateMainPanel() {
	currentModule := a.modules[a.currentModule]
	// 更新主面板标题为当前选中的模块
	a.mainPanel.SetTitle(fmt.Sprintf("%s 连接管理", currentModule))

	if a.inTreeView {
		content := a.renderTreeView()
		a.mainPanel.SetText(content)
	} else {
		content := a.renderOverview()
		a.mainPanel.SetText(content)
	}
}

// 渲染概览视图（非树状导航模式）
func (a *App) renderOverview() string {
	currentModule := a.modules[a.currentModule]
	content := fmt.Sprintf("[yellow]%s 连接管理概览[-]\n\n", currentModule)
	content += "按 [white:blue]Enter[-] 或 [white:blue]Space[-] 进入树状导航模式\n\n"

	switch currentModule {
	case "SSH":
		content += "📁 可用项目:\n"
		content += "  • Web服务器项目 (3个环境, 9个连接)\n"
		content += "  • 数据库项目 (2个环境, 6个连接)\n"
		content += "  • 开发环境项目 (2个环境, 4个连接)\n\n"
	case "MySQL":
		content += "📁 可用项目:\n"
		content += "  • 生产数据库 (3个环境, 9个实例)\n"
		content += "  • 分析数据库 (2个环境, 6个实例)\n"
		content += "  • 测试数据库 (1个环境, 3个实例)\n\n"
	case "PostgreSQL":
		content += "📁 可用项目:\n"
		content += "  • 主业务数据库 (3个环境, 9个实例)\n"
		content += "  • 报表数据库 (2个环境, 6个实例)\n"
		content += "  • 备份数据库 (1个环境, 3个实例)\n\n"
	case "Redis":
		content += "📁 可用项目:\n"
		content += "  • 缓存集群 (3个环境, 9个实例)\n"
		content += "  • 会话存储 (2个环境, 6个实例)\n"
		content += "  • 消息队列 (2个环境, 4个实例)\n\n"
	}

	content += "[dim]按 Enter 进入树状导航，在树状模式中可以管理具体的连接[-]"
	return content
}

// 渲染树状视图
func (a *App) renderTreeView() string {
	currentModule := a.modules[a.currentModule]
	content := fmt.Sprintf("[yellow]%s 树状导航模式[-]\n\n", currentModule)

	// 获取项目列表
	projects := a.getProjectList()

	for i, project := range projects {
		// 左侧箭头指示器（始终在最左侧）
		arrowIndicator := ""
		if a.treeLevel == 0 && i == a.selectedProject {
			arrowIndicator = "[yellow]►[-] "
		} else {
			arrowIndicator = "  "
		}

		// 项目展开状态
		projectKey := fmt.Sprintf("%s-proj-%d", currentModule, i)
		isProjectExpanded := a.expandedNodes[projectKey]
		expandIcon := "+"
		if isProjectExpanded {
			expandIcon = "-"
		}

		content += fmt.Sprintf("%s\t[%s] %s\n", arrowIndicator, expandIcon, project.Name)

		// 如果项目展开，显示环境
		if isProjectExpanded {
			environments := a.getEnvironmentList(i)
			for j, env := range environments {
				// 左侧箭头指示器（始终在最左侧）
				arrowIndicator := ""
				if a.treeLevel == 1 && i == a.selectedProject && j == a.selectedEnv {
					arrowIndicator = "[yellow]►[-] "
				} else {
					arrowIndicator = "  "
				}

				// 环境展开状态
				envKey := fmt.Sprintf("%s-proj-%d-env-%d", currentModule, i, j)
				isEnvExpanded := a.expandedNodes[envKey]
				envExpandIcon := "+"
				if isEnvExpanded {
					envExpandIcon = "-"
				}

				content += fmt.Sprintf("%s\t\t[%s] %s\n", arrowIndicator, envExpandIcon, env.Name)

				// 如果环境展开，显示连接
				if isEnvExpanded {
					connections := a.getConnectionList(i, j)
					for k, conn := range connections {
						// 左侧箭头指示器（始终在最左侧）
						connArrowIndicator := ""
						if a.treeLevel == 2 && i == a.selectedProject && j == a.selectedEnv && k == a.selectedConn {
							connArrowIndicator = "[yellow]►[-] "
						} else {
							connArrowIndicator = "  "
						}

						statusColor := "green"
						statusText := "已连接"
						switch conn.Status {
						case "connected":
							statusColor = "green"
							statusText = "已连接"
						case "disconnected":
							statusColor = "red"
							statusText = "断开"
						case "connecting":
							statusColor = "yellow"
							statusText = "连接中"
						}

						content += fmt.Sprintf("%s\t\t\t%s ([%s]%s[-])\n", connArrowIndicator, conn.Name, statusColor, statusText)
					}
				}
			}
		}
	}

	// 添加操作提示
	content += "\n[dim]"
	switch a.treeLevel {
	case 0:
		content += "项目级别 - ↑↓/JK: 导航, →/L: 进入环境, Space: 展开/收缩, ESC/Q: 退出"
	case 1:
		content += "环境级别 - ↑↓/JK: 导航, ←/H: 返回项目, →/L: 进入连接, Space: 展开/收缩"
	case 2:
		content += "连接级别 - ↑↓/JK: 导航, ←/H: 返回环境, Enter: 连接/断开"
	}
	content += "[-]"

	return content
}

// 项目数据结构
type Project struct {
	Name string
}

type Environment struct {
	Name string
}

type Connection struct {
	Name   string
	Status string
}

// 获取项目列表
func (a *App) getProjectList() []Project {
	currentModule := a.modules[a.currentModule]
	switch currentModule {
	case "SSH":
		return []Project{
			{Name: "Web服务器项目"},
			{Name: "数据库项目"},
			{Name: "开发环境项目"},
		}
	case "MySQL":
		return []Project{
			{Name: "生产数据库"},
			{Name: "分析数据库"},
			{Name: "测试数据库"},
		}
	case "PostgreSQL":
		return []Project{
			{Name: "主业务数据库"},
			{Name: "报表数据库"},
			{Name: "备份数据库"},
		}
	case "Redis":
		return []Project{
			{Name: "缓存集群"},
			{Name: "会话存储"},
			{Name: "消息队列"},
		}
	}
	return []Project{}
}

// 获取环境列表
func (a *App) getEnvironmentList(projectIndex int) []Environment {
	if projectIndex == 2 { // 第三个项目只有1个环境
		return []Environment{{Name: "开发环境"}}
	}
	return []Environment{
		{Name: "生产环境"},
		{Name: "测试环境"},
	}
}

// 获取连接列表
func (a *App) getConnectionList(projectIndex, envIndex int) []Connection {
	currentModule := a.modules[a.currentModule]
	baseConnections := []Connection{
		{Name: fmt.Sprintf("%s-01", currentModule), Status: "connected"},
		{Name: fmt.Sprintf("%s-02", currentModule), Status: "disconnected"},
		{Name: fmt.Sprintf("%s-03", currentModule), Status: "connecting"},
	}
	return baseConnections
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

	var statusText string
	if a.inTreeView {
		levelNames := []string{"项目", "环境", "连接"}
		currentLevel := levelNames[a.treeLevel]
		statusText = fmt.Sprintf("[yellow]状态: %s[-] | [blue]模块: %s[-] | [green]层级: %s[-] | [gray]↑↓/JK: 导航, ←→/HL: 层级, ESC: 退出[-]",
			stateText, a.modules[a.currentModule], currentLevel)
	} else {
		statusText = fmt.Sprintf("[yellow]状态: %s[-] | [blue]当前模块: %s[-] | [green]悬停: %s[-] | [gray]←→/H/L: 导航, Enter/Space: 选择, Q: 退出[-]",
			stateText, a.modules[a.currentModule], a.modules[a.hoveredModule])
	}

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

	if a.inTreeView {
		// 树状视图中的导航
		return a.handleTreeNavigation(event)
	} else {
		// 模块栏导航
		switch event.Key() {
		case tcell.KeyLeft:
			a.moveToPreviousHover()
			return nil
		case tcell.KeyRight:
			a.moveToNextHover()
			return nil
		case tcell.KeyEnter:
			a.enterTreeView()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h', 'H':
				a.moveToPreviousHover()
				return nil
			case 'l', 'L':
				a.moveToNextHover()
				return nil
			case ' ': // 空格键也可以进入树状视图
				a.enterTreeView()
				return nil
			case 'q', 'Q':
				a.showExitConfirmation()
				return nil
			}
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

// 移动到上一个模块（悬停状态）
func (a *App) moveToPreviousHover() {
	if a.hoveredModule > 0 {
		a.hoveredModule--
		a.updateModuleBar()
	}
}

// 移动到下一个模块（悬停状态）
func (a *App) moveToNextHover() {
	if a.hoveredModule < len(a.modules)-1 {
		a.hoveredModule++
		a.updateModuleBar()
	}
}

// 进入树状视图
func (a *App) enterTreeView() {
	a.currentModule = a.hoveredModule
	a.inTreeView = true
	a.treeLevel = 0
	a.selectedProject = 0
	a.selectedEnv = 0
	a.selectedConn = 0
	a.updateMainPanel()
	a.updateStatusBar()
	a.updateModuleBar()
}

// 退出树状视图
func (a *App) exitTreeView() {
	a.inTreeView = false
	a.updateStatusBar()
}

// 处理树状视图中的键盘导航
func (a *App) handleTreeNavigation(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		a.moveTreeUp()
		return nil
	case tcell.KeyDown:
		a.moveTreeDown()
		return nil
	case tcell.KeyLeft:
		a.collapseOrMoveUp()
		return nil
	case tcell.KeyRight:
		a.expandOrMoveDown()
		return nil
	case tcell.KeyEsc:
		a.exitTreeView()
		return nil
	case tcell.KeyEnter:
		a.activateTreeItem()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'k', 'K':
			a.moveTreeUp()
			return nil
		case 'j', 'J':
			a.moveTreeDown()
			return nil
		case 'h', 'H':
			a.collapseOrMoveUp()
			return nil
		case 'l', 'L':
			a.expandOrMoveDown()
			return nil
		case 'q', 'Q':
			a.exitTreeView()
			return nil
		case ' ':
			a.toggleExpansion()
			return nil
		}
	}
	return event
}

// 在树状视图中向上移动
func (a *App) moveTreeUp() {
	switch a.treeLevel {
	case 0: // 项目级别
		if a.selectedProject > 0 {
			a.selectedProject--
			a.updateMainPanel()
		}
	case 1: // 环境级别
		if a.selectedEnv > 0 {
			a.selectedEnv--
		} else {
			a.treeLevel = 0
		}
		a.updateMainPanel()
	case 2: // 连接级别
		if a.selectedConn > 0 {
			a.selectedConn--
		} else {
			a.treeLevel = 1
		}
		a.updateMainPanel()
	}
}

// 在树状视图中向下移动
func (a *App) moveTreeDown() {
	switch a.treeLevel {
	case 0: // 项目级别
		maxProjects := a.getProjectCount() - 1
		if a.selectedProject < maxProjects {
			a.selectedProject++
			a.updateMainPanel()
		}
	case 1: // 环境级别
		maxEnvs := a.getEnvironmentCount() - 1
		if a.selectedEnv < maxEnvs {
			a.selectedEnv++
		} else if a.hasConnections() {
			a.treeLevel = 2
			a.selectedConn = 0
		}
		a.updateMainPanel()
	case 2: // 连接级别
		maxConns := a.getConnectionCount() - 1
		if a.selectedConn < maxConns {
			a.selectedConn++
			a.updateMainPanel()
		}
	}
}

// 收缩节点或向上移动层级
func (a *App) collapseOrMoveUp() {
	switch a.treeLevel {
	case 2: // 从连接回到环境
		a.treeLevel = 1
		// 收缩当前环境
		envKey := fmt.Sprintf("%s-proj-%d-env-%d", a.modules[a.currentModule], a.selectedProject, a.selectedEnv)
		a.expandedNodes[envKey] = false
		a.updateMainPanel()
	case 1: // 从环境回到项目
		a.treeLevel = 0
		// 收缩当前项目
		projectKey := fmt.Sprintf("%s-proj-%d", a.modules[a.currentModule], a.selectedProject)
		a.expandedNodes[projectKey] = false
		a.updateMainPanel()
	case 0: // 从项目退出树状视图
		a.exitTreeView()
	}
}

// 展开节点或向下移动层级
func (a *App) expandOrMoveDown() {
	switch a.treeLevel {
	case 0: // 从项目进入环境
		// 展开当前项目
		projectKey := fmt.Sprintf("%s-proj-%d", a.modules[a.currentModule], a.selectedProject)
		a.expandedNodes[projectKey] = true

		if a.getEnvironmentCount() > 0 {
			a.treeLevel = 1
			a.selectedEnv = 0
			a.updateMainPanel()
		}
	case 1: // 从环境进入连接
		// 展开当前环境
		envKey := fmt.Sprintf("%s-proj-%d-env-%d", a.modules[a.currentModule], a.selectedProject, a.selectedEnv)
		a.expandedNodes[envKey] = true

		if a.hasConnections() {
			a.treeLevel = 2
			a.selectedConn = 0
			a.updateMainPanel()
		}
	}
}

// 切换节点展开状态
func (a *App) toggleExpansion() {
	nodeKey := a.getCurrentNodeKey()
	a.expandedNodes[nodeKey] = !a.expandedNodes[nodeKey]
	a.updateMainPanel()
}

// 激活当前选中的树项目
func (a *App) activateTreeItem() {
	// 这里可以实现连接操作等
	a.updateStatusBar()
}

// 获取当前节点的唯一标识符
func (a *App) getCurrentNodeKey() string {
	return fmt.Sprintf("%s-%d-%d-%d", a.modules[a.currentModule], a.selectedProject, a.selectedEnv, a.selectedConn)
}

// 获取项目数量
func (a *App) getProjectCount() int {
	return len(a.getProjectList())
}

// 获取环境数量
func (a *App) getEnvironmentCount() int {
	return len(a.getEnvironmentList(a.selectedProject))
}

// 获取连接数量
func (a *App) getConnectionCount() int {
	return len(a.getConnectionList(a.selectedProject, a.selectedEnv))
}

// 检查是否有连接
func (a *App) hasConnections() bool {
	return a.getConnectionCount() > 0
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
