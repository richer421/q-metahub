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

	// 迁移 deploy_plans 字段：instance_config_id -> instance_oam_id
	if db.Migrator().HasTable("deploy_plans") &&
		db.Migrator().HasColumn("deploy_plans", "instance_config_id") &&
		!db.Migrator().HasColumn("deploy_plans", "instance_oam_id") {
		if err := db.Migrator().RenameColumn("deploy_plans", "instance_config_id", "instance_oam_id"); err != nil {
			return fmt.Errorf("rename deploy_plans.instance_config_id to instance_oam_id failed: %w", err)
		}
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.Project{},
		&model.BusinessUnit{},
		&model.CIConfig{},
		&model.CDConfig{},
		&model.InstanceOAM{},
		&model.DeployPlan{},
		&model.Dependency{},
		&model.DependencyBinding{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	// 移除已废弃的旧实例配置表（instance_configs）。
	if db.Migrator().HasTable("instance_configs") {
		if err := db.Migrator().DropTable("instance_configs"); err != nil {
			return fmt.Errorf("drop deprecated table instance_configs failed: %w", err)
		}
	}

	// instance_oams 仅持久化 oam_application，清理已废弃的 frontend_payload 列。
	if db.Migrator().HasTable("instance_oams") && db.Migrator().HasColumn("instance_oams", "frontend_payload") {
		if err := db.Migrator().DropColumn("instance_oams", "frontend_payload"); err != nil {
			return fmt.Errorf("drop deprecated column instance_oams.frontend_payload failed: %w", err)
		}
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
