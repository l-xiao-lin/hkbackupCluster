package callThirdPartyAPI

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"io"
	"net/http"
	"time"
)

type SyncResult struct {
	ReleaseTaskID     string `json:"releaseTaskId"`
	ReleaseTaskResult string `json:"releaseTaskResult"`
	ReturnReason      string `json:"returnReason,omitempty"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var APIURL = "https://automated.pvt.tongtool.com/tool-service/releaseTask/api/syncResult"

func JenkinsBuildResultRsync(taskID string, status string, buildErr error, jenkinsAction *string) error {
	var result SyncResult

	if buildErr != nil {
		result = SyncResult{
			ReleaseTaskID:     taskID,
			ReleaseTaskResult: status,
			ReturnReason:      fmt.Sprintf("reason:%v", buildErr),
		}
	} else {
		result = SyncResult{
			ReleaseTaskID:     taskID,
			ReleaseTaskResult: status,
			ReturnReason:      *jenkinsAction,
		}
	}

	jsonData, err := json.Marshal(&result)
	if err != nil {
		logger.SugarLog.Errorf("Failed to marshal sync result for TaskID %s:%v", taskID, err)
		return err
	}

	req, err := http.NewRequest("POST", APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.SugarLog.Errorf("Failed to create HTTP request for TaskID %s:%v", taskID, err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("appid", "96e79218965eb72c92a549dd5a330112")

	client := http.Client{Timeout: time.Second * 10}
	respAPI, err := client.Do(req)
	if err != nil {
		logger.SugarLog.Errorf("Failed to send POST request for TaskID %s:%v", taskID, err)
		return err
	}
	defer respAPI.Body.Close()

	body, err := io.ReadAll(respAPI.Body)
	if err != nil {
		logger.SugarLog.Errorf("Failed to read response body for TaskID %s:%v", taskID, err)
		return err
	}
	var apiResponse APIResponse
	if err = json.Unmarshal(body, &apiResponse); err != nil {
		logger.SugarLog.Errorf("Failed to unmarshal response body for TaskID %s:%v", taskID, err)
		return err
	}

	if apiResponse.Code != 0 {
		logger.SugarLog.Errorf("call callThirdPartyAPI failed,err:%v", apiResponse.Message)
		return errors.New("call callThirdPartyAPI failed")
	}

	logger.SugarLog.Infof("Successfully synced result for TaskID %s", taskID)

	return nil

}
