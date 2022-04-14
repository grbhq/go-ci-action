// -*- coding: utf-8 -*-
// @Time    : 2022/4/2 14:03
// @Author  : Nuanxinqing
// @Email   : nuanxinqing@gmail.com
// @File    : code.go

package response

type ResCode int64

const (
	CodeSuccess ResCode = 2000 + iota
	CodeEnvIsNull

	CodeInvalidParam = 5000 + iota
	CodeServerBusy
	CodeInvalidRouterRequested

	CodeInvalidToken
	CodeNeedLogin
	CodeRegistrationClosed
	CodeInvalidPassword
	CodeEmailFormatError
	CodeEmailNotExist
	CodeEmailExist

	CodeEnvNameExist

	CodeConnectionTimedOut
	CodeDataError
	CodeErrorOccurredInTheRequest
	CodeStorageFailed

	CodeCheckDataNotExist
	CodeOldPassWordError
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:   "Success",
	CodeEnvIsNull: "当前面板没有变量,请直接添加",

	CodeInvalidParam:           "请求参数错误",
	CodeServerBusy:             "服务繁忙",
	CodeInvalidRouterRequested: "请求无效路由",

	CodeInvalidToken:       "无效的Token",
	CodeNeedLogin:          "未登录",
	CodeRegistrationClosed: "已关闭注册",
	CodeInvalidPassword:    "邮箱或密码错误",
	CodeEmailFormatError:   "邮箱格式错误",
	CodeEmailNotExist:      "邮箱不存在",
	CodeEmailExist:         "邮箱已存在",

	CodeEnvNameExist: "变量名已存在",

	CodeConnectionTimedOut:        "面板地址连接超时",
	CodeDataError:                 "面板信息有错误",
	CodeErrorOccurredInTheRequest: "请求发生错误",
	CodeStorageFailed:             "信息储存失败",

	CodeCheckDataNotExist: "查询信息为空",
	CodeOldPassWordError:  "旧密码错误",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}