package weex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GetOpenOrderList 获取持仓中订单列表
func GetOpenOrderList(traderUserId uint) ([]OpenOrder, error) {
	url := "https://http-gateway1.janapw.com/api/v1/public/trace/getOpenOrderList"
	method := "POST"

	// 构建请求体，使用传入的 traderUserId
	request := GetOpenOrderListRequest{
		TraderUserID: strconv.FormatUint(uint64(traderUserId), 10),
		PageNo:       1,
		PageSize:     9999,
		ContractId:   "",
		LanguageType: 1,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	payload := strings.NewReader(string(requestBytes))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("appversion", "2.0.0")
	req.Header.Add("vs", "A5a7fdv8uvY0GK93vYra79kVV4dQ76ir")
	req.Header.Add("content-type", "application/json;charset=UTF-8")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var response GetOpenOrderListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查API响应状态
	if response.Code != "SUCCESS" {
		return nil, fmt.Errorf("API returned error: code=%s, msg=%s", response.Code, response.Msg)
	}

	return response.Data.Rows, nil
}

// GetHistoryOrderList 获取历史订单列表
func GetHistoryOrderList(traderUserId string) ([]OpenOrder, error) {
	url := "https://http-gateway1.janapw.com/api/v1/public/trace/getHistoryOrderList"
	method := "POST"

	// 构建请求体，使用传入的 traderUserId
	request := GetHistoryOrderListRequest{
		TraderUserID: traderUserId,
		CurrentUserID: traderUserId,
		PageSize:     9999,
		LanguageType: 1,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	payload := strings.NewReader(string(requestBytes))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("appversion", "2.0.0")
	req.Header.Add("vs", "A5a7fdv8uvY0GK93vYra79kVV4dQ76ir")
	req.Header.Add("content-type", "application/json;charset=UTF-8")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var response GetHistoryOrderListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查API响应状态
	if response.Code != "SUCCESS" {
		return nil, fmt.Errorf("API returned error: code=%s, msg=%s", response.Code, response.Msg)
	}

	return response.Data.Rows, nil
}

// GetMetaDataV2 获取合约id -> 交易对
func GetMetaDataV2() (map[string]string, error) {
	url := "https://http-gateway1.janapw.com/api/v1/public/meta/getMetaDataV2?languageType=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("appversion", "2.0.0")
	req.Header.Add("vs", "A5a7fdv8uvY0GK93vYra79kVV4dQ76ir")
	req.Header.Add("content-type", "application/json;charset=UTF-8")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var response GetMetaDataV2Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查API响应状态
	if response.Code != "SUCCESS" {
		return nil, fmt.Errorf("API returned error: code=%s, msg=%s", response.Code, response.Msg)
	}

	code2Name := make(map[string]string)
	for _, item := range response.Data.ContractList {
		code2Name[item.Ci] = item.Cn
	}

	return code2Name, nil
}
