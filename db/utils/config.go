package utils

import "github.com/spf13/viper"

// Config stored all configuration of the application
// The values are read by viper from a config file or environment variable
// we add mapstructure to allow for marshalling to be done when viper read this config file
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig reads configuration from file or environment variables
// The path is where the env variables are located
// It will read the configs inside the path, if it exists, or override their values
// with env variables if provided
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  // this is to tell Viper the location of the config file
	viper.SetConfigName("app") // this we tell Viper to look for a config with a specific name
	// ours it's app.env so the name is app
	viper.SetConfigType("env") // This is to tell Viper the type of the config file, which is env for our case
	// It can also be JSON, XML, ...

	viper.AutomaticEnv() // we use this so tha Viper can read values from env variables
	// It will automatically override values that it has read from config file with the values  of the
	// corresponding environment variables if they exist

	// Now we start reading config values
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// Now we can unmarshall the values into the target config object
	err = viper.Unmarshal(&config)
	return

}
