// -*- coding: utf-8 -*-
// @Time    : 2022/4/7 19:19
// @Author  : Nuanxinqing
// @Email   : nuanxinqing@gmail.com
// @File    : openLogic.go

package logic

import (
	"QLPanelTools/server/model"
	"QLPanelTools/server/sqlite"
	"QLPanelTools/tools/goja"
	"QLPanelTools/tools/panel"
	"QLPanelTools/tools/requests"
	res "QLPanelTools/tools/response"
	"encoding/json"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
)

// EnvData 获取all data
func EnvData() (res.ResCode, model.EnvStartServer) {
	var sd model.EnvStartServer

	// 获取所有服务器信息
	sData := sqlite.GetServerCount()
	data, err := json.Marshal(sData)
	if err != nil {
		zap.L().Error(err.Error())
		return res.CodeServerBusy, sd
	}
	// 转化JSON脱敏
	err = json.Unmarshal(data, &sd.ServerData)
	if err != nil {
		zap.L().Error(err.Error())
		return res.CodeServerBusy, sd
	}

	for i := 0; i < len(sd.ServerData); i++ {
		// 获取变量数据
		envData := sqlite.GetEnvAllByID(sd.ServerData[i].ID)
		if len(envData) != 0 {
			eData, err := json.Marshal(envData)
			if err != nil {
				zap.L().Error(err.Error())
				return res.CodeServerBusy, sd
			}

			// 数据绑定
			err = json.Unmarshal(eData, &sd.ServerData[i].EnvData)
			if err != nil {
				zap.L().Error(err.Error())
				return res.CodeServerBusy, sd
			}

			// 获取面板已存在变量数量
			url := panel.StringHTTP(sData[i].URL) + "/open/envs?searchValue=&t=" + strconv.Itoa(sData[i].Params)
			allData, err := requests.Requests("GET", url, "", sData[i].Token)
			if err != nil {
				zap.L().Error("面板：" + sData[i].PanelName + ", 已无法连接，请管理员尽快处理")
				sd.ServerData[i].Name = "服务器已失去连接（禁止提交）"
				sd.ServerData[i].ID = -100
				continue
				//return res.CodeServerBusy, sd
			}
			var token model.EnvData
			err = json.Unmarshal(allData, &token)
			if err != nil {
				zap.L().Error(err.Error())
				return res.CodeServerBusy, sd
			}

			// 判断返回状态
			if token.Code != 200 {
				// 尝试获取授权
				go panel.GetPanelToken(sData[i].URL, sData[i].ClientID, sData[i].ClientSecret)

				// 未授权或Token失效
				continue
				//return res.CodeDataError, sd
			}

			// 计算变量剩余限额
			for x := 0; x < len(sd.ServerData[i].EnvData); x++ {
				if envData[x].Mode == 1 {
					// 新建模式
					sd.ServerData[i].EnvData[x].Quantity, _, _ = CalculateQuantity(sd.ServerData[i].ID, envData[x].Mode, sd.ServerData[i].EnvData[x].Name, "")
				} else {
					// 合并模式
					sd.ServerData[i].EnvData[x].Quantity, _, _ = CalculateQuantity(sd.ServerData[i].ID, envData[x].Mode, sd.ServerData[i].EnvData[x].Name, envData[x].Division)
				}
			}
		}
	}
	return res.CodeSuccess, sd
}

// EnvAdd 添加变量
func EnvAdd(p *model.EnvAdd) (res.ResCode, string) {
	var err error
	var s2 string

	s2 = p.EnvData

	// 不允许内容为空
	if p.EnvData == "" {
		return res.CodeDataIsNull, ""
	}

	var token model.PanelRes
	// 校验服务器ID
	result, sData := sqlite.CheckServerDoesItExist(p.ServerID)
	if result != true {
		// 服务器不存在
		zap.L().Debug("提交容器不存在：" + strconv.Itoa(p.ServerID))
		return res.CodeErrorOccurredInTheRequest, ""
	}

	// 校验变量名是否存在
	resultEnv, eData := sqlite.CheckEnvNameDoesItExist(p.EnvName)
	if resultEnv != true {
		// 变量不存在
		zap.L().Debug("提交变量名不存在：" + p.EnvName)
		return res.CodeErrorOccurredInTheRequest, ""
	}

	// 转换切片
	envBind := strings.Split(sData.EnvBinding, "@")
	// 校验变量是否处于容器白名单
	zap.L().Debug("校验变量是否处于容器白名单")
	zap.L().Debug("提交值：" + strconv.Itoa(int(eData.ID)))
	num := 0
	for i := 0; i < len(envBind); i++ {
		zap.L().Debug("变量绑定数据：" + envBind[i])
		if envBind[i] != "" {
			if envBind[i] == strconv.Itoa(int(eData.ID)) {
				num++
			}
		}
	}
	if num == 0 {
		return res.CodeErrorOccurredInTheRequest, ""
	}

	// 校验CDK是否受限
	if eData.IsCDK != false {
		// 需要校验CKD, 校验CDK是否存在
		if p.EnvCDK != "" {
			// 查询CDK是否存在
			cdk := sqlite.GetCDKData(p.EnvCDK)
			if cdk.CdKey == "" {
				// CDK查询为空
				return res.CodeCDKError, "您的CDK已失效"
			}

			// CDK是否被禁用
			if cdk.State == false {
				// CDK已被管理员禁用
				return res.CodeCDKError, "您的CDK已被禁用"
			}

			if cdk.AvailableTimes <= 0 {
				// 当前CDK使用次数已耗尽
				return res.CodeCDKError, "您CDK使用次数已耗尽"
			}
		} else {
			return res.CodeCDKError, "需要校验CDK，请点击右上角按钮填写您的CDK"
		}
	}

	// 正则处理(检查是否符合规则)
	var s [][]string
	if eData.Regex != "" {
		// 需要处理正则
		zap.L().Debug("需要处理正则")
		reg := regexp.MustCompile(eData.Regex)
		// 匹配内容
		if reg != nil {
			s = reg.FindAllStringSubmatch(p.EnvData, -1)
			if len(s) == 0 {
				return res.CodeEnvDataMismatch, ""
			}
			s2 = s[0][0]
		} else {
			return res.CodeServerBusy, ""
		}
	}

	// 正则处理(检查是否属于黑名单)
	list, err := sqlite.GetSetting("blacklist")
	if err != nil {
		zap.L().Error(err.Error())
		return res.CodeServerBusy, ""
	}
	if list.Value != "" {
		// 如果黑名单不为空,正则匹配是否属于黑名单
		breakList := strings.Split(list.Value, "@")
		for i := 0; i < len(breakList); i++ {
			reg := regexp.MustCompile(breakList[i])
			s = reg.FindAllStringSubmatch(p.EnvData, -1)
			if len(s) != 0 {
				return res.CodeBlackListEnv, ""
			}
		}
	}

	// 校验变量配额
	c, t, code := CalculateQuantity(p.ServerID, eData.Mode, p.EnvName, eData.Division)
	if code == res.CodeServerBusy {
		zap.L().Debug("处理正则失败")
		return res.CodeServerBusy, ""
	} else if c <= 0 {
		if eData.Mode != 3 {
			zap.L().Debug("限额已满，禁止提交")
			return res.CodeLocationFull, ""
		}
	}

	// 检查重复提交
	var bol bool
	var QCount int

	bol, QCount = CheckRepeat(t, s2, p.EnvName, eData)
	if bol == true {
		return res.CodeNoDuplicateSubmission, ""
	}

	// 是否启用插件
	if eData.IsPlugin != false {
		// 启用插件, 传入插件名称和变量
		zap.L().Debug("变量：" + eData.Name + "  启用插件：" + eData.PluginName)
		js, s3, err := goja.RunJS(eData.PluginName, s2)
		if err != nil {
			return res.CodeCustomError, err.Error()
		}
		if js != true {
			return res.CodeNoAdmittance, s3
		}
		s2 = s3
	}

	// 提交到服务器
	var data string
	var idDate string
	var qlVersion string
	url := panel.StringHTTP(sData.URL) + "/open/envs?t=" + strconv.Itoa(sData.Params)
	idDateUrl := panel.StringHTTP(sData.URL) + "/open/envs/enable?t=" + strconv.Itoa(sData.Params)
	zap.L().Debug(url)

	// 将字符串里面的双引号添加转义
	zap.L().Debug("转义前：" + s2)
	s2 = strings.Replace(s2, `"`, `\"`, -1)
	zap.L().Debug("转义后：" + s2)

	// 指定上传数据
	if eData.Mode == 1 {
		// 新建模式
		zap.L().Debug("上传变量：新建模式")
		data = `[{"value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}]`
	} else if eData.Mode == 2 {
		// 合并模式
		zap.L().Debug("上传变量：合并模式")
		if QCount != -1 {
			//vv := t.Data[QCount].Value + eData.Division + s2
			p.EnvRemarks = t.Data[QCount].Name
			vv := ""
			sList := strings.Split(t.Data[QCount].Value, "\n")
			zap.L().Debug("sList的数量是：" + strconv.Itoa(len(sList)))
			if len(sList) != 1 {
				for _, str := range sList {
					vv += str + "\\n"
				}
				zap.L().Debug("合并前：" + vv)
				vv = vv + s2
				zap.L().Debug("合并后：" + vv)
			} else {
				vv = t.Data[QCount].Value + eData.Division + s2
			}
			if t.Data[QCount].OId != "" {
				data = `{"_id": "` + t.Data[QCount].OId + `", "value": "` + vv + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
			} else {
				data = `{"id": "` + strconv.Itoa(t.Data[QCount].ID) + `", "value": "` + vv + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
			}
		} else {
			data = `[{"value": "` + s2 + `","name": "` + p.EnvName + `"}]`
		}
	} else {
		// 更新模式
		zap.L().Debug("上传变量：更新模式")
		/*
			1、获取传入变量的正则
			2、循环匹配正则
			3、匹配成功：更新、匹配失败：新建
		*/
		reg := regexp.MustCompile(eData.ReUpdate)
		s3 := reg.FindAllStringSubmatch(s2, -1)
		co := 0
		for i := 0; i < len(t.Data); i++ {
			// 循环匹配正则, 判断面板变量名和传入变量名是否一致
			if t.Data[i].Name == p.EnvName {
				// 一致, 获取变量正则部分
				envData := reg.FindAllStringSubmatch(t.Data[i].Value, -1)
				// 判断匹配结果是否为空
				if len(envData) != 0 {
					// 判断两个正则值是否一致
					zap.L().Debug("-----更新模式：匹配变量-----")
					zap.L().Debug(envData[0][0])
					zap.L().Debug(s3[0][0])
					if envData[0][0] == s3[0][0] {
						// 一致，更新变量
						QCount = 100
						co = 0
						// 如果原有变量的备注是否为空
						if t.Data[i].Remarks != "" {
							// 用户提交备注是否为空
							if p.EnvRemarks == "" {
								if t.Data[i].OId != "" {
									idDate = t.Data[i].OId
									qlVersion = "旧版本"
									data = `{"_id": "` + t.Data[i].OId + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + t.Data[i].Remarks + `"}`
								} else {
									idDate = strconv.Itoa(t.Data[i].ID)
									qlVersion = "新版本"
									data = `{"id": "` + strconv.Itoa(t.Data[i].ID) + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + t.Data[i].Remarks + `"}`
								}
							} else {
								if t.Data[i].OId != "" {
									idDate = t.Data[i].OId
									qlVersion = "旧版本"
									data = `{"_id": "` + t.Data[i].OId + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
								} else {
									idDate = strconv.Itoa(t.Data[i].ID)
									qlVersion = "新版本"
									data = `{"id": "` + strconv.Itoa(t.Data[i].ID) + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
								}
							}
						} else {
							if t.Data[i].OId != "" {
								idDate = t.Data[i].OId
								qlVersion = "旧版本"
								data = `{"_id": "` + t.Data[i].OId + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
							} else {
								idDate = strconv.Itoa(t.Data[i].ID)
								qlVersion = "新版本"
								data = `{"id": "` + strconv.Itoa(t.Data[i].ID) + `", "value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}`
							}
						}
						break
					} else {
						// 不一致，新建变量
						co++
					}
				}
			} else {
				// 面板没存在此变量
				co++
			}
		}

		if len(t.Data) == 0 {
			co++
		}
		zap.L().Debug("更新模式：是否需要新建变量：" + strconv.Itoa(co))
		if co != 0 {
			qlVersion = "新版本"
			data = `[{"value": "` + s2 + `","name": "` + p.EnvName + `","remarks": "` + p.EnvRemarks + `"}]`
			QCount = -1
		}
	}

	zap.L().Debug(data)
	var r []byte
	if eData.Mode == 1 {
		// 新建模式(POST)
		r, err = requests.Requests("POST", url, data, sData.Token)
	} else if eData.Mode == 2 {
		// 合并模式(PUT)
		if QCount != -1 {
			r, err = requests.Requests("PUT", url, data, sData.Token)
		} else {
			// 面板不存在合并模式变量时(POST)
			r, err = requests.Requests("POST", url, data, sData.Token)
		}
	} else {
		// 更新模式(PUT)
		if QCount != -1 {
			r, err = requests.Requests("PUT", url, data, sData.Token)
			// 启用禁用变量
			var EnableID string
			if qlVersion == "新版本" {
				EnableID = "[" + idDate + "]"
			} else {
				// 旧版本
				EnableID = `["` + idDate + `"]`
			}
			zap.L().Debug("启用禁用变量：" + EnableID)
			go func() {
				_, _ = requests.Requests("PUT", idDateUrl, EnableID, sData.Token)
			}()
		} else {
			// 面板不存在变量时新建(POST)
			if c <= 0 {
				zap.L().Debug("限额已满，禁止提交")
				return res.CodeLocationFull, ""
			} else {
				r, err = requests.Requests("POST", url, data, sData.Token)
			}
		}
	}

	if err != nil {
		return res.CodeServerBusy, ""
	}

	// 序列化内容
	err = json.Unmarshal(r, &token)
	if err != nil {
		zap.L().Error(err.Error())
		return res.CodeServerBusy, ""
	}

	if token.Code > 400 && token.Code <= 500 {
		// 尝试更新Token
		zap.L().Warn("上传错误警告：" + token.Message)
		go panel.GetPanelToken(sData.URL, sData.ClientID, sData.ClientSecret)
		return res.CodeStorageFailed, token.Message
	} else if token.Code >= 500 || token.Code == 400 {
		return res.CodeStorageFailed, "提交数据发生【500】错误，错误原因：" + token.Message
	}

	if token.StatusCode == 400 {
		return res.CodeStorageFailed, "提交数据发生【400】错误，错误原因：" + token.Message
	}

	return res.CodeSuccess, ""
}

// CalculateQuantity 计算变量剩余位置
func CalculateQuantity(id, mode int, name string, division string) (int, model.EnvData, res.ResCode) {
	var token model.EnvData
	// 获取变量数据
	count := sqlite.GetEnvNameCount(name)

	// 获取容器信息
	sData := sqlite.GetPanelDataByID(id)

	// 获取面板已存在变量数量
	url := panel.StringHTTP(sData.URL) + "/open/envs?searchValue=&t=" + strconv.Itoa(sData.Params)
	allData, err := requests.Requests("GET", url, "", sData.Token)
	if err != nil {
		return 0, token, res.CodeServerBusy
	}

	err = json.Unmarshal(allData, &token)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, token, res.CodeServerBusy
	}

	// 计算变量剩余限额
	c := count
	if mode == 1 || mode == 3 {
		// 新建模式
		for i := 0; i < len(token.Data); i++ {
			if token.Data[i].Name == name {
				c--
			}
		}
	} else {
		// 合并模式
		if len(token.Data) != 0 {
			for i := 0; i < len(token.Data); i++ {
				if token.Data[i].Name == name {
					c -= len(strings.Split(token.Data[i].Value, division))
					break
				}
			}
		}
	}

	return c, token, res.CodeSuccess
}

// CheckRepeat 校验是否重复上传
func CheckRepeat(p model.EnvData, env, name string, data model.EnvName) (bool, int) {
	var QCount = -1
	// 通过变量名获取上传模式
	if data.Mode == 1 {
		// 新建模式
		var count = 0
		for i := 0; i < len(p.Data); i++ {
			if p.Data[i].Value == env {
				count++
				break
			}
		}
		if count != 0 {
			return true, 0
		}
	} else {
		// 合并模式
		var count = 0
		// 遍历所有表获取合并表
		if len(p.Data) == 0 {
			return false, QCount
		}
		for i := 0; i < len(p.Data); i++ {
			if p.Data[i].Name == name {
				count = i
				QCount = i
				break
			}
		}

		// 根据分隔符处理面板上的数据
		var up = 0
		envList := strings.Split(p.Data[count].Value, data.Division)
		for i := 0; i < len(envList); i++ {
			if envList[i] == env {
				up++
				break
			}
		}
		if up != 0 {
			return true, 0
		}
	}
	return false, QCount
}

// CheckIPIfItNormal 校验IP是否受限
func CheckIPIfItNormal(ip string) res.ResCode {
	list, err := sqlite.GetSetting("ipCount")
	if err != nil {
		zap.L().Error(err.Error())
		return res.CodeServerBusy
	}
	if list.Value != "0" {
		// 计算此IP今日上传次数
		bol := sqlite.CheckIPCount(ip, list.Value)
		if bol != false {
			return res.CodeNumberDepletion
		}
	}

	return res.CodeSuccess
}
