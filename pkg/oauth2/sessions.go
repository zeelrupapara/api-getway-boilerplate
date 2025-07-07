package oauth2

import (
	"context"
	"time"
)

var (
	SessionIdleWait = 5 * time.Minute
	SessionInterval = 1 * time.Minute
)

// map[session_id]=*config
type ActiveSessionsList map[string]*Config

func (o *OAuth2) GetActiveSessions() ActiveSessionsList {
	o.RLock()
	defer o.RUnlock()

	return o.SessionsList
}

func (o *OAuth2) GetActiveSessionById(id string) (*Config, bool) {
	o.RLock()
	defer o.RUnlock()

	for k := range o.SessionsList {
		if o.SessionsList[k].Id == id {
			return o.SessionsList[k], true
		}
	}

	return nil, false
}

func (o *OAuth2) GetActiveSessionsCountByClientId(clientId int32) (*Config, int) {
	o.RLock()
	defer o.RUnlock()

	count := 0
	var cfg *Config
	for k := range o.SessionsList {
		if o.SessionsList[k].ClientId == clientId {
			count++
			if cfg == nil {
				cfg = o.SessionsList[k]
			} else {
				if o.SessionsList[k].StartedAt.Before(cfg.StartedAt) {
					cfg = o.SessionsList[k]
				}
			}
		}
	}

	return cfg, count
}

// get all active sessions by client id and retuns the newest one as the first element
func (o *OAuth2) GetAllActiveSessionsCountByClientId(clientId int32) []*Config {
	o.RLock()
	defer o.RUnlock()

	cfg := []*Config{}
	oldestOne := &Config{}
	index := -1
	for k := range o.SessionsList {
		if o.SessionsList[k].ClientId == clientId {
			cfg = append(cfg, o.SessionsList[k])
			if o.SessionsList[k].StartedAt.After(oldestOne.StartedAt) {
				oldestOne = o.SessionsList[k]
				index = len(cfg) - 1
			}
		}
	}

	if index != -1 {
		cfg[index] = cfg[0]
		cfg[0] = oldestOne
	}

	return cfg
}

func (o *OAuth2) NewActiveSession(cfg *Config) {
	defer o.Unlock()
	o.Lock()

	o.SessionsList[cfg.SessionId] = cfg
}

func (o *OAuth2) NewActivity(sessionId string) {
	o.Lock()
	defer o.Unlock()

	if _, ok := o.getActiveSession(sessionId); ok {
		o.SessionsList[sessionId].LastActivity = time.Now()
	}
}

func (o *OAuth2) DeleteActiveSession(sessionId string) {
	defer o.Unlock()
	o.Lock()

	if _, ok := o.getActiveSession(sessionId); ok {
		o.Log.Logger.Infof("removing %s session", sessionId)
		delete(o.SessionsList, sessionId)
	}
}

func (o *OAuth2) WSConnected(sessionId string) {
	defer o.Unlock()
	o.Lock()

	if _, ok := o.getActiveSession(sessionId); ok {
		o.SessionsList[sessionId].Ws = true
	}
}

func (o *OAuth2) WSDisconnected(sessionId string) {
	defer o.Unlock()
	o.Lock()

	if _, ok := o.getActiveSession(sessionId); ok {
		o.SessionsList[sessionId].Ws = false
		o.SessionsList[sessionId].LastActivity = time.Now()
	}
}

var onSessionDelete = func(cfg *Config) {}

func (o *OAuth2) MonitoryActivity() {
	for {
		time.Sleep(SessionInterval)
		oldSessions := make([]*Config, 0)
		o.RLock()
		for _, v := range o.SessionsList {
			if time.Since(v.LastActivity) >= SessionIdleWait && !v.Ws {
				oldSessions = append(oldSessions, v)
			}
		}
		o.RUnlock()
		for i := range oldSessions {
			onSessionDelete(oldSessions[i])

			if !oldSessions[i].RememberMe {
				ctx := context.Background()
				err := o.Logout(ctx, oldSessions[i].AccessToken, oldSessions[i].SessionId)
				if err != nil {
					o.Log.Logger.Errorf("error removing %s session", oldSessions[i].SessionId)
				}
			} else {
				o.DeleteActiveSession(oldSessions[i].SessionId)
			}
		}
	}
}

func (o *OAuth2) SetOnSessionDelete(f func(cfg *Config)) {
	onSessionDelete = f
}

// not concurrent safe
func (o *OAuth2) getActiveSession(sessionId string) (*Config, bool) {

	v, ok := o.SessionsList[sessionId]
	return v, ok
}
