package peers

import (
	"os"
)

type Status string

const (
	Online  Status = "Online"
	Offline Status = "Offline"
	Pairing Status = "Pairing"
	Syncing Status = "Syncing"
)

type Peer struct {
	ID        string
	Hostname  string
	IP        string
	Status    Status
	VaultSync string
	Latency   string
	LastSeen  string
	Paired    bool
	Trusted   bool
}

var DefaultPeers = []Peer{
	{"peer_01JABC", "thinkpad-x1", "192.168.1.42", Online, "Synced", "18ms", "now", true, true},
	{"peer_02JDEF", "macbook-pro", "192.168.1.88", Syncing, "Syncing 64%", "42ms", "2m ago", true, true},
	{"peer_03JGHI", "desk-linux", "10.0.0.12", Offline, "Offline", "—", "1h ago", true, false},
	{"peer_04JJKL", "huginn-remote", "100.64.1.5", Pairing, "Pairing…", "—", "—", false, false},
}

var LocalPeerID = "peer_LOCAL_7F3A"

var LocalHostname = func() string {
	h, _ := os.Hostname()
	if h == "" {
		return "huginn-local"
	}
	return h
}()
