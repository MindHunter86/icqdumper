package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MindHunter86/icqdumper/system/mongodb"
	"github.com/rs/zerolog"
)

var (
	gLogger  *zerolog.Logger
	gMongoDB *mongodb.MongoDB
	gQueue   chan *job
)

type (
	App struct {
		params          *AppParams
		icqClient       *ICQApi
		queueDispatcher *dispatcher
	}
	AppParams struct {
		Silent                               bool
		AimSid, MongoConn                    string
		Workers, QueueBuffer, WorkerCapacity int
	}
)

func NewApp(l *zerolog.Logger, params *AppParams) *App {
	gLogger = l

	return &App{
		params: params,
	}
}

func (m *App) Bootstrap() (e error) {

	gLogger.Debug().Msg("Starting App initialization...")

	gLogger.Debug().Msg("MongoDB bootstrap...")
	if gMongoDB, e = mongodb.NewMongoDriver(gLogger, m.params.MongoConn); e != nil {
		return
	}

	gLogger.Debug().Msg("MongoDB database connect...")
	if e = gMongoDB.Construct(); e != nil {
		return e
	}

	gLogger.Debug().Msg("Queue bootstrap...")
	m.queueDispatcher = newDispatcher(m.params.QueueBuffer, m.params.WorkerCapacity)
	gQueue = m.queueDispatcher.getQueueChan()

	// bootstrap part
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	var waitGroup sync.WaitGroup
	var errorPipe chan error

	go func(ep chan error, wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		gLogger.Debug().Msg("Queue worker spawn && Queue dispatch...")
		ep <- m.queueDispatcher.bootstrap(m.params.Workers)
	}(errorPipe, waitGroup)

LOOP:
	for {
		select {
		case <-kernSignal:
			gLogger.Info().Msg("Syscall.SIG* has been detected! Closing application...")
			break LOOP
		case e = <-errorPipe:
			gLogger.Error().Err(e).Msg("Runtime error! Abnormal application closing!")
			break LOOP
		}
	}

	waitGroup.Wait()

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
