// Developer: zeelrupapara@gmail.com
// Description: Database seeding for GreenLync API Gateway Boilerplate
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"greenlync-api-gateway/config"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/db"
	"greenlync-api-gateway/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

var (
	username = flag.String("user", "admin", "Admin username")
	password = flag.String("password", "", "Admin password (will prompt if empty)")
)

func main() {
	flag.Parse()

	if *password == "" {
		fmt.Print("Enter admin password: ")
		fmt.Scanln(password)
		if *password == "" {
			log.Fatal("Password cannot be empty")
		}
	}

	// Initialize config
	cfg := config.NewConfig()

	// Initialize logger
	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	dbSess, err := db.NewMysqDB(cfg)
	if err != nil {
		log.Logger.Fatalf("Failed to connect to database: %v", err)
	}

	log.Logger.Info("Starting database seeding...")

	// Migrate tables
	err = dbSess.Migrate()
	if err != nil {
		log.Logger.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed default admin user
	err = seedAdminUser(dbSess, *username, *password)
	if err != nil {
		log.Logger.Fatalf("Failed to seed admin user: %v", err)
	}

	// Seed default roles and permissions
	err = seedRolesAndPermissions(dbSess)
	if err != nil {
		log.Logger.Fatalf("Failed to seed roles and permissions: %v", err)
	}

	// Seed default configuration
	err = seedDefaultConfig(dbSess)
	if err != nil {
		log.Logger.Fatalf("Failed to seed default configuration: %v", err)
	}

	// Seed Casbin policies
	err = seedCasbinPolicies(dbSess)
	if err != nil {
		log.Logger.Fatalf("Failed to seed Casbin policies: %v", err)
	}

	log.Logger.Info("Database seeding completed successfully!")
}

func seedAdminUser(dbSess *db.MysqlDB, username, password string) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user using GORM
	user := &model.User{
		Username:     username,
		Email:        username + "@greenlync.com",
		FirstName:    "Admin",
		LastName:     "User",
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		IsActive:     true,
	}

	// Check if user already exists
	var existingUser model.User
	err = dbSess.DB.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		// User exists, update password
		existingUser.PasswordHash = string(hashedPassword)
		err = dbSess.DB.Save(&existingUser).Error
		if err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}
		fmt.Printf("Admin user '%s' updated successfully\n", username)
	} else {
		// User doesn't exist, create new
		err = dbSess.DB.Create(user).Error
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Printf("Admin user '%s' created successfully\n", username)
	}

	return nil
}

func seedRolesAndPermissions(dbSess *db.MysqlDB) error {
	// Create default roles
	roles := []struct {
		roleType    model.RoleType
		description string
		status      string
	}{
		{model.RoleType_Admin, "System Administrator", "active"},
		{model.RoleType_Dispensary, "Basic User Access", "active"},
	}

	for _, roleData := range roles {
		role := &model.Role{
			RoleType: roleData.roleType,
			Desc:     roleData.description,
			Original: true,
			Status:   roleData.status,
		}

		// Check if role already exists
		var existingRole model.Role
		err := dbSess.DB.Where("role_type = ?", roleData.roleType).First(&existingRole).Error
		if err != nil {
			// Role doesn't exist, create new
			err = dbSess.DB.Create(role).Error
			if err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleData.description, err)
			}
			fmt.Printf("Role '%s' created successfully\n", roleData.description)
		} else {
			// Role exists, update description
			existingRole.Desc = roleData.description
			existingRole.Status = roleData.status
			err = dbSess.DB.Save(&existingRole).Error
			if err != nil {
				return fmt.Errorf("failed to update role %s: %w", roleData.description, err)
			}
			fmt.Printf("Role '%s' updated successfully\n", roleData.description)
		}
	}

	// Create default permissions for boilerplate
	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		{"users.read", "Read user data", "users", "read"},
		{"users.write", "Create/update users", "users", "write"},
		{"users.delete", "Delete users", "users", "delete"},
		{"sessions.read", "Read session data", "sessions", "read"},
		{"sessions.delete", "Delete sessions", "sessions", "delete"},
		{"config.read", "Read configuration", "config", "read"},
		{"config.write", "Update configuration", "config", "write"},
		{"logs.read", "Read system logs", "logs", "read"},
	}

	for _, permData := range permissions {
		permission := &model.Permission{
			Name:        permData.name,
			Description: permData.description,
			Resource:    permData.resource,
			Action:      permData.action,
			IsActive:    true,
		}

		// Check if permission already exists
		var existingPerm model.Permission
		err := dbSess.DB.Where("name = ?", permData.name).First(&existingPerm).Error
		if err != nil {
			// Permission doesn't exist, create new
			err = dbSess.DB.Create(permission).Error
			if err != nil {
				return fmt.Errorf("failed to create permission %s: %w", permData.name, err)
			}
			fmt.Printf("Permission '%s' created successfully\n", permData.name)
		} else {
			// Permission exists, update fields
			existingPerm.Description = permData.description
			existingPerm.Resource = permData.resource
			existingPerm.Action = permData.action
			existingPerm.IsActive = true
			err = dbSess.DB.Save(&existingPerm).Error
			if err != nil {
				return fmt.Errorf("failed to update permission %s: %w", permData.name, err)
			}
			fmt.Printf("Permission '%s' updated successfully\n", permData.name)
		}
	}

	fmt.Println("Roles and permissions seeded successfully")
	return nil
}

func seedDefaultConfig(dbSess *db.MysqlDB) error {
	// Create default config groups
	configGroups := []struct {
		name        string
		description string
		isSystem    bool
	}{
		{"general", "General application settings", true},
		{"security", "Security and authentication settings", true},
		{"email", "Email service configuration", true},
	}

	for _, groupData := range configGroups {
		configGroup := &model.ConfigGroup{
			Name:        groupData.name,
			Description: groupData.description,
			IsSystem:    groupData.isSystem,
		}

		// Check if config group already exists
		var existingGroup model.ConfigGroup
		err := dbSess.DB.Where("name = ?", groupData.name).First(&existingGroup).Error
		if err != nil {
			// Config group doesn't exist, create new
			err = dbSess.DB.Create(configGroup).Error
			if err != nil {
				return fmt.Errorf("failed to create config group %s: %w", groupData.name, err)
			}
			fmt.Printf("Config group '%s' created successfully\n", groupData.name)
		} else {
			// Config group exists, update description
			existingGroup.Description = groupData.description
			err = dbSess.DB.Save(&existingGroup).Error
			if err != nil {
				return fmt.Errorf("failed to update config group %s: %w", groupData.name, err)
			}
			fmt.Printf("Config group '%s' updated successfully\n", groupData.name)
		}
	}

	// Get the general config group ID for default configs
	var generalGroup model.ConfigGroup
	err := dbSess.DB.Where("name = ?", "general").First(&generalGroup).Error
	if err != nil {
		return fmt.Errorf("failed to find general config group: %w", err)
	}

	// Create default configurations
	configs := []struct {
		key         string
		value       string
		description string
		valueType   model.ValueType
		isPublic    bool
	}{
		{"app.name", "GreenLync API Gateway", "Application name", model.ValueType_String, true},
		{"app.version", "1.0.0", "Application version", model.ValueType_String, true},
		{"auth.session_timeout", "3600", "Session timeout in seconds", model.ValueType_Integer, false},
		{"auth.max_login_attempts", "5", "Maximum login attempts before lockout", model.ValueType_Integer, false},
		{"system.maintenance_mode", "false", "System maintenance mode", model.ValueType_Boolean, false},
	}

	for _, configData := range configs {
		config := &model.Config{
			Key:           configData.key,
			Value:         configData.value,
			Description:   configData.description,
			ValueType:     configData.valueType,
			ConfigGroupId: generalGroup.Id,
			IsPublic:      configData.isPublic,
			RecordType:    model.RecordType_System,
		}

		// Check if config already exists  
		var existingConfig model.Config
		err := dbSess.DB.Where("`key` = ?", configData.key).First(&existingConfig).Error
		if err != nil {
			// Config doesn't exist, create new
			err = dbSess.DB.Create(config).Error
			if err != nil {
				return fmt.Errorf("failed to create config %s: %w", configData.key, err)
			}
			fmt.Printf("Config '%s' created successfully\n", configData.key)
		} else {
			// Config exists, update if needed
			existingConfig.Description = configData.description
			existingConfig.ValueType = configData.valueType
			existingConfig.IsPublic = configData.isPublic
			err = dbSess.DB.Save(&existingConfig).Error
			if err != nil {
				return fmt.Errorf("failed to update config %s: %w", configData.key, err)
			}
			fmt.Printf("Config '%s' updated successfully\n", configData.key)
		}
	}

	fmt.Println("Default configuration seeded successfully")
	return nil
}

func seedCasbinPolicies(dbSess *db.MysqlDB) error {
	// This function would typically integrate with Casbin to set up policies
	// For now, we'll just create a note that policies should be managed through Casbin
	fmt.Println("Casbin policies should be configured through the RBAC system")
	fmt.Println("Default policy: Admin role has access to all resources")
	fmt.Println("Default policy: User role has limited access to basic resources")
	return nil
}

