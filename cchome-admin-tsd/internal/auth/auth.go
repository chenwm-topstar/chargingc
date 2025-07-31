package auth

import (
	"fmt"

	"github.com/astaxie/beego/context"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/lib"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdmodels"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

func GetSidebar(ctx *context.Context, user models.Manager, fixedPage string, refererUrl string) (menu, nav string, err error) {
	var selected, referer map[string]interface{}
	rules := models.GetRules(user)
	newRules := []map[string]interface{}{}
	for _, v := range rules {
		m := lib.Struct2Map(v)
		m["icon"] = v.Icon + " fa-fw"
		m["py"] = ""
		m["pinyin"] = ""
		m["url"] = "/" + v.Name
		delete(m, "Base")
		newRules = append(newRules, m)
		if v.Name == fixedPage {
			selected = m
		}
		if v.Name == refererUrl {
			referer = m
		}
	}

	if referer["url"] == selected["url"] {
		referer = nil
	}

	m := NewTree()
	m.Init(newRules)
	menu = m.GetTreeMenu(uuid.ID(0), `<li class="@class"><a href="@url@addtabs" addtabs="@id" url="@url" py="@py" pinyin="@pinyin"><i class="@icon"></i> <span>@title</span> <span class="pull-right-container">@caret @badge</span></a> @childlist</li>`, []interface{}{selected["id"]}, []interface{}{}, "ul", `class="treeview-menu"`, 0)
	if selected != nil {
		class := ""
		if referer == nil {
			class = "active"
		}
		nav = fmt.Sprintf(`<li role="presentation" id="tab_%v" class="%s"><a href="#con_%v" node-id="%v" aria-controls="%v" role="tab" data-toggle="tab"><i class="%v fa-fw"></i> <span>%v</span> </a></li>`, selected["id"], class, selected["id"], selected["id"], selected["id"], selected["icon"], selected["title"])
	}
	if referer != nil {
		nav = nav + fmt.Sprintf(`<li role="presentation" id="tab_%v" class="active"><a href="#con_%v" node-id="%v" aria-controls="%v" role="tab" data-toggle="tab"><i class="%v fa-fw"></i> <span>%v</span></a> <i class="close-tab fa fa-remove"></i></li>`, referer["id"], referer["id"], referer["id"], referer["id"], referer["icon"], referer["title"])
	}

	return
}
