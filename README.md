# weatherstem-cli

When you just want a quick peek at the local weather, here's a tool to query weatherSTEM's API
and return just a short page of weather data. Or, you can return the full JSON that got returned
and mangle that all you want.

## Dependencies

   - github.com/loraxipam/compassrose
   - github.com/loraxipam/havers2

## Installation

Like most golang tools, you need at least 'go' 1.11 installed to compile it. Here's the easy
peasy steps that should get you running in five minutes, once Go is working.

```
go get github.com/loraxipam/weatherstem-cli
cd weatherstem-cli
go run weatherstem.go
<copy and paste the boilerplate config example JSON to your weatherstem.json file>
<edit the JSON per the below Setup directions>
go run weatherstem.go
```

If you want to compile the code and put it on your path, just run `go build weatherstem.go`
then move the `weatherstem` binary to your bin directory. If you have Go configured for your
machine, you could just run `go install` and it will put it in the usual $GOBIN directory.

## Setup

You'll need to create a small config file with your weatherstem.com API key and the local domain
URL that you will use. You might have one or two favorite stations to query, so you will need
their short name. Here's the example config file. Call it `weatherstem.json` and put it in your
`.config` directory. Or leave it in your current directory. Or put it in `~/.weatherstem.json`
and it will work.

```
{"version": "3.0",
"api_key": "yourKeyGoesHere",
"stations":
   ["firstOne@domain.weatherstem.com",
   "maybeTwo@domain.weatherstem.com",
   "someThird@domain.weatherstem.com"],
"api_url": "https://api.weatherstem.com/api",
"me": {"lat": 45.0, "lon": -123.0}}
```

FYI, if you run it with no config file, it will complain and show you an example as above. Cut
and paste for the win.

## Options

If you want to see it on the screen, just run it.  
If you want to output JSON, you can use the `-json` flag.  
If you want distances in kilometers, use `-kilo`; for miles use `-mile`.  
If you want compact output (few units), use `-lite`.  
If you want to see the full gory details of the complete API call, use the `-orig` flag.  
If you want boring compass rose directions, use `-rose`.  

```
  -json  Output cooked data as JSON
  -kilo  Output station distances in kilometers
  -lite  Output lightweight cooked data
  -mile  Output station distances in statute miles
  -orig  Output original API results
  -rose  Output boring compass rose directions
```

#### Notes

I use the alternate compass rose because I love to say the word "Tramontana."

For a list of the Wet Bulb Globe Temperature icons, just pass it any argument, for example, `weatherstem help` will list out the WBGT alert levels being used.

```
Current WBGT flags:
   <82°F       - normal
 ⚊ 82°F - 87°F - Level 1
 ⚌ 87°F - 90°F - Level 2
 ☰ 90°F - 92°F - Level 3
 ⚑ >92°F       - Level 4
```

The new (Aug 2020) v1 API station ID format allows querying stations in different domains using
the same config file. :raised_hands:
