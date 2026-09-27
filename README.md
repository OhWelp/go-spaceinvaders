# Go-SpaceInvaders: A simple emulator of the Intel 8080-based arcade cabinet

## About

I just built this because I wanted to do some emulator development after finishing a CHIP-8 emulator. It's (mostly) hand-built, but I had Claude handle the test generation.

## Playing

This has only been tested and built on Linux, so there is no guarantee it works on Windows or MacOS.

Press "c" to insert a coin. Press "1" to start for player one, "2" for player 2. Arrow keys to move, spacebar to shoot.

## Testing harness

There are a lot of Space Invaders emulators, and this one is the same as the others. There is one thing you might find helpful, however: 1,321,472 CPU tests exist in i808_tests.json, in the style of SingleStepTests.

These were derived from a reference emulator, [superzazu/808](https://github.com/superzazu/8080). If you've developing your own emulator, this is helpful because you can make it correct as you go along rather than needing to get to the point where a test ROM is loaded and then having to go back and fix your errors. The tests file is very large (480 MB).
