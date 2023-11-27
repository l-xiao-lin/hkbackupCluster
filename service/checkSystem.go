package service

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"hkbackupCluster/logger"
	"sync"
	"time"
)

var m sync.Mutex

func CheckSystem(MerchantsID []string) (Resp map[string]interface{}) {
	start := time.Now()
	retries := 3
	Resp = make(map[string]interface{})
	//将检测失败的商户号添加到map中
	for _, merchantID := range MerchantsID {
		err := AutomateTesting(merchantID)
		if err != nil {
			Resp[merchantID] = err
		}
	}

	logger.SugarLog.Infof("需要重试的商户号Resp:%v", Resp)

	var retryMerchants []string
	if len(Resp) > 0 {
		for retryMerchantID := range Resp {
			retryMerchants = append(retryMerchants, retryMerchantID)
		}

		//重试功能校验
		for i := 0; i < retries; i++ {
			var updateRetryMerchants []string
			for _, merchantID := range retryMerchants {
				err := AutomateTesting(merchantID)
				if err == nil {
					logger.SugarLog.Infof("重试成功的商户号:%s", merchantID)
					delete(Resp, merchantID)
				} else {
					updateRetryMerchants = append(updateRetryMerchants, merchantID)
				}
			}
			retryMerchants = updateRetryMerchants
			if len(retryMerchants) == 0 {
				break
			}
		}
	}
	elapsed := time.Since(start)
	logger.SugarLog.Infof("总共花费时间:%v", elapsed)
	return Resp

}

func AutomateTesting(merchantID string) error {
	m.Lock()
	defer m.Unlock()
	opts := []selenium.ServiceOption{
		selenium.Output(nil),
	}

	service, err := selenium.NewChromeDriverService("C:\\Program Files\\Google\\Chrome\\Application\\chromedriver.exe", 9515, opts...)
	if err != nil {
		logger.SugarLog.Errorf("无法启动 WebDriver 服务 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		logger.SugarLog.Errorf("无法打开会话 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	defer webDriver.Quit()

	err = webDriver.Get("https://passport.tongtool.com/")
	if err != nil {
		logger.SugarLog.Errorf("无法导航到网页 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	usernameInput, err := webDriver.FindElement(selenium.ByCSSSelector, "input[name='username']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到用户名输入框 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}
	err = usernameInput.SendKeys("yingbingbing@isunor.com")
	if err != nil {
		logger.SugarLog.Errorf("无法填充用户名 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	passwordInput, err := webDriver.FindElement(selenium.ByID, "password")
	if err != nil {
		logger.SugarLog.Errorf("无法找到密码输入框 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}
	err = passwordInput.SendKeys("xgrXVwhCVrmspl&C")
	if err != nil {
		logger.SugarLog.Errorf("无法填充密码 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	loginButton, err := webDriver.FindElement(selenium.ByXPATH, "//button[contains(text(), '立即登录')]")
	if err != nil {

		logger.SugarLog.Errorf("无法找到登录按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	err = loginButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击登录按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	var textInput selenium.WebElement
	waitTimeout := 5 * time.Second
	maxWaitTime := time.Now().Add(waitTimeout)

	for time.Now().Before(maxWaitTime) {
		textInput, err = webDriver.FindElement(selenium.ByCSSSelector, "input[name='q']")
		if err == nil && textInput != nil {
			break // 找到元素，退出循环
		}

		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}
	if textInput == nil {
		logger.SugarLog.Errorf("没有找到搜索框，或者textInput值为nil %s", merchantID)
		return errors.New("没有找到搜索框，或者textInput值为nil")
	}

	err = textInput.SendKeys(merchantID)
	if err != nil {
		logger.SugarLog.Errorf("无法输入商户号 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	submitButton, err := webDriver.FindElement(selenium.ByCSSSelector, "input[type='submit'][value='筛选']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到筛选提交按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}
	err = submitButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击筛选提交按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	erpLink, err := webDriver.FindElement(selenium.ByLinkText, "ERP2.0")
	if err != nil {
		logger.SugarLog.Errorf("无法找到进入ERP链接 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	err = erpLink.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击进入ERP链接 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 获取当前窗口句柄
	currentWindowHandle, err := webDriver.CurrentWindowHandle()
	if err != nil {
		logger.SugarLog.Errorf("无法获取当前窗口句柄 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 获取所有窗口句柄
	windowHandles, err := webDriver.WindowHandles()
	if err != nil {
		logger.SugarLog.Errorf("无法获取窗口句柄列表 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 切换到新打开的窗口
	for _, handle := range windowHandles {
		if handle != currentWindowHandle {
			err = webDriver.SwitchWindow(handle)
			if err != nil {
				logger.SugarLog.Errorf("无法切换到新窗口 %s:%v", merchantID, err)
				return fmt.Errorf("%s:%v", merchantID, err)
			}
			break
		}
	}

	var businessDocLink selenium.WebElement

	waitTimeoutBusiness := 5 * time.Second
	maxWaitTimeBusiness := time.Now().Add(waitTimeoutBusiness)

	for time.Now().Before(maxWaitTimeBusiness) {
		businessDocLink, err = webDriver.FindElement(selenium.ByPartialLinkText, "业务单据状态查询")
		if err == nil && businessDocLink != nil {
			break // 找到元素，退出循环
		}

		time.Sleep(500 * time.Millisecond) // 等待一段时间后重试
	}

	if businessDocLink == nil {
		logger.SugarLog.Errorf("未找到 业务单据状态查询元素，或businessDocLink值为nil %s", merchantID)
		return errors.New("未找到 业务单据状态查询元素，或businessDocLink值为nil")
	}

	err = businessDocLink.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击业务单据状态查询链接 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	element, err := webDriver.FindElement(selenium.ByCSSSelector, "textarea[name='contextField']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到元素 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 输入文本
	err = element.SendKeys("0000000")
	if err != nil {
		logger.SugarLog.Errorf("无法输入文本 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 找到查询按钮，并点击查询
	queryButton, err := webDriver.FindElement(selenium.ByCSSSelector, "a.toggle_btn[onclick='queryBusinessDoc()']")
	if err != nil {
		logger.SugarLog.Errorf("无法找到查询按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 点击按钮
	err = queryButton.Click()
	if err != nil {
		logger.SugarLog.Errorf("无法点击查询按钮 %s:%v", merchantID, err)
		return fmt.Errorf("%s:%v", merchantID, err)
	}

	// 等待 ico_agreen 元素的出现
	waitDataTimeout := 5 * time.Second

	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed > waitDataTimeout {
			logger.SugarLog.Errorf("超时：无数据提示未出现")
			break
		}

		_, err := webDriver.FindElement(selenium.ByCSSSelector, "span.ico_agreen.left")
		if err == nil {
			logger.SugarLog.Infof("页面正常，%s环境检测正常", merchantID)
			break
		}

		time.Sleep(500 * time.Millisecond) // 暂停一段时间后重新尝试查找元素
	}

	return nil

}
