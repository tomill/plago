local must_env = std.native('must_env');

local list_flatten(s) =
  local lines = std.split(s, "\n");
  local uncomment = std.map(function(v) std.splitLimit(v, "#", 1)[0], lines);
  local items = std.map(function(v) std.split(v, ","), uncomment);
  std.join(" ", std.flattenArrays(items));

{
  FunctionName: 'plago',
  Description: 'xxxx to gmail',
  Role: must_env('AWS_LAMBDA_ROLE'),
  Runtime: 'nodejs24.x',
  Handler: 'index.handler',
  MemorySize: 256,
  Timeout: 900,
  Environment: {
    Variables: {
      TZ: 'Asia/Tokyo',
      LOG_FORMAT: 'json',
      BLUESKY_HANDLE: must_env('BLUESKY_HANDLE'),
      BLUESKY_APPPASS: must_env('BLUESKY_APPPASS'),
      DISCORD_TOKEN: must_env('DISCORD_TOKEN'),
      DISCORD_CHANNELS: list_flatten(must_env('DISCORD_CHANNELS')),
      FEEDREADER_TOKEN: must_env('FEEDREADER_TOKEN'),
      TWITTER_CONSUMER_KEY: must_env('TWITTER_CONSUMER_KEY'),
      TWITTER_CONSUMER_SECRET: must_env('TWITTER_CONSUMER_SECRET'),
      TWITTER_OAUTH1_TOKEN: must_env('TWITTER_OAUTH1_TOKEN'),
      TWITTER_OAUTH1_TOKEN_SECRET: must_env('TWITTER_OAUTH1_TOKEN_SECRET'),
      TWITTER_USERID: must_env('TWITTER_USERID'),
      TWITTER_LISTS: must_env('TWITTER_LISTS'),
      SLACK_TOKEN: must_env('SLACK_TOKEN'),
      SLACK_WORKSPACE: must_env('SLACK_WORKSPACE'),
      SLACK_CHANNELS: list_flatten(must_env('SLACK_CHANNELS')),
      YOUTUBE_API_KEY: must_env('YOUTUBE_API_KEY'),
      YOUTUBE_CHANNELS: list_flatten(must_env('YOUTUBE_CHANNELS')),
      GMAIL_ADDRESS: must_env('GMAIL_ADDRESS'),
      GMAIL_APPPASS: must_env('GMAIL_APPPASS'),
    }
  }
}
