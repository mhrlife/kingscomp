package cmd

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"
	"kingscomp/internal/config"
	"kingscomp/internal/events"
	"kingscomp/internal/gameserver"
	"kingscomp/internal/matchmaking"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"kingscomp/internal/scoreboard"
	"kingscomp/internal/service"
	"kingscomp/internal/telegram"
	"kingscomp/internal/telegram/teleprompt"
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
	if os.Getenv("ENV") != "local" {
		logrus.SetLevel(logrus.ErrorLevel)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// set up repositories
	redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to the redis server")
	}
	accountRepository := repository.NewAccountRedisRepository(redisClient)
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	questionRepository := repository.NewQuestionRedisRepository(redisClient)

	eventsQueue := events.NewRedisQueue(ctx, "events", redisClient)
	// set up app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
		service.NewLobbyService(lobbyRepository),
	)

	mm := matchmaking.NewRedisMatchmaking(
		redisClient,
		lobbyRepository,
		questionRepository,
		accountRepository,
	)
	gs := gameserver.NewGameServer(
		app,
		events.NewRedisPubSub(ctx, redisClient, "lobby.*"),
		eventsQueue,
		gameserver.DefaultGameServerConfig(),
	)

	tg, err := telegram.NewTelegram(
		ctx,
		app,
		mm,
		gs,
		teleprompt.NewTelePrompt(ctx, redisClient),
		scoreboard.NewScoreboard(redisClient),
		config.Default.BotToken,
	)
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to the telegram server")
	}

	go tg.Start()

	wa := webapp.NewWebApp(app, gs, config.Default.ServerAddr, tg.Bot)

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
