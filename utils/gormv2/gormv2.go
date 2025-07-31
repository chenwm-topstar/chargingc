package gormv2

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cchome-admin-topstar/utils/lg"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type Base struct {
	CreatedAt time.Time      `gorm:"index"`
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	isExists bool `gorm:"-"`
}

func (b *Base) AfterFind(tx *gorm.DB) (err error) {
	// todo: 批量查询的时候，可能有问题，需要验证
	if tx.Error != gorm.ErrRecordNotFound {
		b.isExists = true
	}
	return
}

// IsNew return true 表示是一条新数据
func (b *Base) IsNew() bool {
	return !b.isExists
}

// IsExists return true 表示是数据记录在数据库已经存在
func (b *Base) IsExists() bool {
	return b.isExists
}
func (b *Base) SetExists() {
	b.isExists = true
	return
}
func (b *Base) SetNew() {
	b.isExists = false
	return
}
func AutoMigrate(dst ...interface{}) error {
	return DB.AutoMigrate(dst...)
}

// todo: 后续使用选项模式，加入其他语法解析

func Debug() {
	DB = DB.Debug()
}
func GetDB() *gorm.DB {
	return DB
}

type IGrom interface {
	DBName() string
}

// 链接DB配置
type DBConfig struct {
	User            string
	Passwd          string
	Host            string
	Port            int
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	DBResolvers     []DBResolverConfig
}

// 多个实例或者多个库数据库连接配置
type DBResolverConfig struct {
	DBName string        // 数据库名称
	Models []interface{} // 数据库下面的表对象
}

type InitAfterCB func(context.Context) error

var DB *gorm.DB
var registerModels []IGrom
var initAftercbs []InitAfterCB

func RegisterAfterCBS(fs ...InitAfterCB) {
	initAftercbs = append(initAftercbs, fs...)
}

// RegisterModel 将所有要使用的models注册到这里，
// 当初始化时，会根据这里的model对象以及配置文件的对应关系，
// 设置成， 不同的model访问不同的DBInstance
func RegisterModel(models ...IGrom) {
	registerModels = append(registerModels, models...)
}

func Init(autoMigrate ...bool) (err error) {
	dbconf := DBConfig{
		Host:            mysqlHost(),
		Port:            mysqlPort(),
		User:            mysqlUser(),
		Passwd:          mysqlPasswd(),
		MaxIdleConns:    mysqlMaxIdleConns(),
		MaxOpenConns:    mysqlMaxOpenConns(),
		ConnMaxLifetime: mysqlConnMaxLifetime(),
		DBResolvers:     parseRegisterModels(),
	}

	if err = ManualInit(dbconf); err != nil {
		return
	}
	if len(autoMigrate) > 0 && autoMigrate[0] {
		var tmps []interface{}
		for i := range registerModels {
			tmps = append(tmps, registerModels[i])
		}
		lg.Infof("mysql auto migrate start...")
		if err = DB.AutoMigrate(tmps...); err != nil {
			return
		}
		lg.Infof("mysql auto migrate success...")

		for _, cb := range initAftercbs {
			if err = cb(context.TODO()); err != nil {
				return err
			}
		}
	}

	return
}

// Init 初始化多个实例数据库的时候使用
func ManualInit(conf DBConfig) (err error) {
	defDBName := ""
	if len(conf.DBResolvers) > 0 {
		defDBName = conf.DBResolvers[0].DBName
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: 1 * time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Silent,   // Log level
			Colorful:      false,           // 禁用彩色打印
		},
	)

	lg.Infof("init mysql start...")

	// 设置第一个库，当称一个数据源，来初始化db对象
	db, err := gorm.Open(mysql.Open(getdns(conf, defDBName)), &gorm.Config{Logger: newLogger})
	if err != nil {
		return err
	}

	// 设置其他库的数据源
	for i := 1; i < len(conf.DBResolvers); i++ {
		dbconfig := dbresolver.Config{
			Policy:  dbresolver.RandomPolicy{},
			Sources: []gorm.Dialector{mysql.Open(getdns(conf, conf.DBResolvers[i].DBName))},
		}
		db.Use(dbresolver.Register(dbconfig, conf.DBResolvers[i].Models...))
	}

	DB = db
	sqlDB, _ := db.DB()

	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	// sqlDB.SetConnMaxIdleTime(5 * time.Minute) // go v1.14 不支持这个方法，编译机无法编译通过, 后续支持
	sqlDB.SetConnMaxLifetime(conf.ConnMaxLifetime)

	lg.Infof("init mysql success")

	return nil
}

func getdns(dbconf DBConfig, DBName string) string {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbconf.User, dbconf.Passwd, dbconf.Host, dbconf.Port, DBName)
	condition := "timeout=60s&parseTime=true&charset=utf8mb4,utf8&loc=Local"
	if strings.Contains(dns, "?") {
		dns = dns + "&" + condition
	} else {
		dns = dns + "?" + condition
	}
	return dns
}

func parseRegisterModels() (rets []DBResolverConfig) {
_CONTINUE_DB:
	for i := range registerModels {
		for j := range rets {
			if rets[j].DBName == registerModels[i].DBName() {
				rets[j].Models = append(rets[j].Models, registerModels[i])
				continue _CONTINUE_DB
			}
		}
		rets = append(rets, DBResolverConfig{
			DBName: registerModels[i].DBName(),
			Models: []interface{}{registerModels[i]},
		})
	}
	return
}
