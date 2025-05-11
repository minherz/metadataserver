/*
Package metadataserver provides a metadata server simulator.
Call [New] to create an instance of [Server].
Use [Server.Start] and [Server.Stop] to run and stop the instance.

# Configuration

Customize metadata and set up [Server] using one of [Option]'s:

* [WithAddress] to set IP address to serve metadata
* [WithPort] to set port to serve metadata
* [WithHandlers] to set metadata handlers
* [WithConfigFile] to set [Configuration] loaded from JSON file

# Unit testing

You can use it in unit testing for long running tests:

	  if testing.Short() {
		t.Skip()
	  }
	  s, _ := metadataserver.New(WithConfigFile("path/to/file"))
	  s.Start()
	  defer s.Stop()
	  ...
	  // your test
*/
package metadataserver
