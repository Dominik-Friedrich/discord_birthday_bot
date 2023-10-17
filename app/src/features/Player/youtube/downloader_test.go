package youtube

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
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

func TestDownloader_GetAudio(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "aa",
			args: args{
				url: "https://www.youtube.com/watch?v=n5lt-y8RcVc",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Downloader{}
			got, err := d.GetAudio(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAudio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAudio() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloader_Query(t *testing.T) {
	dl := Downloader{}
	const testQuery = "badgers"

	queryResult, err := dl.Query(testQuery)
	assert.Nil(t, err)
	assert.NotNil(t, queryResult)
}
