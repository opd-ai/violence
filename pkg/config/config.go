// Package config handles loading and storing game configuration.
package config

import (
	"errors"

	"github.com/spf13/viper"
)

// Config holds all game configuration values.
type Config struct {
	WindowWidth      int            `mapstructure:"WindowWidth"`
	WindowHeight     int            `mapstructure:"WindowHeight"`
	InternalWidth    int            `mapstructure:"InternalWidth"`
	InternalHeight   int            `mapstructure:"InternalHeight"`
	FOV              float64        `mapstructure:"FOV"`
	MouseSensitivity float64        `mapstructure:"MouseSensitivity"`
	MasterVolume     float64        `mapstructure:"MasterVolume"`
	MusicVolume      float64        `mapstructure:"MusicVolume"`
	SFXVolume        float64        `mapstructure:"SFXVolume"`
	DefaultGenre     string         `mapstructure:"DefaultGenre"`
	VSync            bool           `mapstructure:"VSync"`
	FullScreen       bool           `mapstructure:"FullScreen"`
	KeyBindings      map[string]int `mapstructure:"KeyBindings"`
}

// C is the global configuration instance.
var C Config

// Load reads configuration from file and environment, populating C.
func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.violence")

	viper.SetDefault("WindowWidth", 960)
	viper.SetDefault("WindowHeight", 600)
	viper.SetDefault("InternalWidth", 320)
	viper.SetDefault("InternalHeight", 200)
	viper.SetDefault("FOV", 66.0)
	viper.SetDefault("MouseSensitivity", 1.0)
	viper.SetDefault("MasterVolume", 0.8)
	viper.SetDefault("MusicVolume", 0.7)
	viper.SetDefault("SFXVolume", 0.8)
	viper.SetDefault("DefaultGenre", "fantasy")
	viper.SetDefault("VSync", true)
	viper.SetDefault("FullScreen", false)
	viper.SetDefault("KeyBindings", map[string]int{})

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return err
		}
	}

	return viper.Unmarshal(&C)
}

// Save writes the current configuration to file.
func Save() error {
	viper.Set("WindowWidth", C.WindowWidth)
	viper.Set("WindowHeight", C.WindowHeight)
	viper.Set("InternalWidth", C.InternalWidth)
	viper.Set("InternalHeight", C.InternalHeight)
	viper.Set("FOV", C.FOV)
	viper.Set("MouseSensitivity", C.MouseSensitivity)
	viper.Set("MasterVolume", C.MasterVolume)
	viper.Set("MusicVolume", C.MusicVolume)
	viper.Set("SFXVolume", C.SFXVolume)
	viper.Set("DefaultGenre", C.DefaultGenre)
	viper.Set("VSync", C.VSync)
	viper.Set("FullScreen", C.FullScreen)
	viper.Set("KeyBindings", C.KeyBindings)

	return viper.WriteConfig()
}
