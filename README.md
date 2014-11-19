Twitterweizen
==============

<img src='https://magnum.travis-ci.com/knspriggs/twitterweizen.svg?token=zZCoL2DxeY3FuDqHfbp7&branch=master'/>

# Setup

## Credentials
There are two ways to pass your credentials to Twitterweizen:
#### CREDENTIALS file
This file must be in the same directory as the application.
Example:
```
<oauth_consumer_key>
<oauth_consumer_secret>
<oauth_token>
<oauth_token_secret>
```

#### Environment Variables
provide the following environment variables to the application:
- `OAUTH_CONSUMER_KEY`
- `OAUTH_CONSUMER_SECRET`
- `OAUTH_TOKEN`
- `OAUTH_TOKEN_SECRET`

## Environment Variables
You must also pass your twitter handle as an environment variable using `TWITTER_USER_NAME`.

# Use
## Asking a Question
To ask a question you must include one of the following hashtags in your question tweet:
`#yesno` or `#yesorno`

## Voting on a Question
For votes to be tracked your responders must reply to your original tweet specifying the appropriate hashtag:
`#yes` or `#no`
