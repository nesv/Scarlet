0.7.0 &mdash; 2012-11-02
*	Added a handler function for the "/" location

0.6.1 &mdash; 2012-11-02
*   Fixed some overly-specific parsing logic in `redis.go`

0.6.0
*	Added some command-line flags that let you override the upstream Redis
	host and password.

0.5.0
*	Fixed a bug where non-existent keys returned type "none", and the error
	in the response said it was an "unknown key type". The error is now
	reported properly.
*	Errors on create operations (through HTTP POST) are now displaying
	properly.
*	HTTP DELETE operations are now supported

0.0.4
*	You can now create keys using the HTTP POST method.

0.0.3
*	You can now use HTTP PUT requests to update existing keys.

0.0.2
*	A minor change involving the `.travis.yml` file.

0.0.1
*	The first development release.
*	You can use HTTP GET requests to fetch values from keys.
