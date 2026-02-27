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

	// Load base layer and additional layers based on genre
	baseData := e.getMusicData(name, 0)
	if baseData == nil {
		return nil
	}

	basePlayer, err := e.createPlayer(baseData)
	if err != nil {
		return err
	}
	basePlayer.SetVolume(1.0)
	basePlayer.Play()
	e.musicLayers = append(e.musicLayers, basePlayer)

	// Load intensity layers (up to 3 additional layers)
	for i := 1; i <= 3; i++ {
		layerData := e.getMusicData(name, i)
		if layerData == nil {
			break
		}

		layerPlayer, err := e.createPlayer(layerData)
		if err != nil {
			continue
		}

		// Calculate layer volume based on intensity
		layerVolume := e.calculateLayerVolume(i, e.intensity)
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

// getMusicData returns embedded music data for a track and layer.
// In a full implementation, this would use //go:embed directives.
func (e *Engine) getMusicData(name string, layer int) []byte {
	// Stub: return generated silence for now
	return generateSilence(sampleRate * 2)
}

// getSFXData returns embedded SFX data by name.
// In a full implementation, this would use //go:embed directives.
func (e *Engine) getSFXData(name string) []byte {
	// Stub: return generated blip for now
	return generateBlip(sampleRate / 10)
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
