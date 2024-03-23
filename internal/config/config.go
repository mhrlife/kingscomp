package config

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type Config struct {
	BotToken string

	AppURL     string
	ServerAddr string

	LobbyMaxPlayer     int
	LobbyQuestionCount int
}

var Default Config

func init() {
	_ = godotenv.Load()
	Default = Config{
		BotToken: os.Getenv("BOT_TOKEN"),

		AppURL: os.Getenv("APP_URL"),

		LobbyMaxPlayer:     getInt("LOBBY_MAX_PLAYER"),
		LobbyQuestionCount: getInt("LOBBY_QUESTION_COUNT"),
		ServerAddr:         getDefault("SERVER_ADDR", ":8080"),
	}
}

func getInt(key string) int {
	num, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		logrus.WithError(err).WithField("key", key).Fatal("couldn't get env value")
	}
	return num
}
func getDefault(key, def string) string {
	k := os.Getenv(key)
	if k == "" {
		return def
	}
	return k
}
