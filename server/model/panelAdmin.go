// -*- coding: utf-8 -*-
// @Time    : 2022/4/6 17:45
// @Author  : Nuanxinqing
// @Email   : nuanxinqing@gmail.com
// @File    : panelAdmin.go

package model

import "gorm.io/gorm"

// QLPanel QL面板数据
type QLPanel struct {
	gorm.Model
	PanelName    string `binding:"required"` // 面板名称
	URL          string `binding:"required"` // 面板连接地址
	ClientID     string `binding:"required"` // 面板Client_ID
	ClientSecret string `binding:"required"` // 面板Client_Secret
	Enable       bool   `binding:"required"` // 是否启用面板
	PanelVersion bool   `binding:"required"` // 面板版本（0 - 旧版本 / 1 - 新版本）
	Token        string // 面板Token
	Params       int    // 面板Params
	EnvBinding   string // 绑定变量
}

// PanelAll 全部面板信息
type PanelAll struct {
	ID           uint   `json:"ID"`
	PanelName    string `json:"name"`
	URL          string `json:"url"`
	ClientID     string `json:"id"`
	ClientSecret string `json:"secret"`
	Enable       bool   `json:"enablePanel"`  // 是否启用面板
	PanelVersion bool   `json:"panelVersion"` // 面板版本（0 - 新版本 / 1 - 旧版本）
	EnvBinding   string `json:"envBinding"`
}

// PanelData 创建面板数据
type PanelData struct {
	Name         string `json:"name"`                      // 面板名称
	URL          string `json:"url" binding:"required"`    // 面板连接地址
	ID           string `json:"id" binding:"required"`     // 面板Client_ID
	Secret       string `json:"secret" binding:"required"` // 面板Client_Secret
	Enable       bool   `json:"enablePanel"`               // 是否启用面板
	PanelVersion bool   `json:"panelVersion"`              // 面板版本（0 - 新版本 / 1 - 旧版本）
}

// UpPanelData 更新面板数据
type UpPanelData struct {
	UID          int    `json:"uid" binding:"required"`    // 数据库ID值
	Name         string `json:"name" binding:"required"`   // 面板名称
	URL          string `json:"url" binding:"required"`    // 面板连接地址
	ID           string `json:"id" binding:"required"`     // 面板Client_ID
	Secret       string `json:"secret" binding:"required"` // 面板Client_Secret
	Enable       bool   `json:"enablePanel"`               // 是否启用面板
	PanelVersion bool   `json:"panelVersion"`              // 面板版本（0 - 新版本 / 1 - 旧版本）
}

// DelPanelData 删除面板数据
type DelPanelData struct {
	UID int `json:"uid" binding:"required"` // 数据库ID值
}

// PanelEnvData 修改面板绑定变量
type PanelEnvData struct {
	UID        int      `json:"uid" binding:"required"`        // 数据库ID值
	EnvBinding []string `json:"envBinding" binding:"required"` // 变量值
}

type EnvStartServer struct {
	// 可用服务器组
	ServerData []envSData `json:"serverData"`
}

type envSData struct {
	// 容器ID
	ID int `json:"ID"`
	// 容器名称
	Name string `json:"PanelName"`
	// 容器绑定变量
	EnvData []envNameData `json:"envData"`
}

// Token 面板Token数据
type Token struct {
	Code int `json:"code"`
	Data struct {
		Token      string `json:"token"`
		TokenType  string `json:"token_type"`
		Expiration int    `json:"expiration"`
	} `json:"data"`
	Message string
}

// PanelRes 面板Token数据
type PanelRes struct {
	Code       int `json:"code"`
	StatusCode int `json:"statusCode"`
	Message    string
}

// EnvData 面板变量数据
type EnvData struct {
	Code int       `json:"code"`
	Data []envData `json:"data"`
}

type envData struct {
	ID      int    `json:"id"`
	OId     string `json:"_id"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Remarks string `json:"remarks"`
}
