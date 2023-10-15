import argparse
import json

import yt_dlp

# Initialize the yt-dlp downloader
ydl_opts = {
    'quiet': True
}


def search_video_and_get_id_duration(search_query):
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        try:
            # Search for the video
            search_results = ydl.extract_info(f"ytsearch:{search_query}", download=False)

            if 'entries' in search_results:
                # Get the video ID and duration of the first search result
                video_info = search_results['entries'][0]
                video_id = video_info['id']
                video_duration = video_info['duration']
                return video_id, video_duration, None
            else:
                print("No search results found.")
                return "", 0, None

        except yt_dlp.DownloadError as e:
            return "", 0, str(e)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-query", type=str, help="youtube query", required=True)
    args = parser.parse_args()

    search_query = args.query
    video_id, video_duration, error = None, None, None

    try:
        video_id, video_duration, error = search_video_and_get_id_duration(search_query)

    finally:
        return_struct = {
            "video": {
                "id": video_id,
                "duration": video_duration,
            },
            "error": error,
        }
        print(json.dumps(return_struct))


if __name__ == "__main__":
    main()
