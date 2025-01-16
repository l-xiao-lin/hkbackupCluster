package service

import (
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

func CheckWwwYms(website, username, password string) (err error) {
	for i := 0; i < 10; i++ {
		err = autoCheckWwwYms(website, username, password)
		if err == nil {
			break
		}
		logger.SugarLog.Errorf("CheckWwwYmsHandler返回error,执行第%d次检测,err:%v", i+1, err)
	}

	return
}

func autoCheckWwwYms(website, username, password string) (err error) {
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

	scriptButton := `document.querySelector('div.ivu-modal:not([style*="display: none"]) .ivu-modal-footer button.ivu-btn.ivu-btn-primary span').click();`
	_, err = webDriver.ExecuteScript(scriptButton, nil)
	if err != nil {
		logger.SugarLog.Errorf("无法执行javascript,无法关闭对话框")
		return
	} else {
		logger.SugarLog.Infof("成功关闭首页对话框")

	}

	time.Sleep(15 * time.Second)
	//查找登录元素

	loginWaitTimeout := 5 * time.Second
	loginMaxTimeout := time.Now().Add(loginWaitTimeout)
	numIterations := 0
	loginFound := false
	for time.Now().Before(loginMaxTimeout) {
		numIterations++
		loginButtonElem, err := webDriver.FindElement(selenium.ByCSSSelector, "div.head-info-noLogin button.ivu-btn-primary span")
		if loginButtonElem != nil && err == nil {
			err = loginButtonElem.Click()
			if err == nil {
				loginFound = true
				logger.SugarLog.Infof("找到首页登录元素,并成功点击登录按钮")
				break
			}
		}
		logger.SugarLog.Infof("未找到首页登录按钮元素,执行下一次查找:%d", numIterations)
		time.Sleep(500 * time.Millisecond)
	}

	if !loginFound {
		logger.SugarLog.Infof("未找到首页登录按钮元素,继续后面操作")
	}

	usernameInput, err := webDriver.FindElement(selenium.ByCSSSelector, "input.ivu-input.ivu-input-default.ivu-input-with-prefix[placeholder*='手机号码']")
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

	//关闭登录页面中的对话框
	waitTimeout := 10 * time.Second
	maxWaitTime := time.Now().Add(waitTimeout)
	found := false
	for time.Now().Before(maxWaitTime) {
		innerScriptButton := `document.querySelector('div.ivu-modal:not([style*="display: none"]) .ivu-modal-footer button.ivu-btn.ivu-btn-primary span').click();`
		_, err = webDriver.ExecuteScript(innerScriptButton, nil)

		if err == nil {
			found = true
			logger.SugarLog.Infof("找到页面内的对话框元素,并成功关闭对话框")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		logger.SugarLog.Infof("没有找到页面内的对话框元素")
	}

	//查找是否登录成功，检测页面页面右上角是否有'商户号'

	elements, err := webDriver.FindElements(selenium.ByCSSSelector, "div.right-box span")

	if err != nil {
		logger.SugarLog.Errorf("未找到商户号元素")
		return errors.New("未找到商户号元素")
	}

	var merchantNumberText string
	for _, elem := range elements {
		text, err := elem.Text()
		if err != nil {
			continue
		}

		if strings.Contains(text, "商户号：060422") {
			merchantNumberText = text
			break
		}
	}

	if merchantNumberText != "" {
		logger.SugarLog.Infof("success 成功找到页面中的商户号")
	} else {
		logger.SugarLog.Errorf("未找到页面中的商户号")
		return errors.New("未找到页面中的商户号")
	}

	return

}
