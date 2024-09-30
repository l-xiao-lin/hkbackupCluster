package service

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"hkbackupCluster/logger"
	"strings"
	"time"
)

func CheckAdminYms(website, username, password string) (err error) {
	for i := 0; i < 10; i++ {
		err = autoCheckAdminYms(website, username, password)
		if err == nil {
			break
		}
		logger.SugarLog.Errorf("CheckAdminYmsHandler返回error,执行第%d次检测,err:%v", i+1, err)
	}

	return
}

func WaitForElement(driver selenium.WebDriver, selector string, timeout time.Duration) (selenium.WebElement, error) {
	startTime := time.Now()
	var element selenium.WebElement
	var err error

	for time.Since(startTime) < timeout {
		element, err = driver.FindElement(selenium.ByCSSSelector, selector)
		if err == nil && element != nil {
			return element, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, errors.New("超时无法找到元素")
}

func autoCheckAdminYms(website, username, password string) (err error) {
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

	//查找是否登录成功，检测页面右上角是否有'通途客户编号'

	waitTimeout := 20 * time.Second
	maxWaitTime := time.Now().Add(waitTimeout)
	numIterations := 0
	found := false
	for time.Now().Before(maxWaitTime) {
		numIterations++
		knowButton, err := webDriver.FindElement(selenium.ByXPATH, "//button[contains(., '知道了')]")
		if err == nil && knowButton != nil {
			err = knowButton.Click()
			if err == nil {
				found = true
				logger.SugarLog.Infof("成功关闭 知道了 对话框")
				break
			}
		}
		logger.SugarLog.Infof("未找到对话框元素，继续执行第%d次查找:%v", numIterations, err)
		time.Sleep(1 * time.Second)
	}

	if !found {
		logger.SugarLog.Infof("未能找到话框元素")
	} else {
		logger.SugarLog.Infof("共查找%d次,并且关闭知道了对话框", numIterations)
	}

	loginElement, err := webDriver.FindElement(selenium.ByCSSSelector, "span.user_info.mr15")
	if err != nil {
		logger.SugarLog.Errorf("没有登录到页面中:%v", err)
		return
	}
	text, err := loginElement.Text()
	if err != nil {
		logger.SugarLog.Errorf("无法获取文本内容:%v", err)
		return
	}

	if strings.Contains(text, "通途客户编号：059353") {
		logger.SugarLog.Infof("登录成功，并且找到客户编号")
	} else {
		logger.SugarLog.Errorf("未找到客户编号")
		return errors.New("未找到客户编号")
	}
	return

}
