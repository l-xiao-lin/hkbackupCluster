package service

import (
	"fmt"
	"github.com/tebeka/selenium"
	"hkbackupCluster/logger"
	"strings"
	"time"
)

func CheckSupplierYms(website, username, password string) (err error) {
	for i := 0; i < 10; i++ {
		err = autoCheckSupplierYms(website, username, password)
		if err == nil {
			break
		}
		logger.SugarLog.Errorf("autoCheckSupplierYms返回error,执行第%d次检测,err:%v", i+1, err)
	}

	return
}

func autoCheckSupplierYms(website, username, password string) (err error) {
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
	//等待页面加载完成,再查找元素
	timeout := 15 * time.Second

	usernameInput, err := WaitForElement(webDriver, "input.ivu-input.ivu-input-default.ivu-input-with-prefix[placeholder*='手机号码']", timeout)
	if err != nil {
		logger.SugarLog.Errorf("无法找到用户名输入框:%v", err)
		return
	}

	err = usernameInput.SendKeys(username)
	if err != nil {
		logger.SugarLog.Errorf("无法填充用户名:%v", err)
		return
	}

	passwordInput, err := webDriver.FindElement(selenium.ByCSSSelector, "input[placeholder='密码']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到密码输入框:%v", err)
		return
	}
	err = passwordInput.SendKeys(password)
	if err != nil {
		logger.SugarLog.Errorf("无法填充密码:%v", err)
		return
	}

	loginButton, err := webDriver.FindElement(selenium.ByXPATH, "//button[contains(., '立即登录')]")
	if err != nil {
		logger.SugarLog.Errorf("无法找到登录按钮:%v", err)
		return
	}

	err = loginButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击登录按钮:%v", err)
		return
	}

	//检查是否登录成功，检测页面右上角是否有'测试供应商03'信息

	waitTimeout := 10 * time.Second
	maxWaitTime := time.Now().Add(waitTimeout)
	found := false

	for time.Now().Before(maxWaitTime) {
		accountElem, err := webDriver.FindElement(selenium.ByCSSSelector, "div.mr15.ivu-dropdown span.text_name")
		if accountElem != nil && err == nil {
			text, err := accountElem.Text()
			if err == nil {
				if strings.Contains(text, "测试供应商03") {
					found = true
					logger.SugarLog.Infof("找到供应商03用户信息")
					break
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !found {
		logger.SugarLog.Errorf("未找到测试供应商03用户信息")
	} else {
		logger.SugarLog.Infof("success登录并且获取到供应商03用户信息")
	}

	return
}
