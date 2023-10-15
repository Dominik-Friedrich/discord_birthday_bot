package Player

import (
	"fmt"
	log "github.com/chris-dot-exe/AwesomeLog"
)

type StateMachine struct {
	currentState State
	//stateMutex   sync.RWMutex

	stateStopped State
	statePaused  State
	statePlaying State
	stateIdle    State
}

func NewStateMachine(p *player) *StateMachine {
	sm := new(StateMachine)

	sm.stateStopped = stateStopped{p}
	sm.statePaused = statePaused{p}
	sm.statePlaying = statePlaying{p}
	sm.stateIdle = stateIdle{p}
	sm.currentState = sm.stateStopped

	return sm
}

func (p *StateMachine) getState() State {
	//p.stateMutex.RLock()
	state := p.currentState
	//p.stateMutex.RUnlock()
	return state
}

func (p *StateMachine) setState(newStateName StateName) {
	//p.stateMutex.Lock()

	if p.currentState.State() == newStateName {
		return
	}
	log.Printf(log.INFO, "player state change '%s' -> '%s'", p.currentState.State(), newStateName)

	oldState := p.currentState
	oldState.OnExit()

	p.currentState = p.fromName(newStateName)
	p.currentState.OnEntry(oldState)

	//p.stateMutex.Unlock()
}

func (p *StateMachine) fromName(stateName StateName) State {
	switch stateName {
	case Stopped:
		return p.stateStopped
	case Paused:
		return p.statePaused
	case Playing:
		return p.statePlaying
	case Idle:
		return p.stateIdle
	}

	panic(fmt.Sprintf("tried to change to unknown state: '%s'", stateName))
}
