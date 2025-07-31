package autoupgrade

import (
	"context"

	"github.com/sirupsen/logrus"
	"gitlab.goiot.net/chargingc/cchome-admin/internal/evsectl"
	"gitlab.goiot.net/chargingc/cchome-admin/models"
	"gitlab.goiot.net/chargingc/utils/gormv2"
)

func ListenEvseAutoUpgrade() {
	for {
		autoUpgrade := <-models.OTACH

		err := evsectl.Upgrade(autoUpgrade.SN, autoUpgrade.Address)
		if err != nil {
			autoUpgrade.Result = err.Error()
		} else {
			autoUpgrade.IsUpgrade = true
		}

		if err := gormv2.Model(context.Background(), &autoUpgrade).Updates(map[string]interface{}{
			"result":     autoUpgrade.Result,
			"is_upgrade": autoUpgrade.IsUpgrade,
		}).Error; err != nil {
			logrus.Warnf("update autoupgrade record fail: " + err.Error())
		}
	}
}
