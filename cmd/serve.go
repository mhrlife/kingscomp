package cmd

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.ngrok.com/ngrok"
	ngrokconfig "golang.ngrok.com/ngrok/config"
	"kingscomp/internal/config"
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
		logrus.WithError(err).Fatalln("couldn't connect to the redis server")
	}
	accountRepository := repository.NewAccountRedisRepository(redisClient)
	lobbyRepository := repository.NewLobbyRedisRepository(redisClient)
	questionRepository := repository.NewQuestionRedisRepository(redisClient)
	// set up app
	app := service.NewApp(
		service.NewAccountService(accountRepository),
		service.NewLobbyService(lobbyRepository),
	)

	mm := matchmaking.NewRedisMatchmaking(redisClient, lobbyRepository, questionRepository)
	gs := gameserver.NewGameServer(app)

	tg, err := telegram.NewTelegram(app, mm, gs, os.Getenv("BOT_API"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't connect to the telegram server")
	}

	go tg.Start()

	wa := webapp.NewWebApp(app, gs, ":8080")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// use ngrok if its local
	if os.Getenv("ENV") == "local" {
		listener, err := ngrok.Listen(ctx,
			ngrokconfig.HTTPEndpoint(ngrokconfig.WithDomain(os.Getenv("NGROK_DOMAIN"))),
			ngrok.WithAuthtokenFromEnv(),
		)
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't set up ngrok")
		}
		defer listener.Close()
		config.Default.WebAppAddr = "https://" + listener.Addr().String()
		logrus.WithField("ngrok_addr", config.Default.WebAppAddr).Info("local server is now online")
		logrus.WithError(wa.StartDev(listener)).Errorln("http server error")
	} else {
		wa.Start()
	}

	defer wa.Shutdown(context.Background())
	defer tg.Shutdown()

	<-ctx.Done()
	logrus.Info("shutting down the server ... please wait ...")
	<-time.After(time.Second)
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
