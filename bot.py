import io
import os
import pytz
import random
import tweepy
import json
from datetime import datetime
from html.parser import HTMLParser
from googleapiclient.http import MediaIoBaseDownload
from googleapiclient.discovery import build
from google.oauth2 import service_account

# ----------------------------------------------------------------
# Message
# ----------------------------------------------------------------
class MessageManager(HTMLParser):
    def __init__(self, data: str|None):
        super().__init__()
        self.entering_tag = ''
        self.tweet = None
        self.alt = None
        if data is not None:
            self.feed(data)
    
    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        self.entering_tag = tag
    
    def handle_endtag(self, tag: str) -> None:
        self.entering_tag = None
    
    def handle_data(self, data: str) -> None:
        match self.entering_tag:
            case 'tweet':
                self.tweet = data
            case 'alt':
                self.alt = data
    
    def generate(self) -> tuple[str, str | None]:
        if self.tweet is not None:
            return [self.tweet, self.alt]

        now = datetime.now(pytz.timezone('Asia/Tokyo'))
        message = os.environ.get('MESSAGE_FOOTER', '#邪神ちゃん今日の１枚 をどうぞ。')
        if 4 <= now.hour < 11:
            message = os.environ.get('MESSAGE_HEADER_MORNING', 'フォロワーの皆さま、おはようございます！ ') + message
        elif 11 <= now.hour < 15:
            message = os.environ.get('MESSAGE_HEADER_NOON', 'フォロワーの皆さま、ランチタイムです！ ') + message
        else :
            message = os.environ.get('MESSAGE_HEADER_NIGHT', 'フォロワーの皆さま、今日も１日おつかれさまでした。お休み前に ') + message
        
        return [message, self.alt]

# ----------------------------------------------------------------
# Google Drive
# ----------------------------------------------------------------
drive_service_account_key = os.environ['GOOGLE_DRIVE_SERVICE_ACCOUNT_KEY']
drive_directory_id = os.environ['GOOGLE_DRIVE_DIRECTORY_ID']

drive = build('drive', 'v3', credentials = service_account.Credentials.from_service_account_info(
    json.loads(drive_service_account_key)).with_scopes(['https://www.googleapis.com/auth/drive']
))

items = []
page_token = None
while True:
    response = drive.files().list(
        q = "'" + drive_directory_id + "' in parents and trashed = false",
        spaces = 'drive',
        fields = 'nextPageToken, files(id)',
        pageToken = page_token
    ).execute()
    for item in response.get('files'):
        items.append(item["id"])
    page_token = response.get('nextPageToken', None)
    if page_token is None:
        break

# Pick a picture randomly
id = random.choice(items)

# Download file from Google Drive
downloader = MediaIoBaseDownload(io.FileIO('temporary', 'wb'), drive.files().get_media(fileId = id))
finished = False
while finished is False:
    _, finished = downloader.next_chunk()

# Get a metadata from Google Drive
metadata = drive.files().get(fileId = id, fields = 'name, mimeType, description').execute()

# ----------------------------------------------------------------
# X (formerly Twitter)
# ----------------------------------------------------------------
message, alt = MessageManager(metadata.get('description', None)).generate()

# Setup tweepy
consumer_key = os.environ['TWITTER_CONSUMER_KEY']
consumer_secret = os.environ['TWITTER_CONSUMER_SECRET']
access_token = os.environ['TWITTER_ACCESS_TOKEN']
access_token_secret = os.environ['TWITTER_ACCESS_TOKEN_SECRET']

oauth1 = tweepy.OAuthHandler(consumer_key, consumer_secret)
oauth1.set_access_token(access_token, access_token_secret)

# Create OAuth 1.1 API object for Media Upload
api_1_1 = tweepy.API(oauth1)

# Create OAuth 2.0 Client object for Post
client_2_0 = tweepy.Client(
    consumer_key = consumer_key, 
    consumer_secret = consumer_secret, 
    access_token = access_token, 
    access_token_secret = access_token_secret
)

# Upload media
mime = metadata.get('mimeType', 'image/')
if mime.startswith('image/'):
    media = api_1_1.simple_upload('temporary')
    if alt is not None:
        api_1_1.create_media_metadata(media.media_id, alt)
else:
    media = api_1_1.chunked_upload('temporary', file_type = mime)

# Post
client_2_0.create_tweet(text = message, media_ids = [media.media_id])
