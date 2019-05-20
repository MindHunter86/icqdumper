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
	gLogger     *zerolog.Logger
	gMongoDB    *mongodb.MongoDB
	gChatsQueue chan *job
	gDBQueue    chan *job
)

type (
	App struct {
		params             *AppParams
		icqClient          *ICQApi
		chatsDispatcher    *dispatcher
		databaseDispatcher *dispatcher
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

func (m *App) Bootstrap(chatId string) (e error) {

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
	m.chatsDispatcher = newDispatcher(m.params.QueueBuffer, m.params.WorkerCapacity)
	gChatsQueue = m.chatsDispatcher.getQueueChan()

	m.databaseDispatcher = newDispatcher(m.params.QueueBuffer, m.params.WorkerCapacity)
	gDBQueue = m.databaseDispatcher.getQueueChan()

	// bootstrap part
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	var waitGroup sync.WaitGroup
	var errorPipe chan error

	go func(ep chan error, wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		gLogger.Debug().Msg("Queue CHAT worker spawn && Queue dispatch...")
		ep <- m.chatsDispatcher.bootstrap(m.params.Workers)
	}(errorPipe, waitGroup)

	go func(ep chan error, wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		gLogger.Debug().Msg("Queue DB worker spawn && Queue dispatch...")
		ep <- m.databaseDispatcher.bootstrap(m.params.Workers)
	}(errorPipe, waitGroup)

	go func(ep chan error, wg sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		gLogger.Debug().Msg("Starting chats && messages parsing...")
		ep <- m.CliGetHistory(m.params.AimSid, chatId)
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
	return m.Destroy()
}

func (m *App) Destroy() (e error) {
	m.databaseDispatcher.destroy()
	m.chatsDispatcher.destroy()
	gMongoDB.Destruct()
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
