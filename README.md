# Chicago Bill Bot

Twitter bot for following Chicago City Council legislation. Tweet a bill identifier at the bot and it will track the bill and post any updates.

## Setup

You'll need GNU Make, Go and node.js installed as well as credentials for AWS and the Twitter API. Copy `.env.example` to `.env` and fill in the blank values.

To install dependencies, build functions and deploy:

```bash
make install
make build
make deploy
```
