package gameserver

import (
	"context"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
	"kingscomp/internal/service"
	"slices"
	"time"
)

type Game struct {
	Config

	app     *service.App
	server  *GameServer
	LobbyId entity.ID
	Events  *Events

	ctx    context.Context
	cancel context.CancelFunc

	lobby entity.Lobby
}

func NewGame(lobbyId string, app *service.App, server *GameServer, config Config) *Game {
	return &Game{
		Config:  config,
		LobbyId: entity.NewID("lobby", lobbyId),
		app:     app,
		server:  server,
		Events:  NewEvents(),
	}
}

func (g *Game) Start(ctx context.Context) {
	g.ctx, g.cancel = context.WithCancel(ctx)
	for {
		g.loadLobby()

		select {
		case <-g.ctx.Done():
			return
		default:
		}

		logrus.WithFields(logrus.Fields{
			"lobbyId":    g.lobby.ID,
			"lobbyState": g.lobby.State,
		}).Info("running sub-state for game")

		var err error
		switch g.lobby.State {
		case "created":
			err = g.created()
		case "get-ready":
			err = g.getReady()
		case "started":
			err = g.started()
		default:
			logrus.WithFields(logrus.Fields{
				"lobbyId": g.lobby.ID,
				"state":   g.lobby.State,
			}).Errorln("bad state, not found")
			return
		}

		if err != nil {
			logrus.WithError(err).Errorln("error crashed the game lobby")
			return
		}
	}
}

func (g *Game) created() error {
	readyCh := make(chan int64)
	cleanAny := g.Events.Register(EventAny, func(info EventInfo) {
		if !info.IsType(EventUserReady, EventUserResigned) {
			return
		}
		readyCh <- info.AccountID
	})

	defer cleanAny()

	noticeSent := false
	defer g.Events.Clean(EventJoinReminder)
	defer g.Events.Clean(EventLateResign)

	defer g.reloadClientLobbies()

	deadline, cancel := context.WithTimeout(context.Background(), g.ReminderToReadyAfter)
	for {
		select {
		case <-g.ctx.Done():
			cancel()
			return g.ctx.Err()
		case _ = <-readyCh:
			g.loadLobby()
			if !g.lobby.EveryoneReady() {
				g.reloadClientLobbies()
				continue
			}
			cancel()
			g.lobby.State = "get-ready"
			g.saveLobby()
			g.reloadClientLobbies()
			return nil
		case <-deadline.Done():
			cancel()
			g.loadLobby()
			if !noticeSent {
				noticeSent = true
				deadline, cancel = context.WithTimeout(context.Background(), g.ReadyDeadline-g.ReminderToReadyAfter)

				for accountId, state := range g.lobby.UserState {
					if state.IsResigned || state.IsReady {
						continue
					}
					g.Events.Dispatch(EventJoinReminder, EventInfo{AccountID: accountId})
				}
			} else {
				g.lobby.State = "get-ready"
				g.saveLobby()

				for accountId, state := range g.lobby.UserState {
					if state.IsResigned || state.IsReady {
						continue
					}
					state.IsResigned = true
					g.lobby.UserState[accountId] = state
					if err := g.app.Account.SetField(g.ctx,
						entity.NewID("account", accountId),
						"current_lobby", ""); err != nil {
						logrus.WithError(err).Errorln("couldn't save resigned user after timeout")
					}
					logrus.WithField("userId", accountId).Info("user late resigned")
					g.Events.Dispatch(EventLateResign, EventInfo{AccountID: accountId})
					g.reloadClientLobbies()
				}
				return nil
			}
		}

	}
}

func (g *Game) getReady() error {
	defer g.reloadClientLobbies()

	<-time.After(g.GetReadyDuration)
	g.lobby.State = "started"
	g.lobby.GameInfo.CorrectAnswers = make(map[int64][]bool)
	g.lobby.GameInfo.CurrentQuestion = 0
	g.lobby.GameInfo.CurrentQuestionStartedAt = time.Now()
	g.saveLobby()
	return nil
}

func (g *Game) started() error {
	chUpdate := make(chan struct{}, 10)
	eCancel := g.Events.Register(EventAny, func(info EventInfo) {
		if !info.IsType(EventUserAnswer, EventUserResigned) {
			return
		}
		g.loadLobby()
		switch info.Type {
		case EventUserAnswer:
			accountId := info.AccountID
			answerIndex := info.UserAnswer
			questionIndex := info.QuestionIndex
			if questionIndex != g.lobby.GameInfo.CurrentQuestion {
				return
			}

			// check has answered questionIndex of questionIndex+1 questions
			if len(g.lobby.GameInfo.CorrectAnswers[accountId]) != questionIndex {
				return
			}

			answer := g.lobby.Questions[questionIndex].CorrectAnswer == answerIndex
			g.lobby.GameInfo.CorrectAnswers[accountId] = append(g.lobby.GameInfo.CorrectAnswers[accountId], answer)
			userState := g.lobby.UserState[accountId]
			userState.LastAnsweredQuestionIndex = questionIndex
			g.lobby.UserState[accountId] = userState
			g.saveLobby()
			chUpdate <- struct{}{}

		case EventUserResigned:
			accountId := info.AccountID
			if !slices.Contains(g.lobby.Participants, accountId) {
				return
			}

			userState := g.lobby.UserState[accountId]
			if userState.IsResigned {
				return
			}
			userState.IsResigned = true
			g.lobby.UserState[accountId] = userState
			g.saveLobby()
			g.reloadClientLobbies()
		}
	})
	defer eCancel()

	for {
		timeout, cancel := context.WithTimeout(g.ctx,
			g.QuestionTimeout-time.Since(g.lobby.GameInfo.CurrentQuestionStartedAt))

		select {
		case <-g.ctx.Done():
			cancel()
			return nil
		case <-chUpdate: // one user has made their answer
			g.loadLobby()
			if len(g.lobby.NotAnsweredUsers()) != 0 {
				g.reloadClientLobbies()
				continue
			}
			g.nextQuestion()
		case <-timeout.Done(): // timeout 30s, finding user's didn't answer
			notAnsweredUsers := g.lobby.NotAnsweredUsers()
			for _, userId := range notAnsweredUsers {
				g.lobby.GameInfo.CorrectAnswers[userId] = append(g.lobby.GameInfo.CorrectAnswers[userId], false)
			}
			g.nextQuestion()
		}
	}
}

func (g *Game) nextQuestion() {
	g.lobby.GameInfo.CurrentQuestion += 1
	g.lobby.GameInfo.CurrentQuestionStartedAt = time.Now()
	g.saveLobby()
	g.reloadClientLobbies()
}

func (g *Game) reloadClientLobbies() {
	g.Events.Dispatch(EventForceLobbyReload, EventInfo{})
}

func (g *Game) loadLobby() {
	lobby, err := g.app.Lobby.Get(g.ctx, g.LobbyId)
	if err != nil {
		logrus.WithError(err).WithField("id", g.LobbyId.ID()).Errorln("couldn't load the game's lobby")
		g.cancel()
		return
	}
	g.lobby = lobby
}

func (g *Game) saveLobby() {
	err := g.app.Lobby.Save(g.ctx, g.lobby)
	if err != nil {
		logrus.WithError(err).WithField("id", g.LobbyId.ID()).Errorln("couldn't load the game's lobby")
		g.cancel()
		return
	}
}
