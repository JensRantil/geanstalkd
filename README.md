geanstalkd
==========
[![Build Status](https://secure.travis-ci.org/JensRantil/geanstalkd.png?branch=master)](http://travis-ci.org/JensRantil/geanstalkd)

A [beanstalkd](http://kr.github.io/beanstalkd/) clone written in Go. beanstalkd
is a small distributed task queue that supports multiple task queue (called
"tubes"), task priority and task delay.

Why a reimplementation?
-----------------------
Don't get me wrong - beanstalkd is amazing, fast and stable! This is somewhat
of a hobby project where I implement beanstalkd in Go, but long-term this
project might also allow [some neat stuff that beanstalkd doesn't
want](https://github.com/JensRantil/geanstalkd/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement).

What differs from beanstalkd?
-----------------------------
 * A single client can reserve multiple jobs.
 * Uses multiple cores.
