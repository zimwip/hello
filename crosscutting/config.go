package crosscutting

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type config struct {
	isDev  bool
	env    string
	change chan bool
	mu     sync.RWMutex
}

var (
	c    *config
	once sync.Once
)

func Config() *config {
	once.Do(func() {
		fmt.Println("Application Configuration initialisation")
		change := make(chan bool)
		env := os.Getenv("ENVIRONMENT")
		isDev := env == "DEV"
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
			c.change <- true
		})
		c = &config{
			change: change,
			isDev:  isDev,
			env:    env,
		}
	})
	return c
}

func (c *config) GetString(key string) string {
	return viper.GetString(key)
}

func (c *config) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (c *config) IsDev() bool {
	return c.isDev
}
