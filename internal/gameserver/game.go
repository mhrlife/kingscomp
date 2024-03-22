package gameserver

import (
	"context"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
	"kingscomp/internal/gameserver/events"
	"kingscomp/internal/service"
	"slices"
	"time"
)

type Game struct {
	Config

	app     *service.App
	server  *GameServer
	LobbyId entity.ID
	Events  events.Eventer

	Ctx        context.Context
	CancelFunc context.CancelFunc

	lobby entity.Lobby
}

func NewGame(lobbyId string, app *service.App, server *GameServer, config Config) *Game {
	return &Game{
		Config:  config,
		LobbyId: entity.NewID("lobby", lobbyId),
		app:     app,
		server:  server,
		Events:  events.NewInMemoryEvents(),
	}
}

func (g *Game) Start(ctx context.Context) {
	g.Ctx, g.CancelFunc = context.WithCancel(ctx)
	for {
		g.loadLobby()

		select {
		case <-g.Ctx.Done():
			return
		default:
		}

		logrus.WithFields(logrus.Fields{
			"lobbyId":    g.lobby.ID,
			"lobbyState": g.lobby.State,
		}).Info("running sub-state for game")

		var err error
		switch g.lobby.State {
		case "created": // just created, waiting for other users to join
			err = g.created()
		case "get-ready": // showing count down of game start
			err = g.getReady()
		case "started": // users are answering to questions
			err = g.started()
		case "ended":
			g.close()
			return

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
	cleanAny, _ := g.Events.Register(events.EventAny, func(info events.EventInfo) {
		if !info.IsType(events.EventUserReady, events.EventUserResigned) {
			return
		}
		readyCh <- info.AccountID
	})

	defer cleanAny()

	defer g.Events.Clean(events.EventJoinReminder)
	defer g.Events.Clean(events.EventLateResign)
	defer g.reloadClientLobbies()

	noticeSent := false
	deadline, cancel := context.WithTimeout(context.Background(), g.ReminderToReadyAfter)
	for {
		select {
		case <-g.Ctx.Done():
			cancel()
			return g.Ctx.Err()
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
					g.Events.Dispatch(events.EventJoinReminder, events.EventInfo{AccountID: accountId})
				}
			} else {
				for accountId, state := range g.lobby.UserState {
					if state.IsResigned || state.IsReady {
						continue
					}
					state.IsResigned = true
					g.lobby.UserState[accountId] = state
					if err := g.app.Account.SetField(g.Ctx,
						entity.NewID("account", accountId),
						"current_lobby", ""); err != nil {
						logrus.WithError(err).Errorln("couldn't save resigned user after timeout")
					}
					logrus.WithField("userId", accountId).Info("user late resigned")
					g.Events.Dispatch(events.EventLateResign, events.EventInfo{AccountID: accountId})
				}

				g.lobby.State = "get-ready"
				g.saveLobby()
				g.reloadClientLobbies()
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
	g.lobby.GameInfo.CurrentQuestionEndsAt = time.Now().Add(g.Config.QuestionTimeout)
	g.saveLobby()
	return nil
}

func (g *Game) started() error {
	chUpdate := make(chan events.EventInfo, 10)
	eCancel, _ := g.Events.Register(events.EventAny, func(info events.EventInfo) {
		if !info.IsType(events.EventUserAnswer, events.EventUserResigned) {
			return
		}
		chUpdate <- info
	})
	defer eCancel()

	for {

		timeout, cancel := context.WithTimeout(g.Ctx,
			g.lobby.GameInfo.CurrentQuestionEndsAt.Sub(time.Now()))

		if g.lobby.State == "ended" {
			cancel()
			return nil
		}

		select {
		case <-g.Ctx.Done():
			cancel()
			return nil
		case info := <-chUpdate: // one user has made their answer
			g.loadLobby()
			switch info.Type {
			case events.EventUserResigned:
				//todo: check if all users have answered except the resigned user
				accountId := info.AccountID
				if !slices.Contains(g.lobby.Participants, accountId) {
					continue
				}

				userState := g.lobby.UserState[accountId]
				if userState.IsResigned {
					continue
				}
				userState.IsResigned = true
				g.lobby.UserState[accountId] = userState
				g.saveLobby()
				g.reloadClientLobbies()
			case events.EventUserAnswer:
				accountId := info.AccountID
				answerIndex := info.UserAnswer
				questionIndex := info.QuestionIndex

				if questionIndex != g.lobby.GameInfo.CurrentQuestion {
					continue
				}
				// check has answered questionIndex of questionIndex+1 questions
				if len(g.lobby.GameInfo.CorrectAnswers[accountId]) != questionIndex {
					continue
				}

				answer := g.lobby.Questions[questionIndex].CorrectAnswer == answerIndex
				g.lobby.GameInfo.CorrectAnswers[accountId] = append(g.lobby.GameInfo.CorrectAnswers[accountId], answer)
				userState := g.lobby.UserState[accountId]
				userState.LastAnsweredQuestionIndex = questionIndex
				g.lobby.UserState[accountId] = userState
				g.saveLobby()
			}

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
	// they have answered to all questions
	if g.lobby.GameInfo.CurrentQuestion == len(g.lobby.Questions)-1 {
		g.lobby.State = "ended"
		//todo: find who is the winner and create the scoreboard
		g.saveLobby()
		g.reloadClientLobbies()
		return
	}

	g.lobby.GameInfo.CurrentQuestion += 1
	g.lobby.GameInfo.CurrentQuestionStartedAt = time.Now()
	g.lobby.GameInfo.CurrentQuestionEndsAt = time.Now().Add(g.Config.QuestionTimeout)
	g.saveLobby()
	g.reloadClientLobbies()
}

func (g *Game) reloadClientLobbies() {
	g.Events.Dispatch(events.EventForceLobbyReload, events.EventInfo{})
}

func (g *Game) loadLobby() {
	lobby, err := g.app.Lobby.Get(g.Ctx, g.LobbyId)
	if err != nil {
		logrus.WithError(err).WithField("id", g.LobbyId.ID()).Errorln("couldn't load the game's lobby")
		g.CancelFunc()
		return
	}
	g.lobby = lobby
}

func (g *Game) saveLobby() {
	err := g.app.Lobby.Save(g.Ctx, g.lobby)
	if err != nil {
		logrus.WithError(err).WithField("id", g.LobbyId.ID()).Errorln("couldn't load the game's lobby")
		g.CancelFunc()
		return
	}
}

func (g *Game) close() {
	for userId, state := range g.lobby.UserState {
		if !state.IsResigned {
			g.Events.Dispatch(events.EventGameClosed, events.EventInfo{AccountID: userId})
		}
	}
	<-time.After(1 * time.Second)
	g.Events.Close()
	g.CancelFunc()
}
