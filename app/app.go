package app

import (
	"github.com/MindHunter86/icqdumper/system/mongodb"
	"github.com/rs/zerolog"
)

var (
	gLogger  *zerolog.Logger
	gMongoDB *mongodb.MongoDB
)

type App struct {
	icqClient *ICQApi
}

func NewApp(l *zerolog.Logger) *App {
	gLogger = l
	return &App{}
}

func (m *App) Create() (*App, error) {
	var e error
	return m, e
}

func (m *App) Bootstrap() (e error) {
	return e
}

func (*App) Destroy() (e error) {
	return e
}

func (m *App) CliGetHistory(aimsid, chatid string) (e error) {
	m.icqClient = NewICQApi(aimsid)
	return m.parseChatId(chatid)
}

func (m *App) parseChatId(chatId string) (e error) {
	switch chatId {
	case "all":
		if chats, e := m.icqClient.getChats(); e != nil {
			return e
		} else {
			return m.icqClient.getChatsMessages(chats)
		}
	default:
		return m.icqClient.getChatMessages(chatId, 1)
	}
}
