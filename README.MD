# WavePower

WavePower is a demo audio player that allows the uploading and streaming audio files.

## Setup

Generate a Google Cloud OAuth 2.0-Client-ID and set the callback url to something like this: `https://domain.com/auth/google/callback"`.
Create a .env file with your google ID & secret like this:
```bash
GOOGLE_CLIENT_ID="420.googleuser.com"
GOOGLE_CLIENT_SECRET="s3cR3t-69"
CALLBACK_LINK="https://localhost:8080/auth/google/callback"
```

You also need to add your Turso auth Token and URL.

```bash
TURSO_DATABASE_URL="libsql://name.location.turso.io"
TURSO_AUTH_TOKEN="key"
```

You can add DEV to enable dev mode
```bash
DEV=""
```
*by default the server runs on the Port `:8080`*

