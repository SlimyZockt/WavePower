package user_test

import (
	"server/internal/user"
	"slices"
	"testing"
)

var tracks = user.Tracks{
	{
		Title:  "Test 1",
		Artist: "Artist 1",
		Id:     "0",
	},
	{
		Title:  "Test 2",
		Artist: "Artist 2",
		Id:     "1",
	},
	{
		Title:  "Test 3",
		Artist: "Artist 3",
		Id:     "2",
	},
	{
		Title:  "Test 4",
		Artist: "Artist 4",
		Id:     "3",
	},
}

func TestMove(t *testing.T) {
	test_tracks := slices.Clone(tracks)

	to := 2
	from := 0

	test_tracks.Move(from, to)

	if test_tracks[to] != tracks[from] {
		t.Error("Track Move failed")
	}
}

func TestGetTrack(t *testing.T) {
	test_tracks := slices.Clone(tracks)

	id := "1"

	out := user.AudioTrack{
		Title:  "Test 2",
		Artist: "Artist 2",
		Id:     "1",
	}

	track, _ := test_tracks.GetTrack(id)

	if *track != out {
		t.Error("Get Track by id failed")
	}
}
