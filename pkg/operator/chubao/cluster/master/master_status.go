package master

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type responseData struct {
	code int          `json:"code"`
	msg  string       `json:"msg"`
	data MasterStatus `json:"data"`
}

type MasterStatus struct {
	dataNodeStatInfo dataNodeStatInfo     `json:"dataNodeStatInfo"`
	metaNodeStatInfo metaNodeStatInfo     `json:"metaNodeStatInfo"`
	zoneStatInfo     map[string]*zoneStat `json:"ZoneStatInfo"`
}

type dataNodeStatInfo struct {
	totalGB     float32 `json:"TotalGB"`
	usedGB      float32 `json:"UsedGB"`
	increasedGB float32 `json:"IncreasedGB"`
	usedRatio   float32 `json:"UsedRatio"`
}

type metaNodeStatInfo struct {
	totalGB     float32 `json:"TotalGB"`
	usedGB      float32 `json:"UsedGB"`
	increasedGB float32 `json:"IncreasedGB"`
	usedRatio   float32 `json:"UsedRatio"`
}

type zoneStat struct {
	dataNodeStat dataNodeStat `json:"dataNodeStat"`
	metaNodeStat metaNodeStat `json:"metaNodeStat"`
}

type dataNodeStat struct {
	totalGB       float32 `json:"TotalGB"`
	usedGB        float32 `json:"UsedGB"`
	availGB       float32 `json:"AvailGB"`
	usedRatio     float32 `json:"UsedRatio"`
	totalNodes    int     `json:"TotalNodes"`
	writableNodes int     `json:"WritableNodes"`
}

type metaNodeStat struct {
	totalGB       float32 `json:"TotalGB"`
	usedGB        float32 `json:"UsedGB"`
	availGB       float32 `json:"AvailGB"`
	usedRatio     float32 `json:"UsedRatio"`
	totalNodes    int     `json:"TotalNodes"`
	writableNodes int     `json:"WritableNodes"`
}

func getMasterStatusURL(serviceName, namespace string, port int32) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local/cluster/stat:%d", serviceName, namespace, port)
}

func (m *Master) GetStatus() (*MasterStatus, error) {
	masterStatusURL := getMasterStatusURL(ServiceName, m.namespace, m.port)
	return getStatus(masterStatusURL)
}

func getStatus(masterStatusURL string) (*MasterStatus, error) {
	req, err := http.NewRequest(http.MethodGet, masterStatusURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to build the request for query chubao cluster status. url:%s", masterStatusURL))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to query chubao cluster status. url:%s", masterStatusURL))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, fmt.Sprintf("get failed StatusCode[%d] from chubao cluster. url:%s", resp.StatusCode, masterStatusURL))
	}

	responseData := &responseData{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read chubao cluster status i/o. url:%s", masterStatusURL))
	}

	err = json.Unmarshal(bytes, responseData)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal masterStatus. url:%s content:%s", masterStatusURL, string(bytes)))
	}

	if responseData.code != 0 {
		return nil, fmt.Errorf("chubao cluster inner error, code:%d msg:%s", responseData.code, responseData.msg)
	}

	return &responseData.data, nil
}

func (ms *MasterStatus) IsAvailable() bool {
	for _, value := range ms.zoneStatInfo {
		if value.dataNodeStat.writableNodes >= 3 && value.metaNodeStat.writableNodes >= 3 {
			return true
		}
	}

	return false
}
