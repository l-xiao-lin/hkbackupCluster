package service

import (
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

func CheckListing(website, username, password string) (err error) {
	for i := 0; i < 10; i++ {
		err = autoCheckListing(website, username, password)
		if err == nil {
			break
		}
		logger.SugarLog.Errorf("autoCheckListing返回error,执行第%d次检测,err:%v", i+1, err)
	}
	return
}

func autoCheckListing(website, username, password string) (err error) {
	m.Lock()
	defer m.Unlock()
	opts := []selenium.ServiceOption{
		selenium.Output(nil),
	}

	service, err := selenium.NewChromeDriverService("C:\\Program Files\\Google\\Chrome\\Application\\chromedriver.exe", 9515, opts...)
	if err != nil {
		logger.SugarLog.Errorf("无法启动 WebDriver 服务:%v", err)
		return
	}
	defer service.Stop()

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		logger.SugarLog.Errorf("无法打开会话:%v", err)
		return
	}

	defer webDriver.Quit()

	err = webDriver.Get(website)
	if err != nil {
		logger.SugarLog.Errorf("无法导航到网页:%v", err)
		return
	}

	usernameInput, err := webDriver.FindElement(selenium.ByCSSSelector, "input[name='username']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到用户名输入框:%v", err)
		return
	}
	err = usernameInput.SendKeys(username)
	if err != nil {
		logger.SugarLog.Errorf("无法填充用户名:%v", err)
		return
	}

	passwordInput, err := webDriver.FindElement(selenium.ByID, "password")
	if err != nil {
		logger.SugarLog.Errorf("无法找到密码输入框:%v", err)
		return
	}
	err = passwordInput.SendKeys(password)
	if err != nil {
		logger.SugarLog.Errorf("无法填充密码:%v", err)
		return
	}

	loginButton, err := webDriver.FindElement(selenium.ByXPATH, "//button[contains(text(), '立即登录')]")
	if err != nil {
		logger.SugarLog.Errorf("无法找到登录按钮:%v", err)
		return
	}

	err = loginButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击登录按钮:%v", err)
		return
	}

	// 关闭系统提示框
	var systemCloseButton selenium.WebElement
	for i := 0; i < 5; i++ {
		// 在容器元素上下文中定位关闭按钮
		systemCloseButton, err = webDriver.FindElement(selenium.ByID, "messageWinBtn")
		if err == nil && systemCloseButton != nil {
			err1 := systemCloseButton.Click()
			if err1 != nil {
				logger.SugarLog.Errorf("无法点击系统提醒框关闭按钮:%v", err)
				return err1
			}
			break
		}

		logger.SugarLog.Infof("没有找到系统提醒框，继续下一次查找:%v", err)
		time.Sleep(500 * time.Millisecond)
	}

	logger.SugarLog.Infof("没有系统提示框")

	var closeButton selenium.WebElement
	waitTimeout := 10 * time.Second
	maxWaitTime := time.Now().Add(waitTimeout)
	for time.Now().Before(maxWaitTime) {
		outerDialogContainer, err := webDriver.FindElement(selenium.ByCSSSelector, "div.panel.window[style*='display: block']")
		if err == nil {
			closeButton, err = outerDialogContainer.FindElement(selenium.ByCSSSelector, ".btn-w.msgBtn")
			if err == nil && closeButton != nil {
				err1 := closeButton.Click()
				if err1 != nil {
					logger.SugarLog.Errorf("无法点击外层对话框按钮:%v", err)
					return err1
				}
				break
			}
		}
		logger.SugarLog.Infof("没有找到外层对话框容器元素，继续下一次查找:%v", err)
		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}

	logger.SugarLog.Infof("没有外层公告框")

	// 定位内层对话框的容器元素
	var innerCloseButton selenium.WebElement
	for i := 0; i < 5; i++ {
		innerDialogContainer, err := webDriver.FindElement(selenium.ByCSSSelector, "div.panel.window[style*='display: block']")
		if err == nil {
			// 在容器元素上下文中定位关闭按钮
			innerCloseButton, err = innerDialogContainer.FindElement(selenium.ByCSSSelector, ".btn-w.ml10.msgBtn")
			if err == nil && innerCloseButton != nil {
				err1 := innerCloseButton.Click()
				if err1 != nil {
					logger.SugarLog.Errorf("无法点击内层对话框关闭按钮:%v", err)
					return err1
				}
				break
			}
		}
		logger.SugarLog.Infof("没有找到内层对话框容器元素，继续下一次查找:%v", err)
		time.Sleep(500 * time.Millisecond)
	}

	logger.SugarLog.Infof("没有内层公告框")

	//找到左侧菜果的产品，并点击

	productButton, err := webDriver.FindElement(selenium.ByCSSSelector, "[onclick='toggleLeftMenu(\\'#_product-ul\\')']")
	if err == nil && productButton != nil {
		err = productButton.Click()
		if err == nil {
			// 成功点击 "产品" 按钮

			salesBtn, err := webDriver.FindElement(selenium.ByID, "left_menu_product")
			if err == nil {
				err = salesBtn.Click()
				if err == nil {
					// 成功点击 "售卖产品资料" 按钮
				}
			}
		}
	} else {
		logger.SugarLog.Errorf("not found productButton or exist err:%v", err)
		return errors.New("not found productButton or exist err")
	}

	//查找sku
	var searchProduct selenium.WebElement
	waitTimeoutProduct := 10 * time.Second
	maxWaitTimeProduct := time.Now().Add(waitTimeoutProduct)
	for time.Now().Before(maxWaitTimeProduct) {
		searchProduct, err = webDriver.FindElement(selenium.ByCSSSelector, "input#searchInput")
		if err == nil && searchProduct != nil {
			break // 找到元素，退出循环
		}

		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}

	err = searchProduct.SendKeys("运维测试专用")
	if err != nil {
		logger.SugarLog.Errorf("填充搜索条件报错:%v", err)
		return
	}

	//查找 搜索并点击

	searchButton, err := webDriver.FindElement(selenium.ByCSSSelector, "a.btn.ml10")
	if err != nil && searchButton != nil {
		logger.SugarLog.Errorf("无法找到搜索元素:%v", err)
		return
	}
	err = searchButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("不能点击搜索按钮:%v", err)
		return
	}
	//找到编辑按钮，并点击
	var editButton selenium.WebElement
	editWaitTimeout := 10 * time.Second
	editMaxWaitTime := time.Now().Add(editWaitTimeout)
	for time.Now().Before(editMaxWaitTime) {
		editButton, err = webDriver.FindElement(selenium.ByCSSSelector, ".iconfont.icon-bianji")
		if err == nil {
			href, err := editButton.GetAttribute("href")
			if err == nil {
				logger.SugarLog.Infof("url地址为:%v", href)
				break
			}
		}

		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}

	// 获取当前窗口句柄
	currentWindowHandle, err := webDriver.CurrentWindowHandle()
	if err != nil {
		logger.SugarLog.Errorf("无法获取当前窗口句柄:%v", err)
		return
	}

	if editButton != nil {
		err = editButton.Click()
		if err != nil {
			logger.SugarLog.Errorf("不能点击edit按钮:%v", err)
			return
		}
	} else {
		logger.SugarLog.Errorf("没有找到运维测试专用产品,err")
		return fmt.Errorf("没有找到运维测试专用产品")
	}

	// 获取所有窗口句柄
	windowHandles, err := webDriver.WindowHandles()
	if err != nil {
		logger.SugarLog.Errorf("无法获取窗口句柄列表:%v", err)
		return
	}

	fmt.Println("currentWindowHandle:", currentWindowHandle)
	fmt.Println("windowHandles: ", windowHandles)

	// 切换到新打开的窗口
	for _, handle := range windowHandles {
		if handle != currentWindowHandle {
			err = webDriver.SwitchWindow(handle)
			if err != nil {
				logger.SugarLog.Errorf("无法切换到新窗口:%v", err)
				return
			}
			break
		}
	}

	//关闭编辑页面中右上角的对话框

	var closeButtonFlag selenium.WebElement
	waitTimeoutFlag := 10 * time.Second
	maxWaitTimeFlag := time.Now().Add(waitTimeoutFlag)
	for time.Now().Before(maxWaitTimeFlag) {
		closeButtonFlag, err = webDriver.FindElement(selenium.ByCSSSelector, ".btn-w.ml5.msgBtn")
		if err == nil && closeButtonFlag != nil {
			logger.SugarLog.Infof("找到closeButtonFlag元素")
			break // 找到元素，退出循环
		}

		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}

	for i := 0; i < 5; i++ {
		err = closeButtonFlag.Click()
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		logger.SugarLog.Infof("无法点击关闭不再提示按钮,尝试第%d次关闭", i+1)
	}
	if err != nil {
		logger.SugarLog.Errorf("无法点击关闭不再提示按钮:%v", err)
		return
	}

	//找到 保存基本资料元素

	button, err := webDriver.FindElement(selenium.ByPartialLinkText, "保存基本资料")
	if err != nil {
		logger.SugarLog.Errorf("无法找到保存基本资料按钮:%v", err)
		return

	}
	err = button.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击保存基本资料按钮:%v", err)
		return
	}

	// 等待 保存成功 元素的出现
	waitDataTimeout := 10 * time.Second

	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed > waitDataTimeout {
			logger.SugarLog.Errorf("超时：无数据提示未出现")
			break
		}

		element, err := webDriver.FindElement(selenium.ByCSSSelector, "div.ts_txt")
		if err == nil {
			text, _ := element.Text()
			if text == "保存成功" || strings.Contains(text, "保存成功") {
				logger.SugarLog.Infof("保存成功")
				break
			}
		}

		time.Sleep(500 * time.Millisecond) // 暂停一段时间后重新尝试查找元素
	}
	return

}
