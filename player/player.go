package player

import (
	"context"
	"errors"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/ffmpeg-audio"
	"github.com/disgoorg/snowflake/v2"
	"google.golang.org/api/youtube/v3"
	"io"
)

var ErrNotConnected = errors.New("player not connected")
var ErrNoSkip = errors.New("no song to skip to")

type Song struct {
	stream  io.Reader
	snippet *youtube.VideoSnippet
	stopped bool
}

type Player struct {
	connected, playing bool
	connection         voice.Conn
	guildId, channelId snowflake.ID

	provider voice.OpusFrameProvider

	queue []*Song
}

func NewSong(stream io.Reader, snippet *youtube.VideoSnippet) *Song {
	return &Song{stream: stream, snippet: snippet}
}

var channels = make(map[snowflake.ID]*Player)

func GetPlayer(guildId snowflake.ID, mgr voice.Manager) *Player {
	if p, ok := channels[guildId]; ok {
		return p
	}
	conn := mgr.CreateConn(guildId)
	p := &Player{connection: conn, guildId: guildId}
	channels[guildId] = p
	return p
}

func (p *Player) Connect(channelId snowflake.ID) error {
	if p.connected {
		return nil
	}
	p.connected = true
	p.channelId = channelId
	return p.connection.Open(context.TODO(), channelId, false, true)
}

func (p *Player) AddToQueue(s *Song) {
	p.queue = append(p.queue, s)
}

func (p *Player) Playing() bool {
	return p.playing
}

func (p *Player) Channel() snowflake.ID {
	return p.channelId
}

func (p *Player) Connected() bool {
	return p.connected
}

func (p *Player) stopCurrent() {
	p.provider.Close()
	p.connection.SetOpusFrameProvider(nil)
	p.queue[0].stopped = true
}

func (p *Player) Skip() error {
	if !p.connected {
		return ErrNotConnected
	}
	if len(p.queue) < 2 {
		return ErrNoSkip
	}
	p.stopCurrent()
	p.queue = p.queue[1:]
	return p.play(p.queue[0])
}

func (p *Player) QueueLen() int {
	return len(p.queue)
}

func (p *Player) play(song *Song) error {
	f, err := ffmpeg.New(context.TODO(), song.stream)
	if err != nil {
		return err
	}

	p.connection.SetOpusFrameProvider(f)
	p.provider = f
	p.playing = true
	go func() {
		_ = f.Wait()

		if song.stopped {
			return
		}

		p.queue = p.queue[1:]

		if len(p.queue) == 0 {
			_ = p.connection.SetSpeaking(context.TODO(), voice.SpeakingFlagNone)
			p.Kill()
		} else {
			_ = p.play(p.queue[0])
		}
	}()

	return nil
}

func (p *Player) Play(song *Song) error {
	if err := p.connection.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone); err != nil {
		return err
	}

	p.queue = append([]*Song{song}, p.queue...)

	return p.play(song)
}

func (p *Player) Kill() {
	p.connection.Close(context.TODO())
	p.connected = false
	delete(channels, p.guildId)
}
