# Scripts to gather info on hanging connections to api.stripe.com

We are seeing issues with commands in curl hanging. Initially I was testing against https://api.stripe.com/, but based on [Stripe's reachability script](https://github.com/stripe/stripe-reachability) I updated the endpoint to https://api.stripe.com/healthcheck . The `check_one_time_things.sh` script is a subset of the stripe-reachability script. The rest was written by me. The stripe-reachability script is good, but it runs everything sequentially, so for a hanging connection issue (like we're encountering), you don't see any information about network status while the issue is happening.

With the scripts in this repo, you can run all three of the active scripts at the same time and they'll gather info in parallel.  They are also set up to only output information when there is an unusually long curl request happening.

# The scripts

+ [`delay.sh`](delay.sh) - Allows for a common delay time in loops
+ [`check_curl_https.sh`](check_curl_https.sh) - Touches a sentinel file, then runs a curl against [the Stripe API health check endpoint](https://api.stripe.com/healthcheck).
+ [`check_mtr.sh`](check_mtr.sh) - If the sentinel file is older than 1 second, it runs `mtr` to show the route(s) between this host and `stripe.api.com`. The options are a count of 4 packets, do not do DNS resolution of the intervening hops, and format output as a report (instead of the 'live updating' it would normally do).
+ [`check_tr.sh`](check_tr.sh) - If the sentinel file is older than 1 second, it runs `traceroute` to show the route(s) between this host and `stripe.api.com`. It runs a TCP traceroute (much more consistent than using ICMP, though it does require `sudo`) with no DNS resolution and a max hop count of 20.
+ [`check_one_time_things.sh`](check_one_time_things.sh) - Run some items that shouldn't change to get information about DNS resolution, OS and network info, etc.

# The process

I manually run `check_one_time_things.sh` and output that to a log file.  For each other `check_*.sh` script, I do the following:
```
hostname> while true; do ./check_FOO.sh; ./delay.sh; done
```

This is effectively an infinite loop.  But the scripts are set up such that they should only log their output if a `curl` has taken more than 2 seconds to run.  Typically they're ~150ms, so 2 seconds is definitely outside our expectations.  The `check_curl_https.sh` script has an internal timer to know how long it's taken, but the two traceroute scripts use the age of the `curl_in_progress` file (which our curl script touches right before the `curl` invocation) to calculate the length of the request.

# Other notes

That's about it. Any questions, feel free to ping me and I'll respond with whatever information I can provide.
