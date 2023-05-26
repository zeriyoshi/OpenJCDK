import io
import os
import random
import tweepy
import json
from googleapiclient.http import MediaIoBaseDownload
from googleapiclient.errors import HttpError
from googleapiclient.discovery import build
from google.oauth2 import service_account

# ----------------------------------------------------------------
# Google Drive
# ----------------------------------------------------------------
drive_service_account_key = os.environ["GOOGLE_DRIVE_SERVICE_ACCOUNT_KEY"]

drive = build('drive', 'v3', credentials = service_account.Credentials.from_service_account_info(
    json.loads(drive_service_account_key)).with_scopes(['https://www.googleapis.com/auth/drive']
))

items = []
page_token = None
while True:
    response = drive.files().list(
        q = "'1Gf1bJ03uo9NKo4mugHkVWlsOZaKtuqoq' in parents and trashed = false",
        spaces = "drive",
        fields = "nextPageToken, files(id)",
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
downloader = MediaIoBaseDownload(io.FileIO("temporary.jpg", "wb"), drive.files().get_media(fileId = id))
finished = False
while finished is False:
    _, finished = downloader.next_chunk()

# ----------------------------------------------------------------
# Twitter
# ----------------------------------------------------------------
consumer_key = os.environ["TWITTER_CONSUMER_KEY"]
consumer_secret = os.environ["TWITTER_CONSUMER_SECRET"]
access_token = os.environ["TWITTER_ACCESS_TOKEN"]
access_token_secret = os.environ["TWITTER_ACCESS_TOKEN_SECRET"]

oauth1 = tweepy.OAuthHandler(consumer_key, consumer_secret)
oauth1.set_access_token(access_token, access_token_secret)

# Create OAuth 1.1 API object for Media Upload
api_1_1 = tweepy.API(oauth1)

# Create OAuth 2.0 Client object for Tweet
client_2_0 = tweepy.Client(
    consumer_key = consumer_key, 
    consumer_secret = consumer_secret, 
    access_token = access_token, 
    access_token_secret = access_token_secret
)

# Upload picture and Tweet
media = api_1_1.media_upload("./temporary.jpg")
client_2_0.create_tweet(media_ids = [media.media_id])
