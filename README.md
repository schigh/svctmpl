# Service Template

This is my opinionated template for Go microservices.

| :warning: WARNING: This is very much a work in progress, which I will continue to update as time permits. |
| For those watching this repo, please be sure to pay attention to the github updates as they roll out. I made the decision to put this out in a _very_ incomplete state, because I feel that the evolution of a service from nothing to fully-formed is just as important as the service itself. |
| --- |

The purpose of this repo? I've been writing services in Go for a _long_ time, and I've seen many permutations and 
revisions of what an idiomatic™ Go service looks like. This service is a representation of what works, in my opinion. 
Note that my [opinions](https://x.com/GoTimeFM/status/1402981188483092480) are not always widely-shared across the Go 
community, so keep that in mind as you go through this. I am sure there will be things you disagree with, and I'd be 
happy to discuss those things in the issues section of this repository.

# Coffee Shop Ordering System

Instead of the normal TODO List application that is the product of many sample apps, I've decided to create a coffee 
shop ordering system. The finished system will have the following features:
- Users can order items from a menu