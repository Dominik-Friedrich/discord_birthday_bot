package Player

import (
    "errors"
    "github.com/bwmarrin/discordgo"
    log "github.com/chris-dot-exe/AwesomeLog"
    "github.com/gammazero/deque"
    "main/src/bot"
    "main/src/features/Player/discord"
    "sync"
    "time"
)

const (
    play = iota
    idle
    togglePause
    stop
    forward
    backwards

    featureMP3Player = "featureMP3Player"

    defaultQueueSize        = 32
    defaultDirectory        = "./resources"
    defaultMaxMediaFiles    = 100
    defaultMaxMediaDuration = 10 * time.Minute
)

// todo: compatibility for multiple guilds
// todo: thread safety
type player struct {
    session   *bot.Session
    currentVc *discordgo.VoiceConnection
    vcMutex   sync.Mutex

    dcPlayer       *discord.AudioPlayer
    dcPlayDoneChan chan discord.PlayerContext

    states *StateMachine

    mediaManager *MediaManager

    currentMedia string
    queue        *deque.Deque[string]
    queueMutex   sync.Mutex
    history      *deque.Deque[string]
    historyMutex sync.Mutex

    togglePause chan struct{}
    play        chan struct {
        dcI   *discordgo.Interaction
        query string
    }
    stop chan struct{}
    idle chan struct{}
}

func Player() bot.Feature {
    b := new(player)
    b.queue = deque.New[string](defaultQueueSize)
    b.history = deque.New[string](defaultQueueSize)

    b.play = make(chan struct {
        dcI   *discordgo.Interaction
        query string
    })
    b.togglePause = make(chan struct{})
    b.stop = make(chan struct{})
    b.idle = make(chan struct{})

    b.dcPlayDoneChan = make(chan discord.PlayerContext)
    dcPlayer, err := discord.NewPlayer(b.dcPlayDoneChan)
    if err != nil {
        log.Fatalf("error initializing discord player: ", err)
    }
    b.dcPlayer = dcPlayer

    b.states = NewStateMachine(b)

    b.mediaManager = NewMediaManager(defaultDirectory, defaultMaxMediaFiles, defaultMaxMediaDuration)

    go b.asyncPlayerStateControlRoutine()

    return b
}

func (p *player) Init(session *bot.Session) error {
    p.session = session

    return nil
}

func (p *player) Name() string {
    return featureMP3Player
}

func (p *player) Commands() []bot.Command {
    return []bot.Command{
        Play(p),
        Pause(p),
        Stop(p),
        //commands.Forward(),
        //commands.Backward(),
    }
}

func (p *player) SupportedSites() []string {
    return []string{ // TODO dynamic generation
        "youtube",
    }
}

// Play pushes the media into the queue. If the player is not currently playing it sends a signal to play the next media
func (p *player) Play(i *discordgo.Interaction, query string) error {
    p.play <- struct {
        dcI   *discordgo.Interaction
        query string
    }{dcI: i, query: query}

    return nil
}

func (p *player) TogglePause() error {
    p.togglePause <- struct{}{}
    return nil
}

func (p *player) Forward() error {
    //return p.getState().Forward()
    return nil
}

func (p *player) Backward() error {
    //return p.getState().Backward()
    return nil
}

func (p *player) Stop() error {
    p.stop <- struct{}{}
    return nil
}

func (p *player) Playing() bool {
    state := p.states.getState().State()
    return state == Stopped || state == Paused
}

func (p *player) asyncPlayerStateControlRoutine() {
    for {
        var err error
        select {
        case ctx := <-p.play:
            err = p.states.getState().Play(ctx.dcI, ctx.query)
        case <-p.togglePause:
            err = p.states.getState().TogglePause()
        case <-p.stop:
            err = p.states.getState().Stop()
        //case <-p.idle:
        case ctx := <-p.dcPlayDoneChan:
            switch ctx.ExitReason {
            case discord.Finished:
                p.states.setState(Idle)
            case discord.Error:
                fallthrough
            case discord.Stopped:
                p.states.setState(Stopped)
            }
        }

        if err != nil {
            log.Println(log.WARN, "error in state control routine: ", err)
        }
    }
}

func (p *player) playNextMedia() {
    log.Println(log.INFO, "[playNextMedia]")
    p.queueMutex.Lock()
    defer p.queueMutex.Unlock()
    if p.queue.Len() == 0 {
        return
    }

    p.currentMedia = p.queue.PopFront()

    p.vcMutex.Lock()
    vc := p.currentVc
    p.vcMutex.Unlock()

    p.states.setState(Playing)

    go func(query string) {
        fileName, err := p.mediaManager.GetMediaFilePathByQuery(query)
        if err != nil {
            log.Printf(log.WARN, "error trying to get media: %s", err)
            return
        }
        log.Println(log.INFO, "playing next media: ", p.currentMedia)
        p.dcPlayer.Play(discord.PlayerContext{
            Vc:        vc,
            MediaName: fileName,
        })
    }(p.currentMedia)
}

func (p *player) initVc(i *discordgo.Interaction) error {
    p.vcMutex.Lock()
    defer p.vcMutex.Unlock()
    if p.currentVc != nil {
        return nil
    }

    g, err := p.session.State.Guild(i.GuildID)

    // Look for the message sender in that guild's current voice states.
    var channelId string
    for _, vs := range g.VoiceStates {
        if vs.UserID == i.Member.User.ID {
            channelId = vs.ChannelID
            break
        }
    }

    if channelId == "" {
        return errors.New("user not in voice channel")
    }

    // Join the provided voice channel.
    vc, err := p.session.ChannelVoiceJoin(i.GuildID, channelId, false, true)
    if err != nil {
        return errors.New("could not join voice channel")
    }

    p.currentVc = vc

    return nil
}

func (p *player) speaking(b bool) {
    p.vcMutex.Lock()
    if p.currentVc != nil {
        err := p.currentVc.Speaking(b)
        if err != nil {
            log.Println(log.WARN, "error setting speaking status: ", err)
        }
    }
    p.vcMutex.Unlock()
}

func (p *player) AddQueueBack(mediaName string) error {
    if mediaName == "" {
        return errors.New("mediaName not set")
    }

    log.Println("added media to back of queue: ", mediaName)
    p.queueMutex.Lock()
    p.queue.PushBack(mediaName)

    // pre download next media
    if p.queue.Len() > 0 {
        nextMedia := p.queue.Front()
        go p.downloadMediaIfNeeded(nextMedia)
    }

    p.queueMutex.Unlock()
    return nil
}

func (p *player) AddQueueFront(mediaName string) error {
    if mediaName == "" {
        return errors.New("mediaName not set")
    }

    log.Println("added media to front of queue: ", mediaName)

    p.queueMutex.Lock()
    p.queue.PushFront(mediaName)
    p.queueMutex.Unlock()

    return nil
}

func (p *player) AddHistoryFront(mediaName string) error {
    if mediaName == "" {
        return errors.New("mediaName not set")
    }

    p.historyMutex.Lock()
    p.history.PushFront(mediaName)
    p.historyMutex.Unlock()

    return nil
}

func (p *player) downloadMediaIfNeeded(media string) {
    if media == "" {
        return
    }

    log.Printf(log.DEBUG, "pre downloading media if needed: %s", media)
    _, err := p.mediaManager.GetMediaFilePathByQuery(media)
    if err != nil {
        log.Printf(log.WARN, "error trying to get media: %s", err)
    }
}
