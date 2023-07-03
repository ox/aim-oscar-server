package main

import (
	"aim-oscar/oscar"
	"sync"
)

// SessionManager maps screen names to user sessions
type SessionManager struct {
	sessions map[string]*oscar.Session
	mutex    *sync.RWMutex
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*oscar.Session),
		mutex:    &sync.RWMutex{},
	}
	return sm
}

func (sm *SessionManager) SetSession(screen_name string, session *oscar.Session) {
	sm.mutex.Lock()
	sm.sessions[screen_name] = session
	sm.mutex.Unlock()
}

func (sm *SessionManager) GetSession(screen_name string) *oscar.Session {
	sm.mutex.RLock()
	s, ok := sm.sessions[screen_name]
	sm.mutex.RUnlock()

	if ok {
		return s
	}
	return nil
}

func (sm *SessionManager) RemoveSession(screen_name string) {
	sm.mutex.Lock()
	sm.sessions[screen_name] = nil
	sm.mutex.Unlock()
}
