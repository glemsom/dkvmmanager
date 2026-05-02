// Package config provides configuration management for DKVM Manager
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Data folder for VM storage
	DataFolder string `mapstructure:"data_folder"`

	// VMs config file (YAML with all VM metadata)
	VMsConfigFile string `mapstructure:"vms_config_file"`

	// Default reserved memory for host (MB)
	ReservedMemMB int `mapstructure:"reserved_mem_mb"`

	// BIOS paths
	BIOSCode string `mapstructure:"bios_code"`
	BIOSVars string `mapstructure:"bios_vars"`

	// Network bridge
	NetworkBridge string `mapstructure:"network_bridge"`

	// QEMU binary path
	QEMUPath string `mapstructure:"qemu_path"`

	// TPM binary path
	TPMBinary string `mapstructure:"tpm_binary"`

	// Log file path
	LogFile string `mapstructure:"log_file"`

	// Grub configuration file path (for vfio-pci.ids kernel parameter)
	GrubConfigPath string `mapstructure:"grub_config_path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DataFolder:    "/media/dkvmdata",
		VMsConfigFile: "/media/dkvmdata/dkvmmanager/config.yaml",
		ReservedMemMB: 4096,
		BIOSCode:      "/usr/share/OVMF/OVMF_CODE.fd",
		BIOSVars:      "/usr/share/OVMF/OVMF_VARS.fd",
		NetworkBridge: "br0",
		QEMUPath:       "/usr/bin/qemu-system-x86_64",
		TPMBinary:      "/usr/bin/swtpm",
		LogFile:        "/var/log/dkvm.log",
		GrubConfigPath: "/media/usb/boot/grub/grub.cfg",
	}
}

// Load loads the configuration from file or returns defaults
func Load() *Config {
	// Try to load from config file
	viper.SetConfigName(".dkvmmanager")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc/dkvm")

	// Set defaults
	cfg := DefaultConfig()
	viper.SetDefault("data_folder", cfg.DataFolder)
	viper.SetDefault("vms_config_file", cfg.VMsConfigFile)
	viper.SetDefault("reserved_mem_mb", cfg.ReservedMemMB)
	viper.SetDefault("bios_code", cfg.BIOSCode)
	viper.SetDefault("bios_vars", cfg.BIOSVars)
	viper.SetDefault("network_bridge", cfg.NetworkBridge)
	viper.SetDefault("qemu_path", cfg.QEMUPath)
	viper.SetDefault("tpm_binary", cfg.TPMBinary)
	viper.SetDefault("log_file", cfg.LogFile)
	viper.SetDefault("grub_config_path", cfg.GrubConfigPath)

	// Read config
	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(cfg); err != nil {
			// Unmarshal failed, defaults will be used
			_ = err
		}
	}

	// Expand environment variables in paths
	cfg.DataFolder = expandPath(cfg.DataFolder)
	cfg.VMsConfigFile = expandPath(cfg.VMsConfigFile)
	cfg.BIOSCode = expandPath(cfg.BIOSCode)
	cfg.BIOSVars = expandPath(cfg.BIOSVars)
	cfg.LogFile = expandPath(cfg.LogFile)

	return cfg
}

// expandPath expands ~ and environment variables in paths
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	} else if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}

	// TODO: Expand environment variables

	return path
}

// Save saves the configuration to file
func (c *Config) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".dkvmmanager.yaml")

	viper.Set("data_folder", c.DataFolder)
	viper.Set("vms_config_file", c.VMsConfigFile)
	viper.Set("reserved_mem_mb", c.ReservedMemMB)
	viper.Set("bios_code", c.BIOSCode)
	viper.Set("bios_vars", c.BIOSVars)
	viper.Set("network_bridge", c.NetworkBridge)
	viper.Set("qemu_path", c.QEMUPath)
	viper.Set("tpm_binary", c.TPMBinary)
	viper.Set("log_file", c.LogFile)
	viper.Set("grub_config_path", c.GrubConfigPath)

	return viper.WriteConfigAs(configPath)
}
