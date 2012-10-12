# Scarlet

An HTTP frontend, for Redis.

## Installing

You can install Scarlet simply by running the following command (assuming you
have a Go environment setup):

    go get github.com/nesv/Scarlet
	
Alternately, if you would like to clone the repository and build it from there,
the sources also ship with a simple Makefile; all you have to do once you get
the repository cloned is run:

    make
	
...and you will have a `Scarlet` executable in your current directory, that you
can move around, to wherever.
	
## Issues?

If you experience any issues, please create an issue [here](https://github.com/nesv/Scarlet/issues),
on Github.

## Development

If you would like to help out with developing Scarlet, that would be awesome!

The first thing to do, would be to fork this repository, do something important
and/or cool, then submit a pull request. If your work is good, and consistent,
then you will be added to the list of contributors.


## Copying & contributing

See the LICENSE.

## Travis CI build status

[![Build Status](https://secure.travis-ci.org/nesv/Scarlet.png)](http://travis-ci.org/nesv/Scarlet)

## Changelog

*0.0.2*
*	A minor change involving the .travis.yml file.

*0.0.1*
*	The first development release.
*	You can use HTTP GET requests to fetch values from keys.
