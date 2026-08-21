package player

import (
	"fmt"
	"log/slog"
)

// StateMachine drives one guildPlayer through Stopped -> Idle ->
// Playing/Paused. Transitions only happen from controlLoop, one at a time.
type StateMachine struct {
	guildID string

	currentState State

	stateStopped State
	stateIdle    State
	statePlaying State
	statePaused  State
}

func NewStateMachine(p *guildPlayer) *StateMachine {
	sm := &StateMachine{guildID: p.guildID}

	sm.stateStopped = stateStopped{p}
	sm.stateIdle = stateIdle{p}
	sm.statePlaying = statePlaying{p}
	sm.statePaused = statePaused{p}
	sm.currentState = sm.stateStopped

	return sm
}

func (sm *StateMachine) getState() State {
	return sm.currentState
}

func (sm *StateMachine) setState(newStateName StateName) {
	if sm.currentState.State() == newStateName {
		return
	}
	slog.Info("player state change", "guild_id", sm.guildID, "from", sm.currentState.State(), "to", newStateName)

	oldState := sm.currentState
	oldState.OnExit()

	sm.currentState = sm.fromName(newStateName)
	sm.currentState.OnEntry(oldState)
}

func (sm *StateMachine) fromName(name StateName) State {
	switch name {
	case Stopped:
		return sm.stateStopped
	case Idle:
		return sm.stateIdle
	case Playing:
		return sm.statePlaying
	case Paused:
		return sm.statePaused
	}

	panic(fmt.Sprintf("tried to change to unknown player state: %q", name))
}
