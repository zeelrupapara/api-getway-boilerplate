package utils

import (
	"fmt"
	"strings"
	model "greenlync-api-gateway/model/common/v1"

	"github.com/mileusna/useragent"
)

type UserAgent struct {
	Device  string
	OS      string
	Channel model.ChannelType
}

func UserAgentParser(s string) *UserAgent {
	ua := useragent.Parse(s)

	var channel model.ChannelType
	if ua.Mobile || ua.Tablet {
		channel = model.ChannelType_Mobile
	} else if ua.IsChrome() || ua.IsEdge() || ua.IsFirefox() || ua.IsSafari() || ua.IsOpera() {
		channel = model.ChannelType_Web
	} else if strings.Contains(ua.Name, "Desktop") {
		channel = model.ChannelType_Desktop
	} else {
		channel = model.ChannelType_API
	}

	return &UserAgent{
		Device:  fmt.Sprintf("%s/%s", ua.Name, ua.Version),
		OS:      ua.OS,
		Channel: channel,
	}
}
