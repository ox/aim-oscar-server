package main

import (
	"aim-oscar/oscar"
	"sync"
)

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

func (sm *SessionManager) SetSession(username string, session *oscar.Session) {
	sm.mutex.Lock()
	sm.sessions[username] = session
	sm.mutex.Unlock()
}

func (sm *SessionManager) GetSession(username string) *oscar.Session {
	sm.mutex.RLock()
	s, ok := sm.sessions[username]
	sm.mutex.RUnlock()

	if ok {
		return s
	}
	return nil
}

func (sm *SessionManager) RemoveSession(username string) {
	sm.mutex.Lock()
	sm.sessions[username] = nil
	sm.mutex.Unlock()
}
