geanstalkd
==========

A [beanstalkd](http://kr.github.io/beanstalkd/) clone written in Go. beanstalkd
is a small distributed task queue that supports multiple task queue (called
"tubes"), task priority and task delay.

Why a reimplementation?
-----------------------
Don't get me wrong - beanstalkd is amazing, fast and stable! This is somewhat
of a hobby project where I implement beanstalkd in Go, but long-term project
might also allow some neat stuff that beanstalkd doesn't want. Things I'm
thinking of:

 * SSL.
 * Password protection.
 * Disk-offloading of larger jobs.
 * Compressed WAL.
 * Compressed job bodies.
 * High-availability using something like RAFT.
 * Burying jobs to disk to offload memory.

What differs from beanstalkd?
-----------------------------
 * A single client can reserve multiple jobs.
 * Uses multiple cores.
