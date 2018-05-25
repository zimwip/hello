package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
)

var (
	isDev  bool
	env    string
	change chan bool
)

func init() {

	change = make(chan bool)
	fmt.Println("Configuration initialisation")
	env = os.Getenv("ENVIRONMENT")
	isDev = env == "DEV"
	if isDev {
		fmt.Println("Loading DEVELOPMENT environment")
	} else {
		fmt.Println("Loading PRODUCTION environment")
	}

	viper.SetConfigName("config")         // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	fmt.Printf("Application [%s] is starting\n", viper.GetString("app.name"))

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		change <- true
	})
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func IsDev() bool {
	return isDev
}
