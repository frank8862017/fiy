package business

import (
	"fiy/app/cmdb/models/business"
	orm "fiy/common/global"
	"fiy/common/models"
	"fiy/common/pagination"
	"fiy/tools"
	"fiy/tools/app"

	"github.com/gin-gonic/gin"
)

/*
  @Author : lanyulei
*/

// 服务模板列表
func ServiceTemplateList(c *gin.Context) {
	var (
		err    error
		result interface{}
		list   []struct {
			business.ServiceTemplate
			SvcClassifyName string `json:"svc_classify_name"`
			ModifyName      string `json:"modify_name"`
		}
	)

	SearchParams := map[string]map[string]interface{}{
		"like": pagination.RequestParams(c),
	}

	db := orm.Eloquent.Model(&business.ServiceTemplate{}).
		Joins("left join cmdb_business_svc_classify as sc on sc.id = cmdb_business_svc_tpl.svc_classify").
		Joins("left join sys_user on sys_user.user_id = cmdb_business_svc_tpl.modifier").
		Select("cmdb_business_svc_tpl.*, sc.name as svc_classify_name, sys_user.nick_name as modify_name")

	result, err = pagination.Paging(&pagination.Param{
		C:  c,
		DB: db,
	}, &list, SearchParams, "cmdb_business_svc_tpl")
	if err != nil {
		app.Error(c, -1, err, "分页查询云账号失败")
		return
	}

	app.OK(c, result, "")
}

// 新建服务模版
func CreateServiceTemplate(c *gin.Context) {
	var (
		err    error
		params struct {
			business.ServiceTemplate
			ProcessList []*business.ServiceTemplateProcess `json:"process_list"`
		}
	)

	err = c.ShouldBind(&params)
	if err != nil {
		app.Error(c, -1, err, "参数绑定失败")
		return
	}

	tx := orm.Eloquent.Begin()

	// 新建服务模版
	currentUser := tools.GetUserId(c)
	svcData := business.ServiceTemplate{
		Name:        params.Name,
		SvcClassify: params.SvcClassify,
		Creator:     currentUser,
		Modifier:    currentUser,
		BaseModel:   models.BaseModel{},
	}
	err = tx.Create(&svcData).Error
	if err != nil {
		tx.Rollback()
		app.Error(c, -1, err, "新建服务模板失败")
		return
	}

	// 新建服务模板的进程
	for _, p := range params.ProcessList {
		p.SvcTpl = svcData.Id
	}
	err = tx.Create(&params.ProcessList).Error
	if err != nil {
		tx.Rollback()
		app.Error(c, -1, err, "新建服务模版失败")
		return
	}

	tx.Commit()

	app.OK(c, nil, "")
}

// 服务模板详情
func ServiceTemplateDetails(c *gin.Context) {
	var (
		err  error
		id   string
		info struct {
			business.ServiceTemplate
			ProcessList []*business.ServiceTemplateProcess `json:"process_list"`
		}
	)

	id = c.Param("id")

	// 查询服务模板
	err = orm.Eloquent.Model(&business.ServiceTemplate{}).Where("id = ?", id).Find(&info).Error
	if err != nil {
		app.Error(c, -1, err, "查询服务模板失败")
		return
	}

	// 查询服务模板进程
	err = orm.Eloquent.Model(&business.ServiceTemplateProcess{}).
		Where("svc_tpl = ?", id).
		Find(&info.ProcessList).Error
	if err != nil {
		app.Error(c, -1, err, "查询服务模板进程列表失败")
		return
	}

	app.OK(c, info, "")
}