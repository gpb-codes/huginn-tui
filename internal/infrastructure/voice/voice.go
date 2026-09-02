package voice

// Voice — TenVAD + MiMo-ASR streaming (Mimo)
type Config struct {
	Enabled bool
	ASR     string // mimo-asr, tenvad
}
