package pack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type NewRespBuild struct {
	BuildNumber int    `json:"build_number"`
	BuildResult string `json:"build_result"`
}

var (
	maxRetries    = 90
	retryInterval = time.Second * 20

	param = map[string][]string{
		"deploy_pods": {"listing-common", "listingsdk", "listing-m"},
	}
	jenkinsURL = "http://jenkins.tongtool.com"
	jobName    = "demo"
	username   = "admin"
	apiToken   = "11fa0303dc916b1611e89f97d469eeb9b3"
)

func Demo() {
	NewJenkinsBuild(jenkinsURL, jobName, username, apiToken, param)
}

func triggerJenkinsBuildWithJSON(jenkinsURL, jobName, username, apiToken string, param map[string][]string) (string, error) {
	URL := fmt.Sprintf("%s/job/%s/buildWithParameters", jenkinsURL, jobName)
	formData := url.Values{}
	for key, values := range param {
		for _, value := range values {
			formData.Add(key, value)
		}
	}

	zap.L().Info("Trigger Jenkins build with params",
		zap.String("jenkinsURL", jenkinsURL),
		zap.String("jobName", jobName),
		zap.String("formData", formData.Encode()),
	)

	req, err := http.NewRequest("POST", URL, bytes.NewBufferString(formData.Encode()))
	req.SetBasicAuth(username, apiToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		zap.L().Error("Failed to create HTTP request", zap.String("URL", URL), zap.Error(err))
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.String("URL", URL), zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code:%d,body:%s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("failed to get build queue from response header")
	}
	parts := strings.Split(location, "/")
	queueID := parts[len(parts)-2]
	zap.L().Info("Successfully triggered Jenkins build", zap.String("queueID", queueID))
	return queueID, nil
}

func getBuildNumberFromQueue(jenkinsURL, username, apiToken, queueID string) (int, error) {
	URL := fmt.Sprintf("%s/queue/item/%s/api/json", jenkinsURL, queueID)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(username, apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.String("URL", URL), zap.Error(err))
		return 0, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	type executable struct {
		Number int `json:"number"`
	}

	type QueueResp struct {
		Executable *executable `json:"executable"`
	}

	var queueResp QueueResp
	if err := json.Unmarshal(body, &queueResp); err != nil {
		zap.L().Error("json Unmarshal failed", zap.Error(err))
		return 0, err
	}

	if queueResp.Executable != nil {
		return queueResp.Executable.Number, nil
	}
	return 0, fmt.Errorf("build number not available  yet")
}

func getBuildResult(jenkinsURL, jobName, username, apiToken string, buildNumber int) (string, error) {
	URL := fmt.Sprintf("%s/job/%s/%d/api/json", jenkinsURL, jobName, buildNumber)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.String("URL", URL), zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type BuildResp struct {
		Result string `json:"result"`
	}
	var buildResp BuildResp
	if err := json.Unmarshal(body, &buildResp); err != nil {
		zap.L().Error("json Unmarshal failed", zap.String("body", string(body)), zap.Error(err))
		return "", err
	}
	if buildResp.Result != "" {
		return buildResp.Result, nil
	}
	return "", fmt.Errorf("build result not available  yet")

}

func NewJenkinsBuild(jenkinsURL, jobName, username, apiToken string, param map[string][]string) (newRespBuild *NewRespBuild, err error) {

	var queueID string
	for i := 0; i < maxRetries; i++ {
		queueID, err = triggerJenkinsBuildWithJSON(jenkinsURL, jobName, username, apiToken, param)
		if err == nil {
			break
		}

		zap.L().Warn("Failed to trigger Jenkins build,retrying...",
			zap.String("jobName", jobName),
			zap.Int("attempt", i+1),
			zap.Error(err))

		time.Sleep(retryInterval)
	}
	if err != nil {
		zap.L().Error("Failed to trigger jenkins build after retries", zap.String("jobName", jobName), zap.Error(err))
		return nil, err
	}

	var buildNumber int
	for i := 0; i < maxRetries; i++ {
		buildNumber, err = getBuildNumberFromQueue(jenkinsURL, username, apiToken, queueID)
		if err == nil {
			break
		}
		zap.L().Warn("Failed to retries build number from queue,retrying...",
			zap.String("queueID", queueID),
			zap.Int("attempt", i+1),
			zap.Error(err))

		time.Sleep(retryInterval)
	}

	if err != nil {
		zap.L().Error("Failed to getBuildNumberFromQueue after retries", zap.String("jobName", jobName),
			zap.String("queue", queueID))
		return nil, err
	}

	var result string
	for i := 0; i < maxRetries; i++ {
		result, err = getBuildResult(jenkinsURL, jobName, username, apiToken, buildNumber)

		if err == nil {
			break
		}
		zap.L().Warn("Get build Result failed,retrying...",
			zap.Int("buildNumber", buildNumber),
			zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(retryInterval)

	}
	if err != nil {
		zap.L().Error("Failed to getBuildResult after retries",
			zap.String("jobName", jobName),
			zap.Int("buildNumber", buildNumber),
			zap.String("buildResult", result))
		return nil, err
	}

	newRespBuild = &NewRespBuild{
		BuildNumber: buildNumber,
		BuildResult: result,
	}
	zap.L().Info("Jenkins build successful.",
		zap.String("jobName", jobName),
		zap.Any("param", param),
		zap.Int("buildNumber", newRespBuild.BuildNumber),
		zap.String("buildResult", newRespBuild.BuildResult))
	return newRespBuild, nil

}
