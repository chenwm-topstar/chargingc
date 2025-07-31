package autoupgrade

import (
	"context"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/internal/evsectl"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/sirupsen/logrus"
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
