package youtube

import (
	"context"
	"errors"
	"fmt"
	ytdl "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"io"
	"net/url"
	"strings"
)

var ErrAudioStream = errors.New("failed to find audio stream")

var client ytdl.Client
var service *youtube.Service

func InitService(youtubeKey string) (err error) {
	service, err = youtube.NewService(context.TODO(), option.WithAPIKey(youtubeKey))
	return
}

func SearchVideos(query string) (*youtube.SearchListResponse, error) {
	call := service.Search.List([]string{"snippet"}).Q(query).MaxResults(10).Type("video")
	res, err := call.Do()
	return res, err
}

func GetVideo(id string) (*ytdl.Video, io.Reader, error) {
	video, err := client.GetVideo(id)
	if err != nil {
		return nil, nil, err
	}

	formats := video.Formats.WithAudioChannels()
	format, err := findOpusFormat(formats)
	if err != nil {
		return nil, nil, err
	}
	stream, _, _ := client.GetStream(video, format)

	return video, stream, nil
}

func findOpusFormat(formats ytdl.FormatList) (*ytdl.Format, error) {
	for _, format := range formats {
		if strings.Contains(format.MimeType, `codecs="opus"`) {
			return &format, nil
		}
	}
	return nil, ErrAudioStream
}

func GetVideoIDFromURL(query string) (string, error) {
	uri, err := url.ParseRequestURI(query)
	if err != nil {
		return "", err
	}
	if id := uri.Query().Get("v"); (uri.Host == "youtube.com" || uri.Host == "www.youtube.com") && uri.Path == "/watch" && id != "" {
		return id, nil
	}
	if uri.Host == "youtu.be" && uri.Path != "" {
		return uri.Path[1:], nil
	}
	return "", fmt.Errorf("no can do")
}

/*func ParseVideoDuration(s string) (time.Duration, error) {
	s = strings.TrimPrefix(s, "PT")
	var (
		hour, minute, second int
	)
	currentNumber := ""
	for _, c := range s {
		if c >= '0' && c <= '9' {
			currentNumber += string(c)
		} else {
			switch c {
			case 'S':
				second, _ = strconv.Atoi(currentNumber)
				currentNumber = ""
			case 'M':
				minute, _ = strconv.Atoi(currentNumber)
				currentNumber = ""
			case 'H':
				hour, _ = strconv.Atoi(currentNumber)
				currentNumber = ""
			default:
				return 0, fmt.Errorf("invalid unit %c", c)
			}
		}
	}
	return time.Duration((hour * 3600000000000) + (minute * 60000000000) + (second * 1000000000)), nil
}
*/
