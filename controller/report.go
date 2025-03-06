package controller

import (
	"bytes"
	"crypto/tls"
	"hkbackupCluster/logger"
	"html/template"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"

	"github.com/gin-gonic/gin"
)

type OperationFaultDetail struct {
	Product          string `json:"product"`
	EnvName          string `json:"env_name"`
	RecoveryMethod   string `json:"recovery_method"`
	ImpactOnBusiness bool   `json:"impact_on_business"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	AffectedTime     string `json:"affected_time"`
	FaultDetail      string `json:"fault_detail"`
}

type VersionCheckItem struct {
	Description    string `json:"description"`
	IsNormal       bool   `json:"is_normal"`
	OccurrenceTime string `json:"occurrence_time"`
	RecoveryTime   string `json:"recovery_time"`
	HandlerDetail  string `json:"handler_detail"`
}

type ReportInfo struct {
	Date       string `json:"date"`
	Ops        string `json:"ops"`
	Failures   int    `json:"failures"`
	Recovery   int    `json:"recovery"`
	ManualHand int    `json:"manual_hand"`
}

type EmailInfo struct {
	To []string `json:"to"`
	Cc []string `json:"cc"`
}

func ReportHandler(c *gin.Context) {
	data := gin.H{
		"OperationFaultDetails": []OperationFaultDetail{
			{Product: "ERP2.0", EnvName: "ERP110", RecoveryMethod: "自动", ImpactOnBusiness: false, StartTime: "2025-02-27 18:00:00", EndTime: "2025-02-27 19:00:00", AffectedTime: "60", FaultDetail: "执行计划任务机器负载过高"},
			{Product: "ERP2.0", EnvName: "ERP115", RecoveryMethod: "手动", ImpactOnBusiness: false, StartTime: "2025-02-27 18:00:00", EndTime: "2025-02-27 19:00:00", AffectedTime: "60", FaultDetail: "执行计划任务机器负载过高"},
			{Product: "ERP2.0", EnvName: "ERP110", RecoveryMethod: "手动", ImpactOnBusiness: false, StartTime: "2025-02-27 18:00:00", EndTime: "2025-02-27 19:00:00", AffectedTime: "60", FaultDetail: "执行计划任务机器负载过高"},
			{Product: "ERP2.0", EnvName: "ERP110", RecoveryMethod: "自动", ImpactOnBusiness: false, StartTime: "2025-02-27 18:00:00", EndTime: "", AffectedTime: "", FaultDetail: "执行计划任务机器负载过高"},
			{Product: "ERP2.0", EnvName: "ERP110", RecoveryMethod: "自动", ImpactOnBusiness: false, StartTime: "2025-02-27 18:00:00", EndTime: "", AffectedTime: "", FaultDetail: "执行计划任务机器负载过高"},
		},
		"VersionCheckItems": []VersionCheckItem{
			{Description: "ERP2.0发版环境校验", IsNormal: false, OccurrenceTime: "2025-03-04 10:30:00", RecoveryTime: "2025-03-04 11:00:00", HandlerDetail: "重新发版market"},
			{Description: "ERP2.0环境服务状态", IsNormal: true},
			{Description: "Listing集群服务状态", IsNormal: true},
			{Description: "Listing独立部署服务状态", IsNormal: true},
		},
		"ReportInfo": ReportInfo{
			Date:       time.Now().Format("2006-01-02 15:04:05"),
			Ops:        "李小林",
			Failures:   5,
			Recovery:   3,
			ManualHand: 2,
		},
	}
	c.HTML(200, "report.html", data)
}

func GenerateReportHTML(data paraReport) (string, error) {
	tmpl, err := template.New("report").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>运维每日巡检报告</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: white;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        h1 {
            color: #333;
            font-size: 24px;
            margin-bottom: 20px;
        }
        .header-info {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 18px;
            margin-bottom: 20px;
        }
        .header-info span {
            padding: 0 5px; /* 添加左右内边距 */
        }
        .summary-title {
            font-size: 24px;
            color: #333;
            margin-bottom: 20px;
        }
        .summary-box {
            background-color: #f1f1f1;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            margin-bottom: 30px;
            text-align: center;
        }
        .summary-content {
            display: flex;
            justify-content: space-around;
            font-size: 20px;
            margin-top: 10px;
        }
        .summary-item {
            text-align: center;
        }
        .summary-item span {
            font-weight: bold;
            font-size: 24px;
            color: #333;
        }
        .section {
            margin-bottom: 40px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 12px;
            text-align: center; /* 添加居中对齐 */
        }
        th {
            background-color: #f2f2f2;
            font-weight: bold;
        }
        .abnormal {
            color: red;
        }
        .empty-cell {
            text-align: center;
            font-style: italic;
            color: #999;
        }

        .manual-recovery {
            color: red;
        }
        tbody tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        tbody tr:hover {
            background-color: #f1f1f1;
        }
        .fault-detail {
            font-size: 14px;
            color: #777;
        }
        .version-check ul {
            list-style-type: none;
            padding: 0;
        }
        .version-check li {
            margin-bottom: 10px;
        }
        .version-check label {
            display: block;
            font-size: 18px;
            margin-bottom: 5px;
        }
    </style>
</head>
<body>
<div class="container">
    <h1 style="text-align: center;">运维巡检报告</h1>
    <div class="header-info">
        <span>时间: {{ .ReportInfo.Date }}</span>
        <span>巡检人: {{ .ReportInfo.Ops}}</span>
    </div>

    <!-- 巡检汇总 -->
    <div class="section">
        <div class="summary-title">巡检汇总</div>
        <div class="summary-box">
            <div class="summary-content">
                <div class="summary-item">
                    故障总数：<br>
                    <span>{{ .ReportInfo.Failures }}个</span>
                </div>
                <div class="summary-item">
                    自动恢复：<br>
                    <span>{{ .ReportInfo.Recovery }}个</span>
                </div>
                <div class="summary-item">
                    手工处理：<br>
                    <span>{{ .ReportInfo.ManualHand }}个</span>
                </div>
            </div>
        </div>
    </div>

    <!-- 故障处理详情 -->
    <div class="section">
        <div class="summary-title">故障处理</div>
        <table>
            <thead>
            <tr>
                <th>项目</th>
                <th>环境</th>
                <th>恢复方式</th>
                <th>影响业务</th>
                <th>告警时间</th>
                <th>恢复时间</th>
                <th>故障时长(min)</th>
                <th>故障处理分析</th>
            </tr>
            </thead>
            <tbody>
            {{range .OperationFaultDetails}}
            <tr>
                <td>{{.Product}}</td>
                <td>{{.EnvName}}</td>
                <td{{if eq .RecoveryMethod "手动"}} class="manual-recovery"{{end}}>{{.RecoveryMethod}}</td>
                <td>{{ if .ImpactOnBusiness}}是{{else}}否{{end}}</td>
                <td>{{.StartTime}}</td>
                <td>{{ if ne .EndTime "" }} {{.EndTime}}  {{else}}--{{end}}</td>
                <td>{{if ne .AffectedTime ""}}{{.AffectedTime}}{{else}}--{{end}}</td>
                <td class="fault-detail">{{.FaultDetail}}</td>
            </tr>
            {{end}}
            </tbody>
        </table>
    </div>

</div>
</body>
</html>
`)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type paraReport struct {
	OperationFaultDetails []OperationFaultDetail `json:"OperationFaultDetails"`
	VersionCheckItems     []VersionCheckItem     `json:"VersionCheckItems"`
	ReportInfo            ReportInfo             `json:"ReportInfo"`
	EmailInfo             EmailInfo              `json:"EmailInfo"`
}

func SendReportHandler(c *gin.Context) {
	p := new(paraReport)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(200, gin.H{
			"msg":   "param invalid",
			"error": err.Error(),
			"code":  1001,
		})
		return
	}
	p.ReportInfo.Date = time.Now().Format("2006-01-02 15:04:05")

	logger.SugarLog.Infof("paraReport:%#v", p)

	htmlContent, err := GenerateReportHTML(*p)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  "Failed to generate report",
			"code": 1001,
		})
		return
	}
	if err := sendEmail(htmlContent, p.EmailInfo); err != nil {
		c.JSON(200, gin.H{
			"msg":  "Failed to sendEmail",
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}

func sendEmail(body string, emailInfo EmailInfo) (err error) {
	smtpServer := "mail.tobosoft.com.cn"
	authEmail := "feedback@isunor.com"
	authPassword := "T12345678"
	from := "feedback@isunor.com"
	e := email.NewEmail()
	e.From = from
	e.To = emailInfo.To
	e.Cc = emailInfo.Cc
	e.Subject = "运维巡检报告"
	e.HTML = []byte(body)
	err = e.SendWithTLS(smtpServer+":465", smtp.PlainAuth("", authEmail, authPassword, smtpServer), &tls.Config{ServerName: smtpServer, InsecureSkipVerify: false})
	if err != nil {
		logger.SugarLog.Errorf("SendWithTLS failed,err%v", err)
		return
	}
	logger.SugarLog.Infof("SendWithTLS success.")
	return nil

}
