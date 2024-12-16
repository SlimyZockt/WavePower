package user

import (
	"errors"
	"slices"

	"github.com/markbates/goth"
)

type User struct {
	goth.User
	CurAudioID    string
	CurPlaylistID int
	Tracks        Tracks
}

type Tracks []AudioTrack

func (playlist Tracks) Move(from int, to int) {
	temp := playlist[from]
	playlist = slices.Delete(playlist, from, from+1)
	playlist = slices.Insert(playlist, to, temp)
}

func (playlist *Tracks) GetTrack(id string) (*AudioTrack, error) {
	for _, val := range *playlist {
		if val.Id == id {
			return &val, nil
		}
	}

	return nil, errors.New("No track found")
}

type AudioTrack struct {
	Title  string
	Artist string
	Id     string
}
