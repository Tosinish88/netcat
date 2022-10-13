// keep logs of current chat in the in-running memory
package netcat

// append new chat to the log
var Chathistory = []string{}

func AddHistory(textmsg string) {
	Chathistory = append(Chathistory, textmsg)
}
