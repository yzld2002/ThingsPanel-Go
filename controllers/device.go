package controllers

import (
	"ThingsPanel-Go/initialize/psql"
	gvalid "ThingsPanel-Go/initialize/validate"
	"ThingsPanel-Go/models"
	"ThingsPanel-Go/services"
	response "ThingsPanel-Go/utils"
	uuid "ThingsPanel-Go/utils"
	valid "ThingsPanel-Go/validate"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cm "ThingsPanel-Go/modules/dataService/mqtt"

	"github.com/beego/beego/v2/core/validation"
	beego "github.com/beego/beego/v2/server/web"
	context2 "github.com/beego/beego/v2/server/web/context"
	"gorm.io/gorm"
)

type DeviceController struct {
	beego.Controller
}

// 设备列表
func (this *DeviceController) Index() {
	this.Data["json"] = "test devices"
	this.ServeJSON()
}

// 设备列表
func (this *DeviceController) Edit() {
	editDeviceValidate := valid.EditDevice{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &editDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(editDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			// 获取字段别称
			alias := gvalid.GetAlias(editDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(this.Ctx))
			break
		}
		return
	}
	var DeviceService services.DeviceService
	f := DeviceService.Edit(editDeviceValidate.ID, editDeviceValidate.Token, editDeviceValidate.Protocol, editDeviceValidate.Port, editDeviceValidate.Publish, editDeviceValidate.Subscribe, editDeviceValidate.Username, editDeviceValidate.Password)
	if f {
		response.SuccessWithMessage(200, "编辑成功", (*context2.Context)(this.Ctx))
		return
	}
	response.SuccessWithMessage(400, "编辑失败", (*context2.Context)(this.Ctx))
	return
}

func (this *DeviceController) Add() {
	addDeviceValidate := valid.AddDevice{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &addDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(addDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(addDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(this.Ctx))
			break
		}
		return
	}
	var DeviceService services.DeviceService
	f, _ := DeviceService.Add(
		addDeviceValidate.Token,
		addDeviceValidate.Protocol,
		addDeviceValidate.Port,
		addDeviceValidate.Publish,
		addDeviceValidate.Subscribe,
		addDeviceValidate.Username,
		addDeviceValidate.Password,
	)
	if f {
		response.SuccessWithMessage(200, "添加成功", (*context2.Context)(this.Ctx))
		return
	}
	response.SuccessWithMessage(400, "添加失败", (*context2.Context)(this.Ctx))
	return
}

type DeviceDash struct {
	ID             string            `json:"id" gorm:"primaryKey,size:36"`
	AssetID        string            `json:"asset_id" gorm:"size:36"`              // 资产id
	Token          string            `json:"token"`                                // 安全key
	AdditionalInfo string            `json:"additional_info" gorm:"type:longtext"` // 存储基本配置
	CustomerID     string            `json:"customer_id" gorm:"size:36"`
	Type           string            `json:"type"` // 插件类型
	Name           string            `json:"name"` // 插件名
	Label          string            `json:"label"`
	SearchText     string            `json:"search_text"`
	Extension      string            `json:"extension" gorm:"size:50"` // 插件( 目录名)
	Protocol       string            `json:"protocol" gorm:"size:50"`
	Port           string            `json:"port" gorm:"size:50"`
	Publish        string            `json:"publish" gorm:"size:255"`
	Subscribe      string            `json:"subscribe" gorm:"size:255"`
	Username       string            `json:"username" gorm:"size:255"`
	Password       string            `json:"password" gorm:"size:255"`
	Dash           []services.Widget `json:"dash"`
}

func (reqDate *DeviceController) AddOnly() {
	addDeviceValidate := valid.Device{}
	err := json.Unmarshal(reqDate.Ctx.Input.RequestBody, &addDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(addDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(addDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(reqDate.Ctx))
			break
		}
		return
	}
	var uuid = uuid.GetUuid()
	var AssetService services.AssetService
	var ResWidgetData []services.Widget
	if addDeviceValidate.Type != "" {
		dd := AssetService.Widget(addDeviceValidate.Type)
		if len(dd) > 0 {
			for _, wv := range dd {
				ResWidgetData = append(ResWidgetData, wv)
			}
		}
	}
	deviceData := models.Device{
		ID:        uuid,
		AssetID:   addDeviceValidate.AssetID,
		Token:     addDeviceValidate.Token,
		Type:      addDeviceValidate.Type,
		Name:      addDeviceValidate.Name,
		Extension: addDeviceValidate.Extension,
		Protocol:  addDeviceValidate.Protocol,
	}

	result := psql.Mydb.Create(&deviceData)
	if result.Error == nil {
		deviceDash := DeviceDash{
			ID:        uuid,
			AssetID:   addDeviceValidate.AssetID,
			Token:     addDeviceValidate.Token,
			Type:      addDeviceValidate.Type,
			Name:      addDeviceValidate.Name,
			Extension: addDeviceValidate.Extension,
			Protocol:  addDeviceValidate.Protocol,
			Dash:      ResWidgetData,
		}
		response.SuccessWithDetailed(200, "success", deviceDash, map[string]string{}, (*context2.Context)(reqDate.Ctx))
	} else {
		errors.Is(result.Error, gorm.ErrRecordNotFound)
		response.SuccessWithMessage(400, "添加失败", (*context2.Context)(reqDate.Ctx))
	}
}

func (reqDate *DeviceController) UpdateOnly() {
	addDeviceValidate := valid.Device{}
	err := json.Unmarshal(reqDate.Ctx.Input.RequestBody, &addDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(addDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(addDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(reqDate.Ctx))
			break
		}
		return
	}
	var AssetService services.AssetService
	var ResWidgetData []services.Widget
	if addDeviceValidate.Type != "" {
		dd := AssetService.Widget(addDeviceValidate.Type)
		if len(dd) > 0 {
			for _, wv := range dd {
				ResWidgetData = append(ResWidgetData, wv)
			}
		}
	}
	deviceData := models.Device{
		ID:        addDeviceValidate.ID,
		AssetID:   addDeviceValidate.AssetID,
		Token:     addDeviceValidate.Token,
		Type:      addDeviceValidate.Type,
		Name:      addDeviceValidate.Name,
		Extension: addDeviceValidate.Extension,
		Protocol:  addDeviceValidate.Protocol,
	}
	result := psql.Mydb.Updates(&deviceData)
	if result.Error == nil {
		deviceDash := DeviceDash{
			ID:        addDeviceValidate.ID,
			AssetID:   addDeviceValidate.AssetID,
			Token:     addDeviceValidate.Token,
			Type:      addDeviceValidate.Type,
			Name:      addDeviceValidate.Name,
			Extension: addDeviceValidate.Extension,
			Protocol:  addDeviceValidate.Protocol,
			Dash:      ResWidgetData,
		}
		response.SuccessWithDetailed(200, "success", deviceDash, map[string]string{}, (*context2.Context)(reqDate.Ctx))
	} else {
		errors.Is(result.Error, gorm.ErrRecordNotFound)
		response.SuccessWithMessage(400, "修改失败", (*context2.Context)(reqDate.Ctx))
	}
}

// 扫码激活设备
func (this *DeviceController) Scan() {
	this.Data["json"] = "Scan success"
	this.ServeJSON()
}

// 获取设备token
func (this *DeviceController) Token() {
	tokenDeviceValidate := valid.TokenDevice{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &tokenDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(tokenDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			// 获取字段别称
			alias := gvalid.GetAlias(tokenDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(this.Ctx))
			break
		}
		return
	}
	var DeviceService services.DeviceService
	d, c := DeviceService.Token(tokenDeviceValidate.ID)
	if c != 0 {
		response.SuccessWithDetailed(200, "获取成功", d.Token, map[string]string{}, (*context2.Context)(this.Ctx))
		return
	}
	response.SuccessWithMessage(400, "设备不存在", (*context2.Context)(this.Ctx))
	return
}

// 删除
func (this *DeviceController) Delete() {
	deleteDeviceValidate := valid.DeleteDevice{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &deleteDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(deleteDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(deleteDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(this.Ctx))
			break
		}
		return
	}
	var DeviceService services.DeviceService
	f := DeviceService.Delete(deleteDeviceValidate.ID)
	if f {
		response.SuccessWithMessage(200, "删除成功", (*context2.Context)(this.Ctx))
		return
	}
	response.SuccessWithMessage(400, "删除失败", (*context2.Context)(this.Ctx))
	return
}

// 获取配置参数
func (this *DeviceController) Configure() {
	configureDeviceValidate := valid.ConfigureDevice{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &configureDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(configureDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(configureDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(this.Ctx))
			break
		}
		return
	}
	//var DeviceService services.DeviceService
	//DeviceService
}

//控制设备
func (request *DeviceController) Operating() {
	operatingDeviceValidate := valid.OperatingDevice{}
	err := json.Unmarshal(request.Ctx.Input.RequestBody, &operatingDeviceValidate)
	if err != nil {
		fmt.Println("参数解析失败", err.Error())
	}
	v := validation.Validation{}
	status, _ := v.Valid(operatingDeviceValidate)
	if !status {
		for _, err := range v.Errors {
			alias := gvalid.GetAlias(operatingDeviceValidate, err.Field)
			message := strings.Replace(err.Message, err.Field, alias, 1)
			response.SuccessWithMessage(1000, message, (*context2.Context)(request.Ctx))
			break
		}
		return
	}
	f := cm.Send(request.Ctx.Input.RequestBody)
	if f == nil {
		response.SuccessWithMessage(200, "发送成功", (*context2.Context)(request.Ctx))
		return
	}
	response.SuccessWithMessage(400, f.Error(), (*context2.Context)(request.Ctx))
	return
}
