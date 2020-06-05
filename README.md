# weatherstem-cli

When you just want a quick peek at the local weather, here's a tool to query weatherSTEM's API and return just a short page of weather data. Or, you can return the full JSON that got returned and mangle that all you want.

## Dependencies

   - github.com/loraxipam/compassrose
   - github.com/loraxipam/haversine

## Setup

You'll need to create a small config file with your weatherstem.com API key and the local domain URL that you will use. You might have one or two favorite stations to query, so you will need their short name. Here's the example config file. Call it `weatherstem.json` and put it in your `.config` directory. Or leave it in your current directory. Or put it in `~/.weatherstem.json` and it will work.

   `{"version":"1.0","api_key":"yourKeyGoesHere","stations":["firstOne","maybeTwo","doubtfulThree"],"api_url":"https://domain.weatherstem.com"}`

FYI, if you run it with no config file, it will complain and show you an example as above. Cut and paste for the win.

## Options

If you want to just see it on the screen, just run it. If you want to see the full gory details of the complete API call, use the `-orig` flag. If for some reason you want to see the cmd line version in JSON, you can use the `-json` flag. If you want boring compass rose directions, use `-rose`.

#### Notes

I included haversine because I plan on showing distances between the stations and my location in future versions.

I use the alternate compass rose because I love to say the word "Tramontana."

If you want to query stations in different domains, you'll need two directories with separate config files.
