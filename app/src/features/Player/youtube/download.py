import argparse
import json
import logging

import yt_dlp

global_error = None


def duration_filter(max_duration: int):
    def filter_func(info, *, incomplete):
        duration = info.get('duration')
        if duration and duration > max_duration:
            err_string = 'The video is too long'

            global global_error
            global_error = err_string

            return err_string

    return filter_func


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-query", type=str, help="youtube query", required=True)
    parser.add_argument("-max_duration", type=int, help="maximum allowed video duration to download", required=True)
    args = parser.parse_args()

    youtube_query = args.query
    max_duration = args.max_duration

    logger = logging.getLogger("custom_logger")
    logger.setLevel(logging.ERROR)

    try:
        ydl_opts = {
            'format': 'bestaudio/best',
            'outtmpl': '%(id)s.opus',
            'default_search': 'ytsearch',
            'noplaylist': True,
            'logger': logger
        }

        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            search_results = ydl.extract_info(f"ytsearch:{youtube_query}", download=False)
            video_info = None
            if 'entries' in search_results:
                # Get the video ID and duration of the first search result
                video_info = search_results['entries'][0]
                video_info['filename'] = ydl.prepare_filename(video_info)

        ydl_opts['outtmpl'] = '%(id)s'
        ydl_opts['match_filter'] = duration_filter(max_duration)
        ydl_opts['postprocessors'] = [{  # Extract audio using ffmpeg
            'key': 'FFmpegExtractAudio',
            'preferredcodec': 'opus',
            'preferredquality': '192',
        }]

        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            ydl.download(youtube_query)

    except Exception as e:
        global global_error
        global_error = "Unknown error occured: " + str(e)

    query_result = {
        "video_info": video_info,
        "error": global_error,
    }
    print(json.dumps(query_result))


if __name__ == '__main__':
    main()
