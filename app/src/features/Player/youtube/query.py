import argparse
import json

import yt_dlp

# Initialize the yt-dlp downloader
ydl_opts = {
    'outtmpl': '%(id)s.%(ext)s',
    'quiet': True
}


def search_get_video_data(search_query):
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        try:
            # Search for the video
            search_results = ydl.extract_info(f"ytsearch:{search_query}", download=False)

            if 'entries' in search_results:
                # Get the video ID and duration of the first search result
                video_info = search_results['entries'][0]
                video_info['filename'] = ydl.prepare_filename(video_info)

                return video_info, None
            else:
                print("No search results found.")
                return None, None

        except yt_dlp.DownloadError as e:
            return None, str(e)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-query", type=str, help="youtube query", required=True)
    args = parser.parse_args()

    search_query = args.query
    video_info, error = None, None

    try:
        video_info, error = search_get_video_data(search_query)
    except Exception as e:
        error = "Unknown error occured: " + str(e)

    query_result = {
        "video_info": video_info,
        "error": error,
    }
    print(json.dumps(query_result))


if __name__ == "__main__":
    main()
