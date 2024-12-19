# WavePower

WavePower is a demo audio player that allows the uploading and streaming audio files.

## Setup

Generate a Google Cloud OAuth 2.0-Client-ID and set the callback url to something like this: `https://domain.com/auth/google/callback"`.
Create a .env file with your google ID & secret like this:
```bash
GOOGLE_CLIENT_ID="420.googleuser.com"
GOOGLE_CLIENT_SECRET="s3cR3t-69"
CALLBACK_LINK="https://localhost:8080/auth/google/callback""
```
In addition genearate a TlS KEY whith `openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650` and add both files to the repo.


At last you can run the application with:
```bash
docker compose up --build
```
*by default the server runs on the Port `:8080`*

