package youtube

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func Test_getVideoId(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid Url",
			args: args{
				uri: "https://www.youtube.com/watch?v=n5lt-y8RcVc",
			},
			want:    "n5lt-y8RcVc",
			wantErr: false,
		},
		{
			name: "No Video Id",
			args: args{
				uri: "https://www.youtube.com/watch?abc=n5lt-y8RcVc",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Malformed id query param",
			args: args{
				uri: "https://www.youtube.com/watch?v=n5lt-y8RcVc&v=testid",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Shorts Url",
			args: args{
				uri: "https://www.youtube.com/shorts/Pjj_tBNh0-I",
			},
			want:    "Pjj_tBNh0-I",
			wantErr: false,
		},
		{
			name: "Shorts Url with Query params",
			args: args{
				uri: "https://www.youtube.com/shorts/Pjj_tBNh0-I?si=3123123",
			},
			want:    "Pjj_tBNh0-I",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVideoId(tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVideoId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getVideoId() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloader_Download_TooLong(t *testing.T) {
	const testQuery = "https://www.youtube.com/watch?v=EIyixC9NsLI"

	queryResult, err := Download(testQuery, time.Second, os.TempDir())
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, "EIyixC9NsLI.opus", queryResult.VideoInfo.Filename)
	assert.Equal(t, "The video is too long", queryResult.Error)
}

func TestDownloader_Download_Ok(t *testing.T) {
	const testQuery = "https://www.youtube.com/watch?v=EIyixC9NsLI"

	queryResult, err := Download(testQuery, time.Second*600, os.TempDir())
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, "EIyixC9NsLI.opus", queryResult.VideoInfo.Filename)
	assert.Equal(t, "", queryResult.Error)
}

func TestDownloader_Download_ComplexUrl(t *testing.T) {
	const testQuery = "https://www.youtube.com/watch?v=7R4Aea6M-5Y&pp=ygUabGVnbyBzdGFyIHdhcnMgZGVhdGggc291bmQ%3D"

	queryResult, err := Download(testQuery, time.Second*600, os.TempDir())
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, "7R4Aea6M-5Y.opus", queryResult.VideoInfo.Filename)
	assert.Equal(t, "", queryResult.Error)
}

func TestDownloader_Download_ShortsUrl(t *testing.T) {
	const testQuery = "https://youtube.com/shorts/SDbm2E2o0-g?si=exLrOSlj90489xYW"

	queryResult, err := Download(testQuery, time.Second*600, os.TempDir())
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, "SDbm2E2o0-g.opus", queryResult.VideoInfo.Filename)
	assert.Equal(t, "", queryResult.Error)
}

func TestDownloader_Query(t *testing.T) {
	const testQuery = "https://www.youtube.com/watch?v=EIyixC9NsLI"

	queryResult, err := Query(testQuery)
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, "EIyixC9NsLI.opus", queryResult.VideoInfo.Filename)
}
