package sendEmail

import (
	"crypto/tls"
	"hkbackupCluster/logger"
	"net/smtp"

	"github.com/jordan-wright/email"
)

func SendEmail() (err error) {
	smtpServer := "mail.tobosoft.com.cn"
	authEmail := "feedback@isunor.com"
	authPassword := "T12345678"
	from := "feedback@isunor.com"
	to := []string{"isunor_opt@tobosoft.com.cn"}

	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = "SQL脚本执行通知"
	e.Text = []byte("有SQL语句待执行")

	err = e.SendWithTLS(smtpServer+":465", smtp.PlainAuth("", authEmail, authPassword, smtpServer), &tls.Config{ServerName: smtpServer, InsecureSkipVerify: false})
	if err != nil {
		logger.SugarLog.Errorf("SendWithTLS failed,err%v", err)
		return
	}
	logger.SugarLog.Infof("SendWithTLS success.")
	return nil
}
