#!/bin/bash

ACTIVE=5
BAIL=20
TOTAL=2000

STRIPE_KEY='your_stripe_key_here'
go build . &&\
  ./stripeRequestDebugger -D=true -deadline=500 -total=${TOTAL} -delayed=${BAIL} -active=${ACTIVE} -key="${STRIPE_KEY}"
echo "Done with running the request debugger"
