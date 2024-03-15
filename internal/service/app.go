package service

type App struct {
	Account *AccountService
}

func NewApp(
	Account *AccountService,
) *App {
	return &App{Account: Account}
}
