// Package audio manages sound effects and music playback with adaptive music layers.
package audio

import (
	"bytes"
	"io"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 48000

var (
	sharedContext     *audio.Context
	sharedContextOnce sync.Once
)

// getAudioContext returns the shared audio context.
func getAudioContext() *audio.Context {
	sharedContextOnce.Do(func() {
		sharedContext = audio.NewContext(sampleRate)
	})
	return sharedContext
}

// Engine handles audio playback with adaptive music intensity and 3D positioning.
type Engine struct {
	musicLayers []*audio.Player
	sfxPlayers  map[string]*audio.Player
	intensity   float64
	genreID     string
	listenerX   float64
	listenerY   float64
	mu          sync.RWMutex
}

// NewEngine creates a new audio engine.
func NewEngine() *Engine {
	return &Engine{
		sfxPlayers: make(map[string]*audio.Player),
		intensity:  0.5,
	}
}

// PlayMusic loads and plays a base music track with additional intensity layers.
// intensity parameter (0.0-1.0) crossfades additional layers on top of the base track.
func (e *Engine) PlayMusic(name string, intensity float64) error {
	// Get genre outside of lock to avoid deadlock
	e.mu.RLock()
	genreID := e.genreID
	e.mu.RUnlock()

	// Generate all music data before acquiring write lock
	baseData := generateMusicForEngine(name, 0, genreID)
	if baseData == nil {
		return nil
	}

	layerDataSlice := make([][]byte, 0, 3)
	for i := 1; i <= 3; i++ {
		layerData := generateMusicForEngine(name, i, genreID)
		if layerData == nil {
			break
		}
		layerDataSlice = append(layerDataSlice, layerData)
	}

	// Now acquire write lock for player management
	e.mu.Lock()
	defer e.mu.Unlock()

	e.intensity = clamp(intensity, 0.0, 1.0)

	// Stop previous music
	for _, player := range e.musicLayers {
		if player != nil {
			player.Pause()
		}
	}
	e.musicLayers = nil

	// Create base player
	basePlayer, err := e.createPlayer(baseData)
	if err != nil {
		return err
	}
	basePlayer.SetVolume(1.0)
	basePlayer.Play()
	e.musicLayers = append(e.musicLayers, basePlayer)

	// Create intensity layer players
	for i, layerData := range layerDataSlice {
		layerPlayer, err := e.createPlayer(layerData)
		if err != nil {
			continue
		}

		// Calculate layer volume based on intensity (i+1 because layer 0 is base)
		layerVolume := e.calculateLayerVolume(i+1, e.intensity)
		layerPlayer.SetVolume(layerVolume)
		layerPlayer.Play()
		e.musicLayers = append(e.musicLayers, layerPlayer)
	}

	return nil
}

// SetIntensity adjusts music intensity dynamically (0.0-1.0).
func (e *Engine) SetIntensity(intensity float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.intensity = clamp(intensity, 0.0, 1.0)

	// Update layer volumes
	for i := 1; i < len(e.musicLayers); i++ {
		if e.musicLayers[i] != nil {
			volume := e.calculateLayerVolume(i, e.intensity)
			e.musicLayers[i].SetVolume(volume)
		}
	}
}

// PlaySFX plays a sound effect by name with 3D positioning.
func (e *Engine) PlaySFX(name string, x, y float64) error {
	e.mu.RLock()
	listenerX, listenerY := e.listenerX, e.listenerY
	e.mu.RUnlock()

	sfxData := e.getSFXData(name)
	if sfxData == nil {
		return nil
	}

	player, err := e.createPlayer(sfxData)
	if err != nil {
		return err
	}

	// Apply 3D positional audio
	distance := math.Sqrt((x-listenerX)*(x-listenerX) + (y-listenerY)*(y-listenerY))
	volume := e.calculateVolume(distance)
	pan := e.calculatePan(x - listenerX)

	player.SetVolume(volume)
	// Ebitengine doesn't have SetPan, so we simulate it by adjusting volume
	// In a full implementation, we'd use a custom audio stream for stereo panning
	_ = pan

	player.Play()
	return nil
}

// SetListenerPosition updates the 3D audio listener position.
func (e *Engine) SetListenerPosition(x, y float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listenerX = x
	e.listenerY = y
}

// SetGenre configures the audio engine for a specific genre.
func (e *Engine) SetGenre(genreID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.genreID = genreID
}

// calculateLayerVolume computes volume for a given music layer based on intensity.
// Layer 1 fades in from 0.3-0.6 intensity
// Layer 2 fades in from 0.5-0.8 intensity
// Layer 3 fades in from 0.7-1.0 intensity
func (e *Engine) calculateLayerVolume(layer int, intensity float64) float64 {
	switch layer {
	case 1:
		return smoothstep(0.3, 0.6, intensity)
	case 2:
		return smoothstep(0.5, 0.8, intensity)
	case 3:
		return smoothstep(0.7, 1.0, intensity)
	default:
		return 0.0
	}
}

// calculateVolume applies distance attenuation (inverse square law).
func (e *Engine) calculateVolume(distance float64) float64 {
	if distance < 0.1 {
		return 1.0
	}
	// Inverse square law with minimum threshold
	attenuation := 1.0 / (1.0 + distance*distance*0.1)
	return clamp(attenuation, 0.0, 1.0)
}

// calculatePan computes stereo pan from horizontal offset (-1.0 left, +1.0 right).
func (e *Engine) calculatePan(dx float64) float64 {
	// Pan based on horizontal distance
	pan := dx / 10.0
	return clamp(pan, -1.0, 1.0)
}

// createPlayer creates an audio player from PCM data.
func (e *Engine) createPlayer(data []byte) (*audio.Player, error) {
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	ctx := getAudioContext()
	player, err := ctx.NewPlayer(stream)
	if err != nil {
		return nil, err
	}

	return player, nil
}

// getMusicData generates procedural music data for a track and layer.
// Returns deterministic audio based on name and layer parameters.
func (e *Engine) getMusicData(name string, layer int) []byte {
	e.mu.RLock()
	genreID := e.genreID
	e.mu.RUnlock()
	return generateMusicForEngine(name, layer, genreID)
}

// generateMusicForEngine is a helper that generates music without locking.
func generateMusicForEngine(name string, layer int, genreID string) []byte {
	// Incorporate genre into seed for variety
	seed := hashString(name+genreID) + uint64(layer)*1000
	duration := sampleRate * 3 // 3 seconds per layer (optimized for performance)
	return generateMusic(seed, duration, genreID, layer)
}

// getSFXData generates procedural SFX data by name.
// Returns deterministic audio based on name parameter.
func (e *Engine) getSFXData(name string) []byte {
	seed := hashString(name)
	return generateSFX(seed, name)
}

// generateSilence creates a silent WAV buffer for testing.
func generateSilence(samples int) []byte {
	buf := &bytes.Buffer{}
	// WAV header (44 bytes)
	buf.Write([]byte("RIFF"))
	writeUint32(buf, uint32(36+samples*4))
	buf.Write([]byte("WAVE"))
	buf.Write([]byte("fmt "))
	writeUint32(buf, 16)
	writeUint16(buf, 1) // PCM
	writeUint16(buf, 2) // Stereo
	writeUint32(buf, sampleRate)
	writeUint32(buf, sampleRate*4)
	writeUint16(buf, 4)
	writeUint16(buf, 16)
	buf.Write([]byte("data"))
	writeUint32(buf, uint32(samples*4))
	// Silence samples
	for i := 0; i < samples*2; i++ {
		writeUint16(buf, 0)
	}
	return buf.Bytes()
}

// generateBlip creates a simple tone WAV buffer for testing.
func generateBlip(samples int) []byte {
	buf := &bytes.Buffer{}
	// WAV header
	buf.Write([]byte("RIFF"))
	writeUint32(buf, uint32(36+samples*4))
	buf.Write([]byte("WAVE"))
	buf.Write([]byte("fmt "))
	writeUint32(buf, 16)
	writeUint16(buf, 1)
	writeUint16(buf, 2)
	writeUint32(buf, sampleRate)
	writeUint32(buf, sampleRate*4)
	writeUint16(buf, 4)
	writeUint16(buf, 16)
	buf.Write([]byte("data"))
	writeUint32(buf, uint32(samples*4))
	// Generate simple tone
	for i := 0; i < samples; i++ {
		val := int16(math.Sin(float64(i)*0.1) * 8000)
		writeUint16(buf, uint16(val))
		writeUint16(buf, uint16(val))
	}
	return buf.Bytes()
}

func writeUint32(w io.Writer, v uint32) {
	w.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func writeUint16(w io.Writer, v uint16) {
	w.Write([]byte{byte(v), byte(v >> 8)})
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// smoothstep provides smooth interpolation between 0 and 1.
func smoothstep(edge0, edge1, x float64) float64 {
	t := clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

// hashString generates a deterministic seed from a string.
func hashString(s string) uint64 {
	h := uint64(0xcbf29ce484222325)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 0x100000001b3
	}
	return h
}

// generateMusic creates procedural music with genre-specific characteristics.
func generateMusic(seed uint64, samples int, genreID string, layer int) []byte {
	rng := newLocalRNG(seed)

	// Genre-specific parameters
	tempo := 120.0
	scale := []int{0, 2, 4, 5, 7, 9, 11} // Major scale default
	baseNote := 48                       // C3

	switch genreID {
	case "fantasy":
		tempo = 100.0
		scale = []int{0, 2, 3, 5, 7, 8, 10} // Natural minor
		baseNote = 55                       // G3
	case "scifi":
		tempo = 128.0
		scale = []int{0, 2, 4, 6, 7, 9, 11} // Lydian mode
		baseNote = 60                       // C4
	case "horror":
		tempo = 80.0
		scale = []int{0, 1, 3, 5, 6, 8, 10} // Locrian mode
		baseNote = 36                       // C2
	case "cyberpunk":
		tempo = 140.0
		scale = []int{0, 2, 3, 5, 7, 8, 10} // Natural minor
		baseNote = 52                       // E3
	case "postapoc":
		tempo = 90.0
		scale = []int{0, 2, 3, 5, 7, 8, 11} // Harmonic minor
		baseNote = 43                       // G2
	}

	// Layer-specific intensity
	complexity := float64(layer + 1)
	noteDensity := 0.3 + float64(layer)*0.15

	beatLength := int(float64(sampleRate) * 60.0 / tempo)
	numBeats := samples / beatLength

	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	pcmData := make([]int16, samples*2)

	// Generate melodic content
	for beat := 0; beat < numBeats; beat++ {
		if rng.Float64() < noteDensity {
			noteIdx := rng.Intn(len(scale))
			midiNote := baseNote + scale[noteIdx] + (rng.Intn(2))*12
			freq := midiToFreq(midiNote)

			startSample := beat * beatLength
			noteLen := beatLength / 2
			if startSample+noteLen > samples {
				noteLen = samples - startSample
			}

			// Layer 0: Simple sine wave (bass)
			// Layer 1: Add harmonics (pad)
			// Layer 2: Add higher octave (lead)
			// Layer 3: Add percussion
			harmonics := []float64{1.0}
			harmAmps := []float64{1.0}

			if layer >= 1 {
				harmonics = append(harmonics, 2.0, 3.0)
				harmAmps = append(harmAmps, 0.3, 0.15)
			}
			if layer >= 2 {
				harmonics = append(harmonics, 4.0)
				harmAmps = append(harmAmps, 0.08)
			}

			for i := 0; i < noteLen; i++ {
				sampleIdx := startSample + i
				if sampleIdx >= samples {
					break
				}

				// ADSR envelope
				env := adsrEnvelope(i, noteLen, 0.05, 0.1, 0.6, 0.2)

				val := 0.0
				for h := 0; h < len(harmonics); h++ {
					val += math.Sin(2*math.Pi*freq*harmonics[h]*float64(sampleIdx)/float64(sampleRate)) * harmAmps[h]
				}
				val *= env * 3000.0 / complexity

				pcmData[sampleIdx*2] += int16(val)
				pcmData[sampleIdx*2+1] += int16(val)
			}
		}

		// Layer 3: Add percussion
		if layer >= 3 && beat%2 == 0 {
			startSample := beat * beatLength
			kickLen := sampleRate / 50
			if startSample+kickLen > samples {
				kickLen = samples - startSample
			}

			for i := 0; i < kickLen; i++ {
				sampleIdx := startSample + i
				if sampleIdx >= samples {
					break
				}

				env := math.Exp(-float64(i) / float64(kickLen/3))
				freq := 80.0 * math.Exp(-float64(i)/float64(kickLen/10))
				val := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 5000.0

				pcmData[sampleIdx*2] += int16(val)
				pcmData[sampleIdx*2+1] += int16(val)
			}
		}
	}

	// Write PCM data
	for i := 0; i < len(pcmData); i++ {
		writeInt16(buf, pcmData[i])
	}

	return buf.Bytes()
}

// generateSFX creates procedural sound effects based on name.
func generateSFX(seed uint64, name string) []byte {
	// Categorize SFX by name pattern
	if containsAny(name, "gun", "shoot", "fire", "pistol", "rifle", "shotgun") {
		return generateGunshot(seed)
	}
	if containsAny(name, "step", "walk", "foot") {
		return generateFootstep(seed)
	}
	if containsAny(name, "door", "open", "close") {
		return generateDoorSound(seed)
	}
	if containsAny(name, "explosion", "boom", "blast") {
		return generateExplosion(seed)
	}
	if containsAny(name, "pickup", "item", "collect") {
		return generatePickup(seed)
	}
	if containsAny(name, "pain", "hurt", "damage") {
		return generatePainSound(seed)
	}
	if containsAny(name, "reload", "click") {
		return generateReload(seed)
	}

	// Default: generate a tone
	return generateBlip(sampleRate / 10)
}

// generateGunshot creates a gunshot sound effect.
func generateGunshot(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 10
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		env := math.Exp(-float64(i) / float64(samples/15))

		// Mix of noise and sine for gunshot character
		noise := (rng.Float64()*2.0 - 1.0) * 0.7
		freq := 120.0 * math.Exp(-float64(i)/float64(samples/20))
		tone := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * 0.3

		val := int16((noise + tone) * env * 20000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// generateFootstep creates a footstep sound effect.
func generateFootstep(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 8
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		env := math.Exp(-float64(i) / float64(samples/5))
		noise := (rng.Float64()*2.0 - 1.0) * env * 8000.0

		val := int16(noise)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// generateDoorSound creates a door open/close sound.
func generateDoorSound(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 2
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samples)
		env := math.Sin(t * math.Pi)

		freq := 80.0 + t*40.0
		tone := math.Sin(2 * math.Pi * freq * float64(i) / float64(sampleRate))
		noise := (rng.Float64()*2.0 - 1.0) * 0.3

		val := int16((tone*0.7 + noise) * env * 10000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// generateExplosion creates an explosion sound effect.
func generateExplosion(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		env := math.Exp(-float64(i) / float64(samples/4))

		noise := (rng.Float64()*2.0 - 1.0) * 0.8
		freq := 60.0 * math.Exp(-float64(i)/float64(samples/10))
		rumble := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * 0.2

		val := int16((noise + rumble) * env * 25000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// generatePickup creates an item pickup sound.
func generatePickup(seed uint64) []byte {
	samples := sampleRate / 6
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samples)
		env := 1.0 - t

		freq := 440.0 * (1.0 + t*0.5)
		val := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 12000.0

		writeInt16(buf, int16(val))
		writeInt16(buf, int16(val))
	}

	return buf.Bytes()
}

// generatePainSound creates a pain/hurt sound effect.
func generatePainSound(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 4
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samples)
		env := math.Sin(t * math.Pi)

		freq := 300.0 - t*100.0
		tone := math.Sin(2 * math.Pi * freq * float64(i) / float64(sampleRate))
		noise := (rng.Float64()*2.0 - 1.0) * 0.2

		val := int16((tone*0.8 + noise) * env * 10000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// generateReload creates a reload/mechanical click sound.
func generateReload(seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 5
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	for i := 0; i < samples; i++ {
		env := 0.0
		if i < samples/10 {
			env = math.Exp(-float64(i) / float64(samples/50))
		}

		noise := (rng.Float64()*2.0 - 1.0) * env * 15000.0
		val := int16(noise)

		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// adsrEnvelope generates an ADSR (Attack-Decay-Sustain-Release) envelope.
func adsrEnvelope(sample, totalSamples int, attack, decay, sustain, release float64) float64 {
	t := float64(sample) / float64(totalSamples)

	attackTime := attack
	decayTime := attack + decay
	releaseTime := 1.0 - release

	if t < attackTime {
		return t / attackTime
	}
	if t < decayTime {
		return 1.0 - (1.0-sustain)*(t-attackTime)/decay
	}
	if t < releaseTime {
		return sustain
	}
	return sustain * (1.0 - (t-releaseTime)/release)
}

// midiToFreq converts MIDI note number to frequency in Hz.
func midiToFreq(midi int) float64 {
	return 440.0 * math.Pow(2.0, float64(midi-69)/12.0)
}

// writeWAVHeader writes a WAV file header for stereo 16-bit PCM.
func writeWAVHeader(buf *bytes.Buffer, samples int) {
	buf.Write([]byte("RIFF"))
	writeUint32(buf, uint32(36+samples*4))
	buf.Write([]byte("WAVE"))
	buf.Write([]byte("fmt "))
	writeUint32(buf, 16)
	writeUint16(buf, 1) // PCM
	writeUint16(buf, 2) // Stereo
	writeUint32(buf, sampleRate)
	writeUint32(buf, sampleRate*4)
	writeUint16(buf, 4)
	writeUint16(buf, 16)
	buf.Write([]byte("data"))
	writeUint32(buf, uint32(samples*4))
}

// writeInt16 writes a signed 16-bit integer in little-endian.
func writeInt16(w io.Writer, v int16) {
	w.Write([]byte{byte(v), byte(v >> 8)})
}

// containsAny checks if string s contains any of the given substrings.
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// localRNG is a simple deterministic RNG for audio generation.
type localRNG struct {
	state uint64
}

func newLocalRNG(seed uint64) *localRNG {
	return &localRNG{state: seed}
}

func (r *localRNG) Intn(n int) int {
	return int(r.Uint64() % uint64(n))
}

func (r *localRNG) Float64() float64 {
	return float64(r.Uint64()&0xFFFFFFF) / float64(0xFFFFFFF)
}

func (r *localRNG) Uint64() uint64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return r.state
}
