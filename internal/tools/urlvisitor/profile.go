package urlvisitor

import "math/rand"

func defaultUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36",
	}
}

func pickUserAgent(userAgents []string) string {
	if len(userAgents) == 0 {
		return defaultUserAgents()[0]
	}
	return userAgents[rand.Intn(len(userAgents))]
}
