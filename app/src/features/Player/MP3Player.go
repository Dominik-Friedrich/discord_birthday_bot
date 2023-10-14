package Player

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
	"github.com/gammazero/deque"
	"io"
	"main/src/bot"
	"os/exec"
	"strconv"
	"sync"
)

const (
	featureMP3Player = "featureMP3Player"

	defaultQueueSize = 32

	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

// todo: compatibility for multiple guilds
// todo: thread safety
type player struct {
	session   *bot.Session
	currentVc *discordgo.VoiceConnection

	currentlyPlaying string
	queue            *deque.Deque[string]
	queueMutex       sync.Mutex
	history          *deque.Deque[string]
	historyMutex     sync.Mutex

	state      PlayerState
	stateMutex sync.RWMutex

	play chan string

	playNext     chan struct{}
	playPrevious chan struct{}
	pause        chan struct{}
	unpause      chan struct{}
	stop         chan struct{}
}

func Player() bot.Feature {
	b := new(player)
	b.queue = deque.New[string](defaultQueueSize)
	b.history = deque.New[string](defaultQueueSize)

	b.play = make(chan string)

	b.playNext = make(chan struct{})
	b.playPrevious = make(chan struct{})
	b.pause = make(chan struct{})
	b.unpause = make(chan struct{})
	b.stop = make(chan struct{})

	b.state = Stopped
	go b.asyncPlayRoutine()
	go b.asyncPlayerRoutine()

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
func (p *player) Play(interaction *discordgo.Interaction, mediaName string) error {
	if p.currentVc == nil {
		err := p.initVc(interaction)
		if err != nil {
			return err
		}
	}

	if mediaName != "" {
		p.queueMutex.Lock()
		p.queue.PushBack(mediaName)
		p.queueMutex.Unlock()
	}

	switch p.getState() {
	case Stopped:
		p.playNext <- struct{}{}
	case Paused:
		p.unpause <- struct{}{}
	}

	return nil
}

func (p *player) Pause() error {
	p.pause <- struct{}{}
	return nil
}

func (p *player) Forward() error {
	p.playNext <- struct{}{}
	return nil
}

func (p *player) Backward() error {
	p.playPrevious <- struct{}{}
	return nil
}

func (p *player) Stop() error {
	if p.getState() != Stopped {
		p.stop <- struct{}{}
	}
	return nil
}

func (p *player) Playing() bool {
	return p.getState() != Stopped
}

func (p *player) asyncPlayRoutine() {
	for {
		select {
		case ctx := <-p.play:
			p.setState(Playing)
			p.playAudioFile(p.currentVc, ctx)
			p.playNext <- struct{}{}
		}
	}
}

func (p *player) asyncPlayerRoutine() {
	var currentMedia string
	for {
		select {
		case <-p.playNext:
			p.queueMutex.Lock()
			p.historyMutex.Lock()

			if p.getState() != Stopped && currentMedia != "" {
				p.history.PushFront(currentMedia) // max history?
			}
			if p.queue.Len() > 0 {
				currentMedia = p.queue.PopFront()
				log.Println(log.INFO, "playing next media: ", currentMedia)
				p.play <- currentMedia
			}

			p.queueMutex.Unlock()
			p.historyMutex.Unlock()
		case <-p.playPrevious:

		}
		log.Print(log.INFO, currentMedia)
	}
}

// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
//
// copied and pasted from dgvoice
func (p *player) playAudioFile(v *discordgo.VoiceConnection, filename string) {

	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		dgvoice.OnError("StdoutPipe Error", err)
		return
	}

	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	err = run.Start()
	if err != nil {
		dgvoice.OnError("RunStart Error", err)
		return
	}

	// prevent memory leak from residual ffmpeg streams
	defer run.Process.Kill()

	// Send "speaking" packet over the voice websocket
	err = v.Speaking(true)
	if err != nil {
		dgvoice.OnError("Couldn't set speaking", err)
	}

	// Send not "speaking" packet over the websocket when we finish
	defer func() {
		err := v.Speaking(false)
		if err != nil {
			dgvoice.OnError("Couldn't stop speaking", err)
		}
	}()

	send := make(chan []int16, 2)
	defer close(send)

	var wg sync.WaitGroup

	close := make(chan bool)
	go func() {
		defer wg.Done()
		dgvoice.SendPCM(v, send)
		close <- true
	}()

	// Producer Goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-close:
				return
			case <-p.stop:
				log.Println(log.INFO, "player stopped")
				return
			case <-p.pause:
				p.setState(Paused)
				log.Println(log.INFO, "player paused")
				<-p.unpause
				p.setState(Playing)

			default:
				// read data from ffmpeg stdout
				audiobuf := make([]int16, frameSize*channels)
				err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return
				}
				if err != nil {
					dgvoice.OnError("error reading from ffmpeg stdout", err)
					return
				}
				// Send received PCM to the sendPCM channel
				send <- audiobuf
			}
		}
	}()

	wg.Wait()
}

func (p *player) getState() PlayerState {
	p.stateMutex.RLock()
	state := p.state
	p.stateMutex.RUnlock()
	return state
}

func (p *player) setState(newState PlayerState) {
	log.Println(log.INFO, "player ", newState.String())

	p.stateMutex.Lock()
	p.state = newState
	p.stateMutex.Unlock()
}

func (p *player) initVc(i *discordgo.Interaction) error {
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
