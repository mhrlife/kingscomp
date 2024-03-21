package cmd

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"kingscomp/internal/service"
	"kingscomp/internal/telegram"
	"kingscomp/internal/webapp"
	"os"
	"os/signal"
	"time"
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
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	questionRepository := repository.NewQuestionRedisRepository(redisClient)
	// set up app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
		service.NewLobbyService(lobbyRepository),
	)

	mm := matchmaking.NewRedisMatchmaking(redisClient, lobbyRepository, questionRepository, accountRepository)
	gs := gameserver.NewGameServer(app, gameserver.DefaultGameServerConfig())

	tg, err := telegram.NewTelegram(app, mm, gs, os.Getenv("BOT_API"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to the telegram server")
	}

	go tg.Start()

	wa := webapp.NewWebApp(app, gs, ":8080")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if os.Getenv("ENV") == "local" {
		go func() {
			logrus.WithError(wa.StartDev()).Errorln("http server error")
		}()
	} else {
		go func() {
			logrus.WithError(wa.Start()).Errorln("http server error")
		}()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer wa.Shutdown(shutdownCtx)
	defer tg.Shutdown()

	logrus.Info("server is up and running")
	<-ctx.Done()
	logrus.Info("shutting down the server ... please wait ...")
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
