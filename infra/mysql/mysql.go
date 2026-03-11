package mysql

import (
	"fmt"

	"github.com/richer421/q-metahub/conf"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	if DB != nil {
		return nil
	}

	if conf.C.MySQL.Database == "" {
		return fmt.Errorf("database config is required")
	}

	dsn := conf.C.MySQL.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 添加 OTel 插件
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return fmt.Errorf("failed to add otelgorm plugin: %w", err)
	}

	// ���接到指定数据库
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", conf.C.MySQL.Database)).Error
	if err != nil {
		return fmt.Errorf("create database failed: %w", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.Project{},
		&model.BusinessUnit{},
		&model.CIConfig{},
		&model.CDConfig{},
		&model.InstanceConfig{},
		&model.DeployPlan{},
		&model.Dependency{},
		&model.DependencyBinding{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	// 设置 DAO 使用的 DB
	dao.SetDefault(db)

	DB = db

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return sqlDB.Close()
}
