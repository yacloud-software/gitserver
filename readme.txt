This service serves git via http.
why?

* It authenticates using "auth" service
* checks permissions to repositories using objectauth service
* keeps a database of commits
* triggers builds for each commit


BUILDRULES:

TARGETS=linux,darwin
will build linux & darwing




--- tech details --

hooks:
on access, the gitserver installs scripts in the git repository hooks directory.
these server-side scripts call into the gitserver

the post-receive uses a dodgy tcp protocol to call into the server
the update hook currently runs within the "git" process space, but an experimental implementation
to run via gRPC on the local gitserver is implemented in update.go (disabled by default)






















