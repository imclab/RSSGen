### The Problem
Turns out Google TV is pretty worthless as a network media player. The only out-of-box functionality is for UPnP DLNA media servers which are painful to work with at best.

A simple solution is to simply host a web server from some system on your network that is online whenever you'd like to watch media from it. From the directory listing of the media folder most files can be played directly in Chrome. Unfortunately the audio and media position bars don't minimize while playing media and the screensaver will come on intermittently while playing media in Chrome.

### The Solution
I quickly discovered that the Podcast app (formerly known as Queue) does a pretty good job of playing media enclosed in an RSS feed. It suspends timing for the screensaver to come on so you're never interrupted. The media position bar along with media controls fade away after a set amount of time. Overall it's better than playing media directly in the browser. You also get the added benefit of keeping track of what you've watched and when it was aired, you can also inject other useful information about the media.

This program will scan a given directory for media and lookup metadata from trakt.tv for tv shows. Information it pulls are things like the air date, episode title and description. This program was written in Google's general purpose programming language [Go](http://golang.org/).

### Networking

There are some tricky networking things we need to do to make the feed usable with the Podcast app. First, the Podcast app stores all of it's information using Google Reader, so in order to subscribe to the feed, the feed must be visible to Google Reader, which means it must be publicly available. In my network I've simply forwarded an arbitrary port from my router to my server, in this case TCP port 2223. I've then created a redirect on lighttpd the web server I'm using to serve the feed. Below is the configuration I'm using for the redirect to the feed folder.

	$SERVER["socket"] == ":2223" {
		server.document-root = "/home/username/path/to/feed/directory/"
	}

Having dynamic DNS for those of us without static WAN IP addresses is really helpful. Simply browse to your external IP address (or dynDNS) at port :2223 on your Google TV and subscribe to the feed. The address used in the configuration to tell the Podcast app where the media lives can and should be a local address such as:

	http://mediaserver/path/to/media/

Where mediaserver is a local hostname (or IP address, either will work) and the path to the media gives direct http access to the files of the media directory.

### Configuration
The sample configuration is stored in json format and is shown below:

	{
		"TraktAPI": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		"MediaPath": "/home/username/path/to/media",
		"FeedPath": "Feed/queue.xml",
		"Host": "Hostname_for_display_only",
		"MediaURL": "http://hostname_of_server/path/to/media"
	}

Adjust the parameters as you see fit. Mind you `Host` is only for aesthetics in the feed to identify what system the feed was generated on. `FeedPath` should be the directory publicly visible to the Google Reader service.

You must have an account with trakt.tv to use this program. Your trakt.tv api key can be found at: [http://trakt.tv/settings/api](http://trakt.tv/settings/api)