package Player

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/chris-dot-exe/AwesomeLog"
)

type stateStopped struct {
	player *player
}

func (s stateStopped) State() StateName {
	return Stopped
}

// OnEntry disconnects and closes voice channel if bot is currently connected to one.
//
// Also stops any currently playing media.
func (s stateStopped) OnEntry(_ State) {
	s.player.dcPlayer.Stop()

	s.player.vcMutex.Lock()
	vc := s.player.currentVc
	if vc != nil {
		err := vc.Disconnect()
		if err != nil {
			log.Println(log.WARN, "error disconnecting from voice channel: ", err)
		}
		vc.Close()
		s.player.currentVc = nil
	}
	s.player.vcMutex.Unlock()

}

func (s stateStopped) OnExit() {
}

func (s stateStopped) Play(i *discordgo.Interaction, mediaName string) error {
	err := s.player.initVc(i)
	if err != nil {
		return err
	}

	err = s.player.AddQueueFront(mediaName)
	if err != nil {
		return err
	}

	s.player.states.setState(Idle)

	return nil
}

func (s stateStopped) Stop() error {
	return nil // todo error for discord reply?
}

func (s stateStopped) TogglePause() error {
	return nil // todo error for discord reply?
}

func (s stateStopped) Forward(forwardCount uint) error {
	return nil // todo error for discord reply?
}

func (s stateStopped) Backward() error {
	return nil // todo error for discord reply?
}

func (s stateStopped) Playing() bool {
	return false
}
