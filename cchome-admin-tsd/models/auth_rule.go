package models

import "github.com/chenwm-topstar/utils/uuid"

type KindAuthRuleStatus int32

const (
	KindAuthRuleStatusHide   KindAuthRuleStatus = 0
	KindAuthRuleStatusNormal KindAuthRuleStatus = 1
)

type AuthRule struct {
	ID        uuid.ID            `gorm:"-" json:"id"`        //规则ID
	PID       uuid.ID            `gorm:"-" json:"pid"`       //父节点ID
	Name      string             `gorm:"-" json:"name"`      //名称
	Title     string             `gorm:"-" json:"title"`     //标题
	Icon      string             `gorm:"-" json:"icon"`      //图标
	Condition string             `gorm:"-" json:"condition"` //条件
	Remark    string             `gorm:"-" json:"remark"`    //备注
	IsMenu    bool               `gorm:"-" json:"is_menu"`   //true为菜单false为权限节点
	Weight    int                `gorm:"-" json:"weight"`    //权重，用于排序
	Status    KindAuthRuleStatus `gorm:"-" json:"status"`    //状态
}

var rules []AuthRule

func init() {
	rules = append(rules, AuthRule{ID: 10, PID: 0, Name: "device/list", Title: "device center", Icon: "fa fa-tasks", Condition: "", Remark: "", IsMenu: true, Weight: 100, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 20, PID: 0, Name: "user/list", Title: "user center", Icon: "fa fa-suitcase", Condition: "", Remark: "", IsMenu: true, Weight: 90, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 30, PID: 0, Name: "order/list", Title: "order center", Icon: "fa fa-get-pocket", Condition: "", Remark: "", IsMenu: true, Weight: 80, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 40, PID: 0, Name: "feedback/list", Title: "feedback", Icon: "fa fa-get-pocket", Condition: "", Remark: "", IsMenu: true, Weight: 80, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 50, PID: 0, Name: "qaa/list", Title: "Q&A", Icon: "fa fa-get-pocket", Condition: "", Remark: "", IsMenu: true, Weight: 80, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 60, PID: 0, Name: "about", Title: "about", Icon: "fa fa-get-pocket", Condition: "", Remark: "", IsMenu: true, Weight: 80, Status: KindAuthRuleStatusNormal})
	rules = append(rules, AuthRule{ID: 70, PID: 0, Name: "dlv/list", Title: "ota version config", Icon: "fa fa-get-pocket", Condition: "", Remark: "", IsMenu: true, Weight: 80, Status: KindAuthRuleStatusNormal})
}

func GetRules(user Manager) []AuthRule {
	return rules
}
