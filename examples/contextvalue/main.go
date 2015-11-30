package main

import (
	"fmt"
	"net/http"

	mv "github.com/delicb/mezvaro"
	"github.com/mssola/user_agent"
)

type userInfo int

const browserInfoKey userInfo = 1

type UserAgentExtractor struct {
	verbose bool
}

type BrowserInfo struct {
	verbose       bool
	Name          string
	Version       string
	OS            string
	Engine        string
	EngineVersion string
	Localization  string
	Platform      string
	Mobile        bool
}

func (bi BrowserInfo) String() string {
	if bi.verbose {
		return fmt.Sprintf(
			`Name: %s, Version: %s, OS: %s,
Engine: %s, Localization: %s,
Platform: %s, Mobile: %t`,
			bi.Name, bi.Version, bi.OS, bi.Engine, bi.Localization,
			bi.Platform, bi.Mobile,
		)
	} else {
		return fmt.Sprintf("Name: %s, Version: %s, OS: %s", bi.Name, bi.Version, bi.OS)
	}
}

func (uae UserAgentExtractor) Handle(c *mv.Context) {
	ua := user_agent.New(c.Request.Header["User-Agent"][0])
	name, version := ua.Browser()
	browserEngine, browserEngineVersion := ua.Engine()
	browserInfo := BrowserInfo{
		verbose:       uae.verbose,
		Name:          name,
		Version:       version,
		OS:            ua.OS(),
		Engine:        browserEngine,
		EngineVersion: browserEngineVersion,
		Localization:  ua.Localization(),
		Platform:      ua.Platform(),
		Mobile:        ua.Mobile(),
	}
	c.WithValue(browserInfoKey, browserInfo)
	c.Next()
}

func BrowserInfoHandler(c *mv.Context) {
	browserInfo := c.Value(browserInfoKey).(BrowserInfo)
	c.Response.Write([]byte(browserInfo.String()))
}

func main() {
	m := mv.New(UserAgentExtractor{verbose: true})
	http.Handle("/", m.HF(BrowserInfoHandler))
	http.ListenAndServe(":8000", nil)
}
