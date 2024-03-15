package cmd

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"kingscomp/internal/service"
	"kingscomp/internal/telegram"
	"os"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the telegram bot",
	Run:   serve,
}

func serve(_ *cobra.Command, _ []string) {
	_ = godotenv.Load()
	// set up repositories
	redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to te redis server")
	}
	accountRepository := repository.NewAccountRedisRepository(redisClient)

	// set up app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
	)

	tg, err := telegram.NewTelegram(app, os.Getenv("BOT_API"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to the telegram server")
	}

	tg.Start()
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
