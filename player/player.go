package player

import (
	"context"
	"errors"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/oq-x/ffmpeg-audio"
	"io"
)

var ErrNotConnected = errors.New("player not connected")
var ErrNoSkip = errors.New("no song to skip to")

type Song struct {
	stream  io.Reader
	Snippet *ytdl.Video
	stopped bool
}

type Player struct {
	connected, playing                        bool
	connection                                voice.Conn
	guildId, channelId, announcementChannelId snowflake.ID

	provider *ffmpeg.AudioProvider

	queue      []*Song
	NowPlaying chan Song
}

func NewSong(stream io.Reader, snippet *ytdl.Video) *Song {
	return &Song{stream: stream, Snippet: snippet}
}

var players = make(map[snowflake.ID]*Player)

func GetPlayer(guildId snowflake.ID, mgr voice.Manager) *Player {
	if p, ok := players[guildId]; ok {
		return p
	}
	conn := mgr.CreateConn(guildId)
	p := &Player{connection: conn, guildId: guildId, NowPlaying: make(chan Song, 1)}
	players[guildId] = p
	return p
}

func HasPlayer(guildId snowflake.ID) bool {
	_, ok := players[guildId]
	return ok
}

func (p *Player) Connect(channelId, announcementChannelId snowflake.ID) error {
	if p.connected {
		return nil
	}
	p.connected = true
	p.channelId = channelId
	p.announcementChannelId = announcementChannelId
	return p.connection.Open(context.TODO(), channelId, false, true)
}

func (p *Player) AddToQueue(s *Song) {
	p.queue = append(p.queue, s)
}

func (p *Player) Pause() {
	p.provider.Paused = true
}

func (p *Player) Resume() {
	p.provider.Paused = false
}

func (p *Player) AnnouncementChannel() snowflake.ID {
	return p.announcementChannelId
}

func (p *Player) Paused() bool {
	return p.provider.Paused
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

func (p *Player) Queue() []*Song {
	return p.queue
}

func (p *Player) play(song *Song) error {
	f, err := ffmpeg.New(context.TODO(), song.stream)
	if err != nil {
		return err
	}

	p.NowPlaying <- *song
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
	p.playing = false
	delete(players, p.guildId)
	close(p.NowPlaying)
}
