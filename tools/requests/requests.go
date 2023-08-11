// -*- coding: utf-8 -*-
// @Time    : 2022/4/6 17:44
// @Author  : Nuanxinqing
// @Email   : nuanxinqing@gmail.com
// @File    : requests.go

package requests

import (
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Requests 封装HTTP请求
func Requests(method, url, data, token string) ([]byte, error) {
	// 创建HTTP实例
	client := &http.Client{Timeout: 3 * time.Second}

	// 添加请求数据
	var ReqData = strings.NewReader(data)
	req, err := http.NewRequest(method, url, ReqData)
	// 添加请求Token
	if token != "" {
		Token := fmt.Sprintf("Bearer %s", token)
		req.Header.Set("Authorization", Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error(err.Error())
		return []byte(""), err
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error(err.Error())
		return []byte(""), err
	}

	zap.L().Debug(fmt.Sprintf("%s\n", bodyText))
	return bodyText, err

}

// Down 下载模块
func Down(url string) (body *http.Response, err error) {
	// 创建HTTP实例
	client := &http.Client{Timeout: 5 * time.Minute}
	// 添加请求数据
	var ReqData = strings.NewReader("")
	req, err := http.NewRequest("GET", url, ReqData)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.67 Safari/537.36")
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error(err.Error())
		return body, err
	}

	return resp, nil
}
