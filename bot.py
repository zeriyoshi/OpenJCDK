import os
import random
import tweepy

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

# Pick a picture randomly
directory = "./images"
file = random.choice([f for f in os.listdir(directory) if os.path.isfile(os.path.join(directory, f))])

# Upload picture and Tweet
media = api_1_1.media_upload(filename = "/".join([directory, file]))
client_2_0.create_tweet(media_ids = [media.media_id])
